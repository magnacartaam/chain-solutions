package elgamal_ec

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type EllipticCurve struct {
	P  *big.Int
	A  *big.Int
	B  *big.Int
	N  *big.Int
	Gx *big.Int
	Gy *big.Int
}

type Point struct {
	X *big.Int
	Y *big.Int
}

type PublicKey struct {
	Curve *EllipticCurve
	Q     *Point
}

type PrivateKey struct {
	Curve  *EllipticCurve
	D      *big.Int
	Public *PublicKey
}

type CipherText struct {
	C1 *Point
	C2 *Point
}

func GetStandardCurveP256() *EllipticCurve {
	curve := &EllipticCurve{}

	curve.P, _ = new(big.Int).SetString("FFFFFFFF00000001000000000000000000000000FFFFFFFFFFFFFFFFFFFFFFFF", 16)
	curve.A, _ = new(big.Int).SetString("FFFFFFFF00000001000000000000000000000000FFFFFFFFFFFFFFFFFFFFFFFC", 16)
	curve.B, _ = new(big.Int).SetString("5AC635D8AA3A93E7B3EBBD55769886BC651D06B0CC53B0F63BCE3C3E27D2604B", 16)
	curve.N, _ = new(big.Int).SetString("FFFFFFFF00000000FFFFFFFFFFFFFFFFBCE6FAADA7179E84F3B9CAC2FC632551", 16)
	curve.Gx, _ = new(big.Int).SetString("6B17D1F2E12C4247F8BCE6E563A440F277037D812DEB33A0F4A13945D898C296", 16)
	curve.Gy, _ = new(big.Int).SetString("4FE342E2FE1A7F9B8EE7EB4A7C0F9E162BCE33576B315ECECBB6406837BF51F5", 16)

	return curve
}

func GetStandardCurveP384() *EllipticCurve {
	curve := &EllipticCurve{}

	curve.P, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFFFF0000000000000000FFFFFFFF", 16)
	curve.A, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFFFF0000000000000000FFFFFFFC", 16)
	curve.B, _ = new(big.Int).SetString("B3312FA7E23EE7E4988E056BE3F82D19181D9C6EFE8141120314088F5013875AC656398D8A2ED19D2A85C8EDD3EC2AEF", 16)
	curve.N, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFC7634D81F4372DDF581A0DB248B0A77AECEC196ACCC52973", 16)
	curve.Gx, _ = new(big.Int).SetString("AA87CA22BE8B05378EB1C71EF320AD746E1D3B628BA79B9859F741E082542A385502F25DBF55296C3A545E3872760AB7", 16)
	curve.Gy, _ = new(big.Int).SetString("3617DE4A96262C6F5D9E98BF9292DC29F8F41DBD289A147CE9DA3113B5F0B8C00A60B1CE1D7E819D7A431D7C90EA0E5F", 16)

	return curve
}

func (curve *EllipticCurve) IsInfinity(p *Point) bool {
	return p == nil || p.X == nil || p.Y == nil
}

func (curve *EllipticCurve) Infinity() *Point {
	return &Point{X: nil, Y: nil}
}

func mod(a, m *big.Int) *big.Int {
	result := new(big.Int).Mod(a, m)
	if result.Sign() < 0 {
		result.Add(result, m)
	}
	return result
}

func modInverse(a, m *big.Int) *big.Int {
	return new(big.Int).ModInverse(a, m)
}

func (curve *EllipticCurve) Add(p1, p2 *Point) *Point {
	if curve.IsInfinity(p1) {
		if curve.IsInfinity(p2) {
			return curve.Infinity()
		}
		return &Point{X: new(big.Int).Set(p2.X), Y: new(big.Int).Set(p2.Y)}
	}
	if curve.IsInfinity(p2) {
		return &Point{X: new(big.Int).Set(p1.X), Y: new(big.Int).Set(p1.Y)}
	}

	var lambda *big.Int

	if p1.X.Cmp(p2.X) == 0 {
		if p1.Y.Cmp(p2.Y) == 0 {
			numerator := new(big.Int).Mul(p1.X, p1.X)
			numerator.Mul(numerator, big.NewInt(3))
			numerator.Add(numerator, curve.A)

			denominator := new(big.Int).Mul(p1.Y, big.NewInt(2))
			denominator = modInverse(denominator, curve.P)

			lambda = new(big.Int).Mul(numerator, denominator)
			lambda = mod(lambda, curve.P)
		} else {
			return curve.Infinity()
		}
	} else {
		numerator := new(big.Int).Sub(p2.Y, p1.Y)
		denominator := new(big.Int).Sub(p2.X, p1.X)
		denominator = modInverse(denominator, curve.P)

		lambda = new(big.Int).Mul(numerator, denominator)
		lambda = mod(lambda, curve.P)
	}

	x3 := new(big.Int).Mul(lambda, lambda)
	x3.Sub(x3, p1.X)
	x3.Sub(x3, p2.X)
	x3 = mod(x3, curve.P)

	y3 := new(big.Int).Sub(p1.X, x3)
	y3.Mul(lambda, y3)
	y3.Sub(y3, p1.Y)
	y3 = mod(y3, curve.P)

	return &Point{X: x3, Y: y3}
}

func (curve *EllipticCurve) ScalarMult(p *Point, k *big.Int) *Point {
	if k.Sign() == 0 || curve.IsInfinity(p) {
		return curve.Infinity()
	}

	result := curve.Infinity()
	addend := &Point{
		X: new(big.Int).Set(p.X),
		Y: new(big.Int).Set(p.Y),
	}

	kBytes := k.Bytes()
	for byteIdx := len(kBytes) - 1; byteIdx >= 0; byteIdx-- {
		for bitIdx := 0; bitIdx < 8; bitIdx++ {
			if (kBytes[byteIdx] & (1 << uint(bitIdx))) != 0 {
				result = curve.Add(result, addend)
			}
			addend = curve.Add(addend, addend)
		}
	}

	return result
}

func GenerateKeyPair(curve *EllipticCurve) (*PrivateKey, error) {
	if curve == nil {
		return nil, fmt.Errorf("curve cannot be nil")
	}

	if curve.P == nil || curve.N == nil || curve.Gx == nil || curve.Gy == nil {
		return nil, fmt.Errorf("invalid curve parameters")
	}

	max := new(big.Int).Sub(curve.N, big.NewInt(1))
	d, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random number: %v", err)
	}
	d.Add(d, big.NewInt(1))

	basePoint := &Point{
		X: new(big.Int).Set(curve.Gx),
		Y: new(big.Int).Set(curve.Gy),
	}

	publicPoint := curve.ScalarMult(basePoint, d)

	if curve.IsInfinity(publicPoint) {
		return nil, fmt.Errorf("public key generation resulted in point at infinity")
	}

	publicKey := &PublicKey{
		Curve: curve,
		Q:     publicPoint,
	}

	privateKey := &PrivateKey{
		Curve:  curve,
		D:      d,
		Public: publicKey,
	}

	return privateKey, nil
}

func (curve *EllipticCurve) EncodeMessage(message []byte) (*Point, error) {
	if len(message) == 0 {
		return nil, fmt.Errorf("message cannot be empty")
	}

	m := new(big.Int).SetBytes(message)

	if m.Cmp(curve.P) >= 0 {
		return nil, fmt.Errorf("message too large for curve")
	}

	maxAttempts := 100
	for i := 0; i < maxAttempts; i++ {
		x := new(big.Int).Add(m, big.NewInt(int64(i)))
		x = mod(x, curve.P)

		y2 := new(big.Int).Mul(x, x)
		y2.Mul(y2, x)

		ax := new(big.Int).Mul(curve.A, x)
		y2.Add(y2, ax)
		y2.Add(y2, curve.B)
		y2 = mod(y2, curve.P)

		y := new(big.Int).ModSqrt(y2, curve.P)
		if y != nil {
			return &Point{X: x, Y: y}, nil
		}
	}

	return nil, fmt.Errorf("failed to encode message to curve point")
}

func (curve *EllipticCurve) DecodeMessage(point *Point, originalLen int) ([]byte, error) {
	if curve.IsInfinity(point) {
		return nil, fmt.Errorf("cannot decode point at infinity")
	}

	// Extract message from x-coordinate
	messageBytes := point.X.Bytes()

	if originalLen > 0 && len(messageBytes) > originalLen {
		messageBytes = messageBytes[:originalLen]
	}

	return messageBytes, nil
}

func Encrypt(publicKey *PublicKey, message []byte) (*CipherText, error) {
	curve := publicKey.Curve

	M, err := curve.EncodeMessage(message)
	if err != nil {
		return nil, fmt.Errorf("failed to encode message: %v", err)
	}

	max := new(big.Int).Sub(curve.N, big.NewInt(1))
	k, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random k: %v", err)
	}
	k.Add(k, big.NewInt(1))

	G := &Point{X: curve.Gx, Y: curve.Gy}
	C1 := curve.ScalarMult(G, k)

	kQ := curve.ScalarMult(publicKey.Q, k)
	C2 := curve.Add(M, kQ)

	return &CipherText{
		C1: C1,
		C2: C2,
	}, nil
}

func Decrypt(privateKey *PrivateKey, ciphertext *CipherText, messageLen int) ([]byte, error) {
	curve := privateKey.Curve

	dC1 := curve.ScalarMult(ciphertext.C1, privateKey.D)

	negDC1 := &Point{
		X: new(big.Int).Set(dC1.X),
		Y: new(big.Int).Neg(dC1.Y),
	}
	negDC1.Y = mod(negDC1.Y, curve.P)

	M := curve.Add(ciphertext.C2, negDC1)

	message, err := curve.DecodeMessage(M, messageLen)
	if err != nil {
		return nil, fmt.Errorf("failed to decode message: %v", err)
	}

	return message, nil
}

func (pub *PublicKey) Bytes() []byte {
	xBytes := pub.Q.X.Bytes()
	yBytes := pub.Q.Y.Bytes()

	// Pad to curve size
	curveByteLen := (pub.Curve.P.BitLen() + 7) / 8
	result := make([]byte, 2*curveByteLen)

	copy(result[curveByteLen-len(xBytes):curveByteLen], xBytes)
	copy(result[2*curveByteLen-len(yBytes):], yBytes)

	return result
}

func PublicKeyFromBytes(curve *EllipticCurve, data []byte) (*PublicKey, error) {
	curveByteLen := (curve.P.BitLen() + 7) / 8

	if len(data) != 2*curveByteLen {
		return nil, fmt.Errorf("invalid public key length")
	}

	x := new(big.Int).SetBytes(data[:curveByteLen])
	y := new(big.Int).SetBytes(data[curveByteLen:])

	return &PublicKey{
		Curve: curve,
		Q:     &Point{X: x, Y: y},
	}, nil
}

func (priv *PrivateKey) Bytes() []byte {
	return priv.D.Bytes()
}

func PrivateKeyFromBytes(curve *EllipticCurve, data []byte) (*PrivateKey, error) {
	d := new(big.Int).SetBytes(data)

	if d.Sign() <= 0 || d.Cmp(curve.N) >= 0 {
		return nil, fmt.Errorf("invalid private key")
	}

	basePoint := &Point{X: curve.Gx, Y: curve.Gy}
	publicPoint := curve.ScalarMult(basePoint, d)

	publicKey := &PublicKey{
		Curve: curve,
		Q:     publicPoint,
	}

	return &PrivateKey{
		Curve:  curve,
		D:      d,
		Public: publicKey,
	}, nil
}

func (ct *CipherText) Bytes(curveByteLen int) []byte {
	c1xBytes := ct.C1.X.Bytes()
	c1yBytes := ct.C1.Y.Bytes()
	c2xBytes := ct.C2.X.Bytes()
	c2yBytes := ct.C2.Y.Bytes()

	result := make([]byte, 4*curveByteLen)

	copy(result[curveByteLen-len(c1xBytes):curveByteLen], c1xBytes)
	copy(result[2*curveByteLen-len(c1yBytes):2*curveByteLen], c1yBytes)
	copy(result[3*curveByteLen-len(c2xBytes):3*curveByteLen], c2xBytes)
	copy(result[4*curveByteLen-len(c2yBytes):], c2yBytes)

	return result
}

func CipherTextFromBytes(curve *EllipticCurve, data []byte) (*CipherText, error) {
	curveByteLen := (curve.P.BitLen() + 7) / 8

	if len(data) != 4*curveByteLen {
		return nil, fmt.Errorf("invalid ciphertext length")
	}

	c1x := new(big.Int).SetBytes(data[:curveByteLen])
	c1y := new(big.Int).SetBytes(data[curveByteLen : 2*curveByteLen])
	c2x := new(big.Int).SetBytes(data[2*curveByteLen : 3*curveByteLen])
	c2y := new(big.Int).SetBytes(data[3*curveByteLen:])

	return &CipherText{
		C1: &Point{X: c1x, Y: c1y},
		C2: &Point{X: c2x, Y: c2y},
	}, nil
}
