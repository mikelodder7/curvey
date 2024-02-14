package internal

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/bits"
)

type Field struct {
	Value      []uint64
	Params     *FieldParams
	Arithmetic FieldArithmetic
}

// FieldParams are the field parameters.
type FieldParams struct {
	// R is 2^N mod Modulus
	R []uint64
	// R2 is 2^2N mod Modulus
	R2 []uint64
	// R3 is 2^3N mod Modulus
	R3 []uint64
	// Modulus of the field
	Modulus []uint64
	// ModulusNegInv of the field, the lowest limb of -(Modulus^-1) mod R
	ModulusNegInv uint64
	// Limbs in the field
	Limbs int
	// Modulus as big.Int
	BiModulus *big.Int
	// SqrtParams is used for computing Tonelli-Shanks ct-sqrt
	SqrtParams SqrtParams
}

// SqrtParams contains the two values needed for computing Tonelli-Shanks ct-sqrt
type SqrtParams struct {
	C1       int
	C3       []uint64
	C5       []uint64
	IsSquare []uint64
}

func (f *FieldParams) newFromHex(modulus string) (*FieldParams, error) {
	modBytes, err := hex.DecodeString(modulus)
	if err != nil {
		return nil, err
	}
	return f.newFromBytes(modBytes)
}

func (f *FieldParams) newFromBytes(beModulusBytes []byte) (*FieldParams, error) {
	f.BiModulus = new(big.Int).SetBytes(beModulusBytes)
	f.Modulus = beBytesToLeUint64Array(beModulusBytes)

	two := new(big.Int).SetInt64(2)
	numBits := int64(len(beModulusBytes) * 8)
	t := new(big.Int).Exp(two, new(big.Int).SetInt64(numBits), f.BiModulus)
	t.FillBytes(beModulusBytes)
	R := new(big.Int).Set(t)
	f.R = beBytesToLeUint64Array(beModulusBytes)

	t.Exp(two, new(big.Int).SetInt64(64), nil)
	modNegInv := new(big.Int).ModInverse(f.BiModulus, t)
	modNegInv.Neg(modNegInv)
	modNegInv.Mod(modNegInv, t)

	t.Exp(two, new(big.Int).SetInt64(numBits*2), f.BiModulus)
	t.FillBytes(beModulusBytes)
	f.R2 = beBytesToLeUint64Array(beModulusBytes)

	t.Exp(two, new(big.Int).SetInt64(numBits*3), f.BiModulus)
	t.FillBytes(beModulusBytes)
	f.R3 = beBytesToLeUint64Array(beModulusBytes)

	f.Limbs = len(beModulusBytes) / 8
	f.ModulusNegInv = modNegInv.Uint64()

	// Set Sqrt params

	// Tonelli-Shanks
	one := new(big.Int).SetInt64(1)

	// c2 = (q - 1) / 2^c1
	c1 := 0
	biC2 := new(big.Int).Sub(f.BiModulus, one)
	for biC2.Bit(0) == 0 {
		c1++
		biC2.Rsh(biC2, 1)
	}
	// c3 = (c2 - 1) / 2
	biC3 := new(big.Int).Sub(biC2, one)
	biC3.Rsh(biC3, 1)

	c3 := make([]uint64, f.Limbs)
	bigIntToLeUint64(&c3, biC3)

	// Non-square value in field
	biC4 := new(big.Int).SetInt64(7)

	biC5 := new(big.Int).Exp(biC4, biC2, f.BiModulus)
	biC5.Mul(biC5, R)
	biC5.Mod(biC5, f.BiModulus)
	c5 := make([]uint64, f.Limbs)
	bigIntToLeUint64(&c5, biC5)

	biIsSquare := new(big.Int).Rsh(f.BiModulus, 1)
	isSquare := make([]uint64, f.Limbs)
	bigIntToLeUint64(&isSquare, biIsSquare)

	f.SqrtParams = SqrtParams{
		c1, c3, c5, isSquare,
	}

	return f, nil
}

func (f *FieldParams) ToMontgomery(out, arg *[]uint64) {
	f.Mul(out, arg, &f.R2)
}

func (f *FieldParams) FromMontgomery(out, arg *[]uint64) {
	tmp := make([]uint64, f.Limbs*2)
	copy(tmp[:f.Limbs], *arg)
	copy(*out, f.montReduce(tmp))
}

func (f *FieldParams) Add(out, arg1, arg2 *[]uint64) {
	t := make([]uint64, f.Limbs)
	var carry uint64

	for i := 0; i < f.Limbs; i++ {
		t[i], carry = adc((*arg1)[i], (*arg2)[i], carry)
	}

	// Subtract the modulus to ensure the value
	// is smaller.
	f.Sub(out, &t, &f.Modulus)
}

func (f *FieldParams) Sub(out, arg1, arg2 *[]uint64) {
	d := make([]uint64, f.Limbs)
	var borrow, carry uint64

	for i := 0; i < f.Limbs; i++ {
		d[i], borrow = sbb((*arg1)[i], (*arg2)[i], borrow)
	}

	// If underflow occurred on the final limb, borrow 0xff...ff, otherwise
	// borrow = 0x00...00. Conditionally mask to add the modulus
	borrow = -borrow

	for i := 0; i < f.Limbs; i++ {
		d[i], carry = adc(d[i], f.Modulus[i]&borrow, carry)
	}

	copy(*out, d)
}

func (f *FieldParams) Mul(out, arg1, arg2 *[]uint64) {
	var carry uint64
	rr := make([]uint64, f.Limbs*2)

	for i := 0; i < f.Limbs; i++ {
		for j := 0; j < f.Limbs; j++ {
			rr[i+j], carry = mac(rr[i+j], (*arg1)[i], (*arg2)[j], carry)
		}
		rr[i+f.Limbs] = carry
		carry = 0
	}

	copy(*out, f.montReduce(rr))
}

func (f *FieldParams) Square(out, arg *[]uint64) {
	rr := make([]uint64, len(*arg)*2)
	carry := uint64(0)

	for i := 0; i < f.Limbs-1; i++ {
		for j := i + 1; j < f.Limbs; j++ {
			rr[i+j], carry = mac(rr[i+j], (*arg)[i], (*arg)[j], carry)
		}
		rr[i+f.Limbs] = carry
		carry = 0
	}

	rr[f.Limbs*2-1] = rr[f.Limbs*2-2] >> 63
	for i := f.Limbs*2 - 2; i > 0; i-- {
		rr[i] = (rr[i] << 1) | rr[i-1]>>63
	}

	rr[0], carry = mac(0, (*arg)[0], (*arg)[0], 0)
	rr[1], carry = adc(0, rr[1], carry)
	j := 2
	for i := 1; i < f.Limbs; i++ {
		rr[j], carry = mac(rr[j], (*arg)[i], (*arg)[i], carry)
		j++
		rr[j], carry = adc(0, rr[j], carry)
		j++
	}

	copy(*out, f.montReduce(rr))
}

func (f *FieldParams) Neg(out, arg *[]uint64) {
	// Subtract `arg` from `modulus`. Ignore final borrow
	// since it can't underflow.
	var borrow, mask uint64
	t := make([]uint64, f.Limbs)

	for i := 0; i < f.Limbs; i++ {
		t[i], borrow = sbb(f.Modulus[i], (*arg)[i], borrow)
		mask |= (*arg)[i]
	}

	// t could be `modulus` if `arg`=0. Set mask=0 if self=0
	// and 0xff..ff if `arg`!=0
	mask = -((mask | -mask) >> 63)

	for i := 0; i < f.Limbs; i++ {
		(*out)[i] = t[i] & mask
	}
}

func (f *FieldParams) FromBytes(out *[]uint64, arg *[]byte) {
	t := leBytesToLeUint64Array(*arg)
	f.ToMontgomery(out, &t)
}

func (f *FieldParams) ToBytes(out *[]byte, arg *[]uint64) {
	rr := make([]uint64, f.Limbs*2)
	copy(rr[:f.Limbs], *arg)
	t := f.montReduce(rr)
	copy(*out, leUint64toLeBytes(t))
}

func (f *FieldParams) Selectznz(out, arg1, arg2 *[]uint64, choice int) {
	mask := uint64(-choice)
	for i := 0; i < f.Limbs; i++ {
		(*out)[i] = (*arg1)[i] ^ (((*arg1)[i] ^ (*arg2)[i]) & mask)
	}
}

func (f *FieldParams) Invert(wasInverted *int, out, arg *[]uint64) {
	pm2 := new(big.Int).Sub(f.BiModulus, new(big.Int).SetInt64(2))
	exp := make([]uint64, f.Limbs)
	bigIntToLeUint64(&exp, pm2)
	t := make([]uint64, f.Limbs)
	z := make([]uint64, f.Limbs)
	Pow(&t, *arg, exp, f, f)
	ff := new(Field).Init(f, f).SetRaw(*arg)
	*wasInverted = ff.IsNonZeroI()
	f.Selectznz(out, &z, &t, *wasInverted)
}

func (f *FieldParams) Sqrt(wasSquare *int, out, arg *[]uint64) {
	z := make([]uint64, f.Limbs)
	t := make([]uint64, f.Limbs)
	b := make([]uint64, f.Limbs)
	c := make([]uint64, f.Limbs)
	tv := make([]uint64, f.Limbs)

	Pow(&z, *arg, f.SqrtParams.C3, f, f)
	f.Square(&t, &z)
	f.Mul(&t, &t, arg)
	f.Mul(&z, &z, arg)

	copy(b[:], t[:])
	copy(c[:], f.SqrtParams.C5[:])

	for i := f.SqrtParams.C1; i >= 2; i-- {
		for j := 1; j <= i-2; j++ {
			f.Square(&b, &b)
		}

		// if b == 1 flag = 0 else flag = 1
		flag := -EqualI(b, f.R) + 1
		f.Mul(&tv, &z, &c)
		f.Selectznz(&z, &z, &tv, flag)
		f.Square(&c, &c)
		f.Mul(&tv, &t, &c)
		f.Selectznz(&t, &t, &tv, flag)
		copy(b[:], t[:])
	}
	f.Square(&c, &z)
	*wasSquare = EqualI(c, *arg)
	f.Selectznz(out, out, &z, *wasSquare)
}

func (f *FieldParams) montReduce(r []uint64) []uint64 {
	// Taken from Algorithm 14.32 in Handbook of Applied Cryptography
	var carry, carry2, k uint64
	out := make([]uint64, f.Limbs)

	for i := 0; i < f.Limbs; i++ {
		k = r[i] * f.ModulusNegInv
		_, carry = mac(r[i], k, f.Modulus[0], 0)
		for j := 1; j < f.Limbs; j++ {
			r[i+j], carry = mac(r[i+j], k, f.Modulus[j], carry)
		}
		r[i+f.Limbs], carry2 = adc(r[i+f.Limbs], carry2, carry)
	}
	copy(out, r[f.Limbs:])
	f.Sub(&out, &out, &f.Modulus)
	return out
}

// FieldArithmetic are the methods that can be done on a field.
type FieldArithmetic interface {
	// ToMontgomery converts this field to montgomery form
	ToMontgomery(out, arg *[]uint64)
	// FromMontgomery converts this field from montgomery form
	FromMontgomery(out, arg *[]uint64)
	// Neg performs modular negation
	Neg(out, arg *[]uint64)
	// Square performs modular square
	Square(out, arg *[]uint64)
	// Mul performs modular multiplication
	Mul(out, arg1, arg2 *[]uint64)
	// Add performs modular addition
	Add(out, arg1, arg2 *[]uint64)
	// Sub performs modular subtraction
	Sub(out, arg1, arg2 *[]uint64)
	// Sqrt performs modular square root
	Sqrt(wasSquare *int, out, arg *[]uint64)
	// Invert performs modular inverse
	Invert(wasInverted *int, out, arg *[]uint64)
	// FromBytes converts a little endian byte array into a field element
	FromBytes(out *[]uint64, arg *[]byte)
	// ToBytes converts a field element to a little endian byte array
	ToBytes(out *[]byte, arg *[]uint64)
	// Selectznz performs conditional select.
	// selects arg1 if choice == 0 and arg2 if choice == 1
	Selectznz(out, arg1, arg2 *[]uint64, choice int)
}

func (f *Field) Init(params *FieldParams, arithmetic FieldArithmetic) *Field {
	f.Params = params
	f.Arithmetic = arithmetic
	f.Value = make([]uint64, f.Params.Limbs)
	return f
}

// SetZero sets Value to all zeros
func (f *Field) SetZero() *Field {
	for i := range f.Value {
		f.Value[i] = 0
	}
	return f
}

// SetOne sets the field to one
func (f *Field) SetOne() *Field {
	copy(f.Value, f.Params.R)
	return f
}

// SetUint64 set this field to value
func (f *Field) SetUint64(value uint64) *Field {
	f.Value[0] = value
	for i := 1; i < f.Params.Limbs; i++ {
		f.Value[i] = 0
	}
	f.Arithmetic.ToMontgomery(&f.Value, &f.Value)
	return f
}

// SetLimbs converts an array into a field element
// by converting to montgomery form.
func (f *Field) SetLimbs(input []uint64) (*Field, error) {
	if len(input) != f.Params.Limbs {
		return nil, fmt.Errorf("invalid length")
	}
	f.Arithmetic.ToMontgomery(&f.Value, &input)
	return f, nil
}

// SetHexLe attempts to convert a little endian byte representation
// into a `Field`, failing if the input is not canonical
func (f *Field) SetHexLe(input string) (*Field, error) {
	bb, err := hex.DecodeString(input)
	if err != nil {
		return nil, err
	}
	return f.SetBytes(bb)
}

// SetHexBe attempts to convert a big endian byte representation
// into a `Field`, failing if the input is not canonical
func (f *Field) SetHexBe(input string) (*Field, error) {
	bb, err := hex.DecodeString(input)
	if err != nil {
		return nil, err
	}
	return f.SetBytes(ReverseBytes(bb))
}

// SetBytes attempts to convert a little endian byte representation
// into a `Field`, failing if input is not canonical.
func (f *Field) SetBytes(input []byte) (*Field, error) {
	if len(input) != f.Params.Limbs*8 {
		return nil, fmt.Errorf("invalid number of bytes")
	}
	f.Arithmetic.FromBytes(&f.Value, &input)
	return f, nil
}

// SetBytesWide attempts to convert a little endian byte sequence
// into a `Field`
func (f *Field) SetBytesWide(input []byte) (*Field, error) {
	if len(input) != f.Params.Limbs*16 {
		return nil, fmt.Errorf("invalid number of bytes")
	}
	d0 := leBytesToLeUint64Array(input[:len(input)/2])
	d1 := leBytesToLeUint64Array(input[len(input)/2:])
	f.Arithmetic.Mul(&d0, &d0, &f.Params.R2)
	f.Arithmetic.Mul(&d1, &d1, &f.Params.R3)
	f.Arithmetic.Add(&f.Value, &d1, &d0)
	return f, nil
}

// SetBigInt initializes an element from big.Int
// The value is reduced by the modulus.
func (f *Field) SetBigInt(input *big.Int) *Field {
	buffer := make([]byte, f.Params.Limbs*8)
	t := new(big.Int).Set(input)
	t.Mod(t, f.Params.BiModulus)
	t.FillBytes(buffer[:])
	copy(buffer[:], ReverseBytes(buffer[:]))
	f.Arithmetic.FromBytes(&f.Value, &buffer)
	return f
}

// Set copies all from the other field into this one
func (f *Field) Set(a *Field) *Field {
	copy(f.Value, a.Value)
	f.Arithmetic = a.Arithmetic
	f.Params = a.Params
	return f
}

// SetRaw initializes the field value directly to the input
// assumes already in montgomery form
func (f *Field) SetRaw(rhs []uint64) *Field {
	copy(f.Value, rhs)
	return f
}

// IsNonZeroI returns 1 if field is not zero
// 0 otherwise
func (f *Field) IsNonZeroI() int {
	return IsNotZeroArrayI(f.Value)
}

// IsNonZero returns true if field is not zero
// false otherwise
func (f *Field) IsNonZero() bool {
	return f.IsNonZeroI() == 1
}

// IsZeroI returns 1 if this field is zero
// 0 otherwise
func (f *Field) IsZeroI() int {
	return IsZeroArrayI(f.Value)
}

// IsZero returns true if this field is zero
// false otherwise
func (f *Field) IsZero() bool {
	return f.IsZeroI() == 1
}

// IsOneI returns 1 if this field is equal to one
// 0 otherwise
func (f *Field) IsOneI() int {
	return EqualI(f.Value, f.Params.R)
}

// IsOne returns true if this field is equal to one
// false otherwise
func (f *Field) IsOne() bool {
	return f.IsOneI() == 1
}

// IsSquareI returns 1 if f.Value is a square in the field
// By Euler's criterion, this can be calculated in constant time as
//
// if x^((q - 1) / 2) is 0 or 1 in F
func (f *Field) IsSquareI() int {
	tmp := make([]uint64, f.Params.Limbs)
	Pow(&tmp, f.Value, f.Params.SqrtParams.IsSquare, f.Params, f.Arithmetic)
	return EqualI(tmp, f.Params.R) | IsZeroArrayI(tmp)
}

// Cmp returns -1 if lhs < rhs
// -1 if lhs == rhs
// 1 if lhs > rhs
func (f *Field) Cmp(rhs *Field) int {
	return CmpI(f.Value, rhs.Value)
}

// EqualI returns 1 if the two fields are equal
// 0 otherwise
func (f *Field) EqualI(rhs *Field) int {
	return EqualI(f.Value, rhs.Value)
}

// Equal returns true if the two fields are equal
// false otherwise
func (f *Field) Equal(rhs *Field) bool {
	return f.EqualI(rhs) == 1
}

// Sgn0I returns the value of the lowest bit
func (f *Field) Sgn0I() int {
	ff := new(Field).Set(f)
	ff.Arithmetic.FromMontgomery(&ff.Value, &f.Value)
	return int(ff.Value[0] & 1)
}

// Sgn0 returns the value of the lowest bit
func (f *Field) Sgn0() bool {
	ff := new(Field).Set(f)
	ff.Arithmetic.FromMontgomery(&ff.Value, &f.Value)
	return int(ff.Value[0]&1) == 1
}

// Bytes converts this element into a byte representation
// in little endian byte order.
func (f *Field) Bytes() []byte {
	output := make([]byte, f.Params.Limbs*8)
	f.Arithmetic.ToBytes(&output, &f.Value)
	return output
}

// BigInt converts this element into the big.Int struct.
func (f *Field) BigInt() *big.Int {
	buffer := f.Bytes()
	return new(big.Int).SetBytes(ReverseBytes(buffer[:]))
}

// Limbs returns the non-montgomery form of the field value
func (f *Field) Limbs() []uint64 {
	output := make([]uint64, f.Params.Limbs)
	f.Arithmetic.FromMontgomery(&output, &f.Value)
	return output
}

// Double the value of this Field
func (f *Field) Double(a *Field) *Field {
	f.Arithmetic.Add(&f.Value, &a.Value, &a.Value)
	return f
}

// Square this Field
func (f *Field) Square(a *Field) *Field {
	f.Arithmetic.Square(&f.Value, &a.Value)
	return f
}

// SqrtI this element, if it exists. If true, then value
// is a square root. If false, value is a QNR.
func (f *Field) SqrtI(a *Field) (*Field, int) {
	wasSquare := 0
	f.Arithmetic.Sqrt(&wasSquare, &f.Value, &a.Value)
	return f, wasSquare
}

// Sqrt this element, if it exists. If true, then value
// is a square root. If false, value is a QNR.
func (f *Field) Sqrt(a *Field) (*Field, bool) {
	wasSquare := 0
	f.Arithmetic.Sqrt(&wasSquare, &f.Value, &a.Value)
	return f, wasSquare == 1
}

// InvertI this element i.e. compute the multiplicative inverse
// return false, zero if this element is zero.
func (f *Field) InvertI(a *Field) (*Field, int) {
	wasInverted := 0
	f.Arithmetic.Invert(&wasInverted, &f.Value, &a.Value)
	return f, wasInverted
}

// Invert this element i.e. compute the multiplicative inverse
// return false, zero if this element is zero.
func (f *Field) Invert(a *Field) (*Field, bool) {
	wasInverted := 0
	f.Arithmetic.Invert(&wasInverted, &f.Value, &a.Value)
	return f, wasInverted == 1
}

// Mul returns the result from multiplying lhs by rhs.
func (f *Field) Mul(lhs, rhs *Field) *Field {
	f.Arithmetic.Mul(&f.Value, &lhs.Value, &rhs.Value)
	return f
}

// Sub returns the result from subtracting rhs from lhs.
func (f *Field) Sub(lhs, rhs *Field) *Field {
	f.Arithmetic.Sub(&f.Value, &lhs.Value, &rhs.Value)
	return f
}

// Add returns the result from adding rhs to lhs.
func (f *Field) Add(lhs, rhs *Field) *Field {
	f.Arithmetic.Add(&f.Value, &lhs.Value, &rhs.Value)
	return f
}

// Neg returns negation of this Field.
func (f *Field) Neg(input *Field) *Field {
	f.Arithmetic.Neg(&f.Value, &input.Value)
	return f
}

// Exp raises base^exp.
func (f *Field) Exp(base, exp *Field) *Field {
	e := make([]uint64, f.Params.Limbs)
	f.Arithmetic.FromMontgomery(&e, &exp.Value)
	Pow(&f.Value, base.Value, e, f.Params, f.Arithmetic)
	return f
}

// CMove sets f = lhs if choice == 0 and f = rhs if choice == 1.
func (f *Field) CMove(lhs, rhs *Field, choice int) *Field {
	f.Arithmetic.Selectznz(&f.Value, &lhs.Value, &rhs.Value, choice)
	return f
}

// CSwap sets f = other and other = f if choice == 1
func (f *Field) CSwap(other *Field, choice int) *Field {
	mask := uint64(-int64(choice))
	for i := 0; i < f.Params.Limbs; i++ {
		f.Value[i] ^= other.Value[i] & mask
		other.Value[i] ^= f.Value[i] & mask
		f.Value[i] ^= other.Value[i] & mask
	}
	return f
}

// Pow raises base^exp. The result is written to out.
// Public only for convenience for some internal implementations.
func Pow(out *[]uint64, base, exp []uint64, params *FieldParams, arithmetic FieldArithmetic) {
	res := make([]uint64, params.Limbs)
	tmp := make([]uint64, params.Limbs)
	copy(res, params.R)

	for i := params.Limbs - 1; i >= 0; i-- {
		for j := 63; j >= 0; j-- {
			arithmetic.Square(&res, &res)
			arithmetic.Mul(&tmp, &res, &base)
			arithmetic.Selectznz(&res, &res, &tmp, int(exp[i]>>j)&1)
		}
	}
	copy(*out, res)
}

// Pow2k raises arg to the power `2^k`. This result is written to out.
// Public only for convenience for some internal implementations.
func Pow2k(out *[]uint64, arg []uint64, k int, arithmetic FieldArithmetic) {
	t := make([]uint64, len(arg))
	copy(t, arg)
	for i := 0; i < k; i++ {
		arithmetic.Square(&t, &t)
	}

	copy(*out, t)
}

func beBytesToLeUint64Array(beBytes []byte) []uint64 {
	limbs := len(beBytes) / 8
	res := make([]uint64, limbs)

	for i := 0; i < limbs; i++ {
		res[limbs-i-1] = binary.BigEndian.Uint64(beBytes[i*8 : i*8+8])
	}

	return res
}

func leBytesToLeUint64Array(leBytes []byte) []uint64 {
	limbs := len(leBytes) / 8
	res := make([]uint64, limbs)

	for i := 0; i < limbs; i++ {
		res[i] = binary.LittleEndian.Uint64(leBytes[i*8 : i*8+8])
	}

	return res
}

func leUint64toLeBytes(arg []uint64) []byte {
	out := make([]byte, len(arg)*8)
	for i, a := range arg {
		binary.LittleEndian.PutUint64(out[i*8:(i+1)*8], a)
	}
	return out
}

func bigIntToLeUint64(out *[]uint64, bi *big.Int) {
	count := len(*out)
	t := new(big.Int).Set(bi)
	for i := len(*out) - 1; t.Sign() > 0; i-- {
		(*out)[count-i-1] = t.Uint64()
		t.Rsh(t, 64)
	}
}

// mac Multiply and Accumulate - compute a + (b * c) + d, return the result and new carry.
func mac(a, b, c, d uint64) (lo, hi uint64) {
	hi, lo = bits.Mul64(b, c)
	carry2, carry := bits.Add64(a, d, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	lo, carry = bits.Add64(lo, carry2, 0)
	hi, _ = bits.Add64(hi, 0, carry)

	return lo, hi
}

// adc Add w/Carry.
func adc(x, y, carry uint64) (sum, carryOut uint64) {
	sum = x + y + carry
	// The sum will overflow if both top bits are set (x & y) or if one of them
	// is (x | y), and a carry from the lower place happened. If such a carry
	// happens, the top bit will be 1 + 0 + 1 = 0 (&^ sum).
	carryOut = ((x & y) | ((x | y) &^ sum)) >> 63
	carryOut |= ((x & carry) | ((x | carry) &^ sum)) >> 63
	carryOut |= ((y & carry) | ((y | carry) &^ sum)) >> 63
	return sum, carryOut
}

// sbb Subtract with borrow.
func sbb(x, y, borrow uint64) (diff, borrowOut uint64) {
	diff = x - (y + borrow)
	borrowOut = ((^x & y) | (^(x ^ y) & diff)) >> 63
	return diff, borrowOut
}
