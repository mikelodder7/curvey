package ed448

import (
	"fmt"
	"github.com/mikelodder7/curvey/internal"
	"github.com/mikelodder7/curvey/native"
	"io"
	"math/big"
)

type Fp struct {
	Value *internal.Field
}

var (
	fpParams = internal.FieldParams{
		Modulus: []uint64{
			0xffffffffffffffff,
			0xffffffffffffffff,
			0xffffffffffffffff,
			0xfffffffeffffffff,
			0xffffffffffffffff,
			0xffffffffffffffff,
			0xffffffffffffffff,
		},
		ModulusNegInv: 1,
		// 2^448 mod p.
		R: []uint64{1, 0, 0, 0x100000000, 0, 0, 0},
		// 2^896 mod p.
		R2: []uint64{2, 0, 0, 0x300000000, 0, 0, 0},
		// 2^1344 mod p.
		R3:    []uint64{5, 0, 0, 0x800000000, 0, 0, 0},
		Limbs: 7,
		BiModulus: new(big.Int).SetBytes([]byte{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		}),
		SqrtParams: internal.SqrtParams{
			C1: 1,
			C3: []uint64{
				0xffffffffffffffff,
				0xffffffffffffffff,
				0xffffffffffffffff,
				0xffffffffbfffffff,
				0xffffffffffffffff,
				0xffffffffffffffff,
				0x3fffffffffffffff,
			},
			C5: []uint64{
				0xfffffffffffffffe,
				0xffffffffffffffff,
				0xffffffffffffffff,
				0xfffffffdffffffff,
				0xffffffffffffffff,
				0xffffffffffffffff,
				0xffffffffffffffff,
			},
			IsSquare: []uint64{
				0xffffffffffffffff,
				0xffffffffffffffff,
				0xffffffffffffffff,
				0xffffffff7fffffff,
				0xffffffffffffffff,
				0xffffffffffffffff,
				0x7fffffffffffffff,
			},
		},
	}
)

func FpNew() *Fp {
	value := new(internal.Field).Init(&fpParams, &fpParams)
	return &Fp{value}
}

// IsZero returns 1 if Fp == 0, 0 otherwise.
func (f *Fp) IsZero() int {
	return f.Value.IsZeroI()
}

// IsNonZero returns 1 if Fp != 0, 0 otherwise.
func (f *Fp) IsNonZero() int {
	return f.Value.IsNonZeroI()
}

// IsOne returns 1 if Fp == 1, 0 otherwise
func (f *Fp) IsOne() int {
	return f.Value.IsOneI()
}

// IsSquare returns 1 if Fp is a quadratic residue, 0 otherwise
func (f *Fp) IsSquare() int {
	return f.Value.IsSquareI()
}

// Cmp returns -1 if f < rhs
// 0 if f == rhs
// 1 if f > rhs.
func (f *Fp) Cmp(rhs *Fp) int {
	return f.Value.Cmp(rhs.Value)
}

// Sgn0I returns the lowest bit value
func (f *Fp) Sgn0I() int {
	return f.Value.Sgn0I()
}

// SetOne f = one
func (f *Fp) SetOne() *Fp {
	f.Value.SetOne()
	return f
}

// SetZero f = one
func (f *Fp) SetZero() *Fp {
	f.Value.SetZero()
	return f
}

// EqualI returns 1 if Fp == rhs, 0 otherwise.
func (f *Fp) EqualI(rhs *Fp) int {
	return f.Value.EqualI(rhs.Value)
}

func (f *Fp) SetUint64(rhs uint64) *Fp {
	f.Value.SetUint64(rhs)
	return f
}

func (f *Fp) Random(reader io.Reader) (*Fp, error) {
	var t [64]byte
	n, err := reader.Read(t[:])
	if err != nil {
		return nil, err
	}
	if n != 64 {
		return nil, fmt.Errorf("can only read %d when %d are needed", n, 64)
	}
	return f.Hash(t[:]), nil
}

func (f *Fp) Hash(input []byte) *Fp {
	dst := []byte("edwards448_XOF:SHAKE256_RO_")
	xof := native.ExpandMsgXof(native.EllipticPointHasherShake256(), input, dst, 88)
	var t [112]byte
	copy(t[:], xof[:])
	return f.SetBytesWide(&t)
}

func (f *Fp) toMontgomery(a *Fp) *Fp {
	f.Value.Arithmetic.ToMontgomery(&f.Value.Value, &a.Value.Value)
	return f
}

func (f *Fp) fromMontgomery(a *Fp) *Fp {
	f.Value.Arithmetic.FromMontgomery(&f.Value.Value, &a.Value.Value)
	return f
}

func (f *Fp) Neg(a *Fp) *Fp {
	f.Value.Neg(a.Value)
	return f
}

func (f *Fp) Square(a *Fp) *Fp {
	f.Value.Square(a.Value)
	return f
}

func (f *Fp) Double(a *Fp) *Fp {
	f.Value.Double(a.Value)
	return f
}

func (f *Fp) Mul(arg1, arg2 *Fp) *Fp {
	f.Value.Mul(arg1.Value, arg2.Value)
	return f
}

func (f *Fp) Add(arg1, arg2 *Fp) *Fp {
	f.Value.Add(arg1.Value, arg2.Value)
	return f
}

// Sub performs modular subtraction.
func (f *Fp) Sub(arg1, arg2 *Fp) *Fp {
	f.Value.Sub(arg1.Value, arg2.Value)
	return f
}

// Sqrt performs modular square root.
func (f *Fp) Sqrt(a *Fp) (*Fp, int) {
	// Shank's method, as p = 3 (mod 4). This means
	// exponentiate by (p+1)/4. This only works for elements
	// that are actually quadratic residue,
	// so check the result at the end.
	z := new(internal.Field).Init(f.Value.Params, f.Value.Arithmetic)
	c := new(internal.Field).Init(f.Value.Params, f.Value.Arithmetic)
	internal.Pow(&z.Value, a.Value.Value, sqrtExp[:], a.Value.Params, a.Value.Arithmetic)

	c.Square(z)
	wasSquare := c.EqualI(a.Value)
	f.Value.CMove(f.Value, z, wasSquare)
	return f, wasSquare
}

// Invert performs modular inverse.
func (f *Fp) Invert(a *Fp) (*Fp, int) {
	// Exponentiate by p - 2
	exp := new(internal.Field).Init(f.Value.Params, f.Value.Arithmetic)
	copy(exp.Value, []uint64{
		0xfffffffffffffffd,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xfffffffeffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
	})
	f.Value.Exp(a.Value, exp)
	wasInverted := a.IsNonZero()
	f.Value.CMove(a.Value, f.Value, wasInverted)
	return f, wasInverted
}

// SetCanonicalBytes converts a little endian byte array into a field element
// returns nil if the bytes are not in the field
func (f *Fp) SetCanonicalBytes(arg *[56]byte) (*Fp, error) {
	_, err := f.Value.SetBytes(arg[:])

	if err != nil {
		return nil, err
	}
	return f, nil
}

// SetBytes converts a little endian byte array into a field element
// returns nil if the bytes are not in the field or the length is invalid
func (f *Fp) SetBytes(arg []byte) (*Fp, error) {
	_, err := f.Value.SetBytes(arg)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// SetBytesWide takes 112 bytes as input and treats them as a 896-bit number.
func (f *Fp) SetBytesWide(a *[112]byte) *Fp {
	_, err := f.Value.SetBytesWide(a[:])
	if err != nil {
		return nil
	}
	return f
}

// SetBigInt initializes an element from big.Int
// The value is reduced by the modulus.
func (f *Fp) SetBigInt(bi *big.Int) *Fp {
	f.Value.SetBigInt(bi)
	return f
}

// Set copies a into Fp.
func (f *Fp) Set(a *Fp) *Fp {
	f.Value.Set(a.Value)
	return f
}

// SetLimbs converts an array into a field element
// by converting to montgomery form.
func (f *Fp) SetLimbs(a *[7]uint64) *Fp {
	_, err := f.Value.SetLimbs(a[:])
	if err != nil {
		return nil
	}
	return f
}

// SetRaw converts a raw array into a field element
// Assumes input is already in montgomery form.
func (f *Fp) SetRaw(a *[7]uint64) *Fp {
	f.Value.SetRaw(a[:])
	return f
}

// Bytes converts a field element to a little endian byte array.
func (f *Fp) Bytes() [56]byte {
	var out [56]byte
	copy(out[:], f.Value.Bytes())
	return out
}

// BigInt converts this element into the big.Int struct.
func (f *Fp) BigInt() *big.Int {
	buffer := f.Bytes()
	return new(big.Int).SetBytes(internal.ReverseBytes(buffer[:]))
}

// Raw converts this element into the a []uint64.
func (f *Fp) Raw() []uint64 {
	var t []uint64
	f.Value.Arithmetic.FromMontgomery(&t, &f.Value.Value)
	return t
}

// CMove performs conditional select.
// selects arg1 if choice == 0 and arg2 if choice == 1.
func (f *Fp) CMove(arg1, arg2 *Fp, choice int) *Fp {
	f.Value.CMove(arg1.Value, arg2.Value, choice)
	return f
}

// CNeg conditionally negates a if choice == 1.
func (f *Fp) CNeg(a *Fp, choice int) *Fp {
	var t Fp
	t.Neg(a)
	return f.CMove(f, &t, choice)
}

// CSwap conditionally swaps this with a if choice == 1
// a is f and f is a if choice == 1
func (f *Fp) CSwap(a *Fp, choice int) *Fp {
	f.Value.CSwap(a.Value, choice)
	return f
}

// Exp raises base^exp.
func (f *Fp) Exp(base, exp *Fp) *Fp {
	f.Value.Exp(base.Value, exp.Value)
	return f
}

func (f *Fp) SquareN(a *Fp, n int) *Fp {
	t := FpNew().Square(a)

	n -= 1
	for ; n > 0; n-- {
		t.Square(t)
	}
	return f.Set(t)
}

// InvSqrt computes 1/sqrt(a)
func (f *Fp) InvSqrt(a *Fp) (*Fp, int) {
	l0 := FpNew()
	l1 := FpNew()
	l2 := FpNew()

	l1.Square(a)
	l2.Mul(l1, a)
	l1.Square(l2)
	l2.Mul(l1, a)
	l1.SquareN(l2, 3)
	l0.Mul(l2, l1)
	l1.SquareN(l0, 3)
	l0.Mul(l2, l1)
	l2.SquareN(l0, 9)
	l1.Mul(l0, l2)
	l0.Square(l1)
	l2.Mul(l0, a)
	l0.SquareN(l2, 18)
	l2.Mul(l1, l0)
	l0.SquareN(l2, 37)
	l1.Mul(l2, l0)
	l0.SquareN(l1, 37)
	l1.Mul(l2, l0)
	l0.SquareN(l1, 111)
	l2.Mul(l1, l0)
	l0.Square(l2)
	l1.Mul(l0, a)
	l0.SquareN(l1, 223)
	l1.Mul(l2, l0)
	l2.Square(l1)
	l0.Mul(l2, a)

	isResidue := l0.IsOne()
	return l1, isResidue
}

// SqrtRatio computes the square root ratio of two elements
func (f *Fp) SqrtRatio(u, v *Fp) (*Fp, int) {
	// Compute sqrt(1/(uv))
	x := FpNew().Mul(u, v)
	invSqrtX, isRes := FpNew().InvSqrt(x)
	// Return u * sqrt(1/(uv)) == sqrt(u/v). However, since this trick only works
	// for u != 0, check for that case explicitly (when u == 0 then inv_sqrt_x
	// will be zero, which is what we want, but is_res will be 0)
	zeroU := u.IsZero()
	invSqrtX.Mul(invSqrtX, u)
	return invSqrtX, zeroU | isRes
}
