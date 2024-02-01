//
// SPDX-License-Identifier: Apache-2.0
//

package curvey

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"github.com/bwesterb/go-ristretto"
	"github.com/mikelodder7/curvey/native"
	"io"
	"math/big"

	"filippo.io/edwards25519"
	"filippo.io/edwards25519/field"
	ed "github.com/bwesterb/go-ristretto/edwards25519"

	"github.com/mikelodder7/curvey/internal"
)

var (
	a, _ = new(field.Element).SetBytes([]byte{
		6, 109, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	})
	minA        = new(field.Element).Negate(a)
	edZero      = new(field.Element).Zero()
	edOne       = new(field.Element).One()
	minOne      = new(field.Element).Negate(edOne)
	two         = new(field.Element).Add(edOne, edOne)
	invsqrtD, _ = new(field.Element).SetBytes([]byte{
		6, 126, 69, 255, 170, 4, 110, 204, 130, 26, 125, 75, 209, 211, 161, 197,
		126, 79, 252, 3, 220, 8, 123, 210, 187, 6, 160, 96, 244, 237, 38, 15,
	})
)

type ScalarEd25519 struct {
	value *edwards25519.Scalar
}

type PointEd25519 struct {
	value *edwards25519.Point
}

type ScalarRistretto25519 struct {
	value *ristretto.Scalar
}

type PointRistretto25519 struct {
	value *ristretto.Point
}

var scOne, _ = edwards25519.NewScalar().SetCanonicalBytes([]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

func (s *ScalarEd25519) Random(reader io.Reader) Scalar {
	if reader == nil {
		return nil
	}
	var seed [64]byte
	_, _ = reader.Read(seed[:])
	return s.Hash(seed[:])
}

func (*ScalarEd25519) Hash(b []byte) Scalar {
	v := new(ristretto.Scalar).Derive(b)
	var data [32]byte
	v.BytesInto(&data)
	value, err := edwards25519.NewScalar().SetCanonicalBytes(data[:])
	if err != nil {
		return nil
	}
	return &ScalarEd25519{value}
}

func (*ScalarEd25519) Zero() Scalar {
	return &ScalarEd25519{
		value: edwards25519.NewScalar(),
	}
}

func (*ScalarEd25519) One() Scalar {
	return &ScalarEd25519{
		value: edwards25519.NewScalar().Set(scOne),
	}
}

func (s *ScalarEd25519) IsZero() bool {
	i := byte(0)
	for _, b := range s.value.Bytes() {
		i |= b
	}
	return i == 0
}

func (s *ScalarEd25519) IsOne() bool {
	data := s.value.Bytes()
	i := byte(0)
	for j := 1; j < len(data); j++ {
		i |= data[j]
	}
	return i == 0 && data[0] == 1
}

func (s *ScalarEd25519) IsOdd() bool {
	return s.value.Bytes()[0]&1 == 1
}

func (s *ScalarEd25519) IsEven() bool {
	return s.value.Bytes()[0]&1 == 0
}

func (*ScalarEd25519) New(input int) Scalar {
	var data [64]byte
	i := input
	if input < 0 {
		i = -input
	}
	data[0] = byte(i)
	data[1] = byte(i >> 8)
	data[2] = byte(i >> 16)
	data[3] = byte(i >> 24)
	value, err := edwards25519.NewScalar().SetUniformBytes(data[:])
	if err != nil {
		return nil
	}
	if input < 0 {
		value.Negate(value)
	}

	return &ScalarEd25519{
		value,
	}
}

func (s *ScalarEd25519) Cmp(rhs Scalar) int {
	r := s.Sub(rhs)
	if r != nil && r.IsZero() {
		return 0
	} else {
		return -2
	}
}

func (s *ScalarEd25519) Square() Scalar {
	value := edwards25519.NewScalar().Multiply(s.value, s.value)
	return &ScalarEd25519{value}
}

func (s *ScalarEd25519) Pow(exp uint64) Scalar {
	out := s.Clone()

	for j := 63; j >= 0; j-- {
		square := out.Square()
		squareMul := square.Mul(square)
		out = cSelect(out, square, squareMul, (exp>>j)&1)
	}

	return out
}

func cSelect(z, x, y Scalar, which uint64) Scalar {
	if which != 0 && which != 1 {
		panic("which must be 0 or 1")
	}

	mask := -byte(which)
	xBytes := x.Bytes()
	yBytes := y.Bytes()
	for i, xByte := range xBytes {
		xBytes[i] ^= (xByte ^ yBytes[i]) & mask
	}
	out, err := z.SetBytes(xBytes)
	if err != nil {
		panic("error setting bytes")
	}

	return out
}

func (s *ScalarEd25519) Double() Scalar {
	return &ScalarEd25519{
		value: edwards25519.NewScalar().Add(s.value, s.value),
	}
}

func (s *ScalarEd25519) Invert() (Scalar, error) {
	return &ScalarEd25519{
		value: edwards25519.NewScalar().Invert(s.value),
	}, nil
}

func (s *ScalarEd25519) Sqrt() (Scalar, error) {
	bi25519, _ := new(big.Int).SetString("1000000000000000000000000000000014DEF9DEA2F79CD65812631A5CF5D3ED", 16)
	x := s.BigInt()
	x.ModSqrt(x, bi25519)
	return s.SetBigInt(x)
}

func (s *ScalarEd25519) Cube() Scalar {
	value := edwards25519.NewScalar().Multiply(s.value, s.value)
	value.Multiply(value, s.value)
	return &ScalarEd25519{value}
}

func (s *ScalarEd25519) Add(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarEd25519)
	if ok {
		return &ScalarEd25519{
			value: edwards25519.NewScalar().Add(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarEd25519) Sub(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarEd25519)
	if ok {
		return &ScalarEd25519{
			value: edwards25519.NewScalar().Subtract(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarEd25519) Mul(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarEd25519)
	if ok {
		return &ScalarEd25519{
			value: edwards25519.NewScalar().Multiply(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarEd25519) MulAdd(y, z Scalar) Scalar {
	yy, ok := y.(*ScalarEd25519)
	if !ok {
		return nil
	}
	zz, ok := z.(*ScalarEd25519)
	if !ok {
		return nil
	}
	return &ScalarEd25519{value: edwards25519.NewScalar().MultiplyAdd(s.value, yy.value, zz.value)}
}

func (s *ScalarEd25519) Div(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarEd25519)
	if ok {
		value := edwards25519.NewScalar().Invert(r.value)
		value.Multiply(value, s.value)
		return &ScalarEd25519{value}
	} else {
		return nil
	}
}

func (s *ScalarEd25519) Neg() Scalar {
	return &ScalarEd25519{
		value: edwards25519.NewScalar().Negate(s.value),
	}
}

func (*ScalarEd25519) SetBigInt(x *big.Int) (Scalar, error) {
	if x == nil {
		return nil, fmt.Errorf("invalid value")
	}

	bi25519, _ := new(big.Int).SetString("1000000000000000000000000000000014DEF9DEA2F79CD65812631A5CF5D3ED", 16)
	var v big.Int
	buf := v.Mod(x, bi25519).Bytes()
	var rBuf [32]byte
	for i := 0; i < len(buf) && i < 32; i++ {
		rBuf[i] = buf[len(buf)-i-1]
	}
	value, err := edwards25519.NewScalar().SetCanonicalBytes(rBuf[:])
	if err != nil {
		return nil, err
	}
	return &ScalarEd25519{value}, nil
}

func (s *ScalarEd25519) BigInt() *big.Int {
	var ret big.Int
	buf := internal.ReverseScalarBytes(s.value.Bytes())
	return ret.SetBytes(buf)
}

func (s *ScalarEd25519) Bytes() []byte {
	return s.value.Bytes()
}

// SetBytes takes input a 32-byte long array and returns a ed25519 scalar.
// The input must be 32-byte long and must be a reduced bytes.
func (*ScalarEd25519) SetBytes(input []byte) (Scalar, error) {
	if len(input) != 32 {
		return nil, fmt.Errorf("invalid byte sequence")
	}
	value, err := edwards25519.NewScalar().SetCanonicalBytes(input)
	if err != nil {
		return nil, err
	}
	return &ScalarEd25519{value}, nil
}

// SetBytesWide takes input a 64-byte long byte array, reduce it and return an ed25519 scalar.
// It uses SetUniformBytes of fillipo.io/edwards25519 - https://github.com/FiloSottile/edwards25519/blob/v1.0.0-rc.1/scalar.go#L85
// If bytes is not of the right length, it returns nil and an error.
func (*ScalarEd25519) SetBytesWide(b []byte) (Scalar, error) {
	value, err := edwards25519.NewScalar().SetUniformBytes(b)
	if err != nil {
		return nil, err
	}
	return &ScalarEd25519{value}, nil
}

// SetBytesClamping uses SetBytesWithClamping of fillipo.io/edwards25519- https://github.com/FiloSottile/edwards25519/blob/v1.0.0-rc.1/scalar.go#L135
// which applies the buffer pruning described in RFC 8032, Section 5.1.5 (also known as clamping)
// and sets bytes to the result. The input must be 32-byte long, and it is not modified.
// If bytes is not of the right length, SetBytesWithClamping returns nil and an error, and the receiver is unchanged.
func (*ScalarEd25519) SetBytesClamping(b []byte) (Scalar, error) {
	value, err := edwards25519.NewScalar().SetBytesWithClamping(b)
	if err != nil {
		return nil, err
	}
	return &ScalarEd25519{value}, nil
}

// SetBytesCanonical uses SetCanonicalBytes of fillipo.io/edwards25519.
// https://github.com/FiloSottile/edwards25519/blob/v1.0.0-rc.1/scalar.go#L98
// This function takes an input x and sets s = x, where x is a 32-byte little-endian
// encoding of s, then it returns the corresponding ed25519 scalar. If the input is
// not a canonical encoding of s, it returns nil and an error.
func (s *ScalarEd25519) SetBytesCanonical(b []byte) (Scalar, error) {
	return s.SetBytes(b)
}

func (*ScalarEd25519) Point() Point {
	return new(PointEd25519).Identity()
}

func (s *ScalarEd25519) Clone() Scalar {
	return &ScalarEd25519{
		value: edwards25519.NewScalar().Set(s.value),
	}
}

func (s *ScalarEd25519) MarshalBinary() ([]byte, error) {
	return ScalarMarshalBinary(s)
}

func (s *ScalarEd25519) UnmarshalBinary(input []byte) error {
	sc, err := ScalarUnmarshalBinary(input)
	if err != nil {
		return err
	}
	ss, ok := sc.(*ScalarEd25519)
	if !ok {
		return fmt.Errorf("invalid scalar")
	}
	s.value = ss.value
	return nil
}

func (s *ScalarEd25519) MarshalText() ([]byte, error) {
	return ScalarMarshalText(s)
}

func (s *ScalarEd25519) UnmarshalText(input []byte) error {
	sc, err := ScalarUnmarshalText(input)
	if err != nil {
		return err
	}
	ss, ok := sc.(*ScalarEd25519)
	if !ok {
		return fmt.Errorf("invalid scalar")
	}
	s.value = ss.value
	return nil
}

func (s *ScalarEd25519) GetEdwardsScalar() *edwards25519.Scalar {
	return edwards25519.NewScalar().Set(s.value)
}

func (*ScalarEd25519) SetEdwardsScalar(sc *edwards25519.Scalar) *ScalarEd25519 {
	return &ScalarEd25519{value: edwards25519.NewScalar().Set(sc)}
}

func (s *ScalarEd25519) MarshalJSON() ([]byte, error) {
	return ScalarMarshalJSON(s)
}

func (s *ScalarEd25519) UnmarshalJSON(input []byte) error {
	sc, err := ScalarUnmarshalJSON(input)
	if err != nil {
		return err
	}
	S, ok := sc.(*ScalarEd25519)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	s.value = S.value
	return nil
}

func (p *PointEd25519) Random(reader io.Reader) Point {
	var seed [64]byte
	_, _ = reader.Read(seed[:])
	return p.Hash(seed[:])
}

func (*PointEd25519) Hash(b []byte) Point {
	// Perform hashing to the group using the Elligator2 map
	//
	// See https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-11#section-6.7.1
	dst := []byte("edwards25519_XMD:SHA-512_ELL2_RO_")
	u := native.ExpandMsgXmd(native.EllipticPointHasherSha512(), b, dst, 96)
	var t [64]byte
	copy(t[:48], internal.ReverseScalarBytes(u[:48]))
	u0, _ := new(field.Element).SetWideBytes(t[:])
	copy(t[:48], internal.ReverseScalarBytes(u[48:96]))
	u1, _ := new(field.Element).SetWideBytes(t[:])

	p0 := mapToEdwards(u0)
	p1 := mapToEdwards(u1)
	p0.Add(p0, p1)
	p0.MultByCofactor(p0)
	return &PointEd25519{
		value: p0,
	}
}

func (*PointEd25519) Identity() Point {
	return &PointEd25519{
		value: edwards25519.NewIdentityPoint(),
	}
}

func (*PointEd25519) Generator() Point {
	return &PointEd25519{
		value: edwards25519.NewGeneratorPoint(),
	}
}

func (p *PointEd25519) IsIdentity() bool {
	return p.Equal(p.Identity())
}

func (*PointEd25519) IsNegative() bool {
	// Negative points don't really exist in ed25519
	return false
}

func (p *PointEd25519) IsOnCurve() bool {
	_, err := edwards25519.NewIdentityPoint().SetBytes(p.ToAffineCompressed())
	return err == nil
}

func (p *PointEd25519) Double() Point {
	return &PointEd25519{value: edwards25519.NewIdentityPoint().Add(p.value, p.value)}
}

func (*PointEd25519) Scalar() Scalar {
	return new(ScalarEd25519).Zero()
}

func (p *PointEd25519) Neg() Point {
	return &PointEd25519{value: edwards25519.NewIdentityPoint().Negate(p.value)}
}

func (p *PointEd25519) Add(rhs Point) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*PointEd25519)
	if ok {
		return &PointEd25519{value: edwards25519.NewIdentityPoint().Add(p.value, r.value)}
	} else {
		return nil
	}
}

func (p *PointEd25519) Sub(rhs Point) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*PointEd25519)
	if ok {
		rTmp := edwards25519.NewIdentityPoint().Negate(r.value)
		return &PointEd25519{value: edwards25519.NewIdentityPoint().Add(p.value, rTmp)}
	} else {
		return nil
	}
}

func (p *PointEd25519) Mul(rhs Scalar) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*ScalarEd25519)
	if ok {
		value := edwards25519.NewIdentityPoint().ScalarMult(r.value, p.value)
		return &PointEd25519{value}
	} else {
		return nil
	}
}

// MangleScalarBitsAndMulByBasepointToProducePublicKey
// is a function for mangling the bits of a (formerly
// mathematically well-defined) "scalar" and multiplying it to produce a
// public key.
func (*PointEd25519) MangleScalarBitsAndMulByBasepointToProducePublicKey(rhs *ScalarEd25519) *PointEd25519 {
	data := rhs.value.Bytes()
	s, err := edwards25519.NewScalar().SetBytesWithClamping(data)
	if err != nil {
		return nil
	}
	value := edwards25519.NewIdentityPoint().ScalarBaseMult(s)
	return &PointEd25519{value}
}

func (p *PointEd25519) Equal(rhs Point) bool {
	r, ok := rhs.(*PointEd25519)
	if ok {
		// We would like to check that the point (X/Z, Y/Z) is equal to
		// the point (X'/Z', Y'/Z') without converting into affine
		// coordinates (x, y) and (x', y'), which requires two inversions.
		// We have that X = xZ and X' = x'Z'. Thus, x = x' is equivalent to
		// (xZ)Z' = (x'Z')Z, and similarly for the y-coordinate.
		return p.value.Equal(r.value) == 1
	} else {
		return false
	}
}

func (p *PointEd25519) Set(x, y *big.Int) (Point, error) {
	// check is identity
	xx := subtle.ConstantTimeCompare(x.Bytes(), []byte{})
	yy := subtle.ConstantTimeCompare(y.Bytes(), []byte{})
	if (xx | yy) == 1 {
		return p.Identity(), nil
	}
	xElem := new(ed.FieldElement).SetBigInt(x)
	yElem := new(ed.FieldElement).SetBigInt(y)

	var data [32]byte
	var affine [64]byte
	xElem.BytesInto(&data)
	copy(affine[:32], data[:])
	yElem.BytesInto(&data)
	copy(affine[32:], data[:])
	return p.FromAffineUncompressed(affine[:])
}

func (p *PointEd25519) ToAffineCompressed() []byte {
	return p.value.Bytes()
}

func (p *PointEd25519) ToAffineUncompressed() []byte {
	x, y, z, _ := p.value.ExtendedCoordinates()
	recip := new(field.Element).Invert(z)
	x.Multiply(x, recip)
	y.Multiply(y, recip)
	var out [64]byte
	copy(out[:32], x.Bytes())
	copy(out[32:], y.Bytes())
	return out[:]
}

func (*PointEd25519) FromAffineCompressed(b []byte) (Point, error) {
	pt, err := edwards25519.NewIdentityPoint().SetBytes(b)
	if err != nil {
		return nil, err
	}
	return &PointEd25519{value: pt}, nil
}

func (*PointEd25519) FromAffineUncompressed(b []byte) (Point, error) {
	if len(b) != 64 {
		return nil, fmt.Errorf("invalid byte sequence")
	}
	if bytes.Equal(b, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		return &PointEd25519{value: edwards25519.NewIdentityPoint()}, nil
	}
	x, err := new(field.Element).SetBytes(b[:32])
	if err != nil {
		return nil, err
	}
	y, err := new(field.Element).SetBytes(b[32:])
	if err != nil {
		return nil, err
	}
	z := new(field.Element).One()
	t := new(field.Element).Multiply(x, y)
	value, err := edwards25519.NewIdentityPoint().SetExtendedCoordinates(x, y, z, t)
	if err != nil {
		return nil, err
	}
	return &PointEd25519{value}, nil
}

func (*PointEd25519) CurveName() string {
	return ED25519Name
}

func (*PointEd25519) SumOfProducts(points []Point, scalars []Scalar) Point {
	nScalars := make([]*edwards25519.Scalar, len(scalars))
	nPoints := make([]*edwards25519.Point, len(points))
	for i, sc := range scalars {
		s, err := edwards25519.NewScalar().SetCanonicalBytes(sc.Bytes())
		if err != nil {
			return nil
		}
		nScalars[i] = s
	}
	for i, pt := range points {
		pp, ok := pt.(*PointEd25519)
		if !ok {
			return nil
		}
		nPoints[i] = pp.value
	}
	pt := edwards25519.NewIdentityPoint().MultiScalarMult(nScalars, nPoints)
	return &PointEd25519{value: pt}
}

func (*PointEd25519) VarTimeDoubleScalarBaseMult(a Scalar, capA Point, b Scalar) Point {
	AA, ok := capA.(*PointEd25519)
	if !ok {
		return nil
	}
	aa, ok := a.(*ScalarEd25519)
	if !ok {
		return nil
	}
	bb, ok := b.(*ScalarEd25519)
	if !ok {
		return nil
	}
	value := edwards25519.NewIdentityPoint().VarTimeDoubleScalarBaseMult(aa.value, AA.value, bb.value)
	return &PointEd25519{value}
}

func (p *PointEd25519) MarshalBinary() ([]byte, error) {
	return PointMarshalBinary(p)
}

func (p *PointEd25519) UnmarshalBinary(input []byte) error {
	pt, err := PointUnmarshalBinary(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointEd25519)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.value = ppt.value
	return nil
}

func (p *PointEd25519) MarshalText() ([]byte, error) {
	return PointMarshalText(p)
}

func (p *PointEd25519) UnmarshalText(input []byte) error {
	pt, err := PointUnmarshalText(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointEd25519)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.value = ppt.value
	return nil
}

func (p *PointEd25519) MarshalJSON() ([]byte, error) {
	return PointMarshalJSON(p)
}

func (p *PointEd25519) UnmarshalJSON(input []byte) error {
	pt, err := PointUnmarshalJSON(input)
	if err != nil {
		return err
	}
	P, ok := pt.(*PointEd25519)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	p.value = P.value
	return nil
}

func (p *PointEd25519) GetEdwardsPoint() *edwards25519.Point {
	return edwards25519.NewIdentityPoint().Set(p.value)
}

func (*PointEd25519) SetEdwardsPoint(pt *edwards25519.Point) *PointEd25519 {
	return &PointEd25519{value: edwards25519.NewIdentityPoint().Set(pt)}
}

func mapToEdwards(e *field.Element) *edwards25519.Point {
	u, v := elligator2Montgomery(e)
	x, y := montgomeryToEdwards(u, v)
	return affineToEdwards(x, y)
}

func elligator2Montgomery(e *field.Element) (x, y *field.Element) {
	t1 := new(field.Element).Square(e) // u^2
	t1.Multiply(t1, two)               // t1 = 2u^2
	e1 := t1.Equal(minOne)             //
	t1.Swap(edZero, e1)                // if 2u^2 == -1, t1 = 0

	x1 := new(field.Element).Add(t1, edOne) // t1 + 1
	x1.Invert(x1)                           // 1 / (t1 + 1)
	x1.Multiply(x1, minA)                   // x1 = -A / (t1 + 1)

	gx1 := new(field.Element).Add(x1, a) // x1 + A
	gx1.Multiply(gx1, x1)                // x1 * (x1 + A)
	gx1.Add(gx1, edOne)                  // x1 * (x1 + A) + 1
	gx1.Multiply(gx1, x1)                // x1 * (x1 * (x1 + A) + 1)

	x2 := new(field.Element).Negate(x1) // -x1
	x2.Subtract(x2, a)                  // -x2 - A

	gx2 := new(field.Element).Multiply(t1, gx1) // t1 * gx1

	root1, _isSquare := new(field.Element).SqrtRatio(gx1, edOne) // root1 = (+) sqrt(gx1)
	negRoot1 := new(field.Element).Negate(root1)                 // negRoot1 = (-) sqrt(gx1)
	root2, _ := new(field.Element).SqrtRatio(gx2, edOne)         // root2 = (+) sqrt(gx2)

	x = new(field.Element)
	y = new(field.Element)

	x.Set(x2)
	y.Set(root2)

	x.Swap(x1, _isSquare)
	y.Swap(negRoot1, _isSquare)

	// if gx1 is square, set the point to (x1, -root1)
	// if not, set the point to (x2, +root2)
	//if _isSquare == 1 {
	//	x = x1
	//	y = negRoot1 // set sgn0(y) == 1, i.e. negative
	//} else {
	//	x = x2
	//	y = root2 // set sgn0(y) == 0, i.e. positive
	//}

	return x, y
}

func montgomeryToEdwards(u, v *field.Element) (x, y *field.Element) {
	x = new(field.Element).Invert(v)
	x.Multiply(x, u)
	x.Multiply(x, invsqrtD)

	u1 := new(field.Element).Subtract(u, edOne)
	u2 := new(field.Element).Add(u, edOne)
	y = u1.Multiply(u1, u2.Invert(u2))

	return
}

func affineToEdwards(x, y *field.Element) *edwards25519.Point {
	t := new(field.Element).Multiply(x, y)

	p, err := new(edwards25519.Point).SetExtendedCoordinates(x, y, new(field.Element).One(), t)
	if err != nil {
		panic(err)
	}
	return p
}

func (s *ScalarRistretto25519) Random(reader io.Reader) Scalar {
	if reader == nil {
		return nil
	}
	var seed [64]byte
	_, _ = reader.Read(seed[:])
	return s.Hash(seed[:])
}

func (*ScalarRistretto25519) Hash(b []byte) Scalar {
	return &ScalarRistretto25519{value: new(ristretto.Scalar).Derive(b)}
}

func (*ScalarRistretto25519) Zero() Scalar {
	return &ScalarRistretto25519{value: new(ristretto.Scalar).SetZero()}
}

func (*ScalarRistretto25519) One() Scalar {
	return &ScalarRistretto25519{value: new(ristretto.Scalar).SetOne()}
}

func (s *ScalarRistretto25519) IsZero() bool {
	i := byte(0)
	for _, b := range s.value.Bytes() {
		i |= b
	}
	return i == 0
}

func (s *ScalarRistretto25519) IsOne() bool {
	data := s.value.Bytes()
	i := byte(0)
	for j := 1; j < len(data); j++ {
		i |= data[j]
	}
	return i == 0 && data[0] == 1
}

func (s *ScalarRistretto25519) IsOdd() bool {
	return s.value.Bytes()[0]&1 == 1
}

func (s *ScalarRistretto25519) IsEven() bool {
	return s.value.Bytes()[0]&1 == 0
}

func (s *ScalarRistretto25519) New(input int) Scalar {
	var data [64]byte
	i := input
	if input < 0 {
		i = -input
	}
	data[0] = byte(i)
	data[1] = byte(i >> 8)
	data[2] = byte(i >> 16)
	data[3] = byte(i >> 24)
	value := new(ristretto.Scalar).SetReduced(&data)
	if input < 0 {
		value.Neg(value)
	}
	return &ScalarRistretto25519{value}
}

func (s *ScalarRistretto25519) Cmp(rhs Scalar) int {
	r := s.Sub(rhs)
	if r != nil && r.IsZero() {
		return 0
	} else {
		return -2
	}
}

func (s *ScalarRistretto25519) Square() Scalar {
	value := new(ristretto.Scalar).Square(s.value)
	return &ScalarRistretto25519{value}
}

func (s *ScalarRistretto25519) Pow(exp uint64) Scalar {
	out := s.Clone()

	for j := 63; j >= 0; j-- {
		square := out.Square()
		squareMul := square.Mul(square)
		out = cSelect(out, square, squareMul, (exp>>j)&1)
	}

	return out
}

func (s *ScalarRistretto25519) Double() Scalar {
	return &ScalarRistretto25519{
		value: new(ristretto.Scalar).Add(s.value, s.value),
	}
}

func (s *ScalarRistretto25519) Invert() (Scalar, error) {
	return &ScalarRistretto25519{
		value: new(ristretto.Scalar).Inverse(s.value),
	}, nil
}

func (s *ScalarRistretto25519) Sqrt() (Scalar, error) {
	bi25519, _ := new(big.Int).SetString("1000000000000000000000000000000014DEF9DEA2F79CD65812631A5CF5D3ED", 16)
	x := s.BigInt()
	x.ModSqrt(x, bi25519)
	return s.SetBigInt(x)
}

func (s *ScalarRistretto25519) Cube() Scalar {
	value := new(ristretto.Scalar).Square(s.value)
	value.Mul(value, s.value)
	return &ScalarRistretto25519{value}
}

func (s *ScalarRistretto25519) Add(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarRistretto25519)
	if ok {
		return &ScalarRistretto25519{
			value: new(ristretto.Scalar).Add(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarRistretto25519) Sub(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarRistretto25519)
	if ok {
		return &ScalarRistretto25519{
			value: new(ristretto.Scalar).Sub(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarRistretto25519) Mul(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarRistretto25519)
	if ok {
		return &ScalarRistretto25519{
			value: new(ristretto.Scalar).Mul(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarRistretto25519) MulAdd(y, z Scalar) Scalar {
	yy, ok := y.(*ScalarRistretto25519)
	if !ok {
		return nil
	}
	zz, ok := z.(*ScalarRistretto25519)
	if !ok {
		return nil
	}
	return &ScalarRistretto25519{value: new(ristretto.Scalar).MulAdd(s.value, yy.value, zz.value)}
}

func (s *ScalarRistretto25519) Div(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarRistretto25519)
	if ok {
		value := new(ristretto.Scalar).Inverse(r.value)
		value.Mul(value, s.value)
		return &ScalarRistretto25519{value}
	} else {
		return nil
	}
}

func (s *ScalarRistretto25519) Neg() Scalar {
	return &ScalarRistretto25519{
		value: new(ristretto.Scalar).Neg(s.value),
	}
}

func (*ScalarRistretto25519) SetBigInt(x *big.Int) (Scalar, error) {
	if x == nil {
		return nil, fmt.Errorf("invalid value")
	}
	value := new(ristretto.Scalar).SetBigInt(x)
	return &ScalarRistretto25519{value}, nil
}

func (s *ScalarRistretto25519) BigInt() *big.Int {
	return s.value.BigInt()
}

func (s *ScalarRistretto25519) Bytes() []byte {
	return s.value.Bytes()
}

func (*ScalarRistretto25519) SetBytes(input []byte) (Scalar, error) {
	if len(input) != 32 {
		return nil, fmt.Errorf("invalid byte sequence")
	}
	var t [32]byte
	copy(t[:], input)
	value := new(ristretto.Scalar).SetBytes(&t)
	return &ScalarRistretto25519{value}, nil
}

func (*ScalarRistretto25519) SetBytesWide(input []byte) (Scalar, error) {
	if len(input) != 64 {
		return nil, fmt.Errorf("invalid bytes sequence")
	}
	var t [64]byte
	copy(t[:], input)
	value := new(ristretto.Scalar).SetReduced(&t)
	return &ScalarRistretto25519{value}, nil
}

func (*ScalarRistretto25519) Point() Point {
	return new(PointRistretto25519).Identity()
}

func (s *ScalarRistretto25519) Clone() Scalar {
	return &ScalarRistretto25519{value: new(ristretto.Scalar).Set(s.value)}
}

func (s *ScalarRistretto25519) MarshalBinary() ([]byte, error) {
	return ScalarMarshalBinary(s)
}

func (s *ScalarRistretto25519) UnmarshalBinary(input []byte) error {
	sc, err := ScalarUnmarshalBinary(input)
	if err != nil {
		return err
	}
	ss, ok := sc.(*ScalarRistretto25519)
	if !ok {
		return fmt.Errorf("invalid scalar")
	}
	s.value = ss.value
	return nil
}

func (s *ScalarRistretto25519) MarshalText() ([]byte, error) {
	return ScalarMarshalText(s)
}

func (s *ScalarRistretto25519) UnmarshalText(input []byte) error {
	sc, err := ScalarUnmarshalText(input)
	if err != nil {
		return err
	}
	ss, ok := sc.(*ScalarRistretto25519)
	if !ok {
		return fmt.Errorf("invalid scalar")
	}
	s.value = ss.value
	return nil
}

func (s *ScalarRistretto25519) MarshalJSON() ([]byte, error) {
	return ScalarMarshalJSON(s)
}

func (s *ScalarRistretto25519) UnmarshalJSON(input []byte) error {
	sc, err := ScalarUnmarshalJSON(input)
	if err != nil {
		return err
	}
	S, ok := sc.(*ScalarRistretto25519)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	s.value = S.value
	return nil
}

func (p *PointRistretto25519) Random(reader io.Reader) Point {
	var seed [64]byte
	_, _ = reader.Read(seed[:])
	return p.Hash(seed[:])
}

func (*PointRistretto25519) Hash(b []byte) Point {
	value := new(ristretto.Point).DeriveDalek(b)
	return &PointRistretto25519{value}
}

func (*PointRistretto25519) Identity() Point {
	return &PointRistretto25519{value: new(ristretto.Point).SetZero()}
}

func (*PointRistretto25519) Generator() Point {
	return &PointRistretto25519{value: new(ristretto.Point).SetBase()}
}

func (p *PointRistretto25519) IsIdentity() bool {
	return p.Equal(p.Identity())
}

func (*PointRistretto25519) IsNegative() bool {
	return false
}

func (*PointRistretto25519) IsOnCurve() bool {
	return true
}

func (p *PointRistretto25519) Double() Point {
	return &PointRistretto25519{value: new(ristretto.Point).Double(p.value)}
}

func (*PointRistretto25519) Scalar() Scalar {
	return new(ScalarRistretto25519).Zero()
}

func (p *PointRistretto25519) Neg() Point {
	return &PointRistretto25519{value: new(ristretto.Point).Neg(p.value)}
}

func (p *PointRistretto25519) Add(rhs Point) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*PointRistretto25519)
	if ok {
		return &PointRistretto25519{value: new(ristretto.Point).Add(p.value, r.value)}
	} else {
		return nil
	}
}

func (p *PointRistretto25519) Sub(rhs Point) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*PointRistretto25519)
	if ok {
		return &PointRistretto25519{value: new(ristretto.Point).Sub(p.value, r.value)}
	} else {
		return nil
	}
}

func (p *PointRistretto25519) Mul(rhs Scalar) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*ScalarRistretto25519)
	if ok {
		return &PointRistretto25519{value: new(ristretto.Point).ScalarMult(p.value, r.value)}
	} else {
		return nil
	}
}

func (p *PointRistretto25519) Equal(rhs Point) bool {
	r, ok := rhs.(*PointRistretto25519)
	if ok {
		return p.value.Equals(r.value)
	} else {
		return false
	}
}

func (p *PointRistretto25519) Set(x, y *big.Int) (Point, error) {
	xx := subtle.ConstantTimeCompare(x.Bytes(), []byte{})
	yy := subtle.ConstantTimeCompare(y.Bytes(), []byte{})
	if (xx | yy) == 1 {
		return p.Identity(), nil
	}

	value := new(ristretto.Point).SetZero()
	value.X.SetBigInt(x)
	value.Y.SetBigInt(y)
	value.Z.SetOne()
	value.T.Mul(&value.X, &value.Y)

	return &PointRistretto25519{value}, nil
}

func (p *PointRistretto25519) ToAffineCompressed() []byte {
	return p.value.Bytes()
}

func (p *PointRistretto25519) ToAffineUncompressed() []byte {
	temp := new(ristretto.Point).SetZero()

	recip := temp.Z.Inverse(&p.value.Z)
	x := temp.X.Mul(&p.value.X, recip)
	y := temp.Y.Mul(&p.value.Y, recip)
	xBytes := x.Bytes()
	yBytes := y.Bytes()
	var out [64]byte
	copy(out[:32], xBytes[:])
	copy(out[32:], yBytes[:])
	return out[:]
}

func (*PointRistretto25519) FromAffineCompressed(b []byte) (Point, error) {
	if len(b) != 32 {
		return nil, fmt.Errorf("invalid length")
	}
	var inBytes [32]byte
	copy(inBytes[:], b)
	value := new(ristretto.Point)
	if value.SetBytes(&inBytes) {
		return &PointRistretto25519{value}, nil
	} else {
		return nil, fmt.Errorf("invalid bytes")
	}
}

func (*PointRistretto25519) FromAffineUncompressed(b []byte) (Point, error) {
	if len(b) != 64 {
		return nil, fmt.Errorf("invalid byte sequence")
	}
	if bytes.Equal(b, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		return &PointRistretto25519{value: new(ristretto.Point).SetZero()}, nil
	}
	var input [32]byte
	value := new(ristretto.Point).SetZero()
	copy(input[:], b[:32])
	value.X.SetBytes(&input)
	copy(input[:], b[32:])
	value.T.SetBytes(&input)
	value.Z.SetOne()
	value.T.Mul(&value.X, &value.Y)
	return &PointRistretto25519{value}, nil
}

func (*PointRistretto25519) CurveName() string {
	return Ristretto25519Name
}

func (p *PointRistretto25519) SumOfProducts(points []Point, scalars []Scalar) Point {
	nScalars := make([]*edwards25519.Scalar, len(scalars))
	nPoints := make([]*edwards25519.Point, len(points))
	for i, sc := range scalars {
		s, err := edwards25519.NewScalar().SetCanonicalBytes(sc.Bytes())
		if err != nil {
			return nil
		}
		nScalars[i] = s
	}
	var inBytes [32]byte
	for i, pt := range points {
		pp, ok := pt.(*PointRistretto25519)
		if !ok {
			return nil
		}
		pp.value.X.BytesInto(&inBytes)
		x, _ := new(field.Element).SetBytes(inBytes[:])
		pp.value.Y.BytesInto(&inBytes)
		y, _ := new(field.Element).SetBytes(inBytes[:])
		pp.value.Z.BytesInto(&inBytes)
		z, _ := new(field.Element).SetBytes(inBytes[:])
		pp.value.T.BytesInto(&inBytes)
		t, _ := new(field.Element).SetBytes(inBytes[:])
		nPoints[i], _ = edwards25519.NewIdentityPoint().SetExtendedCoordinates(x, y, z, t)
	}
	pt := edwards25519.NewIdentityPoint().MultiScalarMult(nScalars, nPoints)
	value := new(ristretto.Point).SetZero()

	x, y, z, t := pt.ExtendedCoordinates()
	copy(inBytes[:], x.Bytes())
	value.X.SetBytes(&inBytes)
	copy(inBytes[:], y.Bytes())
	value.Y.SetBytes(&inBytes)
	copy(inBytes[:], z.Bytes())
	value.Z.SetBytes(&inBytes)
	copy(inBytes[:], t.Bytes())
	value.T.SetBytes(&inBytes)
	return &PointRistretto25519{value}
}

func (p *PointRistretto25519) MarshalBinary() ([]byte, error) {
	return PointMarshalBinary(p)
}

func (p *PointRistretto25519) UnmarshalBinary(input []byte) error {
	pt, err := PointUnmarshalBinary(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointRistretto25519)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.value = ppt.value
	return nil
}

func (p *PointRistretto25519) MarshalText() ([]byte, error) {
	return PointMarshalText(p)
}

func (p *PointRistretto25519) UnmarshalText(input []byte) error {
	pt, err := PointUnmarshalText(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointRistretto25519)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.value = ppt.value
	return nil
}

func (p *PointRistretto25519) MarshalJSON() ([]byte, error) {
	return PointMarshalJSON(p)
}

func (p *PointRistretto25519) UnmarshalJSON(input []byte) error {
	pt, err := PointUnmarshalJSON(input)
	if err != nil {
		return err
	}
	P, ok := pt.(*PointRistretto25519)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	p.value = P.value
	return nil
}
