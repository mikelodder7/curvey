package ed448

import "github.com/mikelodder7/curvey/internal"

type LookupTable []*TwistedProjectiveNielsPoint

func FromTwistedExtendedPoint(t *TwistedExtendedPoint) LookupTable {
	p := t.ToExtensible()

	table := make(LookupTable, 8)
	table[0] = p.ToProjectiveNiels()
	for i := 1; i < 8; i++ {
		table[i] = TwistedExtensiblePointNew().AddProjectiveNiels(p, table[i-1]).ToProjectiveNiels()
	}
	return table
}

// Select point from lookup table in constant time
func (t LookupTable) Select(index uint32) *TwistedProjectiveNielsPoint {
	idx := int(index)
	result := TwistedProjectiveNielsPointNew()
	for i := 1; i < 9; i++ {
		result.CMove(result, t[i-1], internal.IsZeroI(i-idx))
	}
	return result
}
