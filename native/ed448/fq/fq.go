package fq

import (
	"fmt"
	"github.com/mikelodder7/curvey/internal"
	"github.com/mikelodder7/curvey/native"
	"io"
	"math/big"
)

var (
	fqParams = &internal.FieldParams{
		Modulus: []uint64{
			0x2378c292ab5844f3,
			0x216cc2728dc58f55,
			0xc44edb49aed63690,
			0xffffffff7cca23e9,
			0xffffffffffffffff,
			0xffffffffffffffff,
			0x3fffffffffffffff,
		},
		ModulusNegInv: 0x3bd440fae918bc5,
		R: []uint64{
			0x721cf5b5529eec34,
			0x7a4cf635c8e9c2ab,
			0xeec492d944a725bf,
			0x000000020cd77058,
			0x0000000000000000,
			0x0000000000000000,
			0x0000000000000000,
		},
		R2: []uint64{
			0xe3539257049b9b60,
			0x7af32c4bc1b195d9,
			0x0d66de2388ea1859,
			0xae17cf725ee4d838,
			0x1a9cc14ba3c47c44,
			0x2052bcb7e4d070af,
			0x3402a939f823b729,
		},
		R3: []uint64{
			0x62db79e25f9b74ed,
			0x32d533584f61d636,
			0x3e0d0c8b5fa74964,
			0x178769ed878dfcda,
			0xe4c71af86754b842,
			0xed66e7f42bab736d,
			0x0d30a4f69d3af5f1,
		},
		Limbs: 7,
		BiModulus: new(big.Int).SetBytes([]byte{
			0x3f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7c, 0xca, 0x23, 0xe9, 0xc4, 0x4e, 0xdb, 0x49, 0xae, 0xd6, 0x36, 0x90, 0x21, 0x6c, 0xc2, 0x72, 0x8d, 0xc5, 0x8f, 0x55, 0x23, 0x78, 0xc2, 0x92, 0xab, 0x58, 0x44, 0xf3,
		}),
		SqrtParams: internal.SqrtParams{
			C1: 1,
			C3: []uint64{
				0x48de30a4aad6113c,
				0x085b309ca37163d5,
				0x7113b6d26bb58da4,
				0xffffffffdf3288fa,
				0xffffffffffffffff,
				0xffffffffffffffff,
				0x0fffffffffffffff,
			},
			C5: []uint64{
				0xb15bccdd58b958bf,
				0xa71fcc3cc4dbcca9,
				0xd58a48706a2f10d0,
				0xfffffffd6ff2b390,
				0xffffffffffffffff,
				0xffffffffffffffff,
				0x3fffffffffffffff,
			},
		},
	}
	sqrtExp = []uint64{
		0x48de30a4aad6113d,
		0x085b309ca37163d5,
		0x7113b6d26bb58da4,
		0xffffffffdf3288fa,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0x0fffffffffffffff,
	}
	one4th = &internal.Field{
		Params:     fqParams,
		Arithmetic: fqParams,
		Value: []uint64{
			0xcfc4ab375748160e,
			0x94a6ad0ec8bca90e,
			0xad3ee487a7fad19e,
			0xd563c394eff143eb,
			0x3bc2a3bd7cad3c2a,
			0x141729a49275b4ac,
			0x3656fbc6dba69c9d,
		},
	}
	oneHalf = &internal.Field{
		Params:     fqParams,
		Arithmetic: fqParams,
		Value: []uint64{
			0x2ca820fc26068e5d,
			0xcf326895c48ce9e6,
			0x635992a860009b38,
			0x5a3dcaf93803b1d9,
			0x72dfa400d114a36b,
			0x63429040b4ceb5bc,
			0x346bc81445a1b1ed,
		},
	}
)

type Fq struct {
	Value *internal.Field
}

type ed448FqArithmetic struct{}

func (ed448FqArithmetic) ToMontgomery(out, arg *[]uint64) {
	o := *out
	oo := (*MontgomeryDomainFieldElement)(o)
	a := *arg
	aa := (*NonMontgomeryDomainFieldElement)(a)
	ToMontgomery(oo, aa)
}

func (ed448FqArithmetic) FromMontgomery(out, arg *[]uint64) {
	o := *out
	oo := (*NonMontgomeryDomainFieldElement)(o)
	a := *arg
	aa := (*MontgomeryDomainFieldElement)(a)
	FromMontgomery(oo, aa)
}

func (ed448FqArithmetic) Neg(out, arg *[]uint64) {
	o := *out
	oo := (*MontgomeryDomainFieldElement)(o)
	a := *arg
	aa := (*MontgomeryDomainFieldElement)(a)
	Opp(oo, aa)
}

func (ed448FqArithmetic) Square(out, arg *[]uint64) {
	o := *out
	oo := (*MontgomeryDomainFieldElement)(o)
	a := *arg
	aa := (*MontgomeryDomainFieldElement)(a)
	Square(oo, aa)
}

func (ed448FqArithmetic) Mul(out, arg1, arg2 *[]uint64) {
	o := *out
	oo := (*MontgomeryDomainFieldElement)(o)
	a := *arg1
	aa := (*MontgomeryDomainFieldElement)(a)
	b := *arg2
	bb := (*MontgomeryDomainFieldElement)(b)
	Mul(oo, aa, bb)
}

func (ed448FqArithmetic) Add(out, arg1, arg2 *[]uint64) {
	o := *out
	oo := (*MontgomeryDomainFieldElement)(o)
	a := *arg1
	aa := (*MontgomeryDomainFieldElement)(a)
	b := *arg2
	bb := (*MontgomeryDomainFieldElement)(b)
	Add(oo, aa, bb)
}

func (ed448FqArithmetic) Sub(out, arg1, arg2 *[]uint64) {
	o := *out
	oo := (*MontgomeryDomainFieldElement)(o)
	a := *arg1
	aa := (*MontgomeryDomainFieldElement)(a)
	b := *arg2
	bb := (*MontgomeryDomainFieldElement)(b)
	Sub(oo, aa, bb)
}

func (e *ed448FqArithmetic) Sqrt(wasSquare *int, out, arg *[]uint64) {
	// Shank's method, as p = 3 (mod 4). This means
	// exponentiate by (p+1)/4. This only works for elements
	// that are actually quadratic residue,
	// so check the result at the end.
	var c, z [7]uint64
	cc := c[:]
	zz := z[:]
	internal.Pow(&zz, arg, &sqrtExp, fqParams, e)

	e.Square(&cc, &zz)
	*wasSquare = internal.EqualI(cc, *arg)
	e.Selectznz(out, out, &zz, *wasSquare)
}

func (e *ed448FqArithmetic) Invert(wasInverted *int, out, arg *[]uint64) {
	// Exponentiate by p - 2
	exp := []uint64{
		0x2378c292ab5844f1,
		0x216cc2728dc58f55,
		0xc44edb49aed63690,
		0xffffffff7cca23e9,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0x3fffffffffffffff,
	}
	internal.Pow(out, arg, &exp, fqParams, e)
	*wasInverted = internal.IsNotZeroArrayI(*arg)
	e.Selectznz(out, out, arg, *wasInverted)
}

func (ed448FqArithmetic) FromBytes(out *[]uint64, arg *[]byte) {
	o := *out
	oo := (*[7]uint64)(o)
	a := *arg
	aa := (*[56]byte)(a)
	FromBytes(oo, aa)
}

func (ed448FqArithmetic) ToBytes(out *[]byte, arg *[]uint64) {
	o := *out
	oo := (*[56]byte)(o)
	a := *arg
	aa := (*[7]uint64)(a)
	ToBytes(oo, aa)
}

func (ed448FqArithmetic) Selectznz(out, arg1, arg2 *[]uint64, choice int) {
	o := *out
	oo := (*[7]uint64)(o)
	a := *arg1
	aa := (*[7]uint64)(a)
	b := *arg2
	bb := (*[7]uint64)(b)
	Selectznz(oo, uint1(choice), aa, bb)
}

func FqNew() *Fq {
	value := new(internal.Field).Init(fqParams, &ed448FqArithmetic{})
	return &Fq{value}
}

// IsZero returns 1 if Fp == 0, 0 otherwise.
func (f *Fq) IsZero() int {
	return f.Value.IsZeroI()
}

// IsNonZero returns 1 if Fp != 0, 0 otherwise.
func (f *Fq) IsNonZero() int {
	return f.Value.IsNonZeroI()
}

func (f *Fq) IsOne() int {
	return f.Value.IsOneI()
}

// Cmp returns -1 if f < rhs
// 0 if f == rhs
// 1 if f > rhs.
func (f *Fq) Cmp(rhs *Fq) int {
	return f.Value.Cmp(rhs.Value)
}

// Sgn0I returns the lowest bit value
func (f *Fq) Sgn0I() int {
	return f.Value.Sgn0I()
}

// SetOne f = one
func (f *Fq) SetOne() *Fq {
	f.Value.SetOne()
	return f
}

// SetZero f = one
func (f *Fq) SetZero() *Fq {
	f.Value.SetZero()
	return f
}

// EqualI returns 1 if Fp == rhs, 0 otherwise.
func (f *Fq) EqualI(rhs *Fq) int {
	return f.Value.EqualI(rhs.Value)
}

func (f *Fq) SetUint64(rhs uint64) *Fq {
	f.Value.SetUint64(rhs)
	return f
}

func (f *Fq) Random(reader io.Reader) (*Fq, error) {
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

func (f *Fq) Hash(input []byte) *Fq {
	dst := []byte("edwards448_XOF:SHAKE256_RO_")
	xof := native.ExpandMsgXof(native.EllipticPointHasherShake256(), input, dst, 88)
	var t [114]byte
	copy(t[:], xof[:])
	return f.SetBytesWide(&t)
}

func (f *Fq) Neg(a *Fq) *Fq {
	f.Value.Neg(a.Value)
	return f
}

func (f *Fq) Square(a *Fq) *Fq {
	f.Value.Square(a.Value)
	return f
}

func (f *Fq) Double(a *Fq) *Fq {
	f.Value.Double(a.Value)
	return f
}

func (f *Fq) Mul(arg1, arg2 *Fq) *Fq {
	f.Value.Mul(arg1.Value, arg2.Value)
	return f
}

func (f *Fq) Add(arg1, arg2 *Fq) *Fq {
	f.Value.Add(arg1.Value, arg2.Value)
	return f
}

// Sub performs modular subtraction.
func (f *Fq) Sub(arg1, arg2 *Fq) *Fq {
	f.Value.Sub(arg1.Value, arg2.Value)
	return f
}

// Sqrt performs modular square root.
func (f *Fq) Sqrt(a *Fq) (*Fq, int) {
	_, wasSquare := f.Value.SqrtI(a.Value)
	return f, wasSquare
}

// Invert performs modular inverse.
func (f *Fq) Invert(a *Fq) (*Fq, int) {
	_, wasInverted := f.Value.InvertI(a.Value)
	return f, wasInverted
}

func (f *Fq) Halve(a *Fq) *Fq {
	f.Value.Mul(a.Value, oneHalf)
	return f
}

func (f *Fq) Div4(a *Fq) *Fq {
	f.Value.Mul(a.Value, one4th)
	return f
}

func (f *Fq) Mod4(a *Fq) *Fq {
	copy(f.Value.Value, a.Value.Limbs())
	f.Value.Value[0] &= 3
	for i := 1; i < f.Value.Params.Limbs; i++ {
		f.Value.Value[i] = 0
	}
	_, _ = f.Value.SetLimbs(f.Value.Value)
	return f
}

func (f *Fq) ToRadix16() []byte {
	bytes := f.Value.Bytes()
	output := make([]byte, 113)

	// radix-16
	for i := 0; i < 56; i++ {
		output[2*i] = bytes[i] & 0xf
		output[2*i+1] = (bytes[i] >> 4) & 0xf
	}
	// re-center co-efficients to be between [-8, 8)
	for i := 0; i < 112; i++ {
		carry := (output[i] + 8) >> 4
		output[i] -= carry << 4
		output[i+1] += carry
	}
	return output
}

// SetBytes converts a little endian byte array into a field element
// return 0 if the bytes are not in the field, 1 if they are.
func (f *Fq) SetBytes(arg *[57]byte) (*Fq, error) {
	_, err := f.Value.SetBytes(arg[:56])

	// Check that the 10 high bits are not set
	if err != nil || arg[56] != 0 || (arg[55]>>6 != 0) {
		return nil, err
	}
	return f, nil
}

// SetBytesWide takes 112 bytes as input and treats them as a 896-bit number.
func (f *Fq) SetBytesWide(a *[114]byte) *Fq {
	var err error
	lo := new(internal.Field).Init(fqParams, fqParams)
	hi := new(internal.Field).Init(fqParams, fqParams)
	top := new(internal.Field).Init(fqParams, fqParams)

	_, err = lo.SetBytes(a[:56])
	if err != nil {
		return nil
	}
	_, err = hi.SetBytes(a[56:112])
	if err != nil {
		return nil
	}
	var t [56]byte
	t[0] = a[112]
	t[1] = a[113]
	_, err = top.SetBytes(t[:])
	if err != nil {
		return nil
	}
	rr := new(internal.Field).Init(fqParams, fqParams)
	rr2 := new(internal.Field).Init(fqParams, fqParams)

	rr.SetOne()
	copy(rr2.Value, fqParams.R2)

	// lo * R / R == lo mod q
	lo.Mul(lo, rr)
	// hi * R2 / R == hi * R mod q
	hi.Mul(hi, rr2)
	// top * R4 / R^2 == top * R2 mod q
	top.Mul(top, rr2)
	top.Mul(top, rr2)

	// lo + hi * R + top * R2 mod q
	f.Value.Add(lo, hi)
	f.Value.Add(f.Value, top)

	return f
}

// SetBigInt initializes an element from big.Int
// The value is reduced by the modulus.
func (f *Fq) SetBigInt(bi *big.Int) *Fq {
	f.Value.SetBigInt(bi)
	return f
}

// Set copies a into Fp.
func (f *Fq) Set(a *Fq) *Fq {
	f.Value.Set(a.Value)
	return f
}

// SetLimbs converts an array into a field element
// by converting to montgomery form.
func (f *Fq) SetLimbs(a *[7]uint64) *Fq {
	_, err := f.Value.SetLimbs(a[:])
	if err != nil {
		return nil
	}
	return f
}

// SetRaw converts a raw array into a field element
// Assumes input is already in montgomery form.
func (f *Fq) SetRaw(a *[7]uint64) *Fq {
	f.Value.SetRaw(a[:])
	return f
}

// Bytes converts a field element to a little endian byte array.
func (f *Fq) Bytes() [57]byte {
	var out [57]byte
	copy(out[:], f.Value.Bytes())
	return out
}

// BigInt converts this element into the big.Int struct.
func (f *Fq) BigInt() *big.Int {
	buffer := f.Bytes()
	return new(big.Int).SetBytes(internal.ReverseBytes(buffer[:]))
}

// Raw converts this element into the a []uint64.
func (f *Fq) Raw() []uint64 {
	var t []uint64
	f.Value.Arithmetic.FromMontgomery(&t, &f.Value.Value)
	return t
}

// CMove performs conditional select.
// selects arg1 if choice == 0 and arg2 if choice == 1.
func (f *Fq) CMove(arg1, arg2 *Fq, choice int) *Fq {
	f.Value.CMove(arg1.Value, arg2.Value, choice)
	return f
}

// CNeg conditionally negates a if choice == 1.
func (f *Fq) CNeg(a *Fq, choice int) *Fq {
	var t Fq
	t.Neg(a)
	return f.CMove(f, &t, choice)
}

// Exp raises base^exp.
func (f *Fq) Exp(base, exp *Fq) *Fq {
	f.Value.Exp(base.Value, exp.Value)
	return f
}
