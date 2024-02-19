package ed448

import "github.com/mikelodder7/curvey/native/ed448/fp"

type TwistedAffinePoint struct {
	X, Y *fp.Fp
}

func TwistedAffinePointNew() *TwistedAffinePoint {
	return &TwistedAffinePoint{
		X: fp.FpNew(),
		Y: fp.FpNew(),
	}
}

func (t *TwistedAffinePoint) SetIdentity() *TwistedAffinePoint {
	t.X.SetZero()
	t.Y.SetOne()
	return t
}

func (t *TwistedAffinePoint) IsOnCurve() bool {
	x := fp.FpNew().Square(t.X)
	y := fp.FpNew().Square(t.Y)

	lhs := fp.FpNew().Sub(y, x)
	rhs := fp.FpNew().SetOne()
	rhs2 := fp.FpNew().Mul(x, y)
	rhs2.Mul(rhs2, fp.TwistedD)
	rhs.Add(rhs, rhs2)
	return lhs.EqualI(rhs) == 1
}

func (t *TwistedAffinePoint) Neg(a *TwistedAffinePoint) *TwistedAffinePoint {
	t.X.Neg(a.X)
	t.Y.Set(a.Y)
	return t
}

func (t *TwistedAffinePoint) Add(arg1, arg2 *TwistedAffinePoint) *TwistedAffinePoint {
	xx := fp.FpNew().Mul(arg1.X, arg2.X)
	yy := fp.FpNew().Mul(arg1.Y, arg2.Y)
	xy := fp.FpNew().Mul(arg1.X, arg2.Y)
	yx := fp.FpNew().Mul(arg1.Y, arg2.X)
	d := fp.FpNew().Mul(xx, yy)
	d.Mul(d, fp.TwistedD)

	yNum := fp.FpNew().Add(xx, yy)
	yDen := fp.FpNew().Sub(fp.One, d)

	xNum := fp.FpNew().Add(xy, yx)
	xDen := fp.FpNew().Add(fp.One, d)

	t.X, _ = yDen.Invert(yDen)
	t.Y, _ = xDen.Invert(xDen)
	t.X.Mul(t.X, xNum)
	t.Y.Mul(t.Y, yNum)
	return t
}

func (t *TwistedAffinePoint) ToExtensible() *TwistedExtensiblePoint {
	return &TwistedExtensiblePoint{
		X:  fp.FpNew().Set(t.X),
		Y:  fp.FpNew().Set(t.Y),
		Z:  fp.FpNew().SetOne(),
		T1: fp.FpNew().Set(t.X),
		T2: fp.FpNew().Set(t.Y),
	}
}

func (t *TwistedAffinePoint) ToExtended() *TwistedExtendedPoint {
	return t.ToExtensible().ToExtended()
}

type TwistedAffineNielsPoint struct {
	YplusX, YminusX, Td *fp.Fp
}

func AffineNielsPointNew() *TwistedAffineNielsPoint {
	return &TwistedAffineNielsPoint{
		YplusX:  fp.FpNew(),
		YminusX: fp.FpNew(),
		Td:      fp.FpNew(),
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
		X: fp.FpNew().Sub(t.YplusX, t.YminusX),
		Y: fp.FpNew().Add(t.YminusX, t.YplusX),
		Z: fp.FpNew().SetOne(),
		T: fp.FpNew().Mul(t.YplusX, t.YminusX),
	}
}
