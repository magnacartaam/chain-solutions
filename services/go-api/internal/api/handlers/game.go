package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/domain"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/service"
)

type GameHandler struct {
	gameService *service.GameService
}

func NewGameHandler(gameService *service.GameService) *GameHandler {
	return &GameHandler{gameService: gameService}
}

// InitSession POST /game/session
func (h *GameHandler) InitSession(c *gin.Context) {
	var req InitSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	session, err := h.gameService.InitiateSession(c.Request.Context(), req.WalletAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Data: InitSessionResponse{
			SessionID:          session.SessionID.String(),
			NextServerSeedHash: session.NextServerSeedHash,
		},
	})
}

// Spin POST /game/spin
func (h *GameHandler) Spin(c *gin.Context) {
	var req SpinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	spin, nextHash, err := h.gameService.ExecuteSpin(c.Request.Context(), req.WalletAddress, req.BetAmount, req.ClientSeed)
	if err != nil {
		switch err {
		case domain.ErrInsufficientFunds:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Insufficient funds"})
		case domain.ErrSessionInactive:
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Session inactive. Please deposit or re-connect."})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	rawReels := spin.Outcome.Reels
	displayMatrix := make([][]int, 3)

	displayMatrix[0] = rawReels[0:3]
	displayMatrix[1] = rawReels[3:6]
	displayMatrix[2] = rawReels[6:9]

	c.JSON(http.StatusOK, SuccessResponse{
		Data: SpinResponse{
			SpinID:     spin.SpinID.String(),
			SpinNonce:  spin.SpinNonce,
			Outcome:    displayMatrix,
			IsWin:      spin.Outcome.IsWin,
			Payout:     spin.PayoutAmount.String(),
			ServerSeed: spin.ServerSeed,
			NextHash:   nextHash,
		},
	})
}

func (h *GameHandler) GetHistory(c *gin.Context) {
	wallet := c.Query("wallet_address")
	if wallet == "" {
		c.JSON(400, ErrorResponse{Error: "wallet required"})
		return
	}

	spins, err := h.gameService.GetUserHistory(c.Request.Context(), wallet)
	if err != nil {
		c.JSON(500, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(200, SuccessResponse{Data: spins})
}

func (h *GameHandler) GetProof(c *gin.Context) {
	spinID := c.Param("spin_id")
	data, err := h.gameService.GetSpinProof(c.Request.Context(), spinID)
	if err != nil {
		c.JSON(400, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(200, SuccessResponse{Data: data})
}
