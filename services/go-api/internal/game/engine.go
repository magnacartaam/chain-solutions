package game

import (
	"github.com/shopspring/decimal"
)

type SpinResult struct {
	Matrix      [][]Symbol
	PaylineWins []int
	TotalPayout decimal.Decimal
	LeafHash    string
}

// CalculateSpin performs the full slot logic
func CalculateSpin(serverSeed string, clientSeed string, nonce int64, betAmount decimal.Decimal) (*SpinResult, error) {
	reelLength := len(MainReelStrip)
	stops, leafHash, err := GenerateReelStops(serverSeed, clientSeed, nonce, reelLength)
	if err != nil {
		return nil, err
	}

	matrix := make([][]Symbol, 3)
	for row := 0; row < 3; row++ {
		matrix[row] = make([]Symbol, 3)
		for col := 0; col < 3; col++ {
			stripIndex := (stops[col] + row) % reelLength
			matrix[row][col] = MainReelStrip[stripIndex]
		}
	}

	totalMultiplier := 0.0
	var winningLines []int

	for lineIdx, coords := range Paylines {
		s1 := matrix[coords[0][0]][coords[0][1]]
		s2 := matrix[coords[1][0]][coords[1][1]]
		s3 := matrix[coords[2][0]][coords[2][1]]

		if s1 == s2 && s2 == s3 {
			if mult, ok := PayoutMultipliers[s1]; ok {
				totalMultiplier += mult
				winningLines = append(winningLines, lineIdx)
			}
		}
	}

	payout := betAmount.Mul(decimal.NewFromFloat(totalMultiplier))

	return &SpinResult{
		Matrix:      matrix,
		PaylineWins: winningLines,
		TotalPayout: payout,
		LeafHash:    leafHash,
	}, nil
}
