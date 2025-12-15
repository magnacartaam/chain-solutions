package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/google/uuid"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/crypto"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/domain"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/repository"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/solana"
	"github.com/shopspring/decimal"
)

type WalletService struct {
	repo          repository.Repository
	serverPrivKey string
	rpcClient     *rpc.Client
	vaultAddress  solana.PublicKey
}

func NewWalletService(repo repository.Repository, serverPrivKey string, rpcClient *rpc.Client, vaultAddressStr string) (*WalletService, error) {
	vaultPubkey, err := solana.PublicKeyFromBase58(vaultAddressStr)
	if err != nil {
		return nil, fmt.Errorf("invalid vault address: %w", err)
	}

	return &WalletService{
		repo:          repo,
		serverPrivKey: serverPrivKey,
		rpcClient:     rpcClient,
		vaultAddress:  vaultPubkey,
	}, nil
}

// GetBalance returns the current playable balance from the active session
func (s *WalletService) GetBalance(ctx context.Context, walletAddress string) (decimal.Decimal, error) {
	session, err := s.repo.GetActiveSession(ctx, walletAddress)
	if err != nil {
		return decimal.Zero, err
	}
	if session == nil {
		return decimal.Zero, nil
	}
	return session.PlayableBalance, nil
}

// AuthorizeWithdrawal checks funds, increments nonce, signs the message
func (s *WalletService) AuthorizeWithdrawal(ctx context.Context, walletAddress string, amount decimal.Decimal) ([]byte, int, int64, error) {
	user, err := s.repo.GetUser(ctx, walletAddress)
	if err != nil {
		return nil, 0, 0, err
	}
	if user == nil {
		return nil, 0, 0, fmt.Errorf("user not found")
	}

	session, err := s.repo.GetActiveSession(ctx, walletAddress)
	if err != nil {
		return nil, 0, 0, err
	}
	if session == nil {
		return nil, 0, 0, domain.ErrSessionInactive
	}

	if user.PendingWithdrawalAmount.GreaterThan(decimal.Zero) && user.PendingWithdrawalSignature != "" {
		if !user.PendingWithdrawalAmount.Equal(amount) {
			return nil, 0, 0, fmt.Errorf("pending withdrawal exists for %s SOL. Complete it first.", user.PendingWithdrawalAmount.String())
		}

		lamports := amount.Mul(decimal.NewFromInt(1_000_000_000)).BigInt().Uint64()
		sig, recid, err := crypto.SignWithdrawal(s.serverPrivKey, walletAddress, lamports, uint64(user.NextWithdrawalNonce))
		if err != nil {
			return nil, 0, 0, err
		}

		return sig, recid, user.NextWithdrawalNonce, nil
	}

	if session.PlayableBalance.LessThan(amount) {
		return nil, 0, 0, domain.ErrInsufficientFunds
	}

	nonce := user.NextWithdrawalNonce
	lamports := amount.Mul(decimal.NewFromInt(1_000_000_000)).BigInt().Uint64()

	signature, recoveryID, err := crypto.SignWithdrawal(s.serverPrivKey, walletAddress, lamports, uint64(nonce))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("signing failed: %w", err)
	}

	sigHex := hex.EncodeToString(signature)
	err = s.repo.SetPendingWithdrawal(ctx, walletAddress, amount, sigHex)
	if err != nil {
		return nil, 0, 0, err
	}

	return signature, recoveryID, nonce, nil
}

func (s *WalletService) SyncDeposit(ctx context.Context, walletAddress string, txSigStr string) error {
	exists, err := s.repo.CheckDepositProcessed(ctx, txSigStr)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("transaction already processed")
	}

	sig, err := solana.SignatureFromBase58(txSigStr)
	if err != nil {
		return fmt.Errorf("invalid signature format")
	}

	tx, err := s.rpcClient.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch tx from solana: %w", err)
	}
	if tx == nil || tx.Meta == nil || tx.Meta.Err != nil {
		return fmt.Errorf("transaction failed or not found")
	}

	parsedTx, err := tx.Transaction.GetTransaction()
	if err != nil {
		return fmt.Errorf("failed to decode transaction: %w", err)
	}

	vaultIndex := -1
	accountKeys := parsedTx.Message.AccountKeys

	for i, key := range accountKeys {
		if key.Equals(s.vaultAddress) {
			vaultIndex = i
			break
		}
	}

	if vaultIndex == -1 {
		return fmt.Errorf("invalid deposit: casino vault not involved in this transaction")
	}

	if vaultIndex >= len(tx.Meta.PreBalances) || vaultIndex >= len(tx.Meta.PostBalances) {
		return fmt.Errorf("transaction metadata mismatch")
	}

	preBal := int64(tx.Meta.PreBalances[vaultIndex])
	postBal := int64(tx.Meta.PostBalances[vaultIndex])

	amountReceived := postBal - preBal

	if amountReceived <= 0 {
		return fmt.Errorf("invalid deposit: vault balance did not increase")
	}

	seed, err := crypto.GenerateSeed()
	if err != nil {
		return err
	}
	hash := crypto.HashStringSHA256(seed)

	fallbackSession := &domain.Session{
		SessionID:          uuid.New(),
		WalletAddress:      walletAddress,
		PlayableBalance:    decimal.Zero,
		NextServerSeed:     seed,
		NextServerSeedHash: hash,
		IsActive:           true,
	}

	err = s.repo.RecordDeposit(ctx, txSigStr, walletAddress, uint64(amountReceived), fallbackSession)
	if err != nil {
		return err
	}

	return nil
}

func (s *WalletService) AttemptRefund(ctx context.Context, walletAddress string) error {
	user, err := s.repo.GetUser(ctx, walletAddress)
	if err != nil {
		return err
	}

	if user.PendingWithdrawalAmount.IsZero() {
		return fmt.Errorf("no pending withdrawal to refund")
	}

	userPubkey, _ := solana.PublicKeyFromBase58(walletAddress)
	programID, _ := solana.PublicKeyFromBase58(os.Getenv("PROGRAM_ID"))

	userBalancePDA, _, _ := solana.FindProgramAddress(
		[][]byte{[]byte("user_balance"), userPubkey.Bytes()},
		programID,
	)

	accountInfo, err := s.rpcClient.GetAccountInfo(ctx, userBalancePDA)

	var lastOnChainNonce uint64 = 0

	if err == nil && accountInfo != nil && accountInfo.Value != nil {
		onChainData, err := solana_parser.ParseUserBalance(accountInfo.Value.Data.GetBinary())
		if err != nil {
			return fmt.Errorf("failed to parse on-chain data: %w", err)
		}
		lastOnChainNonce = onChainData.LastNonce
	}

	dbNextNonce := uint64(user.NextWithdrawalNonce)
	correctNextNonce := lastOnChainNonce + 1

	if dbNextNonce > correctNextNonce {
		seed, err := crypto.GenerateSeed()
		if err != nil {
			return err
		}
		hash := crypto.HashStringSHA256(seed)

		fallbackSession := &domain.Session{
			SessionID:          uuid.New(),
			WalletAddress:      walletAddress,
			NextServerSeed:     seed,
			NextServerSeedHash: hash,
			IsActive:           true,
		}

		return s.repo.RefundWithdrawal(ctx, walletAddress, user.PendingWithdrawalAmount, int64(correctNextNonce), fallbackSession)
	}

	return fmt.Errorf("cannot refund: transaction appears to have succeeded on-chain")
}

func (s *WalletService) CompleteWithdrawal(ctx context.Context, walletAddress string) error {
	return s.repo.CompleteWithdrawal(ctx, walletAddress)
}
