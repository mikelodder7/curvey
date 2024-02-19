package ed448

import (
	"crypto/subtle"
	"fmt"
	"github.com/mikelodder7/curvey/internal"
	"github.com/mikelodder7/curvey/native"
	"github.com/mikelodder7/curvey/native/ed448/fp"
	"github.com/mikelodder7/curvey/native/ed448/fq"
	"math/big"
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
	yBytes := ([56]byte)(c[:56])
	sign := int(c[56])
	y, err := fp.FpNew().SetCanonicalBytes(&yBytes)
	if err != nil {
		return nil, err
	}
	yy := fp.FpNew().Square(y)
	dyy := fp.FpNew().Mul(fp.EdwardsD, yy)

	numerator := fp.FpNew().SetOne()
	numerator.Sub(numerator, yy)

	denominator := fp.FpNew().SetOne()
	denominator.Sub(denominator, dyy)
	x, isRes := fp.FpNew().SqrtRatio(numerator, denominator)

	signBit := sign >> 7
	isNegative := x.Sgn0I()
	x.CNeg(x, isNegative^signBit)

	pt := (&AffinePoint{X: x, Y: y}).ToEdwards()

	if isRes&pt.IsTorsionFree()&pt.IsOnCurve() == 1 {
		return pt, nil
	} else {
		return nil, fmt.Errorf("invalid point")
	}
}

type EdwardsPoint struct {
	X, Y, Z, T *fp.Fp
}

func EdwardsPointNew() *EdwardsPoint {
	return &EdwardsPoint{
		X: fp.FpNew(),
		Y: fp.FpNew(),
		Z: fp.FpNew(),
		T: fp.FpNew(),
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
	e.X.SetRaw(&fp.GoldilocksBasePointX)
	e.Y.SetRaw(&fp.GoldilocksBasePointY)
	e.Z.SetOne()
	e.T.SetRaw(&fp.GoldilocksBasePointT)
	return e
}

func (e *EdwardsPoint) SetBigInt(x, y *big.Int) (*EdwardsPoint, error) {
	pt := AffinePointNew()
	pt.X.SetBigInt(x)
	pt.Y.SetBigInt(y)
	ept := pt.ToEdwards()
	if ept.IsTorsionFree()&ept.IsOnCurve() == 1 {
		return ept, nil
	} else {
		return nil, fmt.Errorf("invalid point")
	}
}

func (e *EdwardsPoint) IsIdentityI() int {
	return e.EqualI(EdwardsPointNew().SetIdentity())
}

func (e *EdwardsPoint) IsOnCurve() int {
	xy := fp.FpNew().Mul(e.X, e.Y)
	zt := fp.FpNew().Mul(e.Z, e.T)

	// Y^2 + X^2 == Z^2 - T^2 * D

	yy := fp.FpNew().Square(e.Y)
	xx := fp.FpNew().Square(e.X)
	zz := fp.FpNew().Square(e.Z)
	tt := fp.FpNew().Square(e.T)
	lhs := fp.FpNew().Add(yy, xx)
	rhs := fp.FpNew().Mul(tt, fp.EdwardsD)
	rhs.Add(rhs, zz)

	return xy.EqualI(zt) & lhs.EqualI(rhs)
}

func (e *EdwardsPoint) IsTorsionFree() int {
	ss := fq.FqNew().SetRaw(&[7]uint64{
		0x48de30a4aad6113c,
		0x085b309ca37163d5,
		0x7113b6d26bb58da4,
		0xffffffffdf3288fa,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0x0fffffffffffffff,
	})
	pt := e.VariableBase(e.ToTwisted(), ss).ToUntwisted()
	r := EdwardsPointNew().Double(e)
	r.Add(r, e)

	pt.Add(pt, r)
	return pt.EqualI(EdwardsPointNew().SetIdentity())
}

func (e *EdwardsPoint) Set(rhs *EdwardsPoint) *EdwardsPoint {
	e.X.Set(rhs.X)
	e.Y.Set(rhs.Y)
	e.Z.Set(rhs.Z)
	e.T.Set(rhs.T)
	return e
}

func (e *EdwardsPoint) EqualI(rhs *EdwardsPoint) int {
	xz := fp.FpNew().Mul(e.X, rhs.Z)
	zx := fp.FpNew().Mul(e.Z, rhs.X)

	yz := fp.FpNew().Mul(e.Y, rhs.Z)
	zy := fp.FpNew().Mul(e.Z, rhs.Y)

	return xz.EqualI(zx) & yz.EqualI(zy)
}

func (e *EdwardsPoint) Add(arg1, arg2 *EdwardsPoint) *EdwardsPoint {
	tmp := fp.FpNew().Mul(arg1.Y, arg2.X)
	xyXY := fp.FpNew().Mul(arg1.X, arg2.Y)
	// arg1.X * arg2.Y + arg1.Y * arg2.X
	xyXY.Add(xyXY, tmp)

	aXX := fp.FpNew().Mul(arg1.X, arg2.X) // aX1X2
	dTT := fp.FpNew().Mul(fp.EdwardsD, arg1.T)
	dTT.Mul(dTT, arg2.T)                 // dT1T2
	zz := fp.FpNew().Mul(arg1.Z, arg2.Z) // Z1Z2
	yy := fp.FpNew().Mul(arg1.Y, arg2.Y)

	e.X.Sub(zz, dTT)
	e.X.Mul(e.X, xyXY)

	tmp.Sub(yy, aXX)
	e.Y.Add(zz, dTT)
	e.Y.Mul(e.Y, tmp)

	e.T.Sub(yy, aXX)
	e.T.Mul(e.T, xyXY)

	tmp.Add(zz, dTT)
	e.Z.Sub(zz, dTT)
	e.Z.Mul(e.Z, tmp)

	return e
}

func (e *EdwardsPoint) Double(arg *EdwardsPoint) *EdwardsPoint {
	return e.Add(arg, arg)
}

func (e *EdwardsPoint) Negate(arg *EdwardsPoint) *EdwardsPoint {
	e.X.Neg(arg.X)
	e.Y.Set(arg.Y)
	e.Z.Set(arg.Z)
	e.T.Neg(arg.T)
	return e
}

func (e *EdwardsPoint) Sub(arg1, arg2 *EdwardsPoint) *EdwardsPoint {
	return e.Add(arg1, EdwardsPointNew().Negate(arg2))
}

func (e *EdwardsPoint) Mul(arg *EdwardsPoint, s *fq.Fq) *EdwardsPoint {
	l := s.Value.Limbs()
	ll := ([7]uint64)(l)
	sq := fq.FqNew().SetRaw(&ll)

	ss := fq.FqNew().Div4(sq)
	res := EdwardsPointNew().Set(e.VariableBase(arg.ToTwisted(), ss).ToUntwisted())
	tt := EdwardsPointNew().scalarMod4(arg, s)

	return e.Add(res, tt)
}

func (e *EdwardsPoint) scalarMod4(arg *EdwardsPoint, s *fq.Fq) *EdwardsPoint {
	sMod4 := s.Value.Limbs()
	sMod4[0] &= 3

	zeroP := EdwardsPointNew().SetIdentity()
	twoP := EdwardsPointNew().Double(arg)
	threeP := EdwardsPointNew().Add(twoP, arg)

	isZero := internal.IsZeroUint64I(sMod4[0])
	isOne := internal.IsZeroUint64I(sMod4[0] - 1)
	isTwo := internal.IsZeroUint64I(sMod4[0] - 2)
	isThree := internal.IsZeroUint64I(sMod4[0] - 3)

	e.CMove(e, zeroP, isZero)
	e.CMove(e, arg, isOne)
	e.CMove(e, twoP, isTwo)
	e.CMove(e, threeP, isThree)
	return e
}

func (*EdwardsPoint) VariableBase(arg *TwistedExtendedPoint, s *fq.Fq) *TwistedExtendedPoint {
	result := TwistedExtensiblePointNew().SetIdentity()

	// Recode Scalar
	scalar := s.ToRadix16()

	lookup := FromTwistedExtendedPoint(arg)

	for i := 112; i >= 0; i-- {
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
	e.X.Neg(arg.X)
	e.Y.Neg(arg.Y)
	e.Z.Set(arg.Z)
	e.T.Set(arg.T)
	return e
}

func (e *EdwardsPoint) Isogeny(a *fp.Fp) *TwistedExtendedPoint {
	// Convert to affine now, then derive extended version later
	affine := EdwardsPointNew().Set(e).ToAffine()

	// Compute x
	xy := fp.FpNew().Mul(affine.X, affine.Y)
	xNum := fp.FpNew().Double(xy)
	tmp := fp.FpNew().Mul(a, fp.FpNew().Square(affine.X))
	xDen := fp.FpNew().Square(affine.Y)
	xDen.Sub(xDen, tmp)
	newX, _ := fp.FpNew().Invert(xDen)
	newX.Mul(newX, xNum)

	// Compute y
	tmp.Square(affine.X)
	tmp.Mul(tmp, a)
	yNum := fp.FpNew().Square(affine.Y)
	yNum.Add(yNum, tmp)

	yDen := fp.FpNew().Double(fp.One)
	tmp.Square(affine.Y)
	yDen.Sub(yDen, tmp)
	tmp.Square(affine.X)
	tmp.Mul(tmp, a)

	yDen.Sub(yDen, tmp)
	newY, _ := fp.FpNew().Invert(yDen)
	newY.Mul(newY, yNum)

	return &TwistedExtendedPoint{
		X: newX,
		Y: newY,
		Z: fp.FpNew().SetOne(),
		T: fp.FpNew().Mul(newX, newY),
	}
}

func (e *EdwardsPoint) ToAffine() *AffinePoint {
	z, _ := fp.FpNew().Invert(e.Z)
	x := fp.FpNew().Mul(e.X, z)
	y := fp.FpNew().Mul(e.Y, z)
	return &AffinePoint{x, y}
}

func (e *EdwardsPoint) ToMontgomery() *MontgomeryPoint {
	// u = y^2 * [(1-dy^2)/(1-y^2)]

	affine := e.ToAffine()

	yy := fp.FpNew().Square(affine.Y)
	dyy := fp.FpNew().Mul(fp.EdwardsD, yy)

	t1 := fp.FpNew().Sub(fp.One, dyy)
	t2 := fp.FpNew().Sub(fp.One, yy)
	t2.Invert(t2)
	u := fp.FpNew().Mul(yy, t1)
	u.Mul(u, t2)

	bytes := u.Bytes()
	return (*MontgomeryPoint)(&bytes)
}

func (e *EdwardsPoint) ToTwisted() *TwistedExtendedPoint {
	return e.Isogeny(fp.One)
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

func (e *EdwardsPoint) BigInt() (*big.Int, *big.Int) {
	pt := e.ToAffine()
	return pt.X.BigInt(), pt.Y.BigInt()
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
	u0 := fp.FpNew().SetBytesWide(&buf)
	copy(buf[:84], internal.ReverseBytes(u[84:]))
	u1 := fp.FpNew().SetBytesWide(&buf)

	q0 := AffinePointNew().mapToCurveElligator2(u0)
	q1 := AffinePointNew().mapToCurveElligator2(u1)

	q0.isogeny()
	q1.isogeny()

	r := EdwardsPointNew().Add(q0.ToEdwards(), q1.ToEdwards())
	r.Double(r)
	return r.Double(r)
}

func (e *EdwardsPoint) SumOfProducts(points []*EdwardsPoint, scalars []*fq.Fq) (*EdwardsPoint, error) {
	const Upper = 256
	const W = 4
	const Windows = Upper / W // careful--use ceiling division in case this doesn't divide evenly
	if len(points) != len(scalars) {
		return nil, fmt.Errorf("length mismatch")
	}

	bucketSize := 1 << W
	windows := make([]*EdwardsPoint, Windows)
	bytes := make([][57]byte, len(scalars))
	buckets := make([]*EdwardsPoint, bucketSize)

	for i, scalar := range scalars {
		bytes[i] = scalar.Bytes()
	}
	for i := range windows {
		windows[i] = EdwardsPointNew().SetIdentity()
	}

	for i := 0; i < bucketSize; i++ {
		buckets[i] = EdwardsPointNew().SetIdentity()
	}

	sum := EdwardsPointNew().SetIdentity()

	for j := 0; j < len(windows); j++ {
		for i := 0; i < bucketSize; i++ {
			buckets[i].SetIdentity()
		}

		for i := 0; i < len(scalars); i++ {
			// j*W to get the nibble
			// >> 3 to convert to byte, / 8
			// (W * j & W) gets the nibble, mod W
			// 1 << W - 1 to get the offset
			index := bytes[i][j*W>>3] >> (W * j & W) & (1<<W - 1) // little-endian
			buckets[index].Add(buckets[index], points[i])
		}

		sum.SetIdentity()

		for i := bucketSize - 1; i > 0; i-- {
			sum.Add(sum, buckets[i])
			windows[j].Add(windows[j], sum)
		}
	}

	e.SetIdentity()
	for i := len(windows) - 1; i >= 0; i-- {
		for j := 0; j < W; j++ {
			e.Double(e)
		}

		e.Add(e, windows[i])
	}
	return e, nil
}

type AffinePoint struct {
	X, Y *fp.Fp
}

func AffinePointNew() *AffinePoint {
	return &AffinePoint{
		X: fp.FpNew(),
		Y: fp.FpNew(),
	}
}

func (a *AffinePoint) SetIdentity() *AffinePoint {
	a.X.SetZero()
	a.Y.SetOne()
	return a
}

func (a *AffinePoint) isogeny() *AffinePoint {
	t0 := fp.FpNew().Square(a.X)     // x^2
	t1 := fp.FpNew().Add(t0, fp.One) // x^2+1
	t0.Sub(t0, fp.One)               // x^2-1
	t2 := fp.FpNew().Square(a.Y)     // y^2
	t2.Double(t2)                    // 2y^2
	t3 := fp.FpNew().Double(a.X)     // 2x

	t4 := fp.FpNew().Mul(t0, a.Y) // y(x^2-1)
	t4.Double(t4)                 // 2y(x^2-1)
	xNum := fp.FpNew().Double(t4) // xNum = 4y(x^2-1)

	t5 := fp.FpNew().Square(t0)    // x^4-2x^2+1
	t4.Add(t5, t2)                 // x^4-2x^2+1+2y^2
	xDen := fp.FpNew().Add(t4, t2) // xDen = x^4-2x^2+1+4y^2

	t5.Mul(t5, a.X)                // x^5-2x^3+x
	t4.Mul(t2, t3)                 // 4xy^2
	yNum := fp.FpNew().Sub(t4, t5) // yNum = -(x^5-2x^3+x-4xy^2)

	t4.Mul(t1, t2)                 // 2x^2y^2+2y^2
	yDen := fp.FpNew().Sub(t5, t4) // yDen = x^5-2x^3+x-2x^2y^2-2y^2

	_, _ = xDen.Invert(xDen)
	_, _ = yDen.Invert(yDen)
	a.X.Mul(xNum, xDen)
	a.Y.Mul(yNum, yDen)
	return a
}

func (a *AffinePoint) ToEdwards() *EdwardsPoint {
	return &EdwardsPoint{
		X: fp.FpNew().Set(a.X),
		Y: fp.FpNew().Set(a.Y),
		Z: fp.FpNew().SetOne(),
		T: fp.FpNew().Mul(a.X, a.Y),
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

func (a *AffinePoint) mapToCurveElligator2(u *fp.Fp) *AffinePoint {
	t1 := fp.FpNew().Square(u)
	t1.Mul(t1, z)
	e1 := t1.EqualI(fp.MinusOne)
	t1.CMove(t1, fp.Zero, e1)
	x1 := fp.FpNew().Add(t1, fp.One)
	_, _ = x1.Invert(x1)
	x1.Mul(x1, negJ)
	gx1 := fp.FpNew().Add(x1, j)
	gx1.Mul(gx1, x1)
	gx1.Add(gx1, fp.One)
	gx1.Mul(gx1, x1)
	x2 := fp.FpNew().Neg(x1)
	x2.Sub(x2, j)
	gx2 := fp.FpNew().Mul(t1, gx1)
	e2 := gx1.IsSquare()
	a.X.CMove(x2, x1, e2)
	y2 := fp.FpNew().CMove(gx2, gx1, e2)
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

func (m *MontgomeryPoint) Mul(scalar *fq.Fq) *MontgomeryPoint {
	// Algorithm 8 of Costello-Smith 2017
	affineU, err := fp.FpNew().SetCanonicalBytes((*[56]byte)(m))
	if err != nil {
		return nil
	}

	x0 := NewProjectiveMontgomeryPoint().SetIdentity()
	x1 := &ProjectiveMontgomeryPoint{
		U: affineU,
		W: fp.FpNew().SetOne(),
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
	u, _ := fp.FpNew().SetCanonicalBytes((*[56]byte)(m))
	return &ProjectiveMontgomeryPoint{U: u, W: fp.FpNew().SetOne()}
}

type ProjectiveMontgomeryPoint struct {
	U, W *fp.Fp
}

func NewProjectiveMontgomeryPoint() *ProjectiveMontgomeryPoint {
	return &ProjectiveMontgomeryPoint{
		U: fp.FpNew(),
		W: fp.FpNew(),
	}
}

func (p *ProjectiveMontgomeryPoint) SetIdentity() *ProjectiveMontgomeryPoint {
	p.U.SetOne()
	p.W.SetZero()
	return p
}

func (p *ProjectiveMontgomeryPoint) ToAffine() *MontgomeryPoint {
	x, _ := fp.FpNew().Invert(p.W)
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
	affinePmQ *fp.Fp,
) {
	t0 := fp.FpNew().Add(p.U, p.W)
	t1 := fp.FpNew().Sub(p.U, p.W)
	t2 := fp.FpNew().Add(q.U, q.W)
	t3 := fp.FpNew().Sub(q.U, q.W)

	t4 := fp.FpNew().Square(t0) // (U_P + W_P)^2 = U_P^2 + 2 U_P W_P + W_P^2
	t5 := fp.FpNew().Square(t1) // (U_P - W_P)^2 = U_P^2 - 2 U_P W_P + W_P^2

	t6 := fp.FpNew().Sub(t4, t5) // 4 U_P W_P

	t7 := fp.FpNew().Mul(t0, t3) // (U_P + W_P) (U_Q - W_Q) = U_P U_Q + W_P U_Q - U_P W_Q - W_P W_Q
	t8 := fp.FpNew().Mul(t1, t2) // (U_P - W_P) (U_Q + W_Q) = U_P U_Q - W_P U_Q + U_P W_Q - W_P W_Q

	t9 := fp.FpNew().Add(t7, t8)  // 2 (U_P U_Q - W_P W_Q)
	t10 := fp.FpNew().Sub(t7, t8) // 2 (W_P U_Q - U_P W_Q)

	t11 := fp.FpNew().Square(t9)          // 4 (U_P U_Q - W_P W_Q)^2
	t12 := fp.FpNew().Square(t10)         // 4 (W_P U_Q - U_P W_Q)^2
	t13 := fp.FpNew().Mul(fp.Ap2Div4, t6) // (A + 2) U_P U_Q

	t14 := fp.FpNew().Mul(t4, t5)  // ((U_P + W_P)(U_P - W_P))^2 = (U_P^2 - W_P^2)^2
	t15 := fp.FpNew().Add(t13, t5) // (U_P - W_P)^2 + (A + 2) U_P W_P

	t16 := fp.FpNew().Mul(t6, t15)        // 4 (U_P W_P) ((U_P - W_P)^2 + (A + 2) U_P W_P)
	t17 := fp.FpNew().Mul(affinePmQ, t12) // U_D * 4 (W_P U_Q - U_P W_Q)^2
	//t18 := t11; // W_D * 4 (U_P U_Q - W_P W_Q)^2

	p.U = t14 // U_{P'} = (U_P + W_P)^2 (U_P - W_P)^2
	p.W = t16 // W_{P'} = (4 U_P W_P) ((U_P - W_P)^2 + ((A + 2)/4) 4 U_P W_P)
	q.U = t11 // U_{Q'} = W_D * 4 (U_P U_Q - W_P W_Q)^2
	q.W = t17 // W_{Q'} = U_D * 4 (W_P U_Q - U_P W_Q)^2
}
