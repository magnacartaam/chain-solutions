package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	WalletAddress              string          `json:"wallet_address" db:"wallet_address"`
	NextWithdrawalNonce        int64           `json:"next_withdrawal_nonce" db:"next_withdrawal_nonce"`
	PendingWithdrawalAmount    decimal.Decimal `json:"pending_withdrawal_amount" db:"pending_withdrawal_amount"`
	PendingWithdrawalSignature string          `json:"pending_withdrawal_signature" db:"pending_withdrawal_signature"`
	CreatedAt                  time.Time       `json:"created_at" db:"created_at"`
}

type Session struct {
	SessionID       uuid.UUID       `json:"session_id" db:"session_id"`
	WalletAddress   string          `json:"wallet_address" db:"wallet_address"`
	PlayableBalance decimal.Decimal `json:"playable_balance" db:"playable_balance"`

	NextServerSeed     string `json:"-" db:"next_server_seed"`
	NextServerSeedHash string `json:"next_server_seed_hash" db:"next_server_seed_hash"`

	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type SpinOutcome struct {
	Reels []int `json:"reels"`
	IsWin bool  `json:"is_win"`
}

type Spin struct {
	SpinID        uuid.UUID `json:"spin_id" db:"spin_id"`
	SessionID     uuid.UUID `json:"session_id" db:"session_id"`
	WalletAddress string    `json:"wallet_address" db:"wallet_address"`
	SpinNonce     int64     `json:"spin_nonce" db:"spin_nonce"`

	ServerSeed     string `json:"server_seed" db:"server_seed"`
	ClientSeed     string `json:"client_seed" db:"client_seed"`
	ServerSeedHash string `json:"server_seed_hash" db:"server_seed_hash"`

	BetAmount    decimal.Decimal `json:"bet_amount" db:"bet_amount"`
	PayoutAmount decimal.Decimal `json:"payout_amount" db:"payout_amount"`

	Outcome SpinOutcome `json:"outcome" db:"-"`

	LeafHash  string    `json:"leaf_hash" db:"leaf_hash"`
	BatchID   *int64    `json:"batch_id" db:"batch_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Batch struct {
	BatchID     int64      `json:"batch_id" db:"batch_id"`
	Status      string     `json:"status" db:"status"` // 'OPEN', 'COMMITTED', 'FAILED'
	MerkleRoot  *string    `json:"merkle_root" db:"merkle_root"`
	SolanaTxSig *string    `json:"solana_tx_sig" db:"solana_tx_sig"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	CommittedAt *time.Time `json:"committed_at" db:"committed_at"`
}
