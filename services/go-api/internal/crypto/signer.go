package crypto

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/sha3"
)

// SignWithdrawal generates the signature required by the Smart Contract.
// matches: keccak(user_pubkey + amount_le + nonce_le)
func SignWithdrawal(privateKeyHex string, userAddress string, amount uint64, nonce uint64) ([]byte, int, error) {
	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid private key hex: %w", err)
	}

	privKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse ECDSA key: %w", err)
	}

	userPubkeyBytes, err := base58.Decode(userAddress)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user address: %w", err)
	}
	if len(userPubkeyBytes) != 32 {
		return nil, 0, fmt.Errorf("user address must be 32 bytes")
	}

	buf := make([]byte, 0, 32+8+8)
	buf = append(buf, userPubkeyBytes...)

	amountBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountBytes, amount)
	buf = append(buf, amountBytes...)

	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, nonce)
	buf = append(buf, nonceBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(buf)
	messageHash := hasher.Sum(nil)

	sig, err := crypto.Sign(messageHash, privKey)
	if err != nil {
		return nil, 0, fmt.Errorf("signing failed: %w", err)
	}

	if len(sig) != 65 {
		return nil, 0, fmt.Errorf("invalid signature length from crypto.Sign")
	}

	r_s_bytes := sig[:64]
	recoveryID := int(sig[64])

	return r_s_bytes, recoveryID, nil
}
