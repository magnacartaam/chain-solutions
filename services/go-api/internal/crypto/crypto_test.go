package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestSignWithdrawal(t *testing.T) {
	privKeyHex := "3ca2e85e7c3a731d241a50ee4968a672c9bcdf53ecfd335dea4133e829788772"

	userAddress := "AMyC4nrskq9PERnZfFZv3KRhEm23VUpRV4VrggAjYiiU"
	amount := uint64(1000000000)
	nonce := uint64(1)

	sig, recid, err := SignWithdrawal(privKeyHex, userAddress, amount, nonce)
	if err != nil {
		t.Fatalf("Signing failed: %v", err)
	}

	if len(sig) != 64 {
		t.Errorf("Expected signature length 64, got %d", len(sig))
	}

	fmt.Printf("Signature: %s\n", hex.EncodeToString(sig))
	fmt.Printf("Recovery ID: %d\n", recid)
}

func TestMerkleTree(t *testing.T) {
	// 3 Hashes (Odd number test)
	leaves := []string{
		HashStringSHA256("spin1"),
		HashStringSHA256("spin2"),
		HashStringSHA256("spin3"),
	}

	root, err := ComputeMerkleRoot(leaves)
	if err != nil {
		t.Fatalf("Merkle failed: %v", err)
	}
	fmt.Printf("Merkle Root: %s\n", root)
}
