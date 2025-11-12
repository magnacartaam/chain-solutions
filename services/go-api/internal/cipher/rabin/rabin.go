package rabin

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type RabinKeys struct {
	N *big.Int
	P *big.Int
	Q *big.Int
}

func GenerateRabinKeys(bits int) (*RabinKeys, error) {
	p, err := generateBlumPrime(bits / 2)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prime p: %v", err)
	}

	q, err := generateBlumPrime(bits / 2)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prime q: %v", err)
	}

	n := new(big.Int).Mul(p, q)

	return &RabinKeys{
		N: n,
		P: p,
		Q: q,
	}, nil
}

func generateBlumPrime(bits int) (*big.Int, error) {
	four := big.NewInt(4)
	three := big.NewInt(3)

	for {
		prime, err := rand.Prime(rand.Reader, bits)
		if err != nil {
			return nil, err
		}

		mod := new(big.Int).Mod(prime, four)
		if mod.Cmp(three) == 0 {
			return prime, nil
		}
	}
}

func Encrypt(message *big.Int, publicKey *big.Int) *big.Int {
	cipher := new(big.Int).Exp(message, big.NewInt(2), publicKey)
	return cipher
}

func Decrypt(cipher *big.Int, keys *RabinKeys) []*big.Int {
	yp, yq := extendedGCD(keys.P, keys.Q)

	mp := modularSqrt(cipher, keys.P)
	mq := modularSqrt(cipher, keys.Q)

	results := make([]*big.Int, 4)

	term1 := new(big.Int).Mul(yp, keys.P)
	term1.Mul(term1, mq)
	term2 := new(big.Int).Mul(yq, keys.Q)
	term2.Mul(term2, mp)
	r := new(big.Int).Add(term1, term2)
	r.Mod(r, keys.N)
	results[0] = r

	results[1] = new(big.Int).Sub(keys.N, r)

	s := new(big.Int).Sub(term1, term2)
	s.Mod(s, keys.N)
	results[2] = s

	results[3] = new(big.Int).Sub(keys.N, s)

	return results
}

func extendedGCD(a, b *big.Int) (*big.Int, *big.Int) {
	if b.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(1), big.NewInt(0)
	}

	x1, y1 := extendedGCD(b, new(big.Int).Mod(a, b))
	x := y1
	y := new(big.Int).Sub(x1, new(big.Int).Mul(new(big.Int).Div(a, b), y1))

	return x, y
}

func modularSqrt(a, p *big.Int) *big.Int {
	exp := new(big.Int).Add(p, big.NewInt(1))
	exp.Div(exp, big.NewInt(4))
	return new(big.Int).Exp(a, exp, p)
}

func isValidText(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	validCount := 0
	for _, b := range data {
		if (b >= 32 && b <= 126) || b == '\n' || b == '\r' || b == '\t' {
			validCount++
		}
	}
	return float64(validCount)/float64(len(data)) >= 0.8
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
