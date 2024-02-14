package ed448

import (
	"crypto/subtle"
	"fmt"
	"github.com/mikelodder7/curvey/internal"
	"github.com/mikelodder7/curvey/native"
)

const PointLimbs = 7
const PointBytes = 57

type PointBytesSeq [PointBytes]byte

type CompressedEdwardsY = PointBytesSeq

func (c *CompressedEdwardsY) CMove(arg1, arg2 *CompressedEdwardsY, choice int) *CompressedEdwardsY {
	mask := byte(-choice)
	for i := 0; i < PointLimbs; i++ {
		(*c)[i] = (*arg1)[i] ^ (((*arg1)[i] ^ (*arg2)[i]) & mask)
	}
	return c
}

func (c *CompressedEdwardsY) EqualI(rhs *CompressedEdwardsY) int {
	return subtle.ConstantTimeCompare((*c)[:], (*rhs)[:])
}

func (c *CompressedEdwardsY) Decompress() (*EdwardsPoint, error) {
	var yBytes [56]byte
	copy(yBytes[:], c[:56])
	sign := int(c[56])
	y, err := FpNew().SetCanonicalBytes(&yBytes)
	if err != nil {
		return nil, err
	}
	yy := FpNew().Square(y)
	dyy := FpNew().Mul(edwardsD, yy)

	numerator := FpNew().SetOne()
	numerator.Sub(numerator, yy)

	denominator := FpNew().SetOne()
	denominator.Sub(denominator, dyy)
	x, isRes := FpNew().SqrtRatio(numerator, denominator)

	signBit := sign >> 7
	isNegative := x.Sgn0I()
	x.CNeg(x, isNegative|signBit)

	pt := (&AffinePoint{X: x, Y: y}).ToEdwards()

	if isRes&pt.IsTorsionFree()&pt.IsOnCurve() == 1 {
		return pt, nil
	} else {
		return nil, fmt.Errorf("invalid point")
	}
}

type EdwardsPoint struct {
	X, Y, Z, T *Fp
}

func EdwardsPointNew() *EdwardsPoint {
	return &EdwardsPoint{
		X: FpNew(),
		Y: FpNew(),
		Z: FpNew(),
		T: FpNew(),
	}
}

func (e *EdwardsPoint) SetIdentity() *EdwardsPoint {
	e.X.SetZero()
	e.Y.SetOne()
	e.Z.SetOne()
	e.T.SetZero()
	return e
}

func (e *EdwardsPoint) SetGenerator() *EdwardsPoint {
	e.X.SetRaw(&[PointLimbs]uint64{
		0x6d35bf93b17aa383,
		0x65fab7bc2914f8fe,
		0x7e9b28e44cd37ab7,
		0x9be886a7f2ed152a,
		0xc7295e6eb444d6fb,
		0x6ef0905d88b9ee96,
		0x420685f0ea8836d1,
	})
	e.Y.SetRaw(&[PointLimbs]uint64{
		0x04ac119c79a99632,
		0x5486da8e9ac23c21,
		0xa98abb416ef259fc,
		0x012232506ee00948,
		0xe6acaf94714fa9dd,
		0xf7687a33ab50a1f0,
		0xd81f4fba18417765,
	})
	e.Z.SetOne()
	e.T.SetRaw(&[PointLimbs]uint64{
		0x2a8ab420e386ac5c,
		0x481d32474a6b9736,
		0xdf9bfafd84761acf,
		0x445dc2c4a99422e3,
		0x0f71014e6a0f05f3,
		0x5339b7fc98aac411,
		0x70f2d86ecdbe176a,
	})
	return e
}

func (e *EdwardsPoint) IsIdentityI() int {
	return e.X.IsZero() & e.Y.IsOne() & e.Z.IsOne() & e.T.IsZero()
}

func (e *EdwardsPoint) IsOnCurve() int {
	xy := FpNew().Mul(e.X, e.Y)
	zt := FpNew().Mul(e.Z, e.T)

	// Y^2 + X^2 == Z^2 - T^2 * D

	yy := FpNew().Square(e.Y)
	xx := FpNew().Square(e.X)
	zz := FpNew().Square(e.Z)
	tt := FpNew().Square(e.T)
	lhs := FpNew().Add(yy, xx)
	rhs := FpNew().Add(zz, tt)
	rhs.Mul(rhs, edwardsD)

	return xy.EqualI(zt) & lhs.EqualI(rhs)
}

func (e *EdwardsPoint) IsTorsionFree() int {
	return new(EdwardsPoint).Mul(e, basePointOrder).IsIdentityI()
}

func (e *EdwardsPoint) Set(rhs *EdwardsPoint) *EdwardsPoint {
	e.X.Set(rhs.X)
	e.Y.Set(rhs.Y)
	e.Z.Set(rhs.Z)
	e.T.Set(rhs.T)
	return e
}

func (e *EdwardsPoint) EqualI(rhs *EdwardsPoint) int {
	xz := FpNew().Mul(e.X, rhs.Z)
	zx := FpNew().Mul(e.Z, rhs.X)

	yz := FpNew().Mul(e.Y, rhs.Z)
	zy := FpNew().Mul(e.Z, rhs.Y)

	return xz.EqualI(zx) & yz.EqualI(zy)
}

func (e *EdwardsPoint) Add(arg1, arg2 *EdwardsPoint) *EdwardsPoint {
	tmp := FpNew().Mul(arg1.Y, arg2.X)
	xyXY := FpNew().Mul(arg1.X, arg2.Y)
	// arg1.X * arg2.Y + arg1.Y * arg2.X
	xyXY.Add(xyXY, tmp)

	aXX := FpNew().Mul(arg1.X, arg2.X) // aX1X2
	dTT := FpNew().Mul(edwardsD, arg1.T)
	dTT.Mul(dTT, arg2.T)              // dT1T2
	zz := FpNew().Mul(arg1.Z, arg2.Z) // Z1Z2
	yy := FpNew().Mul(arg1.Y, arg2.Y)

	x := FpNew().Sub(zz, dTT)
	x.Mul(xyXY, x)

	tmp.Sub(yy, aXX)
	y := FpNew().Add(zz, dTT)
	y.Mul(y, tmp)

	t := FpNew().Sub(yy, aXX)
	t.Mul(t, xyXY)

	tmp.Add(zz, dTT)
	z := FpNew().Sub(zz, dTT)
	z.Mul(z, tmp)
	return &EdwardsPoint{x, y, z, t}
}

func (e *EdwardsPoint) Double(arg *EdwardsPoint) *EdwardsPoint {
	return e.Add(arg, arg)
}

func (e *EdwardsPoint) Negate(arg *EdwardsPoint) *EdwardsPoint {
	return &EdwardsPoint{
		X: FpNew().Neg(arg.X),
		Y: FpNew().Set(arg.Y),
		Z: FpNew().Set(arg.Z),
		T: FpNew().Neg(arg.T),
	}
}

func (e *EdwardsPoint) Mul(arg *EdwardsPoint, s *Fq) *EdwardsPoint {
	ss := FqNew().Div4(s)
	e.VariableBase(arg.ToTwisted(), ss).ToUntwisted()
	return e.Add(e, arg).scalarMod4(e, s)
}

func (e *EdwardsPoint) scalarMod4(arg *EdwardsPoint, s *Fq) *EdwardsPoint {
	ss := FqNew().fromMontgomery(s)
	sMod4 := ss.Value.Value[0] & 3

	zeroP := EdwardsPointNew().SetIdentity()
	twoP := EdwardsPointNew().Double(arg)
	threeP := EdwardsPointNew().Add(twoP, arg)

	isZero := internal.IsZeroUint64I(sMod4)
	isOne := internal.IsZeroUint64I(sMod4 - 1)
	isTwo := internal.IsZeroUint64I(sMod4 - 2)
	isThree := internal.IsZeroUint64I(sMod4 - 3)

	e.CMove(e, zeroP, isZero)
	e.CMove(e, arg, isOne)
	e.CMove(e, twoP, isTwo)
	e.CMove(e, threeP, isThree)
	return e
}

func (e *EdwardsPoint) VariableBase(arg *TwistedExtendedPoint, s *Fq) *TwistedExtendedPoint {
	result := TwistedExtensiblePointNew().SetIdentity()

	// Recode Scalar
	scalar := s.ToRadix16()

	lookup := FromTwistedExtendedPoint(arg)

	for i := 113; i >= 0; i-- {
		result.Double(result)
		result.Double(result)
		result.Double(result)
		result.Double(result)

		// The mask is the top bit, will be 1 for negative numbers, 0 for positive numbers
		mask := scalar[i] >> 7
		sign := mask & 0x1
		// Use the mask to get the absolute value of scalar
		absValue := uint32((scalar[i] + mask) ^ mask)

		negP := lookup.Select(absValue)
		negP.CNeg(negP, int(sign))

		result.AddProjectiveNiels(result, negP)
	}

	return result.ToExtended()
}

func (e *EdwardsPoint) Torque(arg *EdwardsPoint) *EdwardsPoint {
	return &EdwardsPoint{
		X: FpNew().Neg(arg.X),
		Y: FpNew().Neg(arg.Y),
		Z: FpNew().Set(arg.Z),
		T: FpNew().Set(arg.T),
	}
}

func (e *EdwardsPoint) Isogeny(a *Fp) *TwistedExtendedPoint {
	// Convert to affine now, then derive extended version later
	affine := new(EdwardsPoint).Set(e).ToAffine()

	// Compute x
	xy := FpNew().Mul(affine.X, affine.Y)
	xNum := FpNew().Double(xy)
	tmp := FpNew().Mul(a, FpNew().Square(affine.X))
	xDen := FpNew().Square(affine.Y)
	xDen.Sub(xDen, tmp)
	newX, _ := FpNew().Invert(xDen)
	newX.Mul(newX, xNum)

	// Compute y
	tmp.Square(affine.X)
	tmp.Mul(tmp, a)
	yNum := FpNew().Square(affine.Y)
	yNum.Add(yNum, tmp)

	yDen := FpNew().Double(one)
	tmp.Square(affine.Y)
	yDen.Sub(yDen, tmp)
	tmp.Square(affine.X)
	tmp.Mul(tmp, a)

	yDen.Sub(yDen, tmp)
	newY, _ := FpNew().Invert(yDen)
	newY.Mul(newY, yNum)

	return &TwistedExtendedPoint{
		X: newX,
		Y: newY,
		Z: one,
		T: FpNew().Mul(newX, newY),
	}
}

func (e *EdwardsPoint) ToAffine() *AffinePoint {
	z, _ := FpNew().Invert(e.Z)
	x := FpNew().Mul(e.X, z)
	y := FpNew().Mul(e.Y, z)
	return &AffinePoint{x, y}
}

func (e *EdwardsPoint) ToMontgomery() *MontgomeryPoint {
	// u = y^2 * [(1-dy^2)/(1-y^2)]

	affine := e.ToAffine()

	yy := FpNew().Square(affine.Y)
	dyy := FpNew().Mul(edwardsD, yy)

	t1 := FpNew().Sub(one, dyy)
	t2 := FpNew().Sub(one, yy)
	t2.Invert(t2)
	u := FpNew().Mul(yy, t1)
	u.Mul(u, t2)

	bytes := u.Bytes()
	return (*MontgomeryPoint)(&bytes)
}

func (e *EdwardsPoint) ToTwisted() *TwistedExtendedPoint {
	return e.Isogeny(one)
}

func (e *EdwardsPoint) CMove(a, b *EdwardsPoint, choice int) *EdwardsPoint {
	e.X.CMove(a.X, b.X, choice)
	e.Y.CMove(a.Y, b.Y, choice)
	e.Z.CMove(a.Z, b.Z, choice)
	e.T.CMove(a.T, b.T, choice)
	return e
}

func (e *EdwardsPoint) Compress() *CompressedEdwardsY {
	affine := e.ToAffine()

	var output PointBytesSeq
	sign := affine.X.Sgn0I()

	yBytes := affine.Y.Bytes()
	copy(output[:len(yBytes)], yBytes[:])
	output[PointBytes-1] = byte(sign) << 7
	return &output
}

func (e *EdwardsPoint) HashWithDefaults(msg []byte) *EdwardsPoint {
	return e.Hash(native.EllipticPointHasherShake256(), msg, []byte("edwards448_XOF:SHAKE256_ELL2_RO_"))
}

func (e *EdwardsPoint) Hash(hash *native.EllipticPointHasher, msg, dst []byte) *EdwardsPoint {
	var u []byte
	switch hash.Type() {
	case native.XMD:
		u = native.ExpandMsgXmd(hash, msg, dst, 168)
	case native.XOF:
		u = native.ExpandMsgXof(hash, msg, dst, 168)
	}
	var buf [112]byte
	copy(buf[:84], internal.ReverseBytes(u[:84]))
	u0 := FpNew().SetBytesWide(&buf)
	copy(buf[:84], internal.ReverseBytes(u[84:]))
	u1 := FpNew().SetBytesWide(&buf)
	q0 := AffinePointNew().mapToCurveElligator2(u0)
	q1 := AffinePointNew().mapToCurveElligator2(u1)
	q0.isogeny()
	q1.isogeny()

	r := EdwardsPointNew().Add(q0.ToEdwards(), q1.ToEdwards())
	r.Double(r)
	return r.Double(r)
}

type AffinePoint struct {
	X, Y *Fp
}

func AffinePointNew() *AffinePoint {
	return &AffinePoint{
		X: FpNew(),
		Y: FpNew(),
	}
}

func (a *AffinePoint) SetIdentity() *AffinePoint {
	a.X.Value.SetZero()
	a.Y.Value.SetOne()
	return a
}

func (a *AffinePoint) isogeny() *AffinePoint {
	t0 := FpNew().Square(a.X)  // x^2
	t1 := FpNew().Add(t0, one) // x^2+1
	t0.Sub(t0, one)            // x^2-1
	t2 := FpNew().Square(a.Y)  // y^2
	t2.Double(t2)              // 2y^2
	t3 := FpNew().Double(a.X)  // 2x

	t4 := FpNew().Mul(t0, a.Y) // y(x^2-1)
	t4.Double(t4)              // 2y(x^2-1)
	xNum := FpNew().Double(t3) // xNum = 4y(x^2-1)

	t5 := FpNew().Square(t0)    // x^4-2x^2+1
	t4.Add(t5, t2)              // x^4-2x^2+1+2y^2
	xDen := FpNew().Add(t4, t2) // xDen = x^4-2x^2+1+4y^2

	t5.Mul(t5, a.X)             // x^5-2x^3+x
	t4.Mul(t2, t3)              // 4xy^2
	yNum := FpNew().Sub(t4, t5) // yNum = -(x^5-2x^3+x-4xy^2)

	t4.Mul(t1, t2)              // 2x^2y^2+2y^2
	yDen := FpNew().Sub(t5, t4) // yDen = x^5-2x^3+x-2x^2y^2-2y^2

	_, _ = xDen.Invert(xDen)
	_, _ = yDen.Invert(yDen)
	a.X.Mul(xNum, xDen)
	a.Y.Mul(yNum, yDen)
	return a
}

func (a *AffinePoint) ToEdwards() *EdwardsPoint {
	return &EdwardsPoint{
		X: FpNew().Set(a.X),
		Y: FpNew().Set(a.Y),
		Z: FpNew().SetOne(),
		T: FpNew().Mul(a.X, a.Y),
	}
}

func (a *AffinePoint) CMove(arg1, arg2 *AffinePoint, choice int) *AffinePoint {
	a.X.CMove(arg1.X, arg2.X, choice)
	a.Y.CMove(arg1.Y, arg2.Y, choice)
	return a
}

func (a *AffinePoint) EqualI(rhs *AffinePoint) int {
	return a.X.EqualI(rhs.X) & a.Y.EqualI(rhs.Y)
}

func (a *AffinePoint) mapToCurveElligator2(u *Fp) *AffinePoint {
	t1 := FpNew().Square(u)
	t1.Mul(t1, z)
	e1 := t1.EqualI(minusOne)
	t1.CMove(t1, zero, e1)
	x1 := FpNew().Add(t1, one)
	_, _ = x1.Invert(x1)
	x1.Mul(x1, negJ)
	gx1 := FpNew().Add(x1, j)
	gx1.Mul(gx1, x1)
	gx1.Add(gx1, one)
	gx1.Mul(gx1, x1)
	x2 := FpNew().Neg(x1)
	x2.Sub(x2, j)
	gx2 := FpNew().Mul(t1, gx1)
	e2 := gx1.IsSquare()
	a.X.CMove(x2, x1, e2)
	y2 := FpNew().CMove(gx2, gx1, e2)
	a.Y.Sqrt(y2)
	e3 := a.Y.Sgn0I()
	a.Y.CNeg(a.Y, e2^e3)

	return a
}

type MontgomeryPoint [56]byte

var (
	MontgomeryPointLowA = [56]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	MontgomeryPointLowB = [56]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	MontgomeryPointLowC = [56]byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	}
)

func (m *MontgomeryPoint) SetGenerator() *MontgomeryPoint {
	m[0] = 5
	for i := 1; i < len(m); i++ {
		m[i] = 0
	}
	return m
}

func (m *MontgomeryPoint) EqualI(rhs *MontgomeryPoint) int {
	return subtle.ConstantTimeCompare(m[:], rhs[:])
}

func (m *MontgomeryPoint) IsLowOrder() int {
	return subtle.ConstantTimeCompare(m[:], MontgomeryPointLowA[:]) |
		subtle.ConstantTimeCompare(m[:], MontgomeryPointLowB[:]) |
		subtle.ConstantTimeCompare(m[:], MontgomeryPointLowC[:])
}

func (m *MontgomeryPoint) Mul(scalar *Fq) *MontgomeryPoint {
	// Algorithm 8 of Costello-Smith 2017
	affineU, err := FpNew().SetCanonicalBytes((*[56]byte)(m))
	if err != nil {
		return nil
	}

	x0 := NewProjectiveMontgomeryPoint().SetIdentity()
	x1 := &ProjectiveMontgomeryPoint{
		U: affineU,
		W: FpNew().SetOne(),
	}

	bits := internal.ReverseBytes(scalar.Value.Bytes())
	swap := 0
	for _, s := range bits {
		for i := 7; i >= 0; i-- {
			bit := int((s >> i) & 1)
			choice := swap ^ bit
			(&ProjectiveMontgomeryPoint{}).CSwap(x0, x1, choice)
			(&ProjectiveMontgomeryPoint{}).DifferentialAddAndDouble(x0, x1, affineU)
			swap = bit
		}
	}

	return x0.ToAffine()
}

func (m *MontgomeryPoint) Bytes() []byte {
	out := make([]byte, len(*m))
	copy(out, m[:])
	return out
}

func (m *MontgomeryPoint) ToProjective() *ProjectiveMontgomeryPoint {
	u, _ := FpNew().SetCanonicalBytes((*[56]byte)(m))
	return &ProjectiveMontgomeryPoint{U: u, W: FpNew().SetOne()}
}

type ProjectiveMontgomeryPoint struct {
	U, W *Fp
}

func NewProjectiveMontgomeryPoint() *ProjectiveMontgomeryPoint {
	return &ProjectiveMontgomeryPoint{
		U: FpNew(),
		W: FpNew(),
	}
}

func (p *ProjectiveMontgomeryPoint) SetIdentity() *ProjectiveMontgomeryPoint {
	p.U.SetOne()
	p.W.SetZero()
	return p
}

func (p *ProjectiveMontgomeryPoint) ToAffine() *MontgomeryPoint {
	x, _ := FpNew().Invert(p.W)
	x.Mul(x, p.U)
	b := x.Bytes()
	return (*MontgomeryPoint)(&b)
}

func (p *ProjectiveMontgomeryPoint) CMove(a, b *ProjectiveMontgomeryPoint, choice int) *ProjectiveMontgomeryPoint {
	p.U.CMove(a.U, b.U, choice)
	p.W.CMove(a.W, b.W, choice)
	return p
}

func (*ProjectiveMontgomeryPoint) CSwap(a, b *ProjectiveMontgomeryPoint, choice int) {
	a.U.CSwap(b.U, choice)
	a.W.CSwap(b.W, choice)
}

func (*ProjectiveMontgomeryPoint) DifferentialAddAndDouble(
	p, q *ProjectiveMontgomeryPoint,
	affinePmQ *Fp,
) {
	t0 := FpNew().Add(p.U, p.W)
	t1 := FpNew().Sub(p.U, p.W)
	t2 := FpNew().Add(q.U, q.W)
	t3 := FpNew().Sub(q.U, q.W)

	t4 := FpNew().Square(t0) // (U_P + W_P)^2 = U_P^2 + 2 U_P W_P + W_P^2
	t5 := FpNew().Square(t1) // (U_P - W_P)^2 = U_P^2 - 2 U_P W_P + W_P^2

	t6 := FpNew().Sub(t4, t5) // 4 U_P W_P

	t7 := FpNew().Mul(t0, t3) // (U_P + W_P) (U_Q - W_Q) = U_P U_Q + W_P U_Q - U_P W_Q - W_P W_Q
	t8 := FpNew().Mul(t1, t2) // (U_P - W_P) (U_Q + W_Q) = U_P U_Q - W_P U_Q + U_P W_Q - W_P W_Q

	t9 := FpNew().Add(t7, t8)  // 2 (U_P U_Q - W_P W_Q)
	t10 := FpNew().Sub(t7, t8) // 2 (W_P U_Q - U_P W_Q)

	t11 := FpNew().Square(t9)       // 4 (U_P U_Q - W_P W_Q)^2
	t12 := FpNew().Square(t10)      // 4 (W_P U_Q - U_P W_Q)^2
	t13 := FpNew().Mul(aP2Div4, t6) // (A + 2) U_P U_Q

	t14 := FpNew().Mul(t4, t5)  // ((U_P + W_P)(U_P - W_P))^2 = (U_P^2 - W_P^2)^2
	t15 := FpNew().Add(t13, t5) // (U_P - W_P)^2 + (A + 2) U_P W_P

	t16 := FpNew().Mul(t6, t15)        // 4 (U_P W_P) ((U_P - W_P)^2 + (A + 2) U_P W_P)
	t17 := FpNew().Mul(affinePmQ, t12) // U_D * 4 (W_P U_Q - U_P W_Q)^2
	//t18 := t11; // W_D * 4 (U_P U_Q - W_P W_Q)^2

	p.U = t14 // U_{P'} = (U_P + W_P)^2 (U_P - W_P)^2
	p.W = t16 // W_{P'} = (4 U_P W_P) ((U_P - W_P)^2 + ((A + 2)/4) 4 U_P W_P)
	q.U = t11 // U_{Q'} = W_D * 4 (U_P U_Q - W_P W_Q)^2
	q.W = t17 // W_{Q'} = U_D * 4 (W_P U_Q - U_P W_Q)^2
}
