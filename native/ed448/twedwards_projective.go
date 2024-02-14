package ed448

type TwistedProjectiveNielsPoint struct {
	YplusX, YminusX, Td, Z *Fp
}

func TwistedProjectiveNielsPointNew() *TwistedProjectiveNielsPoint {
	return &TwistedProjectiveNielsPoint{
		YplusX:  FpNew(),
		YminusX: FpNew(),
		Td:      FpNew(),
		Z:       FpNew(),
	}
}

func (t *TwistedProjectiveNielsPoint) SetIdentity() *TwistedProjectiveNielsPoint {
	return TwistedExtensiblePointNew().SetIdentity().ToProjectiveNiels()
}

func (t *TwistedProjectiveNielsPoint) ToExtended() *TwistedExtendedPoint {
	a := FpNew().Sub(t.YplusX, t.YminusX)
	b := FpNew().Add(t.YplusX, t.YminusX)
	return &TwistedExtendedPoint{
		X: FpNew().Mul(t.Z, a),
		Y: FpNew().Mul(t.Z, b),
		Z: FpNew().Square(t.Z),
		T: FpNew().Mul(a, b),
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
