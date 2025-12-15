package game

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// GenerateReelStops uses HMAC-SHA256 to generate deterministic reel positions.
func GenerateReelStops(serverSeed string, clientSeed string, nonce int64, reelLength int) ([]int, string, error) {
	input := fmt.Sprintf("%s:%d", clientSeed, nonce)

	serverKey, err := hex.DecodeString(serverSeed)
	if err != nil {
		return nil, "", fmt.Errorf("invalid server seed hex: %w", err)
	}

	h := hmac.New(sha256.New, serverKey)
	h.Write([]byte(input))
	hash := h.Sum(nil)

	stops := make([]int, 3)
	for i := 0; i < 3; i++ {
		chunk := hash[i*4 : (i+1)*4]

		val := binary.BigEndian.Uint32(chunk)

		stops[i] = int(val) % reelLength
	}

	return stops, hex.EncodeToString(hash), nil
}
