package api

import (
	"log"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/api/handlers"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/repository/postgres"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/service"
)

func RegisterRoutes(router *gin.Engine, dbPool *pgxpool.Pool, serverPrivKey string, rpcClient *rpc.Client, vaultAddress string) {
	repo := postgres.NewPostgresRepo(dbPool)

	gameSvc := service.NewGameService(repo)
	walletSvc, err := service.NewWalletService(repo, serverPrivKey, rpcClient, vaultAddress)
	if err != nil {
		log.Fatalf("Failed to initialize WalletService: %v", err)
	}

	gameH := handlers.NewGameHandler(gameSvc)
	walletH := handlers.NewWalletHandler(walletSvc)

	stegoH := handlers.NewStegoHandler()

	apiGroup := router.Group("/api")
	{
		v1 := apiGroup.Group("/v1")
		{
			gameRoutes := v1.Group("/game")
			{
				gameRoutes.POST("/session", gameH.InitSession)
				gameRoutes.POST("/spin", gameH.Spin)
				gameRoutes.GET("/history", gameH.GetHistory)
				gameRoutes.GET("/proof/:spin_id", gameH.GetProof)
			}

			walletRoutes := v1.Group("/wallet")
			{
				walletRoutes.GET("/balance/:address", walletH.GetBalance)
				walletRoutes.POST("/withdraw", walletH.Withdraw)
				walletRoutes.POST("/sync", walletH.SyncDeposit)
				walletRoutes.POST("/refund", walletH.RequestRefund)
				walletRoutes.POST("/complete-withdraw", walletH.CompleteWithdrawal)
			}

			cipher := v1.Group("/cipher")
			{
				stbGroup := cipher.Group("/stb")
				{
					stbGroup.POST("/cipher", handlers.CipherHandler)
					stbGroup.POST("/decipher", handlers.DecipherHandler)
				}

				rabinGroup := cipher.Group("/rabin")
				{
					rabinGroup.GET("/keygen", handlers.RabinKeygenHandler)
					rabinGroup.POST("/encrypt", handlers.RabinEncryptHandler)
					rabinGroup.POST("/decrypt", handlers.RabinDecryptHandler)
				}

				mcElieceGroup := cipher.Group("/mceliece")
				{
					mcElieceGroup.GET("/keygen", handlers.McElieceKeygenHandler)
					mcElieceGroup.POST("/encrypt", handlers.McElieceEncryptHandler)
					mcElieceGroup.POST("/decrypt", handlers.McElieceDecryptHandler)
				}

				elgamalECGroup := cipher.Group("/elgamal-ec")
				{
					elgamalECGroup.GET("/keygen", handlers.ElGamalECKeygenHandler)
					elgamalECGroup.GET("/keygen-p256", handlers.ElGamalECKeygenP256Handler)
					elgamalECGroup.GET("/keygen-p384", handlers.ElGamalECKeygenP384Handler)
					elgamalECGroup.POST("/encrypt", handlers.ElGamalECEncryptHandler)
					elgamalECGroup.POST("/decrypt", handlers.ElGamalECDecryptHandler)
				}
			}

			hash := v1.Group("/hash")
			{
				gostGroup := hash.Group("/gost3411")
				{
					gostGroup.POST("/hash", handlers.GostHashHandler)
					gostGroup.POST("/hash256", handlers.GostHash256Handler)
					gostGroup.POST("/hash512", handlers.GostHash512Handler)
					gostGroup.POST("/verify", handlers.GostVerifyHandler)
				}

				sha1Group := hash.Group("/sha1")
				{
					sha1Group.POST("/hash", handlers.SHA1HashHandler)
					sha1Group.POST("/verify", handlers.SHA1VerifyHandler)
					sha1Group.POST("/hash-multiple", handlers.SHA1MultipleHashHandler)
					sha1Group.GET("/compare", handlers.SHA1CompareHandler)
				}
			}

			signature := v1.Group("/signature")
			{
				gost3410 := signature.Group("/gost3410")
				{
					gost3410.GET("/keygen", handlers.GOST3410KeygenHandler)
					gost3410.GET("/keygen256", handlers.GOST3410Keygen256Handler)
					gost3410.GET("/keygen512", handlers.GOST3410Keygen512Handler)
					gost3410.POST("/sign", handlers.GOST3410SignHandler)
					gost3410.POST("/verify", handlers.GOST3410VerifyHandler)
				}
			}

			stegoRoutes := v1.Group("/stego")
			{
				// Вход: Form-Data (image: file, message: string)
				stegoRoutes.POST("/hide", stegoH.Hide)

				// Вход: Form-Data (image: file)
				stegoRoutes.POST("/extract", stegoH.Extract)
			}
		}
	}
}
