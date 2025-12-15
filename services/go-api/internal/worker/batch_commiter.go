package worker

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/crypto"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/repository"
)

type BatchCommitter struct {
	repo         repository.Repository
	rpcClient    *rpc.Client
	serverWallet solana.PrivateKey
	programID    solana.PublicKey
	vaultAddress solana.PublicKey
}

// NewBatchCommitter loads the keypair and configures the Solana client
func NewBatchCommitter(repo repository.Repository, rpcURL string, keypairPath string, programIDStr string, vaultAddrStr string) (*BatchCommitter, error) {
	var walletBytes []byte
	var err error

	if rawJSON := os.Getenv("SERVER_WALLET_JSON"); rawJSON != "" {
		walletBytes = []byte(rawJSON)
	} else {
		walletBytes, err = os.ReadFile(keypairPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read server wallet: %w", err)
		}
	}

	var keyInts []uint8
	if err := json.Unmarshal(walletBytes, &keyInts); err != nil {
		return nil, fmt.Errorf("failed to parse wallet json: %w", err)
	}
	serverWallet := solana.PrivateKey(keyInts)

	progID, err := solana.PublicKeyFromBase58(programIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid program ID: %w", err)
	}
	vaultAddr, err := solana.PublicKeyFromBase58(vaultAddrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid vault address: %w", err)
	}

	return &BatchCommitter{
		repo:         repo,
		rpcClient:    rpc.New(rpcURL),
		serverWallet: serverWallet,
		programID:    progID,
		vaultAddress: vaultAddr,
	}, nil
}

// Start runs the background loop
func (b *BatchCommitter) Start(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	log.Println("👷 Batch Committer Worker Started")

	for {
		select {
		case <-ctx.Done():
			log.Println("👷 Worker stopping...")
			return
		case <-ticker.C:
			if err := b.processBatch(ctx); err != nil {
				log.Printf("❌ Batch Error: %v\n", err)
			}
		}
	}
}

// processBatch orchestrates the fetching, hashing, and blockchain submission
func (b *BatchCommitter) processBatch(ctx context.Context) error {
	spins, err := b.repo.GetUnbatchedSpins(ctx, 1000)
	if err != nil {
		return fmt.Errorf("fetching spins: %w", err)
	}
	if len(spins) == 0 {
		return nil
	}

	log.Printf("Processing new batch with %d spins...", len(spins))

	batch, err := b.repo.CreateBatch(ctx)
	if err != nil {
		return fmt.Errorf("creating batch DB record: %w", err)
	}

	for _, s := range spins {
		if err := b.repo.AddSpinToBatch(ctx, s.SpinID.String(), batch.BatchID); err != nil {
			return fmt.Errorf("linking spin to batch: %w", err)
		}
	}

	var leafHashes []string
	for _, s := range spins {
		leafHashes = append(leafHashes, s.LeafHash)
	}
	rootHex, err := crypto.ComputeMerkleRoot(leafHashes)
	if err != nil {
		return fmt.Errorf("calculating merkle root: %w", err)
	}

	txSig, err := b.submitToSolana(ctx, batch.BatchID, rootHex)
	if err != nil {
		return fmt.Errorf("submitting to solana: %w", err)
	}

	if err := b.repo.CloseBatch(ctx, batch.BatchID, rootHex, txSig); err != nil {
		return fmt.Errorf("closing batch in DB: %w", err)
	}

	log.Printf("✅ Batch #%d Committed! Root: %s, Tx: %s", batch.BatchID, rootHex, txSig)
	return nil
}

// submitToSolana constructs the raw Anchor instruction
func (b *BatchCommitter) submitToSolana(ctx context.Context, batchID int64, rootHex string) (string, error) {
	hash := sha256.Sum256([]byte("global:commit_batch_root"))
	discriminator := hash[:8]

	data := make([]byte, 0, 8+8+32)
	data = append(data, discriminator...)

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(batchID))
	data = append(data, buf...)

	rootBytes, err := hex.DecodeString(rootHex)
	if err != nil {
		return "", err
	}
	data = append(data, rootBytes...)

	batchIdLe := make([]byte, 8)
	binary.LittleEndian.PutUint64(batchIdLe, uint64(batchID))

	batchCommitPDA, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("batch_commit"), batchIdLe},
		b.programID,
	)
	if err != nil {
		return "", err
	}

	accounts := []*solana.AccountMeta{
		{PublicKey: b.vaultAddress, IsWritable: true, IsSigner: false},
		{PublicKey: batchCommitPDA, IsWritable: true, IsSigner: false},
		{PublicKey: b.serverWallet.PublicKey(), IsWritable: true, IsSigner: true},
		{PublicKey: solana.SystemProgramID, IsWritable: false, IsSigner: false},
	}

	instruction := solana.NewInstruction(
		b.programID,
		accounts,
		data,
	)

	recent, err := b.rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", err
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		recent.Value.Blockhash,
		solana.TransactionPayer(b.serverWallet.PublicKey()),
	)
	if err != nil {
		return "", err
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(b.serverWallet.PublicKey()) {
			return &b.serverWallet
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	sig, err := b.rpcClient.SendTransaction(ctx, tx)
	if err != nil {
		return "", err
	}

	return sig.String(), nil
}
