package crypto

import (
	"encoding/hex"
	"fmt"
)

// ComputeMerkleRoot takes a list of hex-encoded leaf hashes and returns the hex-encoded root.
func ComputeMerkleRoot(leafHashes []string) (string, error) {
	if len(leafHashes) == 0 {
		return "", nil
	}

	var currentLevel [][]byte
	for _, h := range leafHashes {
		b, err := hex.DecodeString(h)
		if err != nil {
			return "", err
		}
		currentLevel = append(currentLevel, b)
	}

	for len(currentLevel) > 1 {
		var nextLevel [][]byte

		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			var right []byte

			if i+1 < len(currentLevel) {
				right = currentLevel[i+1]
			} else {
				right = left
			}

			combined := append(left, right...)
			hash := HashDataSHA256(combined)
			nextLevel = append(nextLevel, hash)
		}

		currentLevel = nextLevel
	}

	return hex.EncodeToString(currentLevel[0]), nil
}

func GenerateMerkleProof(leafHashes []string, index int) ([]string, error) {
	if index >= len(leafHashes) || index < 0 {
		return nil, fmt.Errorf("index out of bounds")
	}

	var proof []string
	currentLevel := leafHashes
	currentIndex := index

	for len(currentLevel) > 1 {
		var siblingHash string

		if currentIndex%2 == 0 {
			if currentIndex+1 < len(currentLevel) {
				siblingHash = currentLevel[currentIndex+1]
			} else {
				siblingHash = currentLevel[currentIndex]
			}
		} else {
			siblingHash = currentLevel[currentIndex-1]
		}

		proof = append(proof, siblingHash)

		var nextLevel []string
		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			right := left

			if i+1 < len(currentLevel) {
				right = currentLevel[i+1]
			}

			lBytes, _ := hex.DecodeString(left)
			rBytes, _ := hex.DecodeString(right)
			combined := append(lBytes, rBytes...)

			parentHash := hex.EncodeToString(HashDataSHA256(combined))
			nextLevel = append(nextLevel, parentHash)
		}

		currentLevel = nextLevel
		currentIndex = currentIndex / 2
	}

	return proof, nil
}
