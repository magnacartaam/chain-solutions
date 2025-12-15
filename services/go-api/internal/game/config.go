package game

type Symbol int

const (
	SymEmpty   Symbol = iota // 0
	SymCherry                // 1
	SymLemon                 // 2
	SymPlum                  // 3
	SymBar                   // 4
	SymBell                  // 5
	SymSeven                 // 6
	SymDiamond               // 7
	SymWild                  // 8
)

var PayoutMultipliers = map[Symbol]float64{
	SymCherry:  2.0,
	SymLemon:   3.0,
	SymPlum:    5.0,
	SymBar:     10.0,
	SymBell:    20.0,
	SymSeven:   50.0,
	SymDiamond: 100.0,
	SymWild:    500.0,
}

var MainReelStrip = []Symbol{
	SymWild, SymCherry, SymLemon, SymPlum, SymLemon, SymBar, SymCherry,
	SymSeven, SymLemon, SymPlum, SymCherry, SymBell, SymPlum, SymLemon,
	SymDiamond, SymCherry, SymPlum, SymLemon, SymBar, SymCherry, SymLemon,
	SymSeven, SymPlum, SymLemon, SymCherry, SymBar, SymLemon, SymPlum,
	SymCherry, SymBell, SymBar, SymBell,
}

var Paylines = [][][]int{
	{{0, 0}, {0, 1}, {0, 2}}, // Top Row
	{{1, 0}, {1, 1}, {1, 2}}, // Middle Row
	{{2, 0}, {2, 1}, {2, 2}}, // Bottom Row
	{{0, 0}, {1, 1}, {2, 2}}, // Top-Left to Bottom-Right
	{{2, 0}, {1, 1}, {0, 2}}, // Bottom-Left to Top-Right
}
