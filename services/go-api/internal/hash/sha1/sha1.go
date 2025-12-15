package sha1

import (
	"encoding/binary"
)

const (
	Size      = 20
	BlockSize = 64
)

var h0 = []uint32{
	0x67452301,
	0xEFCDAB89,
	0x98BADCFE,
	0x10325476,
	0xC3D2E1F0,
}

func f(t int, b, c, d uint32) uint32 {
	if t < 20 {
		return (b & c) | ((^b) & d)
	} else if t < 40 {
		return b ^ c ^ d
	} else if t < 60 {
		return (b & c) | (b & d) | (c & d)
	} else {
		return b ^ c ^ d
	}
}

func k(t int) uint32 {
	if t < 20 {
		return 0x5A827999
	} else if t < 40 {
		return 0x6ED9EBA1
	} else if t < 60 {
		return 0x8F1BBCDC
	} else {
		return 0xCA62C1D6
	}
}

func leftRotate(value uint32, bits uint) uint32 {
	return (value << bits) | (value >> (32 - bits))
}

func processBlock(block []byte, h []uint32) {
	w := make([]uint32, 80)

	for i := 0; i < 16; i++ {
		w[i] = binary.BigEndian.Uint32(block[i*4 : (i+1)*4])
	}

	for i := 16; i < 80; i++ {
		w[i] = leftRotate(w[i-3]^w[i-8]^w[i-14]^w[i-16], 1)
	}

	a := h[0]
	b := h[1]
	c := h[2]
	d := h[3]
	e := h[4]

	for t := 0; t < 80; t++ {
		temp := leftRotate(a, 5) + f(t, b, c, d) + e + k(t) + w[t]
		e = d
		d = c
		c = leftRotate(b, 30)
		b = a
		a = temp
	}

	h[0] += a
	h[1] += b
	h[2] += c
	h[3] += d
	h[4] += e
}

func padMessage(message []byte) []byte {
	msgLen := len(message)
	bitLen := uint64(msgLen * 8)

	padded := make([]byte, msgLen+1)
	copy(padded, message)
	padded[msgLen] = 0x80

	paddingLen := (56 - (msgLen+1)%64) % 64

	if paddingLen > 0 {
		zeros := make([]byte, paddingLen)
		padded = append(padded, zeros...)
	}

	lengthBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lengthBytes, bitLen)
	padded = append(padded, lengthBytes...)

	return padded
}

func Hash(message []byte) []byte {
	h := make([]uint32, 5)
	copy(h, h0)

	padded := padMessage(message)

	for i := 0; i < len(padded); i += 64 {
		block := padded[i : i+64]
		processBlock(block, h)
	}

	result := make([]byte, Size)
	for i := 0; i < 5; i++ {
		binary.BigEndian.PutUint32(result[i*4:(i+1)*4], h[i])
	}

	return result
}

func Sum(data []byte) [Size]byte {
	hash := Hash(data)
	var result [Size]byte
	copy(result[:], hash)
	return result
}

type Digest struct {
	h   [5]uint32
	x   [BlockSize]byte
	nx  int
	len uint64
}

func New() *Digest {
	d := new(Digest)
	d.Reset()
	return d
}

func (d *Digest) Reset() {
	d.h[0] = h0[0]
	d.h[1] = h0[1]
	d.h[2] = h0[2]
	d.h[3] = h0[3]
	d.h[4] = h0[4]
	d.nx = 0
	d.len = 0
}

func (d *Digest) Write(p []byte) (nn int, err error) {
	nn = len(p)
	d.len += uint64(nn)

	if d.nx > 0 {
		n := copy(d.x[d.nx:], p)
		d.nx += n
		if d.nx == BlockSize {
			processBlock(d.x[:], d.h[:])
			d.nx = 0
		}
		p = p[n:]
	}

	for len(p) >= BlockSize {
		processBlock(p[:BlockSize], d.h[:])
		p = p[BlockSize:]
	}

	if len(p) > 0 {
		d.nx = copy(d.x[:], p)
	}

	return
}

func (d *Digest) Sum(b []byte) []byte {
	d0 := *d
	hash := d0.checkSum()
	return append(b, hash[:]...)
}

func (d *Digest) checkSum() [Size]byte {
	length := d.len
	var tmp [64]byte
	tmp[0] = 0x80

	if length%64 < 56 {
		d.Write(tmp[0 : 56-length%64])
	} else {
		d.Write(tmp[0 : 64+56-length%64])
	}

	length <<= 3
	binary.BigEndian.PutUint64(tmp[:], length)
	d.Write(tmp[0:8])

	if d.nx != 0 {
		panic("d.nx != 0")
	}

	var digest [Size]byte
	binary.BigEndian.PutUint32(digest[0:], d.h[0])
	binary.BigEndian.PutUint32(digest[4:], d.h[1])
	binary.BigEndian.PutUint32(digest[8:], d.h[2])
	binary.BigEndian.PutUint32(digest[12:], d.h[3])
	binary.BigEndian.PutUint32(digest[16:], d.h[4])

	return digest
}

func (d *Digest) Size() int { return Size }

func (d *Digest) BlockSize() int { return BlockSize }
