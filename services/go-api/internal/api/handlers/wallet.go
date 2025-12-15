package handlers

import (
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/domain"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/service"
	"github.com/shopspring/decimal"
)

type WalletHandler struct {
	walletService *service.WalletService
}

func NewWalletHandler(walletService *service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

// GetBalance GET /wallet/balance/:address
func (h *WalletHandler) GetBalance(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Address required"})
		return
	}

	balance, err := h.walletService.GetBalance(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data: BalanceResponse{Balance: balance.String()},
	})
}

// Withdraw POST /wallet/withdraw
func (h *WalletHandler) Withdraw(c *gin.Context) {
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	sig, recid, nonce, err := h.walletService.AuthorizeWithdrawal(c.Request.Context(), req.WalletAddress, req.Amount)
	if err != nil {
		switch err {
		case domain.ErrInsufficientFunds:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Insufficient funds"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	lamports := req.Amount.Mul(decimal.NewFromInt(1_000_000_000)).BigInt().Uint64()

	c.JSON(http.StatusOK, SuccessResponse{
		Data: WithdrawResponse{
			Signature:  hex.EncodeToString(sig),
			RecoveryID: recid,
			Nonce:      nonce,
			Amount:     lamports,
		},
	})
}

func (h *WalletHandler) SyncDeposit(c *gin.Context) {
	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	err := h.walletService.SyncDeposit(c.Request.Context(), req.WalletAddress, req.TxSignature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: "Deposit synced successfully"})
}

func (h *WalletHandler) CompleteWithdrawal(c *gin.Context) {
	var req struct {
		WalletAddress string `json:"wallet_address" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}
	err := h.walletService.CompleteWithdrawal(c.Request.Context(), req.WalletAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Data: "Withdrawal completed"})
}

func (h *WalletHandler) RequestRefund(c *gin.Context) {
	var req WalletOnlyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	err := h.walletService.AttemptRefund(c.Request.Context(), req.WalletAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: "Refund successful. Balance restored."})
}
