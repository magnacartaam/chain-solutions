package game

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateSpin(t *testing.T) {
	serverSeed := "a1a2c3d4e5f6a1f2c3d5e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "user-provdidsded-seesdsddb"
	nonce := int64(1)
	bet := decimal.NewFromFloat(1.0)

	result, err := CalculateSpin(serverSeed, clientSeed, nonce, bet)
	if err != nil {
		t.Fatalf("Engine failed: %v", err)
	}

	fmt.Println("--- 3x3 Slot Result ---")
	for _, row := range result.Matrix {
		fmt.Printf("%v\n", row)
	}

	fmt.Printf("Winning Lines: %v\n", result.PaylineWins)
	fmt.Printf("Total Payout: %s\n", result.TotalPayout.String())
	fmt.Printf("Leaf Hash: %s\n", result.LeafHash)

	if len(result.Matrix) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(result.Matrix))
	}
	if len(result.Matrix[0]) != 3 {
		t.Errorf("Expected 3 cols, got %d", len(result.Matrix[0]))
	}
}
