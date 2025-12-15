package main

import (
	"context"
	"encoding/hex"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/api"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/db"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/repository/postgres"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/worker"
	"golang.org/x/crypto/sha3"
)

func main() {
	_ = godotenv.Load()

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	pool, err := db.ConnectDB(dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	serverPrivKey := os.Getenv("SERVER_SECP_PRIVATE_KEY_HEX")
	if serverPrivKey == "" {
		log.Fatal("SERVER_SECP_PRIVATE_KEY_HEX environment variable is required")
	}

	rpcURL := os.Getenv("SOLANA_RPC_URL")
	if rpcURL == "" {
		log.Fatal("SOLANA_RPC_URL required")
	}
	rpcClient := rpc.New(rpcURL)

	serverWalletPath := os.Getenv("SERVER_WALLET_PATH")
	programID := os.Getenv("PROGRAM_ID")
	vaultAddr := os.Getenv("VAULT_ADDRESS")

	//DEBUG
	pkBytes, _ := hex.DecodeString(serverPrivKey)
	ecdsaKey, _ := crypto.ToECDSA(pkBytes)
	pubKeyBytes := crypto.FromECDSAPub(&ecdsaKey.PublicKey)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(pubKeyBytes[1:])
	identityHash := hex.EncodeToString(hasher.Sum(nil))

	log.Printf("🔑 Server Signing Identity (Hash): %s", identityHash)
	//

	repo := postgres.NewPostgresRepo(pool)

	committer, err := worker.NewBatchCommitter(repo, rpcURL, serverWalletPath, programID, vaultAddr)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to start Batch Committer: %v", err)
		log.Println("Server will run, but spins will NOT be anchored to Solana.")
	} else {
		go committer.Start(context.Background())
	}

	router := gin.Default()

	api.RegisterRoutes(router, pool, serverPrivKey, rpcClient, vaultAddr)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
