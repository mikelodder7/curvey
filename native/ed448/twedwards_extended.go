package ed448

import "github.com/mikelodder7/curvey/native/ed448/fp"

type TwistedExtendedPoint struct {
	X, Y, Z, T *fp.Fp
}

func TwistedExtendedPointNew() *TwistedExtendedPoint {
	return &TwistedExtendedPoint{
		X: fp.FpNew(),
		Y: fp.FpNew(),
		Z: fp.FpNew(),
		T: fp.FpNew(),
	}
}

func (t *TwistedExtendedPoint) SetIdentity() *TwistedExtendedPoint {
	t.X.SetZero()
	t.Y.SetOne()
	t.Z.SetOne()
	t.T.SetZero()
	return t
}

func (t *TwistedExtendedPoint) SetGenerator() *TwistedExtendedPoint {
	t.X.SetRaw(&fp.TwistedBasePointX)
	t.Y.SetRaw(&fp.TwistedBasePointY)
	t.Z.SetRaw(&fp.TwistedBasePointZ)
	t.T.SetRaw(&fp.TwistedBasePointT)
	return t
}

func (t *TwistedExtendedPoint) Double(a *TwistedExtendedPoint) *TwistedExtendedPoint {
	aa := a.ToExtensible()
	aa.Double(aa)
	tmp := aa.ToExtended()
	t.X.Set(tmp.X)
	t.Y.Set(tmp.Y)
	t.Z.Set(tmp.Z)
	t.T.Set(tmp.T)
	return t
}

func (t *TwistedExtendedPoint) Add(arg1, arg2 *TwistedExtendedPoint) *TwistedExtendedPoint {
	aa := arg1.ToExtensible()
	aa.AddExtended(aa, arg2)
	tmp := aa.ToExtended()
	t.X.Set(tmp.X)
	t.Y.Set(tmp.Y)
	t.Z.Set(tmp.Z)
	t.T.Set(tmp.T)
	return t
}

func (t *TwistedExtendedPoint) EqualI(a *TwistedExtendedPoint) int {
	xz := fp.FpNew().Mul(t.X, a.Z)
	zx := fp.FpNew().Mul(t.Z, a.X)
	yz := fp.FpNew().Mul(t.Y, a.Z)
	zy := fp.FpNew().Mul(t.Z, a.Y)

	return xz.EqualI(zx) & yz.EqualI(zy)
}

func (t *TwistedExtendedPoint) CMove(a, b *TwistedExtendedPoint, choice int) *TwistedExtendedPoint {
	t.X.CMove(a.X, b.X, choice)
	t.Y.CMove(a.Y, b.Y, choice)
	t.Z.CMove(a.Z, b.Z, choice)
	t.T.CMove(a.T, b.T, choice)
	return t
}

func (t *TwistedExtendedPoint) ToExtensible() *TwistedExtensiblePoint {
	return &TwistedExtensiblePoint{
		X:  fp.FpNew().Set(t.X),
		Y:  fp.FpNew().Set(t.Y),
		Z:  fp.FpNew().Set(t.Z),
		T1: fp.FpNew().Set(t.T),
		T2: fp.FpNew().SetOne(),
	}
}

func (t *TwistedExtendedPoint) ToAffine() *TwistedAffinePoint {
	z, _ := fp.FpNew().Invert(t.Z)
	return &TwistedAffinePoint{
		X: fp.FpNew().Mul(t.X, z),
		Y: fp.FpNew().Mul(t.Y, z),
	}
}

func (t *TwistedExtendedPoint) Isogeny(a *fp.Fp) *EdwardsPoint {
	affine := t.ToAffine()
	xy := fp.FpNew().Mul(affine.X, affine.Y)
	ax2 := fp.FpNew().Square(affine.X)
	ax2.Mul(ax2, a)
	y2 := fp.FpNew().Square(affine.Y)

	xNum := fp.FpNew().Double(xy)
	xDen := fp.FpNew().Sub(y2, ax2)
	_, _ = xDen.Invert(xDen)
	yNum := fp.FpNew().Add(y2, ax2)
	yDen := fp.FpNew().Double(fp.One)
	yDen.Sub(yDen, y2)
	yDen.Sub(yDen, ax2)
	_, _ = yDen.Invert(yDen)

	xNum.Mul(xNum, xDen)
	yNum.Mul(yNum, yDen)

	return &EdwardsPoint{
		X: xNum,
		Y: yNum,
		Z: fp.FpNew().SetOne(),
		T: fp.FpNew().Mul(xNum, yNum),
	}
}

func (t *TwistedExtendedPoint) ToUntwisted() *EdwardsPoint {
	return t.Isogeny(fp.MinusOne)
}

func (t *TwistedExtendedPoint) IsOnCurveI() int {
	xy := fp.FpNew().Mul(t.X, t.Y)
	zt := fp.FpNew().Mul(t.Z, t.T)
	xx := fp.FpNew().Square(t.X)
	yy := fp.FpNew().Square(t.Y)
	zz := fp.FpNew().Square(t.Z)
	tt := fp.FpNew().Square(t.T)

	lhs := fp.FpNew().Sub(yy, xx)
	rhs := fp.FpNew().Mul(fp.TwistedD, tt)
	rhs.Add(rhs, zz)

	return xy.EqualI(zt) & lhs.EqualI(rhs)
}

func (t *TwistedExtendedPoint) Neg(a *TwistedExtendedPoint) *TwistedExtendedPoint {
	t.X.Neg(a.X)
	t.Y.Set(a.Y)
	t.Z.Set(a.Z)
	t.T.Neg(a.T)
	return t
}

func (t *TwistedExtendedPoint) Torque(a *TwistedExtendedPoint) *TwistedExtendedPoint {
	t.X.Neg(a.X)
	t.Y.Neg(a.Y)
	t.Z.Set(a.Z)
	t.T.Set(a.T)
	return t
}
