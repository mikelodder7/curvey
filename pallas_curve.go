//
// SPDX-License-Identifier: Apache-2.0
//

package curvey

import (
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync"

	"github.com/mikelodder7/curvey/native"
	"github.com/mikelodder7/curvey/native/pasta"
	"github.com/mikelodder7/curvey/native/pasta/fp"
	"github.com/mikelodder7/curvey/native/pasta/fq"
)

var (
	oldPallasInitonce sync.Once
	oldPallas         PallasCurve
)

type PallasCurve struct {
	*elliptic.CurveParams
}

func oldPallasInitAll() {
	oldPallas.CurveParams = new(elliptic.CurveParams)
	oldPallas.P = new(big.Int).SetBytes([]byte{
		0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x22, 0x46, 0x98, 0xfc, 0x09, 0x4c, 0xf9, 0x1b,
		0x99, 0x2d, 0x30, 0xed, 0x00, 0x00, 0x00, 0x01,
	})
	oldPallas.N = new(big.Int).SetBytes([]byte{
		0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x22, 0x46, 0x98, 0xfc, 0x09, 0x94, 0xa8, 0xdd,
		0x8c, 0x46, 0xeb, 0x21, 0x00, 0x00, 0x00, 0x01,
	})
	g := pasta.PointNew().Generator()
	oldPallas.Gx = g.X.BigInt()
	oldPallas.Gy = g.Y.BigInt()
	oldPallas.B = g.Params.B.BigInt()
	oldPallas.BitSize = 255
	pallas.Name = PallasName
}

func Pallas() *PallasCurve {
	oldPallasInitonce.Do(oldPallasInitAll)
	return &oldPallas
}

func (curve *PallasCurve) Params() *elliptic.CurveParams {
	return curve.CurveParams
}

func (*PallasCurve) IsOnCurve(x, y *big.Int) bool {
	_, err := pasta.PointNew().SetBigInt(x, y)
	return err == nil
}

func (*PallasCurve) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	pt1, err := pasta.PointNew().SetBigInt(x1, y1)
	if err != nil {
		return nil, nil
	}
	pt2, err := pasta.PointNew().SetBigInt(x2, y2)
	if err != nil {
		return nil, nil
	}
	return pt1.Add(pt1, pt2).BigInt()
}

func (*PallasCurve) Double(x1, y1 *big.Int) (*big.Int, *big.Int) {
	pt1, err := pasta.PointNew().SetBigInt(x1, y1)
	if err != nil {
		return nil, nil
	}
	return pt1.Double(pt1).BigInt()
}

func (*PallasCurve) ScalarMult(bx, by *big.Int, k []byte) (*big.Int, *big.Int) {
	pt1, err := pasta.PointNew().SetBigInt(bx, by)
	if err != nil {
		return nil, nil
	}
	var blob [native.Field4Bytes]byte
	copy(blob[:], k)
	sc, err := fq.PastaFqNew().SetBytes(&blob)
	if err != nil {
		return nil, nil
	}
	return pt1.Mul(pt1, sc).BigInt()
}

func (*PallasCurve) ScalarBaseMult(k []byte) (*big.Int, *big.Int) {
	var blob [native.Field4Bytes]byte
	copy(blob[:], k)
	sc, err := fq.PastaFqNew().SetBytes(&blob)
	if err != nil {
		return nil, nil
	}
	g := pasta.PointNew().Generator()
	return g.Mul(g, sc).BigInt()
}

// PallasScalar - Old interface.
type PallasScalar struct{}

func NewPallasScalar() *PallasScalar {
	return &PallasScalar{}
}

func (PallasScalar) Add(x, y *big.Int) *big.Int {
	r := new(big.Int).Add(x, y)
	q := fq.PastaFqNew().Params.BiModulus
	return r.Mod(r, q)
}

func (PallasScalar) Sub(x, y *big.Int) *big.Int {
	r := new(big.Int).Sub(x, y)
	q := fq.PastaFqNew().Params.BiModulus
	return r.Mod(r, q)
}

func (PallasScalar) Neg(x *big.Int) *big.Int {
	r := new(big.Int).Neg(x)
	q := fq.PastaFqNew().Params.BiModulus
	return r.Mod(r, q)
}

func (PallasScalar) Mul(x, y *big.Int) *big.Int {
	r := new(big.Int).Mul(x, y)
	q := fq.PastaFqNew().Params.BiModulus
	return r.Mod(r, q)
}

func (PallasScalar) Div(x, y *big.Int) *big.Int {
	q := fq.PastaFqNew().Params.BiModulus
	r := new(big.Int).ModInverse(y, q)
	r.Mul(r, x)
	return r.Mod(r, q)
}

func (PallasScalar) Hash(input []byte) *big.Int {
	hashed, ok := new(ScalarPallas).Hash(input).(*ScalarPallas)
	if !ok {
		return nil
	}
	return hashed.Value.BigInt()
}

func (PallasScalar) Bytes(x *big.Int) []byte {
	return x.Bytes()
}

func (PallasScalar) Random() (*big.Int, error) {
	s, ok := new(ScalarPallas).Random(crand.Reader).(*ScalarPallas)
	if !ok {
		return nil, errors.New("incorrect type conversion")
	}
	return s.Value.BigInt(), nil
}

func (PallasScalar) IsValid(x *big.Int) bool {
	q := fq.PastaFqNew().Params.BiModulus
	return x.Cmp(q) == -1
}

// ScalarPallas - New interface.
type ScalarPallas struct {
	Value *native.Field4
}

func (s *ScalarPallas) Random(reader io.Reader) Scalar {
	if reader == nil {
		return nil
	}
	var seed [64]byte
	_, _ = reader.Read(seed[:])
	return s.Hash(seed[:])
}

func (*ScalarPallas) Hash(bytes []byte) Scalar {
	dst := []byte("pallas_XMD:BLAKE2b_SSWU_RO_")
	xmd := native.ExpandMsgXmd(native.EllipticPointHasherBlake2b(), bytes, dst, 64)
	var t [64]byte
	copy(t[:], xmd)
	return &ScalarPallas{
		Value: fq.PastaFqNew().SetBytesWide(&t),
	}
}

func (*ScalarPallas) Zero() Scalar {
	return &ScalarPallas{
		Value: fq.PastaFqNew().SetZero(),
	}
}

func (*ScalarPallas) One() Scalar {
	return &ScalarPallas{
		Value: fq.PastaFqNew().SetOne(),
	}
}

func (s *ScalarPallas) IsZero() bool {
	return s.Value.IsZero() == 1
}

func (s *ScalarPallas) IsOne() bool {
	return s.Value.IsOne() == 1
}

func (s *ScalarPallas) IsOdd() bool {
	return (s.Value.Bytes()[0] & 1) == 1
}

func (s *ScalarPallas) IsEven() bool {
	return (s.Value.Bytes()[0] & 1) == 0
}

func (*ScalarPallas) New(value int) Scalar {
	v := big.NewInt(int64(value))
	return &ScalarPallas{
		Value: fq.PastaFqNew().SetBigInt(v),
	}
}

func (s *ScalarPallas) Cmp(rhs Scalar) int {
	r, ok := rhs.(*ScalarPallas)
	if ok {
		return s.Value.Cmp(r.Value)
	} else {
		return -2
	}
}

func (s *ScalarPallas) Square() Scalar {
	return &ScalarPallas{
		Value: fq.PastaFqNew().Square(s.Value),
	}
}

func (s *ScalarPallas) Pow(exp uint64) Scalar {
	expFieldLimb := [native.Field4Limbs]uint64{exp, 0, 0, 0}
	out := ScalarPallas{Value: fq.PastaFqNew()}
	native.Pow(&out.Value.Value, &s.Value.Value, &expFieldLimb, s.Value.Params, s.Value.Arithmetic)
	return &ScalarPallas{
		Value: out.Value,
	}
}

func (s *ScalarPallas) Double() Scalar {
	return &ScalarPallas{
		Value: fq.PastaFqNew().Double(s.Value),
	}
}

func (s *ScalarPallas) Invert() (Scalar, error) {
	value, wasInverted := fq.PastaFqNew().Invert(s.Value)
	if !wasInverted {
		return nil, fmt.Errorf("inverse doesn't exist")
	}
	return &ScalarPallas{
		value,
	}, nil
}

func (s *ScalarPallas) Sqrt() (Scalar, error) {
	value, wasSquare := fq.PastaFqNew().Sqrt(s.Value)
	if !wasSquare {
		return nil, fmt.Errorf("not a square")
	}
	return &ScalarPallas{
		value,
	}, nil
}

func (s *ScalarPallas) Cube() Scalar {
	value := fq.PastaFqNew().Square(s.Value)
	value.Mul(value, s.Value)
	return &ScalarPallas{
		value,
	}
}

func (s *ScalarPallas) Add(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarPallas)
	if ok {
		return &ScalarPallas{
			Value: fq.PastaFqNew().Add(s.Value, r.Value),
		}
	} else {
		return nil
	}
}

func (s *ScalarPallas) Sub(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarPallas)
	if ok {
		return &ScalarPallas{
			Value: fq.PastaFqNew().Sub(s.Value, r.Value),
		}
	} else {
		return nil
	}
}

func (s *ScalarPallas) Mul(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarPallas)
	if ok {
		return &ScalarPallas{
			Value: fq.PastaFqNew().Mul(s.Value, r.Value),
		}
	} else {
		return nil
	}
}

func (s *ScalarPallas) MulAdd(y, z Scalar) Scalar {
	return s.Mul(y).Add(z)
}

func (s *ScalarPallas) Div(rhs Scalar) Scalar {
	r, ok := rhs.(*ScalarPallas)
	if ok {
		v, wasInverted := fq.PastaFqNew().Invert(r.Value)
		if !wasInverted {
			return nil
		}
		v.Mul(v, s.Value)
		return &ScalarPallas{Value: v}
	} else {
		return nil
	}
}

func (s *ScalarPallas) Neg() Scalar {
	return &ScalarPallas{
		Value: fq.PastaFqNew().Neg(s.Value),
	}
}

func (*ScalarPallas) SetBigInt(v *big.Int) (Scalar, error) {
	return &ScalarPallas{
		Value: fq.PastaFqNew().SetBigInt(v),
	}, nil
}

func (s *ScalarPallas) BigInt() *big.Int {
	return s.Value.BigInt()
}

func (s *ScalarPallas) Bytes() []byte {
	t := s.Value.Bytes()
	return t[:]
}

func (*ScalarPallas) SetBytes(bytes []byte) (Scalar, error) {
	if len(bytes) != 32 {
		return nil, fmt.Errorf("invalid length")
	}
	var seq [32]byte
	copy(seq[:], bytes)
	value, err := fq.PastaFqNew().SetBytes(&seq)
	if err != nil {
		return nil, err
	}
	return &ScalarPallas{
		value,
	}, nil
}

func (*ScalarPallas) SetBytesWide(bytes []byte) (Scalar, error) {
	if len(bytes) != 64 {
		return nil, fmt.Errorf("invalid length")
	}
	var seq [64]byte
	copy(seq[:], bytes)
	return &ScalarPallas{
		Value: fq.PastaFqNew().SetBytesWide(&seq),
	}, nil
}

func (*ScalarPallas) Point() Point {
	return new(PointPallas).Identity()
}

func (s *ScalarPallas) Clone() Scalar {
	return &ScalarPallas{
		Value: fq.PastaFqNew().Set(s.Value),
	}
}

func (s *ScalarPallas) MarshalBinary() ([]byte, error) {
	return ScalarMarshalBinary(s)
}

func (s *ScalarPallas) UnmarshalBinary(input []byte) error {
	sc, err := ScalarUnmarshalBinary(input)
	if err != nil {
		return err
	}
	ss, ok := sc.(*ScalarPallas)
	if !ok {
		return fmt.Errorf("invalid scalar")
	}
	s.Value = ss.Value
	return nil
}

func (s *ScalarPallas) MarshalText() ([]byte, error) {
	return ScalarMarshalText(s)
}

func (s *ScalarPallas) UnmarshalText(input []byte) error {
	sc, err := ScalarUnmarshalText(input)
	if err != nil {
		return err
	}
	ss, ok := sc.(*ScalarPallas)
	if !ok {
		return fmt.Errorf("invalid scalar")
	}
	s.Value = ss.Value
	return nil
}

func (s *ScalarPallas) MarshalJSON() ([]byte, error) {
	return ScalarMarshalJSON(s)
}

func (s *ScalarPallas) UnmarshalJSON(input []byte) error {
	sc, err := ScalarUnmarshalJSON(input)
	if err != nil {
		return err
	}
	S, ok := sc.(*ScalarPallas)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	s.Value = S.Value
	return nil
}

type PointPallas struct {
	*native.EllipticPoint4
}

func (p *PointPallas) Random(reader io.Reader) Point {
	var seed [2 * native.Field4Bytes]byte
	n, err := reader.Read(seed[:])
	if err != nil {
		return nil
	}
	if n != 2*native.Field4Bytes {
		return nil
	}
	return p.Hash(seed[:])
}

func (*PointPallas) Hash(bytes []byte) Point {
	value, err := pasta.PointNew().Hash(bytes, native.EllipticPointHasherBlake2b())
	if err != nil {
		return nil
	}
	return &PointPallas{value}
}

func (*PointPallas) Identity() Point {
	return &PointPallas{pasta.PointNew().Identity()}
}

func (*PointPallas) Generator() Point {
	return &PointPallas{pasta.PointNew().Generator()}
}

func (p *PointPallas) IsNegative() bool {
	return p.GetY().Bytes()[0]&1 == 1
}

func (p *PointPallas) Double() Point {
	return &PointPallas{pasta.PointNew().Double(p.EllipticPoint4)}
}

func (*PointPallas) Scalar() Scalar {
	return &ScalarPallas{fq.PastaFqNew().SetZero()}
}

func (p *PointPallas) Neg() Point {
	return &PointPallas{pasta.PointNew().Neg(p.EllipticPoint4)}
}

func (p *PointPallas) Add(rhs Point) Point {
	r, ok := rhs.(*PointPallas)
	if !ok {
		return nil
	}
	return &PointPallas{pasta.PointNew().Add(p.EllipticPoint4, r.EllipticPoint4)}
}

func (p *PointPallas) Sub(rhs Point) Point {
	r, ok := rhs.(*PointPallas)
	if !ok {
		return nil
	}
	return &PointPallas{pasta.PointNew().Sub(p.EllipticPoint4, r.EllipticPoint4)}
}

func (p *PointPallas) Mul(rhs Scalar) Point {
	s, ok := rhs.(*ScalarPallas)
	if !ok {
		return nil
	}
	return &PointPallas{pasta.PointNew().Mul(p.EllipticPoint4, s.Value)}
}

func (p *PointPallas) Equal(rhs Point) bool {
	r, ok := rhs.(*PointPallas)
	if !ok {
		return false
	}
	var x1, x2, y1, y2, z1, z2 [native.Field4Limbs]uint64

	u := p.EllipticPoint4.X.Arithmetic

	u.Square(&z1, &p.Z.Value)
	u.Square(&z2, &r.Z.Value)

	u.Mul(&x1, &p.EllipticPoint4.X.Value, &z2)
	u.Mul(&x2, &r.EllipticPoint4.X.Value, &z1)

	u.Mul(&z1, &z1, &p.Z.Value)
	u.Mul(&z2, &z2, &r.Z.Value)

	u.Mul(&y1, &p.EllipticPoint4.Y.Value, &z2)
	u.Mul(&y2, &r.EllipticPoint4.Y.Value, &z1)

	e1 := p.Z.IsZero()
	e2 := r.Z.IsZero()

	tx := (x1[0] ^ x2[0]) | (x1[1] ^ x2[1]) | (x1[2] ^ x2[2]) | (x1[3] ^ x2[3])
	ty := (y1[0] ^ y2[0]) | (y1[1] ^ y2[1]) | (y1[2] ^ y2[2]) | (y1[3] ^ y2[3])

	e3 := int(((int64(tx) | int64(-tx)) >> 63) + 1)
	e4 := int(((int64(ty) | int64(-ty)) >> 63) + 1)

	// Both at infinity or coordinates are the same
	return (e1&e2)|(^e1 & ^e2)&e3&e4 == 1
}

func (*PointPallas) Set(x, y *big.Int) (Point, error) {
	value, err := pasta.PointNew().SetBigInt(x, y)
	if err != nil {
		return nil, err
	}
	return &PointPallas{value}, nil
}

func (p *PointPallas) ToAffineCompressed() []byte {
	// Use ZCash encoding where infinity is all zeros
	// and the top bit represents the sign of y and the
	// remainder represent the x-coordinate
	var inf [32]byte
	p1 := pasta.PointNew().ToAffine(p.EllipticPoint4)
	x := p1.X.Bytes()
	x[31] |= (p1.Y.Bytes()[0] & 1) << 7

	subtle.ConstantTimeCopy(p1.Z.IsZero(), x[:], inf[:])
	return x[:]
}

func (p *PointPallas) ToAffineUncompressed() []byte {
	p1 := pasta.PointNew().ToAffine(p.EllipticPoint4)
	x := p1.X.Bytes()
	y := p1.Y.Bytes()
	return append(x[:], y[:]...)
}

func (p *PointPallas) FromAffineCompressed(bytes []byte) (Point, error) {
	if len(bytes) != 32 {
		return nil, fmt.Errorf("invalid byte sequence")
	}

	var input [32]byte
	copy(input[:], bytes)
	sign := (input[31] >> 7) & 1 // nolint:ifshort // false positive
	input[31] &= 0x7F

	x := fp.PastaFpNew()
	if _, err := x.SetBytes(&input); err != nil {
		return nil, err
	}
	rhs := fp.PastaFpNew()
	p.Arithmetic.RhsEquation(rhs, x)
	if _, square := rhs.Sqrt(rhs); !square {
		return nil, fmt.Errorf("rhs of given x-coordinate is not a square")
	}
	if rhs.Bytes()[0]&1 != sign {
		rhs.Neg(rhs)
	}
	value := pasta.PointNew()
	value.X = x
	value.Y = rhs
	value.Z.SetOne()
	if !value.IsOnCurve() {
		return nil, fmt.Errorf("invalid point")
	}
	return &PointPallas{value}, nil
}

func (*PointPallas) FromAffineUncompressed(bytes []byte) (Point, error) {
	if len(bytes) != 64 {
		return nil, fmt.Errorf("invalid length")
	}
	value := pasta.PointNew()
	value.Z.SetOne()
	var x, y [32]byte
	copy(x[:], bytes[:32])
	copy(y[:], bytes[32:])
	if _, err := value.X.SetBytes(&x); err != nil {
		return nil, err
	}
	if _, err := value.Y.SetBytes(&y); err != nil {
		return nil, err
	}
	if !value.IsOnCurve() {
		return nil, fmt.Errorf("invalid point")
	}
	return &PointPallas{value}, nil
}

func (*PointPallas) CurveName() string {
	return PallasName
}

func (p *PointPallas) SumOfProducts(points []Point, scalars []Scalar) Point {
	eps := make([]*native.EllipticPoint4, len(points))
	for i, pt := range points {
		ps, ok := pt.(*PointPallas)
		if !ok {
			return nil
		}
		eps[i] = ps.EllipticPoint4
	}
	scs := make([]*native.Field4, len(scalars))
	for i, sc := range scalars {
		ss, ok := sc.(*ScalarPallas)
		if !ok {
			return nil
		}
		scs[i] = ss.Value
	}
	value, err := p.EllipticPoint4.SumOfProducts(eps, scs)
	if err != nil {
		return nil
	}
	return &PointPallas{value}
}

func (p *PointPallas) MarshalBinary() ([]byte, error) {
	return PointMarshalBinary(p)
}

func (p *PointPallas) UnmarshalBinary(input []byte) error {
	pt, err := PointUnmarshalBinary(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointPallas)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.EllipticPoint4 = ppt.EllipticPoint4
	return nil
}

func (p *PointPallas) MarshalText() ([]byte, error) {
	return PointMarshalText(p)
}

func (p *PointPallas) UnmarshalText(input []byte) error {
	pt, err := PointUnmarshalText(input)
	if err != nil {
		return err
	}
	ppt, ok := pt.(*PointPallas)
	if !ok {
		return fmt.Errorf("invalid point")
	}
	p.EllipticPoint4 = ppt.EllipticPoint4
	return nil
}

func (p *PointPallas) MarshalJSON() ([]byte, error) {
	return PointMarshalJSON(p)
}

func (p *PointPallas) UnmarshalJSON(input []byte) error {
	pt, err := PointUnmarshalJSON(input)
	if err != nil {
		return err
	}
	P, ok := pt.(*PointPallas)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	p.EllipticPoint4 = P.EllipticPoint4
	return nil
}

func (p *PointPallas) X() *native.Field4 {
	return p.GetX()
}

func (p *PointPallas) Y() *native.Field4 {
	return p.GetY()
}
