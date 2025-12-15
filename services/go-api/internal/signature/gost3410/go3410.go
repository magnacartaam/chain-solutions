package gost3410

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// EllipticCurve y^2 = x^3 + ax + b (mod p)
type EllipticCurve struct {
	P *big.Int
	A *big.Int
	B *big.Int
	Q *big.Int
	X *big.Int
	Y *big.Int
}

type Point struct {
	X *big.Int
	Y *big.Int
}

type PublicKey struct {
	Curve *EllipticCurve
	Point *Point
}

type PrivateKey struct {
	Curve  *EllipticCurve
	D      *big.Int
	Public *PublicKey
}

type Signature struct {
	R *big.Int
	S *big.Int
}

func GetStandardCurve256() *EllipticCurve {
	curve := &EllipticCurve{}

	curve.P, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFD97", 16)
	curve.A, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFD94", 16)
	curve.B, _ = new(big.Int).SetString("00000000000000000000000000000000000000000000000000000000000000a6", 16)
	curve.Q, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF6C611070995AD10045841B09B761B893", 16)
	curve.X, _ = new(big.Int).SetString("0000000000000000000000000000000000000000000000000000000000000001", 16)
	curve.Y, _ = new(big.Int).SetString("8D91E471E0989CDA27DF505A453F2B7635294F2DDF23E3B122ACC99C9E9F1E14", 16)

	return curve
}

func GetStandardCurve512() *EllipticCurve {
	curve := &EllipticCurve{}

	curve.P, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFDC7", 16)
	curve.A, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFDC4", 16)
	curve.B, _ = new(big.Int).SetString("E8C2505DEDFC86DDC1BD0B2B6667F1DA34B82574761CB0E879BD081CFD0B6265EE3CB090F30D27614CB4574010DA90DD862EF9D4EBEE4761503190785A71C760", 16)
	curve.Q, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF27E69532F48D89116FF22B8D4E0560609B4B38ABFAD2B85DCACDB1411F10B275", 16)
	curve.X, _ = new(big.Int).SetString("0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003", 16)
	curve.Y, _ = new(big.Int).SetString("7503CFE87A836AE3A61B8816E25450E6CE5E1C93ACF1ABC1778064FDCBEFA921DF1626BE4FD036E93D75E6A50E3A41E98028FE5FC235F5B889A589CB5215F2A4", 16)

	return curve
}

func (curve *EllipticCurve) IsInfinity(p *Point) bool {
	return p.X == nil || p.Y == nil
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
			addend = curve.Add(addend, addend) // Double
		}
	}

	return result
}

func GenerateKeyPair(curve *EllipticCurve) (*PrivateKey, error) {
	if curve == nil {
		return nil, fmt.Errorf("curve cannot be nil")
	}

	if curve.P == nil || curve.Q == nil || curve.X == nil || curve.Y == nil {
		return nil, fmt.Errorf("invalid curve parameters")
	}

	max := new(big.Int).Sub(curve.Q, big.NewInt(1))
	d, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random number: %v", err)
	}
	d.Add(d, big.NewInt(1))

	if d.Sign() <= 0 || d.Cmp(curve.Q) >= 0 {
		return nil, fmt.Errorf("invalid private key generated")
	}

	basePoint := &Point{
		X: new(big.Int).Set(curve.X),
		Y: new(big.Int).Set(curve.Y),
	}

	publicPoint := curve.ScalarMult(basePoint, d)

	if curve.IsInfinity(publicPoint) {
		return nil, fmt.Errorf("public key generation resulted in point at infinity")
	}

	publicKey := &PublicKey{
		Curve: curve,
		Point: publicPoint,
	}

	privateKey := &PrivateKey{
		Curve:  curve,
		D:      d,
		Public: publicKey,
	}

	return privateKey, nil
}

func Sign(privateKey *PrivateKey, hash []byte) (*Signature, error) {
	curve := privateKey.Curve

	e := new(big.Int).SetBytes(hash)
	e = mod(e, curve.Q)

	if e.Sign() == 0 {
		e = big.NewInt(1)
	}

	var r, s *big.Int

	for {
		k, err := rand.Int(rand.Reader, new(big.Int).Sub(curve.Q, big.NewInt(1)))
		if err != nil {
			return nil, fmt.Errorf("failed to generate random k: %v", err)
		}
		k.Add(k, big.NewInt(1))

		basePoint := &Point{X: curve.X, Y: curve.Y}
		c := curve.ScalarMult(basePoint, k)

		r = mod(c.X, curve.Q)
		if r.Sign() == 0 {
			continue
		}

		s = new(big.Int).Mul(r, privateKey.D)
		ke := new(big.Int).Mul(k, e)
		s.Add(s, ke)
		s = mod(s, curve.Q)

		if s.Sign() != 0 {
			break
		}
	}

	return &Signature{R: r, S: s}, nil
}

func Verify(publicKey *PublicKey, hash []byte, sig *Signature) bool {
	curve := publicKey.Curve

	if sig.R.Sign() <= 0 || sig.R.Cmp(curve.Q) >= 0 {
		return false
	}
	if sig.S.Sign() <= 0 || sig.S.Cmp(curve.Q) >= 0 {
		return false
	}

	e := new(big.Int).SetBytes(hash)
	e = mod(e, curve.Q)

	if e.Sign() == 0 {
		e = big.NewInt(1)
	}

	v := modInverse(e, curve.Q)

	z1 := new(big.Int).Mul(sig.S, v)
	z1 = mod(z1, curve.Q)

	z2 := new(big.Int).Mul(sig.R, v)
	z2.Neg(z2)
	z2 = mod(z2, curve.Q)

	basePoint := &Point{X: curve.X, Y: curve.Y}
	c1 := curve.ScalarMult(basePoint, z1)
	c2 := curve.ScalarMult(publicKey.Point, z2)
	c := curve.Add(c1, c2)

	if curve.IsInfinity(c) {
		return false
	}

	rPrime := mod(c.X, curve.Q)

	return rPrime.Cmp(sig.R) == 0
}

func (pub *PublicKey) Bytes() []byte {
	xBytes := pub.Point.X.Bytes()
	yBytes := pub.Point.Y.Bytes()

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
		Point: &Point{X: x, Y: y},
	}, nil
}

func (priv *PrivateKey) Bytes() []byte {
	return priv.D.Bytes()
}

func PrivateKeyFromBytes(curve *EllipticCurve, data []byte) (*PrivateKey, error) {
	d := new(big.Int).SetBytes(data)

	if d.Sign() <= 0 || d.Cmp(curve.Q) >= 0 {
		return nil, fmt.Errorf("invalid private key")
	}

	basePoint := &Point{X: curve.X, Y: curve.Y}
	publicPoint := curve.ScalarMult(basePoint, d)

	publicKey := &PublicKey{
		Curve: curve,
		Point: publicPoint,
	}

	return &PrivateKey{
		Curve:  curve,
		D:      d,
		Public: publicKey,
	}, nil
}
