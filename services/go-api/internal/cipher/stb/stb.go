// Package cipher STB 34.101.31-2011
package stb

import (
	"encoding/binary"
	"fmt"
)

var hBox = [256]byte{
	0xB1, 0x94, 0xBA, 0xC8, 0x0A, 0x08, 0xF5, 0x3B, 0x36, 0x6D, 0x00, 0x8E, 0x58, 0x4A, 0x5D, 0xE4,
	0x85, 0x04, 0xFA, 0x9D, 0x1B, 0xB6, 0xC7, 0xAC, 0x25, 0x2E, 0x72, 0xC2, 0x02, 0xFD, 0xCE, 0x0D,
	0x5B, 0xE3, 0xD6, 0x12, 0x17, 0xB9, 0x61, 0x81, 0xFE, 0x67, 0x86, 0xAD, 0x71, 0x6B, 0x89, 0x0B,
	0x5C, 0xB0, 0xC0, 0xFF, 0x33, 0xC3, 0x56, 0xB8, 0x35, 0xC4, 0x05, 0xAE, 0xD8, 0xE0, 0x7F, 0x99,
	0xE1, 0x2B, 0xDC, 0x1A, 0xE2, 0x82, 0x57, 0xEC, 0x70, 0x3F, 0xCC, 0xF0, 0x95, 0xEE, 0x8D, 0xF1,
	0xC1, 0xAB, 0x76, 0x38, 0x9F, 0xE6, 0x78, 0xCA, 0xF7, 0xC6, 0xF8, 0x60, 0xD5, 0xBB, 0x9C, 0x4F,
	0xF3, 0x3C, 0x65, 0x7B, 0x63, 0x7C, 0x30, 0x6A, 0xDD, 0x4E, 0xA7, 0x79, 0x9E, 0xB2, 0x3D, 0x31,
	0x3E, 0x98, 0xB5, 0x6E, 0x27, 0xD3, 0xBC, 0xCF, 0x59, 0x1E, 0x18, 0x1F, 0x4C, 0x5A, 0xB7, 0x93,
	0xE9, 0xDE, 0xE7, 0x2C, 0x8F, 0x0C, 0x0F, 0xA6, 0x2D, 0xDB, 0x49, 0xF4, 0x6F, 0x73, 0x96, 0x47,
	0x06, 0x07, 0x53, 0x16, 0xED, 0x24, 0x7A, 0x37, 0x39, 0xCB, 0xA3, 0x83, 0x03, 0xA9, 0x8B, 0xF6,
	0x92, 0xBD, 0x9B, 0x1C, 0xE5, 0xD1, 0x41, 0x01, 0x54, 0x45, 0xFB, 0xC9, 0x5E, 0x4D, 0x0E, 0xF2,
	0x68, 0x20, 0x80, 0xAA, 0x22, 0x7D, 0x64, 0x2F, 0x26, 0x87, 0xF9, 0x34, 0x90, 0x40, 0x55, 0x11,
	0xBE, 0x32, 0x97, 0x13, 0x43, 0xFC, 0x9A, 0x48, 0xA0, 0x2A, 0x88, 0x5F, 0x19, 0x4B, 0x09, 0xA1,
	0x7E, 0xCD, 0xA4, 0xD0, 0x15, 0x44, 0xAF, 0x8C, 0xA5, 0x84, 0x50, 0xBF, 0x66, 0xD2, 0xE8, 0x8A,
	0xA2, 0xD7, 0x46, 0x52, 0x42, 0xA8, 0xDF, 0xB3, 0x69, 0x74, 0xC5, 0x51, 0xEB, 0x23, 0x29, 0x21,
	0xD4, 0xEF, 0xD9, 0xB4, 0x3A, 0x62, 0x28, 0x75, 0x91, 0x14, 0x10, 0xEA, 0x77, 0x6C, 0xDA, 0x1D,
}

type Stb struct {
	key [8]uint32
}

func New(key []byte) (*Stb, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes (256 bits)")
	}
	s := &Stb{}
	for i := 0; i < 8; i++ {
		s.key[i] = binary.LittleEndian.Uint32(key[i*4 : (i+1)*4])
	}
	return s, nil
}

func gTransform(u uint32, r uint) uint32 {
	var bytes [4]byte
	binary.LittleEndian.PutUint32(bytes[:], u)
	for i := 0; i < 4; i++ {
		bytes[i] = hBox[bytes[i]]
	}
	substituted := binary.LittleEndian.Uint32(bytes[:])
	return (substituted << r) | (substituted >> (32 - r))
}

func (s *Stb) encryptBlock(block []byte) []byte {
	a := binary.LittleEndian.Uint32(block[0:4])
	b := binary.LittleEndian.Uint32(block[4:8])
	c := binary.LittleEndian.Uint32(block[8:12])
	d := binary.LittleEndian.Uint32(block[12:16])

	for i := 0; i < 8; i++ {
		k := [8]uint32{
			s.key[(7*i+0)%8], s.key[(7*i+1)%8], s.key[(7*i+2)%8], s.key[(7*i+3)%8],
			s.key[(7*i+4)%8], s.key[(7*i+5)%8], s.key[(7*i+6)%8], s.key[(7*i+7)%8],
		}
		roundNum := uint32(i + 1)

		b ^= gTransform(a+k[0], 5)
		c ^= gTransform(d+k[1], 21)
		a -= gTransform(b+k[2], 13)
		e := gTransform(b+c+k[3], 21) ^ roundNum
		b += e
		c -= e
		d ^= gTransform(c+k[4], 13)
		b ^= gTransform(a+k[5], 21)
		c ^= gTransform(d+k[6], 5)

		a, b = b, a
		c, d = d, c
		b, c = c, b
	}

	result := make([]byte, 16)
	binary.LittleEndian.PutUint32(result[0:4], b)
	binary.LittleEndian.PutUint32(result[4:8], d)
	binary.LittleEndian.PutUint32(result[8:12], a)
	binary.LittleEndian.PutUint32(result[12:16], c)
	return result
}

func (s *Stb) decryptBlock(block []byte) []byte {
	b := binary.LittleEndian.Uint32(block[0:4])
	d := binary.LittleEndian.Uint32(block[4:8])
	a := binary.LittleEndian.Uint32(block[8:12])
	c := binary.LittleEndian.Uint32(block[12:16])

	for i := 7; i >= 0; i-- {
		k := [8]uint32{
			s.key[(7*i+0)%8], s.key[(7*i+1)%8], s.key[(7*i+2)%8], s.key[(7*i+3)%8],
			s.key[(7*i+4)%8], s.key[(7*i+5)%8], s.key[(7*i+6)%8], s.key[(7*i+7)%8],
		}
		roundNum := uint32(i + 1)

		b, c = c, b
		c, d = d, c
		a, b = b, a

		c ^= gTransform(d+k[6], 5)
		b ^= gTransform(a+k[5], 21)
		d ^= gTransform(c+k[4], 13)

		e := gTransform(b+c+k[3], 21) ^ roundNum
		c += e
		b -= e

		a += gTransform(b+k[2], 13)
		c ^= gTransform(d+k[1], 21)
		b ^= gTransform(a+k[0], 5)
	}

	result := make([]byte, 16)
	binary.LittleEndian.PutUint32(result[0:4], a)
	binary.LittleEndian.PutUint32(result[4:8], b)
	binary.LittleEndian.PutUint32(result[8:12], c)
	binary.LittleEndian.PutUint32(result[12:16], d)
	return result
}

func (s *Stb) EncryptECB(plaintext []byte) []byte {
	padded := pkcs7Pad(plaintext, 16)
	ciphertext := make([]byte, len(padded))
	for i := 0; i < len(padded); i += 16 {
		encrypted := s.encryptBlock(padded[i : i+16])
		copy(ciphertext[i:i+16], encrypted)
	}
	return ciphertext
}

func (s *Stb) DecryptECB(ciphertext []byte) ([]byte, error) {
	if len(ciphertext)%16 != 0 {
		return nil, fmt.Errorf("ciphertext length must be multiple of 16")
	}
	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += 16 {
		decrypted := s.decryptBlock(ciphertext[i : i+16])
		copy(plaintext[i:i+16], decrypted)
	}
	return pkcs7Unpad(plaintext)
}

func (s *Stb) EncryptCFB(plaintext, iv []byte) []byte {
	if len(iv) != 16 {
		panic("IV must be 16 bytes (128 bits)")
	}
	ciphertext := make([]byte, len(plaintext))
	feedback := make([]byte, 16)
	copy(feedback, iv)
	for i := 0; i < len(plaintext); i += 16 {
		encrypted := s.encryptBlock(feedback)
		blockSize := 16
		if i+16 > len(plaintext) {
			blockSize = len(plaintext) - i
		}
		for j := 0; j < blockSize; j++ {
			ciphertext[i+j] = plaintext[i+j] ^ encrypted[j]
		}
		copy(feedback, ciphertext[i:i+blockSize])
	}
	return ciphertext
}

func (s *Stb) DecryptCFB(ciphertext, iv []byte) []byte {
	if len(iv) != 16 {
		panic("IV must be 16 bytes (128 bits)")
	}
	plaintext := make([]byte, len(ciphertext))
	feedback := make([]byte, 16)
	copy(feedback, iv)
	for i := 0; i < len(ciphertext); i += 16 {
		encrypted := s.encryptBlock(feedback)
		blockSize := 16
		if i+16 > len(ciphertext) {
			blockSize = len(ciphertext) - i
		}
		for j := 0; j < blockSize; j++ {
			plaintext[i+j] = ciphertext[i+j] ^ encrypted[j]
		}
		copy(feedback, ciphertext[i:i+blockSize])
	}
	return plaintext
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	if padding == 0 {
		padding = blockSize
	}
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("invalid padding: data is empty")
	}
	padding := int(data[length-1])
	if padding > length || padding == 0 {
		return nil, fmt.Errorf("invalid padding: invalid padding value")
	}
	for i := 0; i < padding; i++ {
		if data[length-1-i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding: padding bytes are incorrect")
		}
	}
	return data[:length-padding], nil
}
