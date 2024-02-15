package ed448

import "github.com/mikelodder7/curvey/native/ed448/fp"

type TwistedProjectiveNielsPoint struct {
	YplusX, YminusX, Td, Z *fp.Fp
}

func TwistedProjectiveNielsPointNew() *TwistedProjectiveNielsPoint {
	return &TwistedProjectiveNielsPoint{
		YplusX:  fp.FpNew(),
		YminusX: fp.FpNew(),
		Td:      fp.FpNew(),
		Z:       fp.FpNew(),
	}
}

func (t *TwistedProjectiveNielsPoint) SetIdentity() *TwistedProjectiveNielsPoint {
	return TwistedExtensiblePointNew().SetIdentity().ToProjectiveNiels()
}

func (t *TwistedProjectiveNielsPoint) ToExtended() *TwistedExtendedPoint {
	a := fp.FpNew().Sub(t.YplusX, t.YminusX)
	b := fp.FpNew().Add(t.YplusX, t.YminusX)
	return &TwistedExtendedPoint{
		X: fp.FpNew().Mul(t.Z, a),
		Y: fp.FpNew().Mul(t.Z, b),
		Z: fp.FpNew().Square(t.Z),
		T: fp.FpNew().Mul(a, b),
	}
}

func (t *TwistedProjectiveNielsPoint) CMove(a, b *TwistedProjectiveNielsPoint, choice int) *TwistedProjectiveNielsPoint {
	t.YplusX.CMove(a.YplusX, b.YplusX, choice)
	t.YminusX.CMove(a.YminusX, b.YminusX, choice)
	t.Td.CMove(a.Td, b.Td, choice)
	t.Z.CMove(a.Z, b.Z, choice)
	return t
}

func (t *TwistedProjectiveNielsPoint) CNeg(a *TwistedProjectiveNielsPoint, choice int) *TwistedProjectiveNielsPoint {
	t.YplusX.Set(a.YplusX)
	t.YminusX.Set(a.YminusX)
	t.Td.Set(a.Td)
	t.Z.Set(a.Z)

	t.YplusX.CSwap(t.YminusX, choice)
	t.Td.CNeg(t.Td, choice)
	return t
}

func (t *TwistedProjectiveNielsPoint) EqualI(other *TwistedProjectiveNielsPoint) int {
	return t.ToExtended().EqualI(other.ToExtended())
}
