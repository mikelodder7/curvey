//
// SPDX-License-Identifier: Apache-2.0
//

package curvey

import (
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"sync"

	"github.com/mikelodder7/curvey/native/bls12381"
)

var (
	k256Initonce sync.Once
	k256         Curve

	bls12381g1Initonce sync.Once
	bls12381g1         Curve

	bls12381g2Initonce sync.Once
	bls12381g2         Curve

	p256Initonce sync.Once
	p256         Curve

	ed25519Initonce sync.Once
	ed25519         Curve

	pallasInitonce sync.Once
	pallas         Curve
)

const (
	K256Name       = "secp256k1"
	BLS12381G1Name = "BLS12381G1"
	BLS12381G2Name = "BLS12381G2"
	BLS12831Name   = "BLS12831"
	P256Name       = "P-256"
	ED25519Name    = "ed25519"
	PallasName     = "pallas"
)

const scalarBytes = 32

// Scalar represents an element of the scalar field \mathbb{F}_q
// of the elliptic curve construction.
type Scalar interface {
	// Random returns a random scalar using the provided reader
	// to retrieve bytes
	Random(reader io.Reader) Scalar
	// Hash the specific bytes in a manner to yield a
	// uniformly distributed scalar
	Hash(bytes []byte) Scalar
	// Zero returns the additive identity element
	Zero() Scalar
	// One returns the multiplicative identity element
	One() Scalar
	// IsZero returns true if this element is the additive identity element
	IsZero() bool
	// IsOne returns true if this element is the multiplicative identity element
	IsOne() bool
	// IsOdd returns true if this element is odd
	IsOdd() bool
	// IsEven returns true if this element is even
	IsEven() bool
	// New returns an element with the value equal to `value`
	New(value int) Scalar
	// Cmp returns
	// -2 if this element is in a different field than rhs
	// -1 if this element is less than rhs
	// 0 if this element is equal to rhs
	// 1 if this element is greater than rhs
	Cmp(rhs Scalar) int
	// Square returns element*element
	Square() Scalar
	// Double returns element+element
	Double() Scalar
	// Invert returns element^-1 mod p
	Invert() (Scalar, error)
	// Sqrt computes the square root of this element if it exists.
	Sqrt() (Scalar, error)
	// Cube returns element*element*element
	Cube() Scalar
	// Pow returns the scalar exponentiated to the power of i
	Pow(exp uint64) Scalar
	// Add returns element+rhs
	Add(rhs Scalar) Scalar
	// Sub returns element-rhs
	Sub(rhs Scalar) Scalar
	// Mul returns element*rhs
	Mul(rhs Scalar) Scalar
	// MulAdd returns element * y + z mod p
	MulAdd(y, z Scalar) Scalar
	// Div returns element*rhs^-1 mod p
	Div(rhs Scalar) Scalar
	// Neg returns -element mod p
	Neg() Scalar
	// SetBigInt returns this element set to the value of v
	SetBigInt(v *big.Int) (Scalar, error)
	// BigInt returns this element as a big integer
	BigInt() *big.Int
	// Point returns the associated point for this scalar
	Point() Point
	// Bytes returns the canonical byte representation of this scalar
	Bytes() []byte
	// SetBytes creates a scalar from the canonical representation expecting the exact number of bytes needed to represent the scalar
	SetBytes(bytes []byte) (Scalar, error)
	// SetBytesWide creates a scalar expecting double the exact number of bytes needed to represent the scalar which is reduced by the modulus
	SetBytesWide(bytes []byte) (Scalar, error)
	// Clone returns a cloned Scalar of this value
	Clone() Scalar
}

type PairingScalar interface {
	Scalar
	SetPoint(p Point) PairingScalar
}

func unmarshalScalar(input []byte) (*Curve, []byte, error) {
	sep := byte(':')
	i := 0
	for ; i < len(input); i++ {
		if input[i] == sep {
			break
		}
	}
	name := string(input[:i])
	curve := GetCurveByName(name)
	if curve == nil {
		return nil, nil, fmt.Errorf("unrecognized curve")
	}
	return curve, input[i+1:], nil
}

func ScalarMarshalBinary(scalar Scalar) ([]byte, error) {
	// All scalars are 32 bytes long
	// The last 32 bytes are the actual value
	// The first remaining bytes are the curve name
	// separated by a colon
	name := []byte(scalar.Point().CurveName())
	output := make([]byte, len(name)+1+scalarBytes)
	copy(output[:len(name)], name)
	output[len(name)] = byte(':')
	copy(output[len(name)+1:], scalar.Bytes())
	return output, nil
}

func ScalarUnmarshalBinary(input []byte) (Scalar, error) {
	// All scalars are 32 bytes long
	// The first 32 bytes are the actual value
	// The remaining bytes are the curve name
	if len(input) < scalarBytes+1+len(P256Name) {
		return nil, fmt.Errorf("invalid byte sequence")
	}
	sc, data, err := unmarshalScalar(input)
	if err != nil {
		return nil, err
	}
	return sc.Scalar.SetBytes(data)
}

func ScalarMarshalText(scalar Scalar) ([]byte, error) {
	// All scalars are 32 bytes long
	// For text encoding we put the curve name first for readability
	// separated by a colon, then the hex encoding of the scalar
	// which avoids the base64 weakness with strict mode or not
	name := []byte(scalar.Point().CurveName())
	output := make([]byte, len(name)+1+scalarBytes*2)
	copy(output[:len(name)], name)
	output[len(name)] = byte(':')
	_ = hex.Encode(output[len(name)+1:], scalar.Bytes())
	return output, nil
}

func ScalarUnmarshalText(input []byte) (Scalar, error) {
	if len(input) < scalarBytes*2+len(P256Name)+1 {
		return nil, fmt.Errorf("invalid byte sequence")
	}
	curve, data, err := unmarshalScalar(input)
	if err != nil {
		return nil, err
	}
	var t [scalarBytes]byte
	_, err = hex.Decode(t[:], data)
	if err != nil {
		return nil, err
	}
	return curve.Scalar.SetBytes(t[:])
}

func ScalarMarshalJSON(scalar Scalar) ([]byte, error) {
	m := make(map[string]string, 2)
	m["type"] = scalar.Point().CurveName()
	m["value"] = hex.EncodeToString(scalar.Bytes())
	return json.Marshal(m)
}

func ScalarUnmarshalJSON(input []byte) (Scalar, error) {
	var m map[string]string

	err := json.Unmarshal(input, &m)
	if err != nil {
		return nil, err
	}
	curve := GetCurveByName(m["type"])
	if curve == nil {
		return nil, fmt.Errorf("invalid type")
	}
	s, err := hex.DecodeString(m["value"])
	if err != nil {
		return nil, err
	}
	S, err := curve.Scalar.SetBytes(s)
	if err != nil {
		return nil, err
	}
	return S, nil
}

// Point represents an elliptic curve point.
type Point interface {
	Random(reader io.Reader) Point
	Hash(bytes []byte) Point
	Identity() Point
	Generator() Point
	IsIdentity() bool
	IsNegative() bool
	IsOnCurve() bool
	Double() Point
	Scalar() Scalar
	Neg() Point
	Add(rhs Point) Point
	Sub(rhs Point) Point
	Mul(rhs Scalar) Point
	Equal(rhs Point) bool
	Set(x, y *big.Int) (Point, error)
	ToAffineCompressed() []byte
	ToAffineUncompressed() []byte
	FromAffineCompressed(bytes []byte) (Point, error)
	FromAffineUncompressed(bytes []byte) (Point, error)
	CurveName() string
	SumOfProducts(points []Point, scalars []Scalar) Point
}

type PairingPoint interface {
	Point
	OtherGroup() PairingPoint
	Pairing(rhs PairingPoint) Scalar
	MultiPairing(...PairingPoint) Scalar
}

func PointMarshalBinary(point Point) ([]byte, error) {
	// Always stores points in compressed form
	// The first bytes are the curve name
	// separated by a colon followed by the compressed point
	// bytes
	t := point.ToAffineCompressed()
	name := []byte(point.CurveName())
	output := make([]byte, len(name)+1+len(t))
	copy(output[:len(name)], name)
	output[len(name)] = byte(':')
	copy(output[len(output)-len(t):], t)
	return output, nil
}

func PointUnmarshalBinary(input []byte) (Point, error) {
	if len(input) < scalarBytes+1+len(P256Name) {
		return nil, fmt.Errorf("invalid byte sequence")
	}
	sep := byte(':')
	i := 0
	for ; i < len(input); i++ {
		if input[i] == sep {
			break
		}
	}
	name := string(input[:i])
	curve := GetCurveByName(name)
	if curve == nil {
		return nil, fmt.Errorf("unrecognized curve")
	}
	return curve.Point.FromAffineCompressed(input[i+1:])
}

func PointMarshalText(point Point) ([]byte, error) {
	// Always stores points in compressed form
	// The first bytes are the curve name
	// separated by a colon followed by the compressed point
	// bytes
	t := point.ToAffineCompressed()
	name := []byte(point.CurveName())
	output := make([]byte, len(name)+1+len(t)*2)
	copy(output[:len(name)], name)
	output[len(name)] = byte(':')
	hex.Encode(output[len(output)-len(t)*2:], t)
	return output, nil
}

func PointUnmarshalText(input []byte) (Point, error) {
	if len(input) < scalarBytes*2+1+len(P256Name) {
		return nil, fmt.Errorf("invalid byte sequence")
	}
	sep := byte(':')
	i := 0
	for ; i < len(input); i++ {
		if input[i] == sep {
			break
		}
	}
	name := string(input[:i])
	curve := GetCurveByName(name)
	if curve == nil {
		return nil, fmt.Errorf("unrecognized curve")
	}
	buffer := make([]byte, (len(input)-i)/2)
	_, err := hex.Decode(buffer, input[i+1:])
	if err != nil {
		return nil, err
	}
	return curve.Point.FromAffineCompressed(buffer)
}

func PointMarshalJSON(point Point) ([]byte, error) {
	m := make(map[string]string, 2)
	m["type"] = point.CurveName()
	m["value"] = hex.EncodeToString(point.ToAffineCompressed())
	return json.Marshal(m)
}

func PointUnmarshalJSON(input []byte) (Point, error) {
	var m map[string]string

	err := json.Unmarshal(input, &m)
	if err != nil {
		return nil, err
	}
	curve := GetCurveByName(m["type"])
	if curve == nil {
		return nil, fmt.Errorf("invalid type")
	}
	p, err := hex.DecodeString(m["value"])
	if err != nil {
		return nil, err
	}
	P, err := curve.Point.FromAffineCompressed(p)
	if err != nil {
		return nil, err
	}
	return P, nil
}

// Curve represents a named elliptic curve with a scalar field and point group.
type Curve struct {
	Scalar Scalar
	Point  Point
	Name   string
}

func (c *Curve) ScalarBaseMult(sc Scalar) Point {
	return c.Point.Generator().Mul(sc)
}

func (c *Curve) NewGeneratorPoint() Point {
	return c.Point.Generator()
}

func (c *Curve) NewIdentityPoint() Point {
	return c.Point.Identity()
}

func (c *Curve) NewScalar() Scalar {
	return c.Scalar.Zero()
}

// ToEllipticCurve returns the equivalent of this curve as the go interface `elliptic.Curve`.
func (c *Curve) ToEllipticCurve() (elliptic.Curve, error) {
	err := fmt.Errorf("can't convert %s", c.Name)
	switch c.Name {
	case K256Name:
		return K256Curve(), nil
	case BLS12381G1Name:
		return nil, err
	case BLS12381G2Name:
		return nil, err
	case BLS12831Name:
		return nil, err
	case P256Name:
		return NistP256Curve(), nil
	case ED25519Name:
		return nil, err
	case PallasName:
		return nil, err
	default:
		return nil, err
	}
}

// PairingCurve represents a named elliptic curve
// that supports pairings.
type PairingCurve struct {
	Scalar  PairingScalar
	PointG1 PairingPoint
	PointG2 PairingPoint
	GT      Scalar
	Name    string
}

func (c *PairingCurve) ScalarG1BaseMult(sc Scalar) PairingPoint {
	pairingPoint, ok := c.PointG1.Generator().Mul(sc).(PairingPoint)
	if !ok {
		return nil
	}
	return pairingPoint
}

func (c *PairingCurve) ScalarG2BaseMult(sc Scalar) PairingPoint {
	pairingPoint, ok := c.PointG2.Generator().Mul(sc).(PairingPoint)
	if !ok {
		return nil
	}
	return pairingPoint
}

func (c *PairingCurve) NewG1GeneratorPoint() PairingPoint {
	pairingPoint, ok := c.PointG1.Generator().(PairingPoint)
	if !ok {
		return nil
	}
	return pairingPoint
}

func (c *PairingCurve) NewG2GeneratorPoint() PairingPoint {
	pairingPoint, ok := c.PointG2.Generator().(PairingPoint)
	if !ok {
		return nil
	}
	return pairingPoint
}

func (c *PairingCurve) NewG1IdentityPoint() PairingPoint {
	pairingPoint, ok := c.PointG1.Identity().(PairingPoint)
	if !ok {
		return nil
	}
	return pairingPoint
}

func (c *PairingCurve) NewG2IdentityPoint() PairingPoint {
	pairingPoint, ok := c.PointG2.Identity().(PairingPoint)
	if !ok {
		return nil
	}
	return pairingPoint
}

func (c *PairingCurve) NewScalar() PairingScalar {
	pairingPoint, ok := c.Scalar.Zero().(PairingScalar)
	if !ok {
		return nil
	}
	return pairingPoint
}

// GetCurveByName returns the correct `Curve` given the name.
func GetCurveByName(name string) *Curve {
	switch name {
	case K256Name:
		return K256()
	case BLS12381G1Name:
		return BLS12381G1()
	case BLS12381G2Name:
		return BLS12381G2()
	case BLS12831Name:
		return BLS12381G1()
	case P256Name:
		return P256()
	case ED25519Name:
		return ED25519()
	case PallasName:
		return PALLAS()
	default:
		return nil
	}
}

func GetPairingCurveByName(name string) *PairingCurve {
	switch name {
	case BLS12381G1Name:
		return BLS12381(BLS12381G1().NewIdentityPoint())
	case BLS12381G2Name:
		return BLS12381(BLS12381G2().NewIdentityPoint())
	case BLS12831Name:
		return BLS12381(BLS12381G1().NewIdentityPoint())
	default:
		return nil
	}
}

// BLS12381G1 returns the BLS12-381 curve with points in G1.
func BLS12381G1() *Curve {
	bls12381g1Initonce.Do(bls12381g1Init)
	return &bls12381g1
}

func bls12381g1Init() {
	bls12381g1 = Curve{
		Scalar: &ScalarBls12381{
			Value: bls12381.FqNew(),
			point: new(PointBls12381G1),
		},
		Point: new(PointBls12381G1).Identity(),
		Name:  BLS12381G1Name,
	}
}

// BLS12381G2 returns the BLS12-381 curve with points in G2.
func BLS12381G2() *Curve {
	bls12381g2Initonce.Do(bls12381g2Init)
	return &bls12381g2
}

func bls12381g2Init() {
	bls12381g2 = Curve{
		Scalar: &ScalarBls12381{
			Value: bls12381.FqNew(),
			point: new(PointBls12381G2),
		},
		Point: new(PointBls12381G2).Identity(),
		Name:  BLS12381G2Name,
	}
}

func BLS12381(preferredPoint Point) *PairingCurve {
	return &PairingCurve{
		Scalar: &ScalarBls12381{
			Value: bls12381.FqNew(),
			point: preferredPoint,
		},
		PointG1: &PointBls12381G1{
			Value: new(bls12381.G1).Identity(),
		},
		PointG2: &PointBls12381G2{
			Value: new(bls12381.G2).Identity(),
		},
		GT: &ScalarBls12381Gt{
			Value: new(bls12381.Gt).SetOne(),
		},
		Name: BLS12831Name,
	}
}

// K256 returns the secp256k1 curve.
func K256() *Curve {
	k256Initonce.Do(k256Init)
	return &k256
}

func k256Init() {
	k256 = Curve{
		Scalar: new(ScalarK256).Zero(),
		Point:  new(PointK256).Identity(),
		Name:   K256Name,
	}
}

func P256() *Curve {
	p256Initonce.Do(p256Init)
	return &p256
}

func p256Init() {
	p256 = Curve{
		Scalar: new(ScalarP256).Zero(),
		Point:  new(PointP256).Identity(),
		Name:   P256Name,
	}
}

func ED25519() *Curve {
	ed25519Initonce.Do(ed25519Init)
	return &ed25519
}

func ed25519Init() {
	ed25519 = Curve{
		Scalar: new(ScalarEd25519).Zero(),
		Point:  new(PointEd25519).Identity(),
		Name:   ED25519Name,
	}
}

func PALLAS() *Curve {
	pallasInitonce.Do(pallasInit)
	return &pallas
}

func pallasInit() {
	pallas = Curve{
		Scalar: new(ScalarPallas).Zero(),
		Point:  new(PointPallas).Identity(),
		Name:   PallasName,
	}
}

func bhex(s string) *big.Int {
	r, _ := new(big.Int).SetString(s, 16)
	return r
}
