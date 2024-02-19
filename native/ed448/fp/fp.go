package fp

import (
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"github.com/mikelodder7/curvey/internal"
	"github.com/mikelodder7/curvey/native"
	"io"
	"math/big"
)

const FpLimbs = 8

var (
	Zero     = &Fp{Value: []uint64{0, 0, 0, 0, 0, 0, 0, 0}}
	One      = &Fp{Value: []uint64{1, 0, 0, 0, 0, 0, 0, 0}}
	MinusOne = &Fp{
		Value: []uint64{
			0x1fffffffffffffd,
			0x1fffffffffffffe,
			0x1fffffffffffffe,
			0x1fffffffffffffe,
			0x1fffffffffffffc,
			0x1fffffffffffffe,
			0x1fffffffffffffe,
			0x1fffffffffffffe,
		},
	}
	Neg4TwistedD = []uint64{
		0x000000000262a7,
		0x00000000000000,
		0x00000000000000,
		0x00000000000000,
		0xffffffffffffff,
		0xffffffffffffff,
		0xffffffffffffff,
		0xffffffffffffff,
	}
	EdwardsD = &Fp{Value: []uint64{
		144115188075816789,
		144115188075855870,
		144115188075855870,
		144115188075855870,
		144115188075855868,
		144115188075855870,
		144115188075855870,
		144115188075855870,
	}}
	NegEdwardsD = []uint64{39081, 0, 0, 0, 0, 0, 0, 0}
	TwistedD    = &Fp{Value: []uint64{
		144115188075816788,
		144115188075855870,
		144115188075855870,
		144115188075855870,
		144115188075855868,
		144115188075855870,
		144115188075855870,
		144115188075855870,
	}}
	TwoXTwistedD = &Fp{Value: []uint64{
		144115188075777706,
		144115188075855870,
		144115188075855870,
		144115188075855870,
		144115188075855868,
		144115188075855870,
		144115188075855870,
		144115188075855870,
	}}
	DecafFactor = []uint64{
		0x42ef0f45572736,
		0x7bf6aa20ce5296,
		0xf4fd6eded26033,
		0x968c14ba839a66,
		0xb8d54b64a2d780,
		0x6aa0a1f1a7b8a5,
		0x683bf68d722fa2,
		0x22d962fbeb24f7,
	}
	Ap2Div4              = &Fp{Value: []uint64{9082, 0, 0, 0, 0, 0, 0, 0}}
	GoldilocksBasePointX = []uint64{
		10880955091566686,
		36276784145337894,
		69571282115576635,
		46113124210880026,
		4247859732800292,
		15440021224255559,
		66747077793030847,
		22264495316135181,
	}
	GoldilocksBasePointY = []uint64{
		2385235625966100,
		5396741696826776,
		8134720567442877,
		1584133578609663,
		46047824121994270,
		56121598560924524,
		10283140089599689,
		29624444337960636,
	}
	GoldilocksBasePointT = []uint64{
		1796939199780339,
		45174008172060139,
		40732174862907279,
		63672088496536030,
		37244660935497319,
		41035719659624511,
		30626637035688077,
		56117654178374172,
	}

	TwistedBasePointX = []uint64{
		0,
		72057594037927936,
		72057594037927935,
		36028797018963967,
		72057594037927934,
		72057594037927935,
		72057594037927935,
		36028797018963967,
	}
	TwistedBasePointY = []uint64{
		27155415521118820,
		3410937204744648,
		19376965222209947,
		22594032279754776,
		21520481577673772,
		10141917371396176,
		59827755213158602,
		37445921829569158,
	}

	TwistedBasePointZ = []uint64{1, 0, 0, 0, 0, 0, 0, 0}
	TwistedBasePointT = []uint64{
		64114820220813573,
		27592348249940115,
		21918321435874307,
		45908688348236165,
		34141937727972228,
		63575698147485199,
		22766751209138687,
		30740600843388580,
	}
	SqrtExp, _ = FpNew().SetCanonicalBytes(&[56]byte{
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0xc0,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0x3f,
	})
	InvertExp, _ = FpNew().SetCanonicalBytes(&[56]byte{
		0xfd,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xfe,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
	})
	IsSquareExp, _ = FpNew().SetCanonicalBytes(&[56]byte{
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0x7f,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0x7f,
	})
	IsTorsionMul, _ = FpNew().SetCanonicalBytes(&[56]byte{
		0x3c,
		0x11,
		0xd6,
		0xaa,
		0xa4,
		0x30,
		0xde,
		0x48,
		0xd5,
		0x63,
		0x71,
		0xa3,
		0x9c,
		0x30,
		0x5b,
		0x08,
		0xa4,
		0x8d,
		0xb5,
		0x6b,
		0xd2,
		0xb6,
		0x13,
		0x71,
		0xfa,
		0x88,
		0x32,
		0xdf,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0xff,
		0x0f,
	})
)

type Fp struct {
	Value []uint64
}

var (
	BiModulus = new(big.Int).SetBytes([]byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	})
	//fpParams = internal.FieldParams{
	//	Modulus: []uint64{
	//		0xffffffffffffffff,
	//		0xffffffffffffffff,
	//		0xffffffffffffffff,
	//		0xfffffffeffffffff,
	//		0xffffffffffffffff,
	//		0xffffffffffffffff,
	//		0xffffffffffffffff,
	//	},
	//	ModulusNegInv: 1,
	//	// 2^448 mod p.
	//	R: []uint64{1, 0, 0, 0x100000000, 0, 0, 0},
	//	// 2^896 mod p.
	//	R2: []uint64{2, 0, 0, 0x300000000, 0, 0, 0},
	//	// 2^1344 mod p.
	//	R3:    []uint64{5, 0, 0, 0x800000000, 0, 0, 0},
	//	Limbs: 7,
	//	BiModulus: new(big.Int).SetBytes([]byte{
	//		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	//	}),
	//	SqrtParams: internal.SqrtParams{
	//		C1: 1,
	//		C3: []uint64{
	//			0xffffffffffffffff,
	//			0xffffffffffffffff,
	//			0xffffffffffffffff,
	//			0xffffffffbfffffff,
	//			0xffffffffffffffff,
	//			0xffffffffffffffff,
	//			0x3fffffffffffffff,
	//		},
	//		C5: []uint64{
	//			0xfffffffffffffffe,
	//			0xffffffffffffffff,
	//			0xffffffffffffffff,
	//			0xfffffffdffffffff,
	//			0xffffffffffffffff,
	//			0xffffffffffffffff,
	//			0xffffffffffffffff,
	//		},
	//		IsSquare: []uint64{
	//			0xffffffffffffffff,
	//			0xffffffffffffffff,
	//			0xffffffffffffffff,
	//			0xffffffff7fffffff,
	//			0xffffffffffffffff,
	//			0xffffffffffffffff,
	//			0x7fffffffffffffff,
	//		},
	//	},
	//}

	//sqrtExp = []uint64{
	//	0x0000000000000000,
	//	0x0000000000000000,
	//	0x0000000000000000,
	//	0xffffffffc0000000,
	//	0xffffffffffffffff,
	//	0xffffffffffffffff,
	//	0x3fffffffffffffff,
	//}
)

func FpNew() *Fp {
	return &Fp{
		Value: make([]uint64, FpLimbs),
	}
}

// IsZero returns 1 if Fp == 0, 0 otherwise.
func (f *Fp) IsZero() int {
	var a, b [56]byte
	ff := (*TightFieldElement)(f.Value)
	ToBytes(&b, ff)
	return subtle.ConstantTimeCompare(a[:], b[:])
}

// IsNonZero returns 1 if Fp != 0, 0 otherwise.
func (f *Fp) IsNonZero() int {
	return internal.IsNotZeroArrayI(f.Value)
}

// IsOne returns 1 if Fp == 1, 0 otherwise
func (f *Fp) IsOne() int {
	var a, b [56]byte
	a[0] = 1
	ff := (*TightFieldElement)(f.Value)
	ToBytes(&b, ff)
	return subtle.ConstantTimeCompare(a[:], b[:])
}

// IsSquare returns 1 if Fp is a quadratic residue, 0 otherwise
func (f *Fp) IsSquare() int {
	var a [56]byte
	z := FpNew().Exp(f, IsSquareExp)
	c := z.Bytes()
	return subtle.ConstantTimeCompare(a[1:], c[1:]) & (subtle.ConstantTimeByteEq(0, c[0]) | subtle.ConstantTimeByteEq(1, c[0]))
}

// Cmp returns -1 if f < rhs
// 0 if f == rhs
// 1 if f > rhs.
func (f *Fp) Cmp(rhs *Fp) int {
	a := f.Bytes()
	b := rhs.Bytes()
	return internal.CmpBytesI(a[:], b[:])
}

// Sgn0I returns the lowest bit value
func (f *Fp) Sgn0I() int {
	return int(f.Bytes()[0] & 1)
}

// SetOne f = one
func (f *Fp) SetOne() *Fp {
	copy(f.Value, One.Value)
	return f
}

// SetZero f = one
func (f *Fp) SetZero() *Fp {
	for i := 0; i < FpLimbs; i++ {
		f.Value[i] ^= f.Value[i]
	}
	return f
}

// EqualI returns 1 if Fp == rhs, 0 otherwise.
func (f *Fp) EqualI(rhs *Fp) int {
	a := f.Bytes()
	b := rhs.Bytes()
	return subtle.ConstantTimeCompare(a[:], b[:])
}

func (f *Fp) SetUint64(rhs uint64) *Fp {
	var b [56]byte
	binary.LittleEndian.PutUint64(b[:8], rhs)
	_, _ = f.SetCanonicalBytes(&b)

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

func (f *Fp) Neg(arg *Fp) *Fp {
	o := (*TightFieldElement)(f.Value)
	a := (*TightFieldElement)(arg.Value)

	CarryOpp(o, a)

	return f
}

func (f *Fp) Square(arg *Fp) *Fp {
	o := (*TightFieldElement)(f.Value)
	a := (*LooseFieldElement)(arg.Value)

	CarrySquare(o, a)

	return f
}

func (f *Fp) Double(a *Fp) *Fp {
	f.Add(a, a)
	return f
}

func (f *Fp) Mul(arg1, arg2 *Fp) *Fp {
	o := (*TightFieldElement)(f.Value)
	a := (*LooseFieldElement)(arg1.Value)
	b := (*LooseFieldElement)(arg2.Value)

	CarryMul(o, a, b)

	return f
}

func (f *Fp) Add(arg1, arg2 *Fp) *Fp {
	o := (*TightFieldElement)(f.Value)
	a := (*TightFieldElement)(arg1.Value)
	b := (*TightFieldElement)(arg2.Value)

	CarryAdd(o, a, b)

	return f
}

// Sub performs modular subtraction.
func (f *Fp) Sub(arg1, arg2 *Fp) *Fp {
	o := (*TightFieldElement)(f.Value)
	a := (*TightFieldElement)(arg1.Value)
	b := (*TightFieldElement)(arg2.Value)

	CarrySub(o, a, b)

	return f
}

// Sqrt performs modular square root.
func (f *Fp) Sqrt(arg *Fp) (*Fp, int) {
	// Shank's method, as p = 3 (mod 4). This means
	// exponentiate by (p+1)/4. This only works for elements
	// that are actually quadratic residue,
	// so check the result at the end.
	var wasSquare int
	zz := FpNew().Exp(arg, SqrtExp)

	cc := FpNew().Square(zz)
	wasSquare = cc.EqualI(arg)

	f.CMove(FpNew(), zz, wasSquare)
	return f, wasSquare
}

// Invert performs modular inverse.
func (f *Fp) Invert(a *Fp) (*Fp, int) {
	wasInverted := a.IsNonZero()

	t10 := FpNew().Square(a)
	t11 := FpNew().Mul(a, t10)
	t110 := FpNew().Square(t11)
	t111 := FpNew().Mul(a, t110)
	t111000 := FpNew().SquareN(t111, 3)
	t111111 := FpNew().Mul(t111, t111000)

	x12 := FpNew().SquareN(t111111, 6)
	x12.Mul(x12, t111111)

	x24 := FpNew().SquareN(x12, 12)
	x24.Mul(x24, x12)

	i34 := FpNew().SquareN(x24, 6)

	x30 := FpNew().Mul(i34, t111111)

	x48 := FpNew().SquareN(i34, 18)
	x48.Mul(x48, x24)

	x96 := FpNew().SquareN(x48, 48)
	x96.Mul(x96, x48)

	x192 := FpNew().SquareN(x96, 96)
	x192.Mul(x192, x96)

	x222 := FpNew().SquareN(x192, 30)
	x222.Mul(x222, x30)

	x223 := FpNew().Square(x222)
	x223.Mul(x223, a)

	z := FpNew().SquareN(x223, 223)
	z.Mul(z, x222)
	z.SquareN(z, 2)
	z.Mul(z, a)

	f.CMove(FpNew(), z, wasInverted)
	return f, wasInverted
}

// SetCanonicalBytes converts a little endian byte array into a field element
// returns nil if the bytes are not in the field
func (f *Fp) SetCanonicalBytes(arg *[56]byte) (*Fp, error) {
	i := (*TightFieldElement)(f.Value)
	FromBytes(i, arg)
	return f, nil
}

// SetBytes converts a little endian byte array into a field element
// returns nil if the bytes are not in the field or the length is invalid
func (f *Fp) SetBytes(arg []byte) (*Fp, error) {
	if len(arg) != 56 {
		return nil, fmt.Errorf("invalid length")
	}
	a := [56]byte(arg)
	return f.SetCanonicalBytes(&a)
}

// SetBytesWide takes 112 bytes as input and treats them as a 896-bit number.
func (f *Fp) SetBytesWide(a *[112]byte) *Fp {
	aa := *a
	bi := new(big.Int).SetBytes(internal.ReverseBytes(aa[:]))
	bi.Mod(bi, BiModulus)

	return f.SetBigInt(bi)
}

// SetBigInt initializes an element from big.Int
// The value is reduced by the modulus.
func (f *Fp) SetBigInt(bi *big.Int) *Fp {
	var b [56]byte
	bi.FillBytes(b[:])
	copy(b[:], internal.ReverseBytes(b[:]))
	i := (*TightFieldElement)(f.Value)
	FromBytes(i, &b)
	return f
}

// Set copies a into Fp.
func (f *Fp) Set(a *Fp) *Fp {
	copy(f.Value, a.Value)
	return f
}

// SetLimbs converts an array into a field element
// by converting to montgomery form.
func (f *Fp) SetLimbs(a *[]uint64) *Fp {
	var b [56]byte
	l := len(*a)
	if l != 7 {
		return nil
	}
	for i := 0; i < l; i++ {
		binary.LittleEndian.PutUint64(b[i*8:(i+1)*8], (*a)[i])
	}
	i := (*TightFieldElement)(f.Value)
	FromBytes(i, &b)
	return f
}

// SetRaw converts a raw array into a field element
// Assumes input is already in montgomery form.
func (f *Fp) SetRaw(a *[]uint64) *Fp {
	copy(f.Value, *a)
	return f
}

// Bytes converts a field element to a little endian byte array.
func (f *Fp) Bytes() [56]byte {
	var out [56]byte
	i := (*TightFieldElement)(f.Value)
	ToBytes(&out, i)
	return out
}

// BigInt converts this element into the big.Int struct.
func (f *Fp) BigInt() *big.Int {
	buffer := f.Bytes()
	return new(big.Int).SetBytes(internal.ReverseBytes(buffer[:]))
}

// CMove performs conditional select.
// selects arg1 if choice == 0 and arg2 if choice == 1.
func (f *Fp) CMove(arg1, arg2 *Fp, choice int) *Fp {
	o := (*[FpLimbs]uint64)(f.Value)
	a := (*[FpLimbs]uint64)(arg1.Value)
	b := (*[FpLimbs]uint64)(arg2.Value)
	Selectznz(o, uint1(choice), a, b)
	return f
}

// CNeg conditionally negates a if choice == 1.
func (f *Fp) CNeg(a *Fp, choice int) *Fp {
	t := FpNew().Neg(a)
	return f.CMove(f, t, choice)
}

// CSwap conditionally swaps this with a if choice == 1
// a is f and f is a if choice == 1
func (f *Fp) CSwap(a *Fp, choice int) *Fp {
	mask := uint64(-int64(choice))
	for i := 0; i < FpLimbs; i++ {
		f.Value[i] ^= a.Value[i] & mask
		a.Value[i] ^= f.Value[i] & mask
		f.Value[i] ^= a.Value[i] & mask
	}
	return f
}

// Exp raises base^exp.
func (f *Fp) Exp(base, exp *Fp) *Fp {
	res := make([]uint64, FpLimbs)
	tmp := make([]uint64, FpLimbs)
	copy(res, One.Value)

	bb := exp.Bytes()

	b := (*LooseFieldElement)(base.Value)
	t := (*TightFieldElement)(res)
	r := (*LooseFieldElement)(res)
	tt := (*TightFieldElement)(tmp)
	rr := (*[8]uint64)(res)
	tm := (*[8]uint64)(tmp)
	for i := 56 - 1; i >= 0; i-- {
		for j := 7; j >= 0; j-- {
			CarrySquare(t, r)
			CarryMul(tt, r, b)
			Selectznz(rr, uint1(bb[i]>>j&1), rr, tm)
		}
	}
	copy(f.Value, res)
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
