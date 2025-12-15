package handlers

import (
	"github.com/shopspring/decimal"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

type WithdrawRequest struct {
	WalletAddress string          `json:"wallet_address" binding:"required"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
}

type WithdrawResponse struct {
	Signature  string `json:"signature"`
	RecoveryID int    `json:"recovery_id"`
	Nonce      int64  `json:"nonce"`
	Amount     uint64 `json:"amount_lamports"`
}

type WalletOnlyRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
}

type BalanceResponse struct {
	Balance string `json:"balance_sol"`
}

type InitSessionRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
}

type InitSessionResponse struct {
	SessionID          string `json:"session_id"`
	NextServerSeedHash string `json:"next_server_seed_hash"`
}

type SpinRequest struct {
	WalletAddress string          `json:"wallet_address" binding:"required"`
	BetAmount     decimal.Decimal `json:"bet_amount" binding:"required"`
	ClientSeed    string          `json:"client_seed" binding:"required"`
}

type SpinResponse struct {
	SpinID     string  `json:"spin_id"`
	SpinNonce  int64   `json:"spin_nonce"`
	Outcome    [][]int `json:"outcome"`
	IsWin      bool    `json:"is_win"`
	Payout     string  `json:"payout_sol"`
	ServerSeed string  `json:"server_seed"`
	NextHash   string  `json:"next_server_seed_hash"`
}

type SyncRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	TxSignature   string `json:"tx_signature" binding:"required"`
}
