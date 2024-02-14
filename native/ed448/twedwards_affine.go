package ed448

type TwistedAffinePoint struct {
	X, Y *Fp
}

func TwistedAffinePointNew() *TwistedAffinePoint {
	return &TwistedAffinePoint{
		X: FpNew(),
		Y: FpNew(),
	}
}

func (t *TwistedAffinePoint) SetIdentity() *TwistedAffinePoint {
	t.X.SetZero()
	t.Y.SetOne()
	return t
}

func (t *TwistedAffinePoint) IsOnCurve() bool {
	x := FpNew().Square(t.X)
	y := FpNew().Square(t.Y)

	lhs := FpNew().Sub(y, x)
	rhs := FpNew().SetOne()
	rhs2 := FpNew().Mul(x, y)
	rhs2.Mul(rhs2, twistedD)
	rhs.Add(rhs, rhs2)
	return lhs.EqualI(rhs) == 1
}

func (t *TwistedAffinePoint) Neg(a *TwistedAffinePoint) *TwistedAffinePoint {
	t.X.Neg(a.X)
	t.Y.Set(a.Y)
	return t
}

func (t *TwistedAffinePoint) Add(arg1, arg2 *TwistedAffinePoint) *TwistedAffinePoint {
	xx := FpNew().Mul(arg1.X, arg2.X)
	yy := FpNew().Mul(arg1.Y, arg2.Y)
	xy := FpNew().Mul(arg1.X, arg2.Y)
	yx := FpNew().Mul(arg1.Y, arg2.X)
	d := FpNew().Mul(xx, yy)
	d.Mul(d, twistedD)

	yNum := FpNew().Add(xx, yy)
	yDen := FpNew().Sub(one, d)

	xNum := FpNew().Add(xy, yx)
	xDen := FpNew().Add(one, d)

	t.X, _ = yDen.Invert(yDen)
	t.Y, _ = xDen.Invert(xDen)
	t.X.Mul(t.X, xNum)
	t.Y.Mul(t.Y, yNum)
	return t
}

func (t *TwistedAffinePoint) ToExtensible() *TwistedExtensiblePoint {
	return &TwistedExtensiblePoint{
		X:  FpNew().Set(t.X),
		Y:  FpNew().Set(t.Y),
		Z:  FpNew().SetOne(),
		T1: FpNew().Set(t.X),
		T2: FpNew().Set(t.Y),
	}
}

func (t *TwistedAffinePoint) ToExtended() *TwistedExtendedPoint {
	return t.ToExtensible().ToExtended()
}

type TwistedAffineNielsPoint struct {
	YplusX, YminusX, Td *Fp
}

func AffineNielsPointNew() *TwistedAffineNielsPoint {
	return &TwistedAffineNielsPoint{
		YplusX:  FpNew(),
		YminusX: FpNew(),
		Td:      FpNew(),
	}
}

func (t *TwistedAffineNielsPoint) SetIdentity() *TwistedAffineNielsPoint {
	t.YplusX.SetOne()
	t.YminusX.SetOne()
	t.Td.SetZero()
	return t
}

func (t *TwistedAffineNielsPoint) CMove(a, b *TwistedAffineNielsPoint, choice int) *TwistedAffineNielsPoint {
	t.YplusX.CMove(a.YplusX, b.YplusX, choice)
	t.YminusX.CMove(a.YminusX, b.YminusX, choice)
	t.Td.CMove(a.Td, b.Td, choice)
	return t
}

func (t *TwistedAffineNielsPoint) EqualI(arg *TwistedAffineNielsPoint) int {
	return t.YplusX.EqualI(arg.YplusX) & t.YminusX.EqualI(arg.YminusX) & t.Td.EqualI(arg.Td)
}

func (t *TwistedAffineNielsPoint) ToExtended() *TwistedExtendedPoint {
	return &TwistedExtendedPoint{
		X: FpNew().Sub(t.YplusX, t.YminusX),
		Y: FpNew().Add(t.YminusX, t.YplusX),
		Z: FpNew().SetOne(),
		T: FpNew().Mul(t.YplusX, t.YminusX),
	}
}
