package native

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/mikelodder7/curvey/internal"
)

// Field6Limbs is the number of limbs needed to represent this field.
const Field6Limbs = 6

// Field6Bytes is the number of bytes needed to represent this field.
const Field6Bytes = 48

// WideField6Bytes is the number of bytes needed for safe conversion
// to this field to avoid bias when reduced.
const WideField6Bytes = 96

// Field6 represents a field element.
type Field6 struct {
	// Value is the field elements value
	Value [Field6Limbs]uint64
	// Params are the field parameters
	Params *Field6Params
	// Arithmetic are the field methods
	Arithmetic Field6Arithmetic
}

// Field6Params are the field parameters.
type Field6Params struct {
	// R is 2^256 mod Modulus
	R [Field6Limbs]uint64
	// R2 is 2^512 mod Modulus
	R2 [Field6Limbs]uint64
	// R3 is 2^768 mod Modulus
	R3 [Field6Limbs]uint64
	// Modulus of the field
	Modulus [Field6Limbs]uint64
	// Modulus as big.Int
	BiModulus *big.Int
}

// Field6Arithmetic are the methods that can be done on a field.
type Field6Arithmetic interface {
	// ToMontgomery converts this field to montgomery form
	ToMontgomery(out, arg *[Field6Limbs]uint64)
	// FromMontgomery converts this field from montgomery form
	FromMontgomery(out, arg *[Field6Limbs]uint64)
	// Neg performs modular negation
	Neg(out, arg *[Field6Limbs]uint64)
	// Square performs modular square
	Square(out, arg *[Field6Limbs]uint64)
	// Mul performs modular multiplication
	Mul(out, arg1, arg2 *[Field6Limbs]uint64)
	// Add performs modular addition
	Add(out, arg1, arg2 *[Field6Limbs]uint64)
	// Sub performs modular subtraction
	Sub(out, arg1, arg2 *[Field6Limbs]uint64)
	// Sqrt performs modular square root
	Sqrt(wasSquare *int, out, arg *[Field6Limbs]uint64)
	// Invert performs modular inverse
	Invert(wasInverted *int, out, arg *[Field6Limbs]uint64)
	// FromBytes converts a little endian byte array into a field element
	FromBytes(out *[Field6Limbs]uint64, arg *[Field6Bytes]byte)
	// ToBytes converts a field element to a little endian byte array
	ToBytes(out *[Field6Bytes]byte, arg *[Field6Limbs]uint64)
	// Selectznz performs conditional select.
	// selects arg1 if choice == 0 and arg2 if choice == 1
	Selectznz(out, arg1, arg2 *[Field6Limbs]uint64, choice int)
}

// Cmp returns -1 if f < rhs
// 0 if f == rhs
// 1 if f > rhs.
func (f *Field6) Cmp(rhs *Field6) int {
	return cmp6Helper(&f.Value, &rhs.Value)
}

// cmp6Helper returns -1 if lhs < rhs
// -1 if lhs == rhs
// 1 if lhs > rhs
// Public only for convenience for some internal implementations.
func cmp6Helper(lhs, rhs *[Field6Limbs]uint64) int {
	gt := uint64(0)
	lt := uint64(0)
	for i := 5; i >= 0; i-- {
		// convert to two 64-bit numbers where
		// the leading bits are zeros and hold no meaning
		//  so rhs - fp actually means gt
		// and fp - rhs actually means lt.
		rhsH := rhs[i] >> 32
		rhsL := rhs[i] & 0xffffffff
		lhsH := lhs[i] >> 32
		lhsL := lhs[i] & 0xffffffff

		// Check the leading bit
		// if negative then fp > rhs
		// if positive then fp < rhs
		gt |= (rhsH - lhsH) >> 32 & 1 &^ lt
		lt |= (lhsH - rhsH) >> 32 & 1 &^ gt
		gt |= (rhsL - lhsL) >> 32 & 1 &^ lt
		lt |= (lhsL - rhsL) >> 32 & 1 &^ gt
	}
	// Make the result -1 for <, 0 for =, 1 for >
	return int(gt) - int(lt)
}

// Equal returns 1 if f == rhs, 0 otherwise.
func (f *Field6) Equal(rhs *Field6) int {
	return equal6Helper(&f.Value, &rhs.Value)
}

func equal6Helper(lhs, rhs *[Field6Limbs]uint64) int {
	t := lhs[0] ^ rhs[0]
	t |= lhs[1] ^ rhs[1]
	t |= lhs[2] ^ rhs[2]
	t |= lhs[3] ^ rhs[3]
	t |= lhs[4] ^ rhs[4]
	t |= lhs[5] ^ rhs[5]
	return int(((int64(t) | int64(-t)) >> 63) + 1)
}

// New returns a brand new field
func (f *Field6) New() *Field6 {
	return &Field6{
		Value:      [Field6Limbs]uint64{0, 0, 0, 0, 0, 0},
		Params:     f.Params,
		Arithmetic: f.Arithmetic,
	}
}

// IsZero returns 1 if f == 0, 0 otherwise.
func (f *Field6) IsZero() int {
	t := f.Value[0]
	t |= f.Value[1]
	t |= f.Value[2]
	t |= f.Value[3]
	t |= f.Value[4]
	t |= f.Value[5]
	return int(((int64(t) | int64(-t)) >> 63) + 1)
}

// IsNonZero returns 1 if f != 0, 0 otherwise.
func (f *Field6) IsNonZero() int {
	t := f.Value[0]
	t |= f.Value[1]
	t |= f.Value[2]
	t |= f.Value[3]
	t |= f.Value[4]
	t |= f.Value[5]
	return int(-((int64(t) | int64(-t)) >> 63))
}

// IsOne returns 1 if f == 1, 0 otherwise.
func (f *Field6) IsOne() int {
	return equal6Helper(&f.Value, &f.Params.R)
}

// Set f = rhs.
func (f *Field6) Set(rhs *Field6) *Field6 {
	f.Value[0] = rhs.Value[0]
	f.Value[1] = rhs.Value[1]
	f.Value[2] = rhs.Value[2]
	f.Value[3] = rhs.Value[3]
	f.Value[4] = rhs.Value[4]
	f.Value[5] = rhs.Value[5]
	f.Params = rhs.Params
	f.Arithmetic = rhs.Arithmetic
	return f
}

// SetUint64 f = rhs.
func (f *Field6) SetUint64(rhs uint64) *Field6 {
	t := &[Field6Limbs]uint64{rhs, 0, 0, 0, 0, 0}
	f.Arithmetic.ToMontgomery(&f.Value, t)
	return f
}

// SetOne f = r.
func (f *Field6) SetOne() *Field6 {
	f.Value[0] = f.Params.R[0]
	f.Value[1] = f.Params.R[1]
	f.Value[2] = f.Params.R[2]
	f.Value[3] = f.Params.R[3]
	f.Value[4] = f.Params.R[4]
	f.Value[5] = f.Params.R[5]
	return f
}

// SetZero f = 0.
func (f *Field6) SetZero() *Field6 {
	f.Value[0] = 0
	f.Value[1] = 0
	f.Value[2] = 0
	f.Value[3] = 0
	f.Value[4] = 0
	f.Value[5] = 0
	return f
}

// SetBytesWide takes 96 bytes as input and treats them as a 512-bit number.
// Attributed to https://github.com/zcash/pasta_curves/blob/main/src/fields/Fp.rs#L255
// We reduce an arbitrary 512-bit number by decomposing it into two 256-bit digits
// with the higher bits multiplied by 2^256. Thus, we perform two reductions
//
// 1. the lower bits are multiplied by r^2, as normal
// 2. the upper bits are multiplied by r^2 * 2^256 = r^3
//
// and computing their sum in the field. It remains to see that arbitrary 256-bit
// numbers can be placed into Montgomery form safely using the reduction. The
// reduction works so long as the product is less than r=2^256 multiplied by
// the modulus. This holds because for any `c` smaller than the modulus, we have
// that (2^256 - 1)*c is an acceptable product for the reduction. Therefore, the
// reduction always works so long as `c` is in the field; in this case it is either the
// constant `r2` or `r3`.
func (f *Field6) SetBytesWide(input *[WideField6Bytes]byte) *Field6 {
	d0 := [Field6Limbs]uint64{
		binary.LittleEndian.Uint64(input[:8]),
		binary.LittleEndian.Uint64(input[8:16]),
		binary.LittleEndian.Uint64(input[16:24]),
		binary.LittleEndian.Uint64(input[24:32]),
		binary.LittleEndian.Uint64(input[32:40]),
		binary.LittleEndian.Uint64(input[40:48]),
	}
	d1 := [Field6Limbs]uint64{
		binary.LittleEndian.Uint64(input[48:56]),
		binary.LittleEndian.Uint64(input[56:64]),
		binary.LittleEndian.Uint64(input[64:72]),
		binary.LittleEndian.Uint64(input[72:80]),
		binary.LittleEndian.Uint64(input[80:88]),
		binary.LittleEndian.Uint64(input[88:96]),
	}
	// f.Arithmetic.ToMontgomery(&d0, &d0)
	// f.Arithmetic.Mul(&d1, &d1, &f.Params.R2)
	// f.Arithmetic.Add(&f.Value, &d0, &d0)
	// Convert to Montgomery form
	tv1 := &[Field6Limbs]uint64{}
	tv2 := &[Field6Limbs]uint64{}
	// d0*r2 + d1*r3
	f.Arithmetic.Mul(tv1, &d0, &f.Params.R2)
	f.Arithmetic.Mul(tv2, &d1, &f.Params.R3)
	f.Arithmetic.Add(&f.Value, tv1, tv2)
	return f
}

// SetBytes attempts to convert a little endian byte representation
// of a scalar into a `Fp`, failing if input is not canonical.
func (f *Field6) SetBytes(input *[Field6Bytes]byte) (*Field6, error) {
	d0 := [Field6Limbs]uint64{0, 0, 0, 0, 0, 0}
	f.Arithmetic.FromBytes(&d0, input)

	if cmp6Helper(&d0, &f.Params.Modulus) != -1 {
		return nil, fmt.Errorf("invalid byte sequence")
	}
	return f.SetLimbs(&d0), nil
}

// SetBigInt initializes an element from big.Int
// The value is reduced by the modulus.
func (f *Field6) SetBigInt(bi *big.Int) *Field6 {
	var buffer [Field6Bytes]byte
	t := new(big.Int).Set(bi)
	t.Mod(t, f.Params.BiModulus)
	t.FillBytes(buffer[:])
	copy(buffer[:], internal.ReverseBytes(buffer[:]))
	_, _ = f.SetBytes(&buffer)
	return f
}

// SetRaw converts a raw array into a field element
// Assumes input is already in montgomery form.
func (f *Field6) SetRaw(input *[Field6Limbs]uint64) *Field6 {
	f.Value[0] = input[0]
	f.Value[1] = input[1]
	f.Value[2] = input[2]
	f.Value[3] = input[3]
	f.Value[4] = input[4]
	f.Value[5] = input[5]
	return f
}

// SetLimbs converts an array into a field element
// by converting to montgomery form.
func (f *Field6) SetLimbs(input *[Field6Limbs]uint64) *Field6 {
	f.Arithmetic.ToMontgomery(&f.Value, input)
	return f
}

// Bytes converts this element into a byte representation
// in little endian byte order.
func (f *Field6) Bytes() [Field6Bytes]byte {
	var output [Field6Bytes]byte
	tv := &[Field6Limbs]uint64{}
	f.Arithmetic.FromMontgomery(tv, &f.Value)
	f.Arithmetic.ToBytes(&output, tv)
	return output
}

// BigInt converts this element into the big.Int struct.
func (f *Field6) BigInt() *big.Int {
	buffer := f.Bytes()
	return new(big.Int).SetBytes(internal.ReverseBytes(buffer[:]))
}

// Raw converts this element into the a [Field4Limbs]uint64.
func (f *Field6) Raw() [Field6Limbs]uint64 {
	res := &[Field6Limbs]uint64{}
	f.Arithmetic.FromMontgomery(res, &f.Value)
	return *res
}

// Double this element.
func (f *Field6) Double(a *Field6) *Field6 {
	f.Arithmetic.Add(&f.Value, &a.Value, &a.Value)
	return f
}

// Square this element.
func (f *Field6) Square(a *Field6) *Field6 {
	f.Arithmetic.Square(&f.Value, &a.Value)
	return f
}

func (f *Field6) MulBy3b(arg *Field6) *Field6 {
	a := new(Field6)
	t := new(Field6)

	a.Set(arg)
	t.Set(arg)
	a.Double(arg)
	t.Double(a)
	a.Double(t)
	a.Add(a, t)
	return f.Set(a)
}

// Sqrt this element, if it exists. If true, then value
// is a square root. If false, value is a QNR.
func (f *Field6) Sqrt(a *Field6) (*Field6, bool) {
	wasSquare := 0
	f.Arithmetic.Sqrt(&wasSquare, &f.Value, &a.Value)
	return f, wasSquare == 1
}

// Invert this element i.e. compute the multiplicative inverse
// return false, zero if this element is zero.
func (f *Field6) Invert(a *Field6) (*Field6, bool) {
	wasInverted := 0
	f.Arithmetic.Invert(&wasInverted, &f.Value, &a.Value)
	return f, wasInverted == 1
}

// Mul returns the result from multiplying this element by rhs.
func (f *Field6) Mul(lhs, rhs *Field6) *Field6 {
	f.Arithmetic.Mul(&f.Value, &lhs.Value, &rhs.Value)
	return f
}

// Sub returns the result from subtracting rhs from this element.
func (f *Field6) Sub(lhs, rhs *Field6) *Field6 {
	f.Arithmetic.Sub(&f.Value, &lhs.Value, &rhs.Value)
	return f
}

// Add returns the result from adding rhs to this element.
func (f *Field6) Add(lhs, rhs *Field6) *Field6 {
	f.Arithmetic.Add(&f.Value, &lhs.Value, &rhs.Value)
	return f
}

// Neg returns negation of this element.
func (f *Field6) Neg(input *Field6) *Field6 {
	f.Arithmetic.Neg(&f.Value, &input.Value)
	return f
}

// Exp raises base^exp.
func (f *Field6) Exp(base, exp *Field6) *Field6 {
	e := [Field6Limbs]uint64{}
	f.Arithmetic.FromMontgomery(&e, &exp.Value)
	Pow6(&f.Value, &base.Value, &e, f.Params, f.Arithmetic)
	return f
}

// CMove sets f = lhs if choice == 0 and f = rhs if choice == 1.
func (f *Field6) CMove(lhs, rhs *Field6, choice int) *Field6 {
	f.Arithmetic.Selectznz(&f.Value, &lhs.Value, &rhs.Value, choice)
	return f
}

// Pow6 raises base^exp. The result is written to out.
// Public only for convenience for some internal implementations.
func Pow6(out, base, exp *[Field6Limbs]uint64, params *Field6Params, arithmetic Field6Arithmetic) {
	res := [Field6Limbs]uint64{params.R[0], params.R[1], params.R[2], params.R[3], params.R[4], params.R[5]}
	tmp := [Field6Limbs]uint64{}

	for i := len(exp) - 1; i >= 0; i-- {
		for j := 63; j >= 0; j-- {
			arithmetic.Square(&res, &res)
			arithmetic.Mul(&tmp, &res, base)
			arithmetic.Selectznz(&res, &res, &tmp, int(exp[i]>>j)&1)
		}
	}
	out[0] = res[0]
	out[1] = res[1]
	out[2] = res[2]
	out[3] = res[3]
	out[4] = res[4]
	out[5] = res[5]
}

// Pow2k6 raises arg to the power `2^k`. This result is written to out.
// Public only for convenience for some internal implementations.
func Pow2k6(out, arg *[Field6Limbs]uint64, k int, arithmetic Field6Arithmetic) {
	var t [Field6Limbs]uint64
	t[0] = arg[0]
	t[1] = arg[1]
	t[2] = arg[2]
	t[3] = arg[3]
	t[4] = arg[4]
	t[5] = arg[5]
	for i := 0; i < k; i++ {
		arithmetic.Square(&t, &t)
	}

	out[0] = t[0]
	out[1] = t[1]
	out[2] = t[2]
	out[3] = t[3]
	out[4] = t[4]
	out[5] = t[5]
}
