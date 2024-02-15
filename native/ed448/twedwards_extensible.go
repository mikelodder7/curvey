package ed448

import "github.com/mikelodder7/curvey/native/ed448/fp"

type TwistedExtensiblePoint struct {
	X, Y, Z, T1, T2 *fp.Fp
}

func TwistedExtensiblePointNew() *TwistedExtensiblePoint {
	return &TwistedExtensiblePoint{
		X:  fp.FpNew().SetZero(),
		Y:  fp.FpNew().SetZero(),
		Z:  fp.FpNew().SetZero(),
		T1: fp.FpNew().SetZero(),
		T2: fp.FpNew().SetZero(),
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
	xz := fp.FpNew().Mul(e.X, rhs.Z)
	zx := fp.FpNew().Mul(e.Z, rhs.X)
	yz := fp.FpNew().Mul(e.Y, rhs.Z)
	zy := fp.FpNew().Mul(e.Z, rhs.Y)

	return xz.EqualI(zx) & yz.EqualI(zy)
}

func (e *TwistedExtensiblePoint) ToExtended() *TwistedExtendedPoint {
	return &TwistedExtendedPoint{
		X: fp.FpNew().Set(e.X),
		Y: fp.FpNew().Set(e.Y),
		Z: fp.FpNew().Set(e.Z),
		T: fp.FpNew().Mul(e.T1, e.T2),
	}
}

func (e *TwistedExtensiblePoint) ToProjectiveNiels() *TwistedProjectiveNielsPoint {
	td := fp.FpNew().Mul(e.T1, e.T2)
	return &TwistedProjectiveNielsPoint{
		YplusX:  fp.FpNew().Add(e.X, e.Y),
		YminusX: fp.FpNew().Sub(e.X, e.Y),
		Z:       fp.FpNew().Double(e.Z),
		Td:      td.Mul(td, twoXTwistedD),
	}
}

func (e *TwistedExtensiblePoint) Double(arg *TwistedExtensiblePoint) *TwistedExtensiblePoint {
	a := fp.FpNew().Square(arg.X)
	b := fp.FpNew().Square(arg.Y)
	c := fp.FpNew().Square(arg.Z)
	c.Double(c)

	d := fp.FpNew().Neg(a)
	ee := fp.FpNew().Add(arg.X, arg.Y)
	ee.Sub(ee, a)
	ee.Sub(ee, b)

	g := fp.FpNew().Add(d, b)
	f := fp.FpNew().Sub(g, c)
	h := fp.FpNew().Sub(d, b)

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
	a := fp.FpNew().Mul(arg1.X, arg2.X)
	b := fp.FpNew().Mul(arg1.Y, arg2.X)
	c := fp.FpNew().Mul(arg1.T1, arg1.T2)
	c.Mul(c, arg2.T)
	c.Mul(c, twistedD)
	d := fp.FpNew().Mul(arg1.Z, arg2.Z)
	e1 := fp.FpNew().Add(arg1.X, arg1.Y)
	e2 := fp.FpNew().Add(arg2.X, arg2.Y)
	ee := fp.FpNew().Mul(e1, e2)
	ee.Sub(ee, a)
	ee.Sub(ee, b)
	f := fp.FpNew().Sub(d, c)
	g := fp.FpNew().Add(d, c)
	h := fp.FpNew().Add(b, a)

	e.X.Mul(ee, f)
	e.Y.Mul(g, h)
	e.Z.Mul(f, g)
	e.T1.Set(ee)
	e.T2.Set(h)
	return e
}

func (e *TwistedExtensiblePoint) SubExtended(arg1 *TwistedExtensiblePoint, arg2 *TwistedExtendedPoint) *TwistedExtensiblePoint {
	a := fp.FpNew().Mul(arg1.X, arg2.X)
	b := fp.FpNew().Mul(arg1.Y, arg2.Y)
	c := fp.FpNew().Mul(arg1.T1, arg1.T2)
	c.Mul(c, arg2.T)
	c.Mul(c, twistedD)
	d := fp.FpNew().Mul(arg1.Z, arg2.Z)
	e1 := fp.FpNew().Add(arg1.X, arg1.Y)
	e2 := fp.FpNew().Sub(arg2.Y, arg2.X)
	ee := fp.FpNew().Mul(e1, e2)
	ee.Add(ee, a)
	ee.Sub(ee, b)
	f := fp.FpNew().Add(d, c)
	g := fp.FpNew().Sub(d, c)
	h := fp.FpNew().Sub(b, a)

	e.X.Mul(ee, f)
	e.Y.Mul(g, h)
	e.Z.Mul(f, g)
	e.T1.Set(ee)
	e.T2.Set(h)
	return e
}

func (e *TwistedExtensiblePoint) AddProjectiveNiels(arg1 *TwistedExtensiblePoint, arg2 *TwistedProjectiveNielsPoint) *TwistedExtensiblePoint {
	z := fp.FpNew().Mul(arg1.Z, arg2.Z)

	a := fp.FpNew().Sub(arg1.Y, arg1.X)
	a.Mul(a, arg2.YminusX)
	b := fp.FpNew().Add(arg1.Y, arg1.X)
	b.Mul(b, arg2.YplusX)
	c := fp.FpNew().Mul(arg2.Td, arg1.T1)
	c.Mul(c, arg1.T2)
	d := fp.FpNew().Add(b, a)
	ee := fp.FpNew().Sub(b, a)
	f := fp.FpNew().Sub(z, c)
	g := fp.FpNew().Add(z, c)

	e.X.Mul(ee, f)
	e.Y.Mul(g, d)
	e.Z.Mul(f, g)
	e.T1.Set(ee)
	e.T2.Set(d)
	return e
}
