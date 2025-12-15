package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/domain"
	"github.com/shopspring/decimal"
)

type Repository interface {
	CreateUser(ctx context.Context, walletAddress string) error
	GetUser(ctx context.Context, walletAddress string) (*domain.User, error)
	GetNextWithdrawalNonce(ctx context.Context, walletAddress string) (int64, error)
	IncrementWithdrawalNonce(ctx context.Context, walletAddress string) error
	SetPendingWithdrawal(ctx context.Context, walletAddress string, amount decimal.Decimal, signature string) error
	CompleteWithdrawal(ctx context.Context, walletAddress string) error
	RefundWithdrawal(ctx context.Context, walletAddress string, amount decimal.Decimal, correctNextNonce int64, fallbackSession *domain.Session) error

	CheckDepositProcessed(ctx context.Context, txSig string) (bool, error)
	RecordDeposit(ctx context.Context, txSig string, walletAddress string, amount uint64, fallbackSession *domain.Session) error

	CreateSession(ctx context.Context, session *domain.Session) error
	GetActiveSession(ctx context.Context, walletAddress string) (*domain.Session, error)
	GetLatestSession(ctx context.Context, walletAddress string) (*domain.Session, error)
	UpdateSessionState(ctx context.Context, sessionID string, newBalance string, newSeed string, newHash string) error

	CreateSpin(ctx context.Context, spin *domain.Spin) error
	GetSpinsByWallet(ctx context.Context, walletAddress string, limit int, offset int) ([]domain.Spin, error)
	GetSpinCount(ctx context.Context, sessionID uuid.UUID) (int64, error)
	GetUnbatchedSpins(ctx context.Context, limit int) ([]domain.Spin, error)
	GetBatchSpins(ctx context.Context, batchID int64) ([]domain.Spin, error)
	GetSpin(ctx context.Context, spinIDStr string) (*domain.Spin, error)
	GetBatch(ctx context.Context, batchID int64) (*domain.Batch, error)

	GetOpenBatch(ctx context.Context) (*domain.Batch, error)
	CreateBatch(ctx context.Context) (*domain.Batch, error)
	AddSpinToBatch(ctx context.Context, spinID string, batchID int64) error
	CloseBatch(ctx context.Context, batchID int64, merkleRoot string, txSig string) error
}
