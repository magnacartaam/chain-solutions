package mceliece

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

type Parameters struct {
	N int
	K int
	T int
}

type PublicKey struct {
	G      [][]int
	T      int
	Params Parameters
}

type PrivateKey struct {
	S      [][]int
	G      [][]int
	P      [][]int
	Params Parameters
}

type KeyPair struct {
	Public  *PublicKey
	Private *PrivateKey
}

func GenerateKeys(params Parameters) (*KeyPair, error) {
	if params.K <= 0 || params.N <= 0 || params.T <= 0 {
		return nil, errors.New("invalid parameters")
	}

	G := generateRandomMatrix(params.K, params.N)

	S, err := generateInvertibleMatrix(params.K)
	if err != nil {
		return nil, err
	}

	P := generatePermutationMatrix(params.N)

	temp := matrixMultiply(S, G)
	GPrime := matrixMultiply(temp, P)

	return &KeyPair{
		Public: &PublicKey{
			G:      GPrime,
			T:      params.T,
			Params: params,
		},
		Private: &PrivateKey{
			S:      S,
			G:      G,
			P:      P,
			Params: params,
		},
	}, nil
}

func Encrypt(message []byte, pubKey *PublicKey) ([]byte, error) {
	k := pubKey.Params.K
	n := pubKey.Params.N

	fmt.Printf("[CORE] Encrypt called. Desired message bit length (k) = %d\n", k)
	fmt.Printf("[CORE] Input message bytes: %v (%q)\n", message, string(message))

	if len(message)*8 > k {
		return nil, fmt.Errorf("message length (%d bytes) is too long for k=%d bits", len(message), k)
	}

	requiredBytes := (k + 7) / 8
	paddedMessage := make([]byte, requiredBytes)
	copy(paddedMessage, message)
	fmt.Printf("[CORE] Padded message bytes for conversion: %v\n", paddedMessage)

	M := bytesToBinaryVector(paddedMessage, k)

	isMAllZeros := true
	for _, bit := range M {
		if bit != 0 {
			isMAllZeros = false
			break
		}
	}
	fmt.Printf("[CORE] Converted message vector M (is all zeros? %t): %v\n", isMAllZeros, M)

	Z := generateErrorVector(n, pubKey.T)

	MG := vectorMatrixMultiply(M, pubKey.G)

	fmt.Printf("[CORE] Result of M*G (is all zeros? %t): %v\n", isVectorAllZeros(MG), MG)

	C := addVectors(MG, Z)

	fmt.Printf("[CORE] Final cipher vector C (M*G + Z): %v\n", C)

	finalBytes := binaryVectorToBytes(C)
	fmt.Printf("[CORE] Final cipher bytes: %v\n", finalBytes)
	return finalBytes, nil
}

func isVectorAllZeros(vector []int) bool {
	for _, bit := range vector {
		if bit != 0 {
			return false
		}
	}
	return true
}

func Decrypt(ciphertext []byte, privKey *PrivateKey) ([]byte, error) {
	n := privKey.Params.N

	C := bytesToBinaryVector(ciphertext, n)
	if len(C) != n {
		return nil, errors.New("invalid ciphertext length")
	}

	PInv := invertPermutationMatrix(privKey.P)
	C1 := vectorMatrixMultiply(C, PInv)

	M1, err := decode(C1, privKey.G, privKey.Params.T)
	if err != nil {
		return nil, err
	}

	SInv, err := invertMatrix(privKey.S)
	if err != nil {
		return nil, err
	}
	M := vectorMatrixMultiply(M1, SInv)

	return binaryVectorToBytes(M), nil
}

func generateRandomMatrix(rows, cols int) [][]int {
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
		for j := range matrix[i] {
			if randBit() {
				matrix[i][j] = 1
			}
		}
	}
	return matrix
}

func generateInvertibleMatrix(size int) ([][]int, error) {
	maxAttempts := 100
	for attempt := 0; attempt < maxAttempts; attempt++ {
		matrix := generateRandomMatrix(size, size)
		if isInvertible(matrix) {
			return matrix, nil
		}
	}
	return nil, errors.New("failed to generate invertible matrix")
}

func generatePermutationMatrix(size int) [][]int {
	perm := make([]int, size)
	for i := range perm {
		perm[i] = i
	}

	for i := size - 1; i > 0; i-- {
		j := randInt(i + 1)
		perm[i], perm[j] = perm[j], perm[i]
	}

	matrix := make([][]int, size)
	for i := range matrix {
		matrix[i] = make([]int, size)
		matrix[i][perm[i]] = 1
	}
	return matrix
}

func generateErrorVector(length, maxErrors int) []int {
	vector := make([]int, length)
	errorCount := randInt(maxErrors + 1)

	positions := make([]int, length)
	for i := range positions {
		positions[i] = i
	}

	for i := 0; i < errorCount && i < length; i++ {
		j := i + randInt(length-i)
		positions[i], positions[j] = positions[j], positions[i]
		vector[positions[i]] = 1
	}

	return vector
}

func matrixMultiply(A, B [][]int) [][]int {
	rowsA := len(A)
	colsA := len(A[0])
	colsB := len(B[0])

	result := make([][]int, rowsA)
	for i := range result {
		result[i] = make([]int, colsB)
		for j := 0; j < colsB; j++ {
			sum := 0
			for k := 0; k < colsA; k++ {
				sum += A[i][k] * B[k][j]
			}
			result[i][j] = sum % 2
		}
	}
	return result
}

func vectorMatrixMultiply(v []int, M [][]int) []int {
	cols := len(M[0])
	result := make([]int, cols)

	for j := 0; j < cols; j++ {
		sum := 0
		for i := range v {
			sum += v[i] * M[i][j]
		}
		result[j] = sum % 2
	}
	return result
}

func addVectors(a, b []int) []int {
	result := make([]int, len(a))
	for i := range a {
		result[i] = (a[i] + b[i]) % 2
	}
	return result
}

func invertPermutationMatrix(P [][]int) [][]int {
	n := len(P)
	inv := make([][]int, n)
	for i := range inv {
		inv[i] = make([]int, n)
	}

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			inv[j][i] = P[i][j]
		}
	}
	return inv
}

func invertMatrix(matrix [][]int) ([][]int, error) {
	n := len(matrix)

	augmented := make([][]int, n)
	for i := range augmented {
		augmented[i] = make([]int, 2*n)
		copy(augmented[i][:n], matrix[i])
		augmented[i][n+i] = 1
	}

	for col := 0; col < n; col++ {
		pivotRow := -1
		for row := col; row < n; row++ {
			if augmented[row][col] == 1 {
				pivotRow = row
				break
			}
		}

		if pivotRow == -1 {
			return nil, errors.New("matrix is not invertible")
		}

		if pivotRow != col {
			augmented[col], augmented[pivotRow] = augmented[pivotRow], augmented[col]
		}

		for row := 0; row < n; row++ {
			if row != col && augmented[row][col] == 1 {
				for k := 0; k < 2*n; k++ {
					augmented[row][k] = (augmented[row][k] + augmented[col][k]) % 2
				}
			}
		}
	}

	inverse := make([][]int, n)
	for i := range inverse {
		inverse[i] = make([]int, n)
		copy(inverse[i], augmented[i][n:])
	}

	return inverse, nil
}

func isInvertible(matrix [][]int) bool {
	_, err := invertMatrix(matrix)
	return err == nil
}

func decode(received []int, G [][]int, t int) ([]int, error) {
	k := len(G)
	n := len(G[0])

	bestMessage := make([]int, k)
	minDistance := n + 1

	maxTries := 1 << uint(k)
	if k > 16 {
		maxTries = 1 << 16
	}

	for trial := 0; trial < maxTries; trial++ {
		message := make([]int, k)
		for i := 0; i < k; i++ {
			if trial&(1<<uint(i)) != 0 {
				message[i] = 1
			}
		}

		codeword := vectorMatrixMultiply(message, G)
		distance := hammingDistance(received, codeword)

		if distance < minDistance {
			minDistance = distance
			copy(bestMessage, message)
		}

		if minDistance <= t {
			break
		}
	}

	if minDistance <= t {
		return bestMessage, nil
	}

	return nil, errors.New("decoding failed: too many errors")
}

func hammingDistance(a, b []int) int {
	dist := 0
	for i := range a {
		if a[i] != b[i] {
			dist++
		}
	}
	return dist
}

func bytesToBinaryVector(data []byte, k int) []int {
	vector := make([]int, k)

	for i := 0; i < k; i++ {
		byteIndex := i / 8
		bitIndex := uint(7 - (i % 8))

		if byteIndex < len(data) {
			if (data[byteIndex]>>bitIndex)&1 == 1 {
				vector[i] = 1
			}
		}
	}
	return vector
}

func binaryVectorToBytes(vector []int) []byte {
	numBytes := (len(vector) + 7) / 8
	data := make([]byte, numBytes)

	for i := 0; i < numBytes; i++ {
		var currentByte byte = 0
		for j := 0; j < 8; j++ {
			bitIndex := i*8 + j
			if bitIndex < len(vector) && vector[bitIndex] == 1 {
				currentByte |= 1 << uint(7-j)
			}
		}
		data[i] = currentByte
	}
	return data
}

func randBit() bool {
	b := make([]byte, 1)
	_, err := rand.Read(b)
	if err != nil {
		return false
	}
	return b[0]&1 == 1
}

func randInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}
