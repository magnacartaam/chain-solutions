package stenography

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	_ "image/jpeg"
	"io"

	"lukechampine.com/jsteg"
)

type StegoAlgorithm interface {
	Hide(jpegReader io.Reader, message []byte) ([]byte, error)
	Extract(jpegReader io.Reader) ([]byte, error)
}

type JstegAlgo struct{}

func NewAlgorithm() StegoAlgorithm {
	return &JstegAlgo{}
}

func (a *JstegAlgo) Hide(jpegReader io.Reader, message []byte) ([]byte, error) {
	img, _, err := image.Decode(jpegReader)
	if err != nil {
		return nil, err
	}

	outBuffer := new(bytes.Buffer)

	length := uint32(len(message))

	fullData := make([]byte, 4+len(message))

	binary.BigEndian.PutUint32(fullData[:4], length)

	copy(fullData[4:], message)

	err = jsteg.Hide(outBuffer, img, fullData, nil)
	if err != nil {
		return nil, err
	}

	return outBuffer.Bytes(), nil
}

func (a *JstegAlgo) Extract(jpegReader io.Reader) ([]byte, error) {
	rawBytes, err := jsteg.Reveal(jpegReader)
	if err != nil {
		return nil, err
	}

	if len(rawBytes) < 4 {
		return nil, errors.New("no hidden message found (data too short)")
	}

	msgLen := binary.BigEndian.Uint32(rawBytes[:4])

	if uint64(len(rawBytes)) < 4+uint64(msgLen) {
		return nil, errors.New("message corrupted or incomplete")
	}

	return rawBytes[4 : 4+msgLen], nil
}
