package native

import (
	"fmt"
	"io"
	"math/big"

	"github.com/pkg/errors"
)

// EllipticPoint6 represents a Weierstrauss elliptic curve point.
type EllipticPoint6 struct {
	X          *Field6
	Y          *Field6
	Z          *Field6
	Params     *EllipticPoint6Params
	Arithmetic EllipticPoint6Arithmetic
}

// EllipticPoint6Params are the Weierstrauss curve parameters
// such as the name, the coefficients the generator point,
// and the prime bit size.
type EllipticPoint6Params struct {
	Name    string
	A       *Field6
	B       *Field6
	Gx      *Field6
	Gy      *Field6
	BitSize int
}

// EllipticPoint6Arithmetic are the methods that specific curves
// need to implement for higher abstractions to wrap the point.
type EllipticPoint6Arithmetic interface {
	// Hash a byte sequence to the curve using the specified hasher
	// and dst and store the result in out
	Hash(out *EllipticPoint6, hasher *EllipticPointHasher, bytes, dst []byte) error
	// Double arg and store the result in out
	Double(out, arg *EllipticPoint6)
	// Add arg1 with arg2 and store the result in out
	Add(out, arg1, arg2 *EllipticPoint6)
	// IsOnCurve tests arg if it represents a valid point on the curve
	IsOnCurve(arg *EllipticPoint6) bool
	// ToAffine converts arg to affine coordinates storing the result in out
	ToAffine(out, arg *EllipticPoint6)
	// RhsEquation computes the right-hand side of the ecc equation
	RhsEquation(out, x *Field6)
}

// Random creates a random point on the curve
// from the specified reader.
func (p *EllipticPoint6) Random(reader io.Reader) (*EllipticPoint6, error) {
	var seed [WideField6Bytes]byte
	n, err := reader.Read(seed[:])
	if err != nil {
		return nil, errors.Wrap(err, "random could not read from stream")
	}
	if n != WideField6Bytes {
		return nil, fmt.Errorf("insufficient bytes read %d when %d are needed", n, WideField6Bytes)
	}
	dst := []byte(fmt.Sprintf("%s_XMD:SHA-256_SSWU_RO_", p.Params.Name))
	err = p.Arithmetic.Hash(p, EllipticPointHasherSha256(), seed[:], dst)
	if err != nil {
		return nil, errors.Wrap(err, "ecc hash failed")
	}
	return p, nil
}

// Hash uses the hasher to map bytes to a valid point.
func (p *EllipticPoint6) Hash(bytes []byte, hasher *EllipticPointHasher) (*EllipticPoint6, error) {
	dst := []byte(fmt.Sprintf("%s_%s:%s_SSWU_RO_", p.Params.Name, hasher.hashType, hasher.name))
	err := p.Arithmetic.Hash(p, hasher, bytes, dst)
	if err != nil {
		return nil, errors.Wrap(err, "hash failed")
	}
	return p, nil
}

// Identity returns the identity point.
func (p *EllipticPoint6) Identity() *EllipticPoint6 {
	p.X.SetZero()
	p.Y.SetOne()
	p.Z.SetZero()
	return p
}

// Generator returns the base point for the curve.
func (p *EllipticPoint6) Generator() *EllipticPoint6 {
	p.X.Set(p.Params.Gx)
	p.Y.Set(p.Params.Gy)
	p.Z.SetOne()
	return p
}

// IsIdentity returns true if this point is at infinity.
func (p *EllipticPoint6) IsIdentity() bool {
	return p.Z.IsZero() == 1
}

// Double this point.
func (p *EllipticPoint6) Double(point *EllipticPoint6) *EllipticPoint6 {
	p.Set(point)
	p.Arithmetic.Double(p, point)
	return p
}

// Neg negates this point.
func (p *EllipticPoint6) Neg(point *EllipticPoint6) *EllipticPoint6 {
	p.Set(point)
	p.Y.Neg(p.Y)
	return p
}

// Add adds the two points.
func (p *EllipticPoint6) Add(lhs, rhs *EllipticPoint6) *EllipticPoint6 {
	p.Set(lhs)
	p.Arithmetic.Add(p, lhs, rhs)
	return p
}

// Sub subtracts the two points.
func (p *EllipticPoint6) Sub(lhs, rhs *EllipticPoint6) *EllipticPoint6 {
	p.Set(lhs)
	p.Arithmetic.Add(p, lhs, new(EllipticPoint6).Neg(rhs))
	return p
}

// Mul multiplies this point by the input scalar.
func (p *EllipticPoint6) Mul(point *EllipticPoint6, scalar *Field6) *EllipticPoint6 {
	bytes := scalar.Bytes()
	precomputed := [16]*EllipticPoint6{}
	precomputed[0] = new(EllipticPoint6).Set(point).Identity()
	precomputed[1] = new(EllipticPoint6).Set(point)
	for i := 2; i < 16; i += 2 {
		precomputed[i] = new(EllipticPoint6).Set(point).Double(precomputed[i>>1])
		precomputed[i+1] = new(EllipticPoint6).Set(point).Add(precomputed[i], point)
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
func (p *EllipticPoint6) Equal(rhs *EllipticPoint6) int {
	var x1, x2, y1, y2 Field6

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
func (p *EllipticPoint6) Set(clone *EllipticPoint6) *EllipticPoint6 {
	p.X = new(Field6).Set(clone.X)
	p.Y = new(Field6).Set(clone.Y)
	p.Z = new(Field6).Set(clone.Z)
	p.Params = clone.Params
	p.Arithmetic = clone.Arithmetic
	return p
}

// BigInt returns the x and y as big.Ints in affine.
func (p *EllipticPoint6) BigInt() (x, y *big.Int) {
	t := new(EllipticPoint6).Set(p)
	p.Arithmetic.ToAffine(t, p)
	x = t.X.BigInt()
	y = t.Y.BigInt()
	return x, y
}

// SetBigInt creates a point from affine x, y
// and returns the point if it is on the curve.
func (p *EllipticPoint6) SetBigInt(x, y *big.Int) (*EllipticPoint6, error) {
	xx := &Field6{
		Params:     p.Params.Gx.Params,
		Arithmetic: p.Params.Gx.Arithmetic,
	}
	xx.SetBigInt(x)
	yy := &Field6{
		Params:     p.Params.Gx.Params,
		Arithmetic: p.Params.Gx.Arithmetic,
	}
	yy.SetBigInt(y)
	pp := new(EllipticPoint6).Set(p)

	zero := new(Field6).Set(xx).SetZero()
	one := new(Field6).Set(xx).SetOne()
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
func (p *EllipticPoint6) GetX() *Field6 {
	t := new(EllipticPoint6).Set(p)
	p.Arithmetic.ToAffine(t, p)
	return t.X
}

// GetY returns the affine Y coordinate.
func (p *EllipticPoint6) GetY() *Field6 {
	t := new(EllipticPoint6).Set(p)
	p.Arithmetic.ToAffine(t, p)
	return t.Y
}

// IsOnCurve determines if this point represents a valid curve point.
func (p *EllipticPoint6) IsOnCurve() bool {
	return p.Arithmetic.IsOnCurve(p)
}

// ToAffine converts the point into affine coordinates.
func (p *EllipticPoint6) ToAffine(clone *EllipticPoint6) *EllipticPoint6 {
	p.Arithmetic.ToAffine(p, clone)
	return p
}

// SumOfProducts computes the multi-exponentiation for the specified
// points and scalars and stores the result in `p`.
// Returns an error if the lengths of the arguments is not equal.
func (p *EllipticPoint6) SumOfProducts(points []*EllipticPoint6, scalars []*Field6) (*EllipticPoint6, error) {
	const Upper = 256
	const W = 4
	const Windows = Upper / W // careful--use ceiling division in case this doesn't divide evenly
	if len(points) != len(scalars) {
		return nil, fmt.Errorf("length mismatch")
	}

	bucketSize := 1 << W
	windows := make([]*EllipticPoint6, Windows)
	bytes := make([][48]byte, len(scalars))
	buckets := make([]*EllipticPoint6, bucketSize)

	for i, scalar := range scalars {
		bytes[i] = scalar.Bytes()
	}
	for i := range windows {
		windows[i] = new(EllipticPoint6).Set(p).Identity()
	}

	for i := 0; i < bucketSize; i++ {
		buckets[i] = new(EllipticPoint6).Set(p).Identity()
	}

	sum := new(EllipticPoint6).Set(p)

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
func (*EllipticPoint6) CMove(pt1, pt2 *EllipticPoint6, choice int) *EllipticPoint6 {
	pt1.X.CMove(pt1.X, pt2.X, choice)
	pt1.Y.CMove(pt1.Y, pt2.Y, choice)
	pt1.Z.CMove(pt1.Z, pt2.Z, choice)
	return pt1
}
