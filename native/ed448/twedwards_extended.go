package ed448

type TwistedExtendedPoint struct {
	X, Y, Z, T *Fp
}

func TwistedExtendedPointNew() *TwistedExtendedPoint {
	return &TwistedExtendedPoint{
		X: FpNew(),
		Y: FpNew(),
		Z: FpNew(),
		T: FpNew(),
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
	t.X.Set(twistedBasePoint.X)
	t.Y.Set(twistedBasePoint.Y)
	t.Z.Set(twistedBasePoint.Z)
	t.T.Set(twistedBasePoint.T)
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
	xz := FpNew().Mul(t.X, a.Z)
	zx := FpNew().Mul(t.Z, a.X)
	yz := FpNew().Mul(t.Y, a.Z)
	zy := FpNew().Mul(t.Z, a.Y)

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
		X:  FpNew().Set(t.X),
		Y:  FpNew().Set(t.Y),
		Z:  FpNew().Set(t.Z),
		T1: FpNew().Set(t.T),
		T2: FpNew().SetOne(),
	}
}

func (t *TwistedExtendedPoint) ToAffine() *TwistedAffinePoint {
	z, _ := FpNew().Invert(t.Z)
	return &TwistedAffinePoint{
		X: FpNew().Mul(t.X, z),
		Y: FpNew().Mul(t.Y, z),
	}
}

func (t *TwistedExtendedPoint) Isogeny(a *Fp) *EdwardsPoint {
	affine := t.ToAffine()
	xy := FpNew().Mul(affine.X, affine.Y)
	ax2 := FpNew().Square(affine.X)
	ax2.Mul(ax2, a)
	y2 := FpNew().Square(affine.Y)

	xNum := FpNew().Double(xy)
	xDen := FpNew().Set(y2)
	xDen.Sub(xDen, ax2)
	_, _ = xDen.Invert(xDen)
	yNum := FpNew().Set(y2)
	yNum.Add(yNum, ax2)
	yDen := FpNew().Double(one)
	yDen.Sub(yDen, y2)
	yDen.Sub(yDen, ax2)
	_, _ = yDen.Invert(yDen)

	xNum.Mul(xNum, xDen)
	yNum.Mul(yNum, yDen)

	return &EdwardsPoint{
		X: xNum,
		Y: yNum,
		Z: FpNew().SetOne(),
		T: FpNew().Mul(xNum, yNum),
	}
}

func (t *TwistedExtendedPoint) ToUntwisted() *EdwardsPoint {
	return t.Isogeny(minusOne)
}

func (t *TwistedExtendedPoint) IsOnCurveI() int {
	xy := FpNew().Mul(t.X, t.Y)
	zt := FpNew().Mul(t.Z, t.T)
	xx := FpNew().Square(t.X)
	yy := FpNew().Square(t.Y)
	zz := FpNew().Square(t.Z)
	tt := FpNew().Square(t.T)

	lhs := FpNew().Sub(yy, xx)
	rhs := FpNew().Mul(twistedD, tt)
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
