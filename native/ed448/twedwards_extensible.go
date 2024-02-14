package ed448

type TwistedExtensiblePoint struct {
	X, Y, Z, T1, T2 *Fp
}

func TwistedExtensiblePointNew() *TwistedExtensiblePoint {
	return &TwistedExtensiblePoint{
		X:  FpNew().SetZero(),
		Y:  FpNew().SetZero(),
		Z:  FpNew().SetZero(),
		T1: FpNew().SetZero(),
		T2: FpNew().SetZero(),
	}
}

func (e *TwistedExtensiblePoint) SetIdentity() *TwistedExtensiblePoint {
	e.X.SetZero()
	e.Y.SetOne()
	e.Z.SetOne()
	e.T1.SetZero()
	e.T2.SetOne()
	return e
}

func (e *TwistedExtensiblePoint) EqualI(rhs *TwistedExtensiblePoint) int {
	xz := FpNew().Mul(e.X, rhs.Z)
	zx := FpNew().Mul(e.Z, rhs.X)
	yz := FpNew().Mul(e.Y, rhs.Z)
	zy := FpNew().Mul(e.Z, rhs.Y)

	return xz.EqualI(zx) & yz.EqualI(zy)
}

func (e *TwistedExtensiblePoint) ToExtended() *TwistedExtendedPoint {
	return &TwistedExtendedPoint{
		X: FpNew().Set(e.X),
		Y: FpNew().Set(e.Y),
		Z: FpNew().Set(e.Z),
		T: FpNew().Mul(e.T1, e.T2),
	}
}

func (e *TwistedExtensiblePoint) ToProjectiveNiels() *TwistedProjectiveNielsPoint {
	td := FpNew().Mul(e.T1, e.T2)
	return &TwistedProjectiveNielsPoint{
		YplusX:  FpNew().Add(e.X, e.Y),
		YminusX: FpNew().Sub(e.X, e.Y),
		Z:       FpNew().Double(e.Z),
		Td:      td.Mul(td, twoXTwistedD),
	}
}

func (e *TwistedExtensiblePoint) Double(arg *TwistedExtensiblePoint) *TwistedExtensiblePoint {
	a := FpNew().Square(arg.X)
	b := FpNew().Square(arg.Y)
	c := FpNew().Square(arg.Z)
	c.Double(c)

	d := FpNew().Neg(a)
	ee := FpNew().Add(arg.X, arg.Y)
	ee.Sub(ee, a)
	ee.Sub(ee, b)

	g := FpNew().Add(d, b)
	f := FpNew().Sub(g, c)
	h := FpNew().Sub(d, b)

	e.X.Mul(ee, f)
	e.Y.Mul(g, h)
	e.Z.Mul(f, g)
	e.T1.Set(ee)
	e.T2.Set(h)
	return e
}

func (e *TwistedExtensiblePoint) Add(arg1, arg2 *TwistedExtensiblePoint) *TwistedExtensiblePoint {
	return e.AddExtended(arg1, arg2.ToExtended())
}

func (e *TwistedExtensiblePoint) Sub(arg1, arg2 *TwistedExtensiblePoint) *TwistedExtensiblePoint {
	return e.SubExtended(arg1, arg2.ToExtended())
}

func (e *TwistedExtensiblePoint) AddExtended(arg1 *TwistedExtensiblePoint, arg2 *TwistedExtendedPoint) *TwistedExtensiblePoint {
	a := FpNew().Mul(arg1.X, arg2.X)
	b := FpNew().Mul(arg1.Y, arg2.X)
	c := FpNew().Mul(arg1.T1, arg1.T2)
	c.Mul(c, arg2.T)
	c.Mul(c, twistedD)
	d := FpNew().Mul(arg1.Z, arg2.Z)
	e1 := FpNew().Add(arg1.X, arg1.Y)
	e2 := FpNew().Add(arg2.X, arg2.Y)
	ee := FpNew().Mul(e1, e2)
	ee.Sub(ee, a)
	ee.Sub(ee, b)
	f := FpNew().Sub(d, c)
	g := FpNew().Add(d, c)
	h := FpNew().Add(b, a)

	e.X.Mul(ee, f)
	e.Y.Mul(g, h)
	e.Z.Mul(f, g)
	e.T1.Set(ee)
	e.T2.Set(h)
	return e
}

func (e *TwistedExtensiblePoint) SubExtended(arg1 *TwistedExtensiblePoint, arg2 *TwistedExtendedPoint) *TwistedExtensiblePoint {
	a := FpNew().Mul(arg1.X, arg2.X)
	b := FpNew().Mul(arg1.Y, arg2.Y)
	c := FpNew().Mul(arg1.T1, arg1.T2)
	c.Mul(c, arg2.T)
	c.Mul(c, twistedD)
	d := FpNew().Mul(arg1.Z, arg2.Z)
	e1 := FpNew().Add(arg1.X, arg1.Y)
	e2 := FpNew().Sub(arg2.Y, arg2.X)
	ee := FpNew().Mul(e1, e2)
	ee.Add(ee, a)
	ee.Sub(ee, b)
	f := FpNew().Add(d, c)
	g := FpNew().Sub(d, c)
	h := FpNew().Sub(b, a)

	e.X.Mul(ee, f)
	e.Y.Mul(g, h)
	e.Z.Mul(f, g)
	e.T1.Set(ee)
	e.T2.Set(h)
	return e
}

func (e *TwistedExtensiblePoint) AddProjectiveNiels(arg1 *TwistedExtensiblePoint, arg2 *TwistedProjectiveNielsPoint) *TwistedExtensiblePoint {
	z := FpNew().Mul(arg1.Z, arg2.Z)

	a := FpNew().Sub(arg1.Y, arg1.X)
	a.Mul(a, arg2.YminusX)
	b := FpNew().Add(arg1.Y, arg1.X)
	b.Mul(b, arg2.YplusX)
	c := FpNew().Mul(arg2.Td, arg1.T1)
	c.Mul(c, arg1.T2)
	d := FpNew().Add(b, a)
	ee := FpNew().Sub(b, a)
	f := FpNew().Sub(z, c)
	g := FpNew().Add(z, c)

	e.X.Mul(ee, f)
	e.Y.Mul(g, d)
	e.Z.Mul(f, g)
	e.T1.Set(ee)
	e.T2.Set(d)
	return e
}
