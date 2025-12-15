package service

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/crypto"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/domain"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/game"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/repository"
	"github.com/shopspring/decimal"
)

type GameService struct {
	repo repository.Repository
}

func NewGameService(repo repository.Repository) *GameService {
	return &GameService{
		repo: repo,
	}
}

// InitiateSession creates a new session or rotates seeds if needed
func (s *GameService) InitiateSession(ctx context.Context, walletAddress string) (*domain.Session, error) {
	err := s.repo.CreateUser(ctx, walletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	startingBalance := decimal.Zero

	lastSession, err := s.repo.GetLatestSession(ctx, walletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch last session: %w", err)
	}

	if lastSession != nil {
		startingBalance = lastSession.PlayableBalance
	}

	seed, err := crypto.GenerateSeed()
	if err != nil {
		return nil, err
	}
	hash := crypto.HashStringSHA256(seed)

	session := &domain.Session{
		SessionID:          uuid.New(),
		WalletAddress:      walletAddress,
		PlayableBalance:    startingBalance,
		NextServerSeed:     seed,
		NextServerSeedHash: hash,
		IsActive:           true,
	}

	err = s.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// ExecuteSpin performs the game logic and returns the Spin result and the Next Server Seed Hash
func (s *GameService) ExecuteSpin(ctx context.Context, walletAddress string, betAmount decimal.Decimal, clientSeed string) (*domain.Spin, string, error) {
	session, err := s.repo.GetActiveSession(ctx, walletAddress)
	if err != nil {
		return nil, "", err
	}
	if session == nil {
		return nil, "", domain.ErrSessionInactive
	}
	if session.PlayableBalance.LessThan(betAmount) {
		return nil, "", domain.ErrInsufficientFunds
	}

	currentSeed := session.NextServerSeed
	currentHash := session.NextServerSeedHash

	prevCount, err := s.repo.GetSpinCount(ctx, session.SessionID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to calculate nonce: %w", err)
	}
	spinNonce := prevCount + 1

	result, err := game.CalculateSpin(currentSeed, clientSeed, spinNonce, betAmount)
	if err != nil {
		return nil, "", err
	}

	nextSeed, err := crypto.GenerateSeed()
	if err != nil {
		return nil, "", err
	}
	nextHash := crypto.HashStringSHA256(nextSeed)

	outcomeBytes, _ := json.Marshal(result.Matrix)
	canonicalString := fmt.Sprintf("%s:%d:%s:%s:%s:%s:%s",
		walletAddress,
		spinNonce,
		currentSeed,
		clientSeed,
		betAmount.String(),
		string(outcomeBytes),
		result.TotalPayout.String(),
	)
	leafHash := hex.EncodeToString(crypto.HashDataSHA256([]byte(canonicalString)))

	newBalance := session.PlayableBalance.Sub(betAmount).Add(result.TotalPayout)

	spin := &domain.Spin{
		SpinID:         uuid.New(),
		SessionID:      session.SessionID,
		WalletAddress:  walletAddress,
		SpinNonce:      spinNonce,
		ServerSeed:     currentSeed,
		ClientSeed:     clientSeed,
		ServerSeedHash: currentHash,
		BetAmount:      betAmount,
		PayoutAmount:   result.TotalPayout,
		Outcome:        domain.SpinOutcome{Reels: convertMatrixToFlat(result.Matrix), IsWin: result.TotalPayout.GreaterThan(decimal.Zero)},
		LeafHash:       leafHash,
	}

	err = s.repo.CreateSpin(ctx, spin)
	if err != nil {
		return nil, "", err
	}

	err = s.repo.UpdateSessionState(ctx, session.SessionID.String(), newBalance.String(), nextSeed, nextHash)
	if err != nil {
		return nil, "", err
	}

	return spin, nextHash, nil
}

func convertMatrixToFlat(matrix [][]game.Symbol) []int {
	var flat []int
	for _, row := range matrix {
		for _, sym := range row {
			flat = append(flat, int(sym))
		}
	}
	return flat
}

func (s *GameService) GetUserHistory(ctx context.Context, walletAddress string) ([]domain.Spin, error) {
	return s.repo.GetSpinsByWallet(ctx, walletAddress, 50, 0)
}

func (s *GameService) GetSpinProof(ctx context.Context, spinIDStr string) (interface{}, error) {
	spin, err := s.repo.GetSpin(ctx, spinIDStr)
	if err != nil {
		return nil, err
	}
	if spin.BatchID == nil {
		return nil, fmt.Errorf("spin is not anchored to blockchain yet")
	}

	batch, err := s.repo.GetBatch(ctx, *spin.BatchID)
	if err != nil {
		return nil, err
	}

	batchSpins, err := s.repo.GetBatchSpins(ctx, *spin.BatchID)
	if err != nil {
		return nil, err
	}

	var leafHashes []string
	targetIndex := -1

	for i, s := range batchSpins {
		leafHashes = append(leafHashes, s.LeafHash)
		if s.SpinID.String() == spinIDStr {
			targetIndex = i
		}
	}

	if targetIndex == -1 {
		return nil, fmt.Errorf("spin not found in batch")
	}

	proof, _ := crypto.GenerateMerkleProof(leafHashes, targetIndex)

	return map[string]interface{}{
		"spin":        spin,
		"batch_id":    batch.BatchID,
		"merkle_root": batch.MerkleRoot,
		"solana_tx":   batch.SolanaTxSig,
		"proof":       proof,
	}, nil
}
