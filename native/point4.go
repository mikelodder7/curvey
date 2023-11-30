package native

import (
	"fmt"
	"io"
	"math/big"

	"github.com/pkg/errors"
)

// EllipticPoint4 represents a Weierstrauss elliptic curve point.
type EllipticPoint4 struct {
	X          *Field4
	Y          *Field4
	Z          *Field4
	Params     *EllipticPoint4Params
	Arithmetic EllipticPoint4Arithmetic
}

// EllipticPoint4Params are the Weierstrauss curve parameters
// such as the name, the coefficients the generator point,
// and the prime bit size.
type EllipticPoint4Params struct {
	Name    string
	A       *Field4
	B       *Field4
	Gx      *Field4
	Gy      *Field4
	BitSize int
}

// EllipticPoint4Arithmetic are the methods that specific curves
// need to implement for higher abstractions to wrap the point.
type EllipticPoint4Arithmetic interface {
	// Hash a byte sequence to the curve using the specified hasher
	// and dst and store the result in out
	Hash(out *EllipticPoint4, hasher *EllipticPointHasher, bytes, dst []byte) error
	// Double arg and store the result in out
	Double(out, arg *EllipticPoint4)
	// Add arg1 with arg2 and store the result in out
	Add(out, arg1, arg2 *EllipticPoint4)
	// IsOnCurve tests arg if it represents a valid point on the curve
	IsOnCurve(arg *EllipticPoint4) bool
	// ToAffine converts arg to affine coordinates storing the result in out
	ToAffine(out, arg *EllipticPoint4)
	// RhsEquation computes the right-hand side of the ecc equation
	RhsEquation(out, x *Field4)
}

// Random creates a random point on the curve
// from the specified reader.
func (p *EllipticPoint4) Random(reader io.Reader) (*EllipticPoint4, error) {
	var seed [WideField4Bytes]byte
	n, err := reader.Read(seed[:])
	if err != nil {
		return nil, errors.Wrap(err, "random could not read from stream")
	}
	if n != WideField4Bytes {
		return nil, fmt.Errorf("insufficient bytes read %d when %d are needed", n, WideField4Bytes)
	}
	dst := []byte(fmt.Sprintf("%s_XMD:SHA-256_SSWU_RO_", p.Params.Name))
	err = p.Arithmetic.Hash(p, EllipticPointHasherSha256(), seed[:], dst)
	if err != nil {
		return nil, errors.Wrap(err, "ecc hash failed")
	}
	return p, nil
}

// Hash uses the hasher to map bytes to a valid point.
func (p *EllipticPoint4) Hash(bytes []byte, hasher *EllipticPointHasher) (*EllipticPoint4, error) {
	dst := []byte(fmt.Sprintf("%s_%s:%s_SSWU_RO_", p.Params.Name, hasher.hashType, hasher.name))
	err := p.Arithmetic.Hash(p, hasher, bytes, dst)
	if err != nil {
		return nil, errors.Wrap(err, "hash failed")
	}
	return p, nil
}

// Identity returns the identity point.
func (p *EllipticPoint4) Identity() *EllipticPoint4 {
	p.X.SetZero()
	p.Y.SetZero()
	p.Z.SetZero()
	return p
}

// Generator returns the base point for the curve.
func (p *EllipticPoint4) Generator() *EllipticPoint4 {
	p.X.Set(p.Params.Gx)
	p.Y.Set(p.Params.Gy)
	p.Z.SetOne()
	return p
}

// IsIdentity returns true if this point is at infinity.
func (p *EllipticPoint4) IsIdentity() bool {
	return p.Z.IsZero() == 1
}

// Double this point.
func (p *EllipticPoint4) Double(point *EllipticPoint4) *EllipticPoint4 {
	p.Set(point)
	p.Arithmetic.Double(p, point)
	return p
}

// Neg negates this point.
func (p *EllipticPoint4) Neg(point *EllipticPoint4) *EllipticPoint4 {
	p.Set(point)
	p.Y.Neg(p.Y)
	return p
}

// Add adds the two points.
func (p *EllipticPoint4) Add(lhs, rhs *EllipticPoint4) *EllipticPoint4 {
	p.Set(lhs)
	p.Arithmetic.Add(p, lhs, rhs)
	return p
}

// Sub subtracts the two points.
func (p *EllipticPoint4) Sub(lhs, rhs *EllipticPoint4) *EllipticPoint4 {
	p.Set(lhs)
	p.Arithmetic.Add(p, lhs, new(EllipticPoint4).Neg(rhs))
	return p
}

// Mul multiplies this point by the input scalar.
func (p *EllipticPoint4) Mul(point *EllipticPoint4, scalar *Field4) *EllipticPoint4 {
	bytes := scalar.Bytes()
	precomputed := [16]*EllipticPoint4{}
	precomputed[0] = new(EllipticPoint4).Set(point).Identity()
	precomputed[1] = new(EllipticPoint4).Set(point)
	for i := 2; i < 16; i += 2 {
		precomputed[i] = new(EllipticPoint4).Set(point).Double(precomputed[i>>1])
		precomputed[i+1] = new(EllipticPoint4).Set(point).Add(precomputed[i], point)
	}
	p.Identity()
	for i := 0; i < 256; i += 4 {
		// Brouwer / windowing method. window size of 4.
		for j := 0; j < 4; j++ {
			p.Double(p)
		}
		window := bytes[32-1-i>>3] >> (4 - i&0x04) & 0x0F
		p.Add(p, precomputed[window])
	}
	return p
}

// Equal returns 1 if the two points are equal 0 otherwise.
func (p *EllipticPoint4) Equal(rhs *EllipticPoint4) int {
	var x1, x2, y1, y2 Field4

	x1.Arithmetic = p.X.Arithmetic
	x2.Arithmetic = p.X.Arithmetic
	y1.Arithmetic = p.Y.Arithmetic
	y2.Arithmetic = p.Y.Arithmetic

	x1.Mul(p.X, rhs.Z)
	x2.Mul(rhs.X, p.Z)

	y1.Mul(p.Y, rhs.Z)
	y2.Mul(rhs.Y, p.Z)

	e1 := p.Z.IsZero()
	e2 := rhs.Z.IsZero()

	// Both at infinity or coordinates are the same
	return (e1 & e2) | (^e1 & ^e2)&x1.Equal(&x2)&y1.Equal(&y2)
}

// Set copies clone into p.
func (p *EllipticPoint4) Set(clone *EllipticPoint4) *EllipticPoint4 {
	p.X = new(Field4).Set(clone.X)
	p.Y = new(Field4).Set(clone.Y)
	p.Z = new(Field4).Set(clone.Z)
	p.Params = clone.Params
	p.Arithmetic = clone.Arithmetic
	return p
}

// BigInt returns the x and y as big.Ints in affine.
func (p *EllipticPoint4) BigInt() (x, y *big.Int) {
	t := new(EllipticPoint4).Set(p)
	p.Arithmetic.ToAffine(t, p)
	x = t.X.BigInt()
	y = t.Y.BigInt()
	return x, y
}

// SetBigInt creates a point from affine x, y
// and returns the point if it is on the curve.
func (p *EllipticPoint4) SetBigInt(x, y *big.Int) (*EllipticPoint4, error) {
	xx := &Field4{
		Params:     p.Params.Gx.Params,
		Arithmetic: p.Params.Gx.Arithmetic,
	}
	xx.SetBigInt(x)
	yy := &Field4{
		Params:     p.Params.Gx.Params,
		Arithmetic: p.Params.Gx.Arithmetic,
	}
	yy.SetBigInt(y)
	pp := new(EllipticPoint4).Set(p)

	zero := new(Field4).Set(xx).SetZero()
	one := new(Field4).Set(xx).SetOne()
	isIdentity := xx.IsZero() & yy.IsZero()
	pp.X = xx.CMove(xx, zero, isIdentity)
	pp.Y = yy.CMove(yy, zero, isIdentity)
	pp.Z = one.CMove(one, zero, isIdentity)
	if !p.Arithmetic.IsOnCurve(pp) && isIdentity == 0 {
		return nil, fmt.Errorf("invalid coordinates")
	}
	return p.Set(pp), nil
}

// GetX returns the affine X coordinate.
func (p *EllipticPoint4) GetX() *Field4 {
	t := new(EllipticPoint4).Set(p)
	p.Arithmetic.ToAffine(t, p)
	return t.X
}

// GetY returns the affine Y coordinate.
func (p *EllipticPoint4) GetY() *Field4 {
	t := new(EllipticPoint4).Set(p)
	p.Arithmetic.ToAffine(t, p)
	return t.Y
}

// IsOnCurve determines if this point represents a valid curve point.
func (p *EllipticPoint4) IsOnCurve() bool {
	return p.Arithmetic.IsOnCurve(p)
}

// ToAffine converts the point into affine coordinates.
func (p *EllipticPoint4) ToAffine(clone *EllipticPoint4) *EllipticPoint4 {
	p.Arithmetic.ToAffine(p, clone)
	return p
}

// SumOfProducts computes the multi-exponentiation for the specified
// points and scalars and stores the result in `p`.
// Returns an error if the lengths of the arguments is not equal.
func (p *EllipticPoint4) SumOfProducts(points []*EllipticPoint4, scalars []*Field4) (*EllipticPoint4, error) {
	const Upper = 256
	const W = 4
	const Windows = Upper / W // careful--use ceiling division in case this doesn't divide evenly
	if len(points) != len(scalars) {
		return nil, fmt.Errorf("length mismatch")
	}

	bucketSize := 1 << W
	windows := make([]*EllipticPoint4, Windows)
	bytes := make([][32]byte, len(scalars))
	buckets := make([]*EllipticPoint4, bucketSize)

	for i, scalar := range scalars {
		bytes[i] = scalar.Bytes()
	}
	for i := range windows {
		windows[i] = new(EllipticPoint4).Set(p).Identity()
	}

	for i := 0; i < bucketSize; i++ {
		buckets[i] = new(EllipticPoint4).Set(p).Identity()
	}

	sum := new(EllipticPoint4).Set(p)

	for j := 0; j < len(windows); j++ {
		for i := 0; i < bucketSize; i++ {
			buckets[i].Identity()
		}

		for i := 0; i < len(scalars); i++ {
			// j*W to get the nibble
			// >> 3 to convert to byte, / 8
			// (W * j & W) gets the nibble, mod W
			// 1 << W - 1 to get the offset
			index := bytes[i][j*W>>3] >> (W * j & W) & (1<<W - 1) // little-endian
			buckets[index].Add(buckets[index], points[i])
		}

		sum.Identity()

		for i := bucketSize - 1; i > 0; i-- {
			sum.Add(sum, buckets[i])
			windows[j].Add(windows[j], sum)
		}
	}

	p.Identity()
	for i := len(windows) - 1; i >= 0; i-- {
		for j := 0; j < W; j++ {
			p.Double(p)
		}

		p.Add(p, windows[i])
	}
	return p, nil
}

// CMove returns arg1 if choice == 0, otherwise returns arg2.
func (*EllipticPoint4) CMove(pt1, pt2 *EllipticPoint4, choice int) *EllipticPoint4 {
	pt1.X.CMove(pt1.X, pt2.X, choice)
	pt1.Y.CMove(pt1.Y, pt2.Y, choice)
	pt1.Z.CMove(pt1.Z, pt2.Z, choice)
	return pt1
}
