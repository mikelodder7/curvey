//
// SPDX-License-Identifier: Apache-2.0
//

package curvey

import (
	"crypto/elliptic"
	"fmt"
	"github.com/mikelodder7/curvey/internal"
	"io"
	"math/big"
	"sync"

	"github.com/mikelodder7/curvey/native"
	p384n "github.com/mikelodder7/curvey/native/p384"
	"github.com/mikelodder7/curvey/native/p384/fp"
	"github.com/mikelodder7/curvey/native/p384/fq"
)

var (
	oldP384InitOnce sync.Once
	oldP384         NistP384
)

type NistP384 struct {
	*elliptic.CurveParams
}

func oldP384InitAll() {
	curve := elliptic.P384()
	oldP384.CurveParams = curve.Params()
	oldP384.P = curve.Params().P
	oldP384.N = curve.Params().N
	oldP384.Gx = curve.Params().Gx
	oldP384.Gy = curve.Params().Gy
	oldP384.B = curve.Params().B
	oldP384.BitSize = curve.Params().BitSize
	oldP384.Name = curve.Params().Name
}

func NistP384Curve() *NistP384 {
	oldP384InitOnce.Do(oldP384InitAll)
	return &oldP384
}

func (curve *NistP384) Params() *elliptic.CurveParams {
	return curve.CurveParams
}

func (*NistP384) IsOnCurve(x, y *big.Int) bool {
	_, err := p384n.PointNew().SetBigInt(x, y)
	return err == nil
}

func (*NistP384) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	p1, err := p384n.PointNew().SetBigInt(x1, y1)
	if err != nil {
		return nil, nil
	}
	p2, err := p384n.PointNew().SetBigInt(x2, y2)
	if err != nil {
		return nil, nil
	}
	return p1.Add(p1, p2).BigInt()
}

func (*NistP384) Double(x1, y1 *big.Int) (*big.Int, *big.Int) {
	p1, err := p384n.PointNew().SetBigInt(x1, y1)
	if err != nil {
		return nil, nil
	}
	return p1.Double(p1).BigInt()
}

func (*NistP384) ScalarMul(capBx, capBy *big.Int, k []byte) (*big.Int, *big.Int) {
	p1, err := p384n.PointNew().SetBigInt(capBx, capBy)
	if err != nil {
		return nil, nil
	}
	var bytes [48]byte
	copy(bytes[:], internal.ReverseScalarBytes(k))
	s, err := fq.P384FqNew().SetBytes(&bytes)
	if err != nil {
		return nil, nil
	}
	return p1.Mul(p1, s).BigInt()
}

func (*NistP384) ScalarBaseMult(k []byte) (*big.Int, *big.Int) {
	var bytes [48]byte
	copy(bytes[:], internal.ReverseScalarBytes(k))
	s, err := fq.P384FqNew().SetBytes(&bytes)
	if err != nil {
		return nil, nil
	}
	p1 := p384n.PointNew().Generator()
	return p1.Mul(p1, s).BigInt()
}

type ScalarP384 struct {
	value *native.Field6
}

type PointP384 struct {
	value *native.EllipticPoint6
}

func (s *ScalarP384) Random(reader io.Reader) Scalar {
	if reader == nil {
		return nil
	}
	var seed [96]byte
	_, _ = reader.Read(seed[:])
	return s.Hash(seed[:])
}

func (*ScalarP384) Hash(bytes []byte) Scalar {
	dst := []byte("P384_XMD:SHA-384_SSWU_RO_")
	xmd := native.ExpandMsgXmd(native.EllipticPointHasherSha384(), bytes, dst, 72)
	var t [96]byte
	copy(t[:72], internal.ReverseScalarBytes(xmd))

	return &ScalarP384{
		value: fq.P384FqNew().SetBytesWide(&t),
	}
}

func (*ScalarP384) Zero() Scalar {
	return &ScalarP384{
		value: fq.P384FqNew().SetZero(),
	}
}

func (*ScalarP384) One() Scalar {
	return &ScalarP384{
		value: fq.P384FqNew().SetOne(),
	}
}

func (s *ScalarP384) IsZero() bool {
	return s.value.IsZero() == 1
}

func (s *ScalarP384) IsOne() bool {
	return s.value.IsOne() == 1
}

func (s *ScalarP384) IsOdd() bool {
	return s.value.Bytes()[0]&1 == 1
}

func (s *ScalarP384) IsEven() bool {
	return s.value.Bytes()[0]&1 == 0
}

func (*ScalarP384) New(value int) Scalar {
	t := fq.P384FqNew()
	v := big.NewInt(int64(value))
	if value < 0 {
		v.Mod(v, t.Params.BiModulus)
	}
	return &ScalarP384{
		value: t.SetBigInt(v),
	}
}

func (s *ScalarP384) Cmp(rhs Scalar) int {
	r, ok := rhs.(*ScalarP384)
	if ok {
		return s.value.Cmp(r.value)
	} else {
		return -2
	}
}

func (s *ScalarP384) Square() Scalar {
	return &ScalarP384{
		value: fq.P384FqNew().Square(s.value),
	}
}

func (s *ScalarP384) Pow(exp uint64) Scalar {
	expFieldLimb := [native.Field6Limbs]uint64{exp, 0, 0, 0, 0, 0}
	out := ScalarP384{value: fq.P384FqNew()}
	native.Pow6(&out.value.Value, &s.value.Value, &expFieldLimb, s.value.Params, s.value.Arithmetic)
	return &ScalarP384{
		value: out.value,
	}
}

func (s *ScalarP384) Double() Scalar {
	return &ScalarP384{
		value: fq.P384FqNew().Double(s.value),
	}
}

func (s *ScalarP384) Invert() (Scalar, error) {
	value, wasInverted := fq.P384FqNew().Invert(s.value)
	if !wasInverted {
		return nil, fmt.Errorf("inverse doesn't exist")
	}
	return &ScalarP384{
		value,
	}, nil
}

func (s *ScalarP384) Sqrt() (Scalar, error) {
	value, wasSquare := fq.P384FqNew().Sqrt(s.value)
	if !wasSquare {
		return nil, fmt.Errorf("not a square")
	}
	return &ScalarP384{
		value,
	}, nil
}

func (s *ScalarP384) Cube() Scalar {
	value := fq.P384FqNew().Square(s.value)
	value.Mul(value, s.value)
	return &ScalarP384{
		value,
	}
}

func (s *ScalarP384) Add(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarP384)
	if ok {
		return &ScalarP384{
			value: fq.P384FqNew().Add(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarP384) Sub(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarP384)
	if ok {
		return &ScalarP384{
			value: fq.P384FqNew().Sub(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarP384) Mul(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarP384)
	if ok {
		return &ScalarP384{
			value: fq.P384FqNew().Mul(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarP384) MulAdd(y, z Scalar) Scalar {
	return s.Mul(y).Add(z)
}

func (s *ScalarP384) Div(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarP384)
	if ok {
		v, wasInverted := fq.P384FqNew().Invert(r.value)
		if !wasInverted {
			return nil
		}
		v.Mul(v, s.value)
		return &ScalarP384{value: v}
	} else {
		return nil
	}
}

func (s *ScalarP384) Neg() Scalar {
	return &ScalarP384{
		value: fq.P384FqNew().Neg(s.value),
	}
}

func (*ScalarP384) SetBigInt(v *big.Int) (Scalar, error) {
	if v == nil {
		return nil, fmt.Errorf("'v' cannot be nil")
	}
	value := fq.P384FqNew().SetBigInt(v)
	return &ScalarP384{
		value,
	}, nil
}

func (s *ScalarP384) BigInt() *big.Int {
	return s.value.BigInt()
}

func (s *ScalarP384) Bytes() []byte {
	t := s.value.Bytes()
	return internal.ReverseScalarBytes(t[:])
}

func (*ScalarP384) SetBytes(bytes []byte) (Scalar, error) {
	if len(bytes) != 48 {
		return nil, fmt.Errorf("invalid length")
	}
	var seq [48]byte
	copy(seq[:], internal.ReverseScalarBytes(bytes))
	value, err := fq.P384FqNew().SetBytes(&seq)
	if err != nil {
		return nil, err
	}
	return &ScalarP384{
		value,
	}, nil
}

func (*ScalarP384) SetBytesWide(bytes []byte) (Scalar, error) {
	if len(bytes) != 96 {
		return nil, fmt.Errorf("invalid length")
	}
	var seq [96]byte
	copy(seq[:], bytes)
	return &ScalarP384{
		value: fq.P384FqNew().SetBytesWide(&seq),
	}, nil
}

func (*ScalarP384) Point() Point {
	return new(PointP384).Identity()
}

func (s *ScalarP384) Clone() Scalar {
	return &ScalarP384{
		value: fq.P384FqNew().Set(s.value),
	}
}

func (s *ScalarP384) MarshalBinary() ([]byte, error) {
	return ScalarMarshalBinary(s)
}

func (s *ScalarP384) UnmarshalBinary(input []byte) error {
	sc, err := ScalarUnmarshalBinary(input)
	if err != nil {
		return err
	}
	ss, ok := sc.(*ScalarP384)
	if !ok {
		return fmt.Errorf("invalid scalar")
	}
	s.value = ss.value
	return nil
}

func (s *ScalarP384) MarshalText() ([]byte, error) {
	return ScalarMarshalText(s)
}

func (s *ScalarP384) UnmarshalText(input []byte) error {
	sc, err := ScalarUnmarshalText(input)
	if err != nil {
		return err
	}
	ss, ok := sc.(*ScalarP384)
	if !ok {
		return fmt.Errorf("invalid scalar")
	}
	s.value = ss.value
	return nil
}

func (s *ScalarP384) MarshalJSON() ([]byte, error) {
	return ScalarMarshalJSON(s)
}

func (s *ScalarP384) UnmarshalJSON(input []byte) error {
	sc, err := ScalarUnmarshalJSON(input)
	if err != nil {
		return err
	}
	S, ok := sc.(*ScalarP384)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	s.value = S.value
	return nil
}

func (p *PointP384) Random(reader io.Reader) Point {
	var seed [64]byte
	_, _ = reader.Read(seed[:])
	return p.Hash(seed[:])
}

func (*PointP384) Hash(bytes []byte) Point {
	value, err := p384n.PointNew().Hash(bytes, native.EllipticPointHasherSha384())
	// TODO: change hash to return an error also
	if err != nil {
		return nil
	}

	return &PointP384{value}
}

func (*PointP384) Identity() Point {
	return &PointP384{
		value: p384n.PointNew().Identity(),
	}
}

func (*PointP384) Generator() Point {
	return &PointP384{
		value: p384n.PointNew().Generator(),
	}
}

func (p *PointP384) IsIdentity() bool {
	return p.value.IsIdentity()
}

func (p *PointP384) IsNegative() bool {
	return p.value.GetY().Value[0]&1 == 1
}

func (p *PointP384) IsOnCurve() bool {
	return p.value.IsOnCurve()
}

func (p *PointP384) Double() Point {
	value := p384n.PointNew().Double(p.value)
	return &PointP384{value}
}

func (*PointP384) Scalar() Scalar {
	return new(ScalarP384).Zero()
}

func (p *PointP384) Neg() Point {
	value := p384n.PointNew().Neg(p.value)
	return &PointP384{value}
}

func (p *PointP384) Add(rhs Point) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*PointP384)
	if ok {
		value := p384n.PointNew().Add(p.value, r.value)
		return &PointP384{value}
	} else {
		return nil
	}
}

func (p *PointP384) Sub(rhs Point) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*PointP384)
	if ok {
		value := p384n.PointNew().Sub(p.value, r.value)
		return &PointP384{value}
	} else {
		return nil
	}
}

func (p *PointP384) Mul(rhs Scalar) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*ScalarP384)
	if ok {
		value := p384n.PointNew().Mul(p.value, r.value)
		return &PointP384{value}
	} else {
		return nil
	}
}

func (p *PointP384) Equal(rhs Point) bool {
	r, ok := rhs.(*PointP384)
	if ok {
		return p.value.Equal(r.value) == 1
	} else {
		return false
	}
}

func (*PointP384) Set(x, y *big.Int) (Point, error) {
	value, err := p384n.PointNew().SetBigInt(x, y)
	if err != nil {
		return nil, err
	}
	return &PointP384{value}, nil
}

func (p *PointP384) ToAffineCompressed() []byte {
	var x [49]byte
	x[0] = byte(2)

	t := p384n.PointNew().ToAffine(p.value)

	x[0] |= t.Y.Bytes()[0] & 1

	xBytes := t.X.Bytes()
	copy(x[1:], internal.ReverseScalarBytes(xBytes[:]))
	return x[:]
}

func (p *PointP384) ToAffineUncompressed() []byte {
	var out [97]byte
	out[0] = byte(4)
	t := p384n.PointNew().ToAffine(p.value)
	arr := t.X.Bytes()
	copy(out[1:49], internal.ReverseScalarBytes(arr[:]))
	arr = t.Y.Bytes()
	copy(out[49:], internal.ReverseScalarBytes(arr[:]))
	return out[:]
}

func (p *PointP384) FromAffineCompressed(bytes []byte) (Point, error) {
	var raw [native.Field6Bytes]byte
	if len(bytes) != 49 {
		return nil, fmt.Errorf("invalid byte sequence")
	}
	sign := int(bytes[0])
	if sign != 2 && sign != 3 {
		return nil, fmt.Errorf("invalid sign byte")
	}
	sign &= 0x1

	copy(raw[:], internal.ReverseScalarBytes(bytes[1:]))
	x, err := fp.P384FpNew().SetBytes(&raw)
	if err != nil {
		return nil, err
	}

	value := p384n.PointNew().Identity()
	rhs := fp.P384FpNew()
	p.value.Arithmetic.RhsEquation(rhs, x)
	// test that rhs is quadratic residue
	// if not, then this Point is at infinity
	y, wasQr := fp.P384FpNew().Sqrt(rhs)
	if wasQr {
		// fix the sign
		sigY := int(y.Bytes()[0] & 1)
		if sigY != sign {
			y.Neg(y)
		}
		value.X = x
		value.Y = y
		value.Z.SetOne()
	}
	return &PointP384{value}, nil
}

func (*PointP384) FromAffineUncompressed(bytes []byte) (Point, error) {
	var arr [native.Field6Bytes]byte
	if len(bytes) != 97 {
		return nil, fmt.Errorf("invalid byte sequence")
	}
	if bytes[0] != 4 {
		return nil, fmt.Errorf("invalid sign byte")
	}

	copy(arr[:], internal.ReverseScalarBytes(bytes[1:49]))
	x, err := fp.P384FpNew().SetBytes(&arr)
	if err != nil {
		return nil, err
	}
	copy(arr[:], internal.ReverseScalarBytes(bytes[49:]))
	y, err := fp.P384FpNew().SetBytes(&arr)
	if err != nil {
		return nil, err
	}
	value := p384n.PointNew()
	value.X = x
	value.Y = y
	value.Z.SetOne()
	return &PointP384{value}, nil
}

func (*PointP384) CurveName() string {
	return elliptic.P384().Params().Name
}

func (*PointP384) SumOfProducts(points []Point, scalars []Scalar) Point {
	nPoints := make([]*native.EllipticPoint6, len(points))
	nScalars := make([]*native.Field6, len(scalars))
	for i, pt := range points {
		ptv, ok := pt.(*PointP384)
		if !ok {
			return nil
		}
		nPoints[i] = ptv.value
	}
	for i, sc := range scalars {
		s, ok := sc.(*ScalarP384)
		if !ok {
			return nil
		}
		nScalars[i] = s.value
	}
	value := p384n.PointNew()
	_, err := value.SumOfProducts(nPoints, nScalars)
	if err != nil {
		return nil
	}
	return &PointP384{value}
}

func (p *PointP384) X() *native.Field6 {
	return p.value.GetX()
}

func (p *PointP384) Y() *native.Field6 {
	return p.value.GetY()
}

func (*PointP384) Params() *elliptic.CurveParams {
	return elliptic.P384().Params()
}

func (p *PointP384) MarshalBinary() ([]byte, error) {
	return PointMarshalBinary(p)
}

func (p *PointP384) UnmarshalBinary(input []byte) error {
	pt, err := PointUnmarshalBinary(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointP384)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.value = ppt.value
	return nil
}

func (p *PointP384) MarshalText() ([]byte, error) {
	return PointMarshalText(p)
}

func (p *PointP384) UnmarshalText(input []byte) error {
	pt, err := PointUnmarshalText(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointP384)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.value = ppt.value
	return nil
}

func (p *PointP384) MarshalJSON() ([]byte, error) {
	return PointMarshalJSON(p)
}

func (p *PointP384) UnmarshalJSON(input []byte) error {
	pt, err := PointUnmarshalJSON(input)
	if err != nil {
		return err
	}
	P, ok := pt.(*PointP384)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	p.value = P.value
	return nil
}
