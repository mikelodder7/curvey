package curvey

import (
	"crypto/elliptic"
	"fmt"
	"github.com/mikelodder7/curvey/internal"
	"github.com/mikelodder7/curvey/native/ed448"
	"github.com/mikelodder7/curvey/native/ed448/fp"
	"github.com/mikelodder7/curvey/native/ed448/fq"
	"io"
	"math/big"
	"sync"
)

var (
	oldEd448Initonce sync.Once
	oldEd448Params   TwistedEdwards448
)

type TwistedEdwards448 struct {
	*elliptic.CurveParams
}

func oldEd448InitAll() {
	pt := ed448.EdwardsPointNew().SetGenerator().ToAffine()
	oldEd448Params.CurveParams = new(elliptic.CurveParams)
	oldEd448Params.P = new(big.Int).Set(fp.BiModulus)
	oldEd448Params.N = new(big.Int).Set(fq.FqNew().Value.Params.BiModulus)
	oldEd448Params.Gx = pt.X.BigInt()
	oldEd448Params.Gy = pt.Y.BigInt()
	oldEd448Params.B = fp.EdwardsD.BigInt()
	oldEd448Params.BitSize = 448
	oldEd448Params.Name = ED448Name
}

func Edwards448Curve() *TwistedEdwards448 {
	oldEd448Initonce.Do(oldEd448InitAll)
	return &oldEd448Params
}

func (curve *TwistedEdwards448) Params() *elliptic.CurveParams {
	return curve.CurveParams
}

func (*TwistedEdwards448) IsOnCurve(x, y *big.Int) bool {
	_, err := ed448.EdwardsPointNew().SetBigInt(x, y)
	return err != nil
}

func (*TwistedEdwards448) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	pt1, err := ed448.EdwardsPointNew().SetBigInt(x1, y1)
	if err != nil {
		return nil, nil
	}
	pt2, err := ed448.EdwardsPointNew().SetBigInt(x2, y2)
	if err != nil {
		return nil, nil
	}
	return pt1.Add(pt1, pt2).BigInt()
}

func (*TwistedEdwards448) Double(x, y *big.Int) (*big.Int, *big.Int) {
	pt, err := ed448.EdwardsPointNew().SetBigInt(x, y)
	if err != nil {
		return nil, nil
	}
	return pt.Double(pt).BigInt()
}

func (*TwistedEdwards448) ScalarMult(x, y *big.Int, k []byte) (*big.Int, *big.Int) {
	pt, err := ed448.EdwardsPointNew().SetBigInt(x, y)
	if err != nil {
		return nil, nil
	}
	kk := ([57]byte)(k)
	ss, err := fq.FqNew().SetBytes(&kk)
	if err != nil {
		return nil, nil
	}
	return pt.Mul(pt, ss).BigInt()
}

func (*TwistedEdwards448) ScalarBaseMult(k []byte) (*big.Int, *big.Int) {
	kk := ([57]byte)(k)
	ss, err := fq.FqNew().SetBytes(&kk)
	if err != nil {
		return nil, nil
	}
	pt := ed448.EdwardsPointNew().SetGenerator()
	return pt.Mul(pt, ss).BigInt()
}

type ScalarEd448 struct {
	value *fq.Fq
}

type PointEd448 struct {
	value *ed448.EdwardsPoint
}

func (s *ScalarEd448) Random(reader io.Reader) Scalar {
	if reader == nil {
		return nil
	}
	var seed [64]byte
	_, _ = reader.Read(seed[:])
	return s.Hash(seed[:])
}

func (s *ScalarEd448) Hash(bytes []byte) Scalar {
	return &ScalarEd448{
		value: fq.FqNew().Hash(bytes),
	}
}

func (s *ScalarEd448) Zero() Scalar {
	return &ScalarEd448{
		value: fq.FqNew().SetZero(),
	}
}

func (s *ScalarEd448) One() Scalar {
	return &ScalarEd448{
		value: fq.FqNew().SetOne(),
	}
}

func (s *ScalarEd448) IsZero() bool {
	return s.value.IsZero() == 1
}

func (s *ScalarEd448) IsOne() bool {
	return s.value.IsOne() == 1
}

func (s *ScalarEd448) IsOdd() bool {
	return s.value.Sgn0I() == 1
}

func (s *ScalarEd448) IsEven() bool {
	return s.value.Sgn0I() == 0
}

func (s *ScalarEd448) New(value int) Scalar {
	t := fq.FqNew()
	v := big.NewInt(int64(value))
	if value < 0 {
		v.Mod(v, t.Value.Params.BiModulus)
	}
	return &ScalarEd448{
		value: t.SetBigInt(v),
	}
}

func (s *ScalarEd448) Cmp(rhs Scalar) int {
	r, ok := rhs.(*ScalarEd448)
	if ok {
		return s.value.Cmp(r.value)
	} else {
		return -2
	}
}

func (s *ScalarEd448) Square() Scalar {
	return &ScalarEd448{
		value: fq.FqNew().Square(s.value),
	}
}

func (s *ScalarEd448) Pow(exp uint64) Scalar {
	v := fq.FqNew().SetZero()
	internal.Pow(&v.Value.Value, &s.value.Value.Value, &[]uint64{exp, 0, 0, 0, 0, 0, 0}, s.value.Value.Params, s.value.Value.Arithmetic)
	return &ScalarEd448{value: v}
}

func (s *ScalarEd448) Double() Scalar {
	return &ScalarEd448{
		value: fq.FqNew().Double(s.value),
	}
}

func (s *ScalarEd448) Invert() (Scalar, error) {
	value, wasInverted := fq.FqNew().Invert(s.value)
	if wasInverted == 0 {
		return nil, fmt.Errorf("inverse doesn't exist")
	}
	return &ScalarEd448{
		value,
	}, nil
}

func (s *ScalarEd448) Sqrt() (Scalar, error) {
	value, wasSquare := fq.FqNew().Sqrt(s.value)
	if wasSquare == 0 {
		return nil, fmt.Errorf("not a square")
	}
	return &ScalarEd448{
		value,
	}, nil
}

func (s *ScalarEd448) Cube() Scalar {
	value := fq.FqNew().Square(s.value)
	value.Mul(value, s.value)
	return &ScalarEd448{
		value,
	}
}

func (s *ScalarEd448) Add(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarEd448)
	if ok {
		return &ScalarEd448{
			value: fq.FqNew().Add(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarEd448) Sub(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarEd448)
	if ok {
		return &ScalarEd448{
			value: fq.FqNew().Sub(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarEd448) Mul(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarEd448)
	if ok {
		return &ScalarEd448{
			value: fq.FqNew().Mul(s.value, r.value),
		}
	} else {
		return nil
	}
}

func (s *ScalarEd448) MulAdd(y, z Scalar) Scalar {
	return s.Mul(y).Add(z)
}

func (s *ScalarEd448) Div(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarEd448)
	if ok {
		v, wasInverted := fq.FqNew().Invert(r.value)
		if wasInverted == 0 {
			return nil
		}
		v.Mul(v, s.value)
		return &ScalarEd448{value: v}
	} else {
		return nil
	}
}

func (s *ScalarEd448) Neg() Scalar {
	return &ScalarEd448{
		value: fq.FqNew().Neg(s.value),
	}
}

func (*ScalarEd448) SetBigInt(v *big.Int) (Scalar, error) {
	if v == nil {
		return nil, fmt.Errorf("'v' cannot be nil")
	}
	value := fq.FqNew().SetBigInt(v)
	return &ScalarEd448{
		value,
	}, nil
}

func (s *ScalarEd448) BigInt() *big.Int {
	return s.value.BigInt()
}

func (s *ScalarEd448) Bytes() []byte {
	t := s.value.Bytes()
	return internal.ReverseBytes(t[:])
}

func (*ScalarEd448) SetBytes(bytes []byte) (Scalar, error) {
	if len(bytes) != 57 {
		return nil, fmt.Errorf("invalid length")
	}
	var seq [57]byte
	copy(seq[:], internal.ReverseBytes(bytes))
	value, err := fq.FqNew().SetBytes(&seq)
	if err != nil {
		return nil, err
	}
	return &ScalarEd448{
		value,
	}, nil
}

func (*ScalarEd448) SetBytesWide(bytes []byte) (Scalar, error) {
	if len(bytes) != 114 {
		return nil, fmt.Errorf("invalid length")
	}
	var seq [114]byte
	copy(seq[:], internal.ReverseBytes(bytes))
	return &ScalarEd448{
		value: fq.FqNew().SetBytesWide(&seq),
	}, nil
}

func (*ScalarEd448) Point() Point {
	return new(PointEd448).Identity()
}

func (s *ScalarEd448) Clone() Scalar {
	return &ScalarEd448{
		value: fq.FqNew().Set(s.value),
	}
}

func (s *ScalarEd448) MarshalBinary() ([]byte, error) {
	return ScalarMarshalBinary(s)
}

func (s *ScalarEd448) UnmarshalBinary(input []byte) error {
	sc, err := ScalarUnmarshalBinary(input)
	if err != nil {
		return err
	}
	ss, ok := sc.(*ScalarEd448)
	if !ok {
		return fmt.Errorf("invalid scalar")
	}
	s.value = ss.value
	return nil
}

func (s *ScalarEd448) MarshalText() ([]byte, error) {
	return ScalarMarshalText(s)
}

func (s *ScalarEd448) UnmarshalText(input []byte) error {
	sc, err := ScalarUnmarshalText(input)
	if err != nil {
		return err
	}
	ss, ok := sc.(*ScalarEd448)
	if !ok {
		return fmt.Errorf("invalid scalar")
	}
	s.value = ss.value
	return nil
}

func (s *ScalarEd448) MarshalJSON() ([]byte, error) {
	return ScalarMarshalJSON(s)
}

func (s *ScalarEd448) UnmarshalJSON(input []byte) error {
	sc, err := ScalarUnmarshalJSON(input)
	if err != nil {
		return err
	}
	S, ok := sc.(*ScalarEd448)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	s.value = S.value
	return nil
}

func (p *PointEd448) Random(reader io.Reader) Point {
	var seed [64]byte
	_, _ = reader.Read(seed[:])
	return p.Hash(seed[:])
}

func (*PointEd448) Hash(bytes []byte) Point {
	value := ed448.EdwardsPointNew().HashWithDefaults(bytes)
	return &PointEd448{value}
}

func (*PointEd448) Identity() Point {
	return &PointEd448{
		value: ed448.EdwardsPointNew().SetIdentity(),
	}
}

func (*PointEd448) Generator() Point {
	return &PointEd448{
		value: ed448.EdwardsPointNew().SetGenerator(),
	}
}

func (p *PointEd448) IsIdentity() bool {
	return p.value.IsIdentityI() == 1
}

func (p *PointEd448) IsNegative() bool {
	return false
}

func (p *PointEd448) IsOnCurve() bool {
	return p.value.IsOnCurve() == 1
}

func (p *PointEd448) Double() Point {
	value := ed448.EdwardsPointNew().Double(p.value)
	return &PointEd448{value}
}

func (*PointEd448) Scalar() Scalar {
	return new(ScalarEd448).Zero()
}

func (p *PointEd448) Neg() Point {
	value := ed448.EdwardsPointNew().Negate(p.value)
	return &PointEd448{value}
}

func (p *PointEd448) Add(rhs Point) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*PointEd448)
	if ok {
		value := ed448.EdwardsPointNew().Add(p.value, r.value)
		return &PointEd448{value}
	} else {
		return nil
	}
}

func (p *PointEd448) Sub(rhs Point) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*PointEd448)
	if ok {
		value := ed448.EdwardsPointNew().Sub(p.value, r.value)
		return &PointEd448{value}
	} else {
		return nil
	}
}

func (p *PointEd448) Mul(rhs Scalar) Point {
	if rhs == nil {
		return nil
	}
	r, ok := rhs.(*ScalarEd448)
	if ok {
		value := ed448.EdwardsPointNew().Mul(p.value, r.value)
		return &PointEd448{value}
	} else {
		return nil
	}
}

func (p *PointEd448) Equal(rhs Point) bool {
	r, ok := rhs.(*PointEd448)
	if ok {
		return p.value.EqualI(r.value) == 1
	} else {
		return false
	}
}

func (*PointEd448) Set(x, y *big.Int) (Point, error) {
	pt := ed448.AffinePointNew()
	pt.X.SetBigInt(x)
	pt.Y.SetBigInt(y)

	value := pt.ToEdwards()
	if value.IsOnCurve()&value.IsTorsionFree() == 1 {
		return &PointEd448{value}, nil
	} else {
		return nil, fmt.Errorf("invalid coordinates")
	}
}

func (p *PointEd448) X() *big.Int {
	return p.value.ToAffine().X.BigInt()
}

func (p *PointEd448) Y() *big.Int {
	return p.value.ToAffine().Y.BigInt()
}

func (p *PointEd448) ToAffineCompressed() []byte {
	value := p.value.Compress()
	return value[:]
}

func (p *PointEd448) ToAffineUncompressed() []byte {
	var out [112]byte
	affine := p.value.ToAffine()
	x := affine.X.Bytes()
	y := affine.Y.Bytes()
	copy(out[:ed448.PointBytes-1], x[:])
	copy(out[ed448.PointBytes-1:], y[:])
	return out[:]
}

func (p *PointEd448) FromAffineCompressed(bytes []byte) (Point, error) {
	if len(bytes) != ed448.PointBytes {
		return nil, fmt.Errorf("invalid length")
	}
	pt := (*ed448.CompressedEdwardsY)(bytes)
	value, err := pt.Decompress()
	if err != nil {
		return nil, err
	}
	return &PointEd448{value}, nil
}

func (*PointEd448) FromAffineUncompressed(bytes []byte) (Point, error) {
	if len(bytes) != 112 {
		return nil, fmt.Errorf("invalid length")
	}
	x := ([ed448.PointBytes - 1]byte)(bytes[:ed448.PointBytes-1])
	pt := ed448.AffinePointNew()
	_, err := pt.X.SetCanonicalBytes(&x)
	if err != nil {
		return nil, err
	}
	y := ([ed448.PointBytes - 1]byte)(bytes[ed448.PointBytes-1:])
	_, err = pt.Y.SetCanonicalBytes(&y)
	if err != nil {
		return nil, err
	}
	return &PointEd448{value: pt.ToEdwards()}, nil
}

func (p *PointEd448) CurveName() string {
	return ED448Name
}

func (*PointEd448) SumOfProducts(points []Point, scalars []Scalar) Point {
	nPoints := make([]*ed448.EdwardsPoint, len(points))
	nScalars := make([]*fq.Fq, len(scalars))
	for i, pt := range points {
		ptv, ok := pt.(*PointEd448)
		if !ok {
			return nil
		}
		nPoints[i] = ptv.value
	}
	for i, sc := range scalars {
		s, ok := sc.(*ScalarEd448)
		if !ok {
			return nil
		}
		nScalars[i] = s.value
	}
	value, err := ed448.EdwardsPointNew().SumOfProducts(nPoints, nScalars)
	if err != nil {
		return nil
	}
	return &PointEd448{value}
}

func (p *PointEd448) MarshalBinary() ([]byte, error) {
	return PointMarshalBinary(p)
}

func (p *PointEd448) UnmarshalBinary(input []byte) error {
	pt, err := PointUnmarshalBinary(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointEd448)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.value = ppt.value
	return nil
}

func (p *PointEd448) MarshalText() ([]byte, error) {
	return PointMarshalText(p)
}

func (p *PointEd448) UnmarshalText(input []byte) error {
	pt, err := PointUnmarshalText(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointEd448)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.value = ppt.value
	return nil
}

func (p *PointEd448) MarshalJSON() ([]byte, error) {
	return PointMarshalJSON(p)
}

func (p *PointEd448) UnmarshalJSON(input []byte) error {
	pt, err := PointUnmarshalJSON(input)
	if err != nil {
		return err
	}
	P, ok := pt.(*PointEd448)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	p.value = P.value
	return nil
}
