package solana_parser

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go"
)

type UserBalanceAccount struct {
	Discriminator [8]byte
	User          solana.PublicKey
	Amount        uint64
	LastNonce     uint64
	Bump          uint8
}

func ParseUserBalance(data []byte) (*UserBalanceAccount, error) {
	if len(data) < 57 {
		return nil, fmt.Errorf("data too short")
	}

	var acc UserBalanceAccount
	buf := bytes.NewReader(data)

	if err := binary.Read(buf, binary.LittleEndian, &acc.Discriminator); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &acc.User); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &acc.Amount); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &acc.LastNonce); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &acc.Bump); err != nil {
		return nil, err
	}

	return &acc, nil
}
