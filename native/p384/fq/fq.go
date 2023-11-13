//
// SPDX-License-Identifier: Apache-2.0
//

package fq

import (
	"math/big"
	"sync"

	"github.com/mikelodder7/curvey/native"
)

var (
	p384FqInitonce sync.Once
	p384FqParams   native.Field6Params
)

func P384FqNew() *native.Field6 {
	return &native.Field6{
		Value:      [native.Field6Limbs]uint64{},
		Params:     getP384FqParams(),
		Arithmetic: p384FqArithmetic{},
	}
}

func p384FqParamsInit() {
	// See FIPS 186-3, section D.2.3
	p384FqParams = native.Field6Params{
		R: [native.Field6Limbs]uint64{
			0x1313e695333ad68d,
			0xa7e5f24db74f5885,
			0x389cb27e0bc8d220,
			0x0000000000000000,
			0x0000000000000000,
			0x0000000000000000,
		},
		R2: [native.Field6Limbs]uint64{
			0x2d319b2419b409a9,
			0xff3d81e5df1aa419,
			0xbc3e483afcb82947,
			0xd40d49174aab1cc5,
			0x3fb05b7a28266895,
			0x0c84ee012b39bf21,
		},
		R3: [native.Field6Limbs]uint64{
			0x302a6faf377c7677,
			0x2a70cb61d26894bc,
			0x0c27ddb8ba8dc4ba,
			0x5dbd3f41edb48eb6,
			0x16d081679522617b,
			0xd558bfbcb33c33c6,
		},
		Modulus: [native.Field6Limbs]uint64{
			0xecec196accc52973,
			0x581a0db248b0a77a,
			0xc7634d81f4372ddf,
			0xffffffffffffffff,
			0xffffffffffffffff,
			0xffffffffffffffff,
		},
		BiModulus: new(big.Int).SetBytes([]byte{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xc7, 0x63, 0x4d, 0x81, 0xf4, 0x37, 0x2d, 0xdf, 0x58, 0x1a, 0x0d, 0xb2, 0x48, 0xb0, 0xa7, 0x7a, 0xec, 0xec, 0x19, 0x6a, 0xcc, 0xc5, 0x29, 0x73,
		}),
	}
}

func getP384FqParams() *native.Field6Params {
	p384FqInitonce.Do(p384FqParamsInit)
	return &p384FqParams
}

type p384FqArithmetic struct{}

// ToMontgomery converts this field to montgomery form.
func (p384FqArithmetic) ToMontgomery(out, arg *[native.Field6Limbs]uint64) {
	ToMontgomery((*MontgomeryDomainFieldElement)(out), (*NonMontgomeryDomainFieldElement)(arg))
}

// FromMontgomery converts this field from montgomery form.
func (p384FqArithmetic) FromMontgomery(out, arg *[native.Field6Limbs]uint64) {
	FromMontgomery((*NonMontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg))
}

// Neg performs modular negation.
func (p384FqArithmetic) Neg(out, arg *[native.Field6Limbs]uint64) {
	Opp((*MontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg))
}

// Square performs modular square.
func (p384FqArithmetic) Square(out, arg *[native.Field6Limbs]uint64) {
	Square((*MontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg))
}

// Mul performs modular multiplication.
func (p384FqArithmetic) Mul(out, arg1, arg2 *[native.Field6Limbs]uint64) {
	Mul((*MontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg1), (*MontgomeryDomainFieldElement)(arg2))
}

// Add performs modular addition.
func (p384FqArithmetic) Add(out, arg1, arg2 *[native.Field6Limbs]uint64) {
	Add((*MontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg1), (*MontgomeryDomainFieldElement)(arg2))
}

// Sub performs modular subtraction.
func (p384FqArithmetic) Sub(out, arg1, arg2 *[native.Field6Limbs]uint64) {
	Sub((*MontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg1), (*MontgomeryDomainFieldElement)(arg2))
}

// Sqrt performs modular square root.
func (p p384FqArithmetic) Sqrt(wasSquare *int, out, arg *[native.Field6Limbs]uint64) {
	// p mod 4 = 3 -> compute sqrt(x) using x^((p+1)/4) =
	// x^9850501549098619803069760025035903451269934817616361666986726319906914849778315892349739077038073728388608413485661
	var t1, t10, t11, t101, t111, t1001 [native.Field6Limbs]uint64
	var t1011, t1101, t1111, t11110 [native.Field6Limbs]uint64
	var t11111, t1111100, t11111000 [native.Field6Limbs]uint64
	var i14, i20, i31, i58, i110, i286, i308 [native.Field6Limbs]uint64
	var x194, i225, i235, i258, i269, i323, i340, i357 [native.Field6Limbs]uint64
	var i369, i387, i397, i413, i427, t, tt, x [native.Field6Limbs]uint64

	copy(t1[:], arg[:])
	p.Square(&t10, &t1)
	p.Mul(&t11, arg, &t10)
	p.Mul(&t101, &t10, &t11)
	p.Mul(&t111, &t10, &t101)
	p.Mul(&t1001, &t10, &t111)
	p.Mul(&t1011, &t10, &t1001)
	p.Mul(&t1101, &t10, &t1011)
	p.Mul(&t1111, &t10, &t1101)
	p.Square(&t11110, &t1111)
	p.Mul(&t11111, &t1, &t11110)
	native.Pow2k6(&t1111100, &t11111, 2, p)
	p.Square(&t11111000, &t1111100)
	p.Square(&i14, &t11111000)

	native.Pow2k6(&t, &i14, 5, p)
	p.Mul(&i20, &t, &i14)

	native.Pow2k6(&t, &i20, 10, p)
	p.Mul(&i31, &t, &i20)

	native.Pow2k6(&t, &i31, 4, p)
	p.Mul(&tt, &t, &t11111000)
	native.Pow2k6(&t, &tt, 21, p)
	p.Mul(&i58, &t, &i31)

	native.Pow2k6(&t, &i58, 3, p)
	p.Mul(&tt, &t, &t1111100)
	native.Pow2k6(&t, &tt, 47, p)
	p.Mul(&i110, &t, &i58)

	native.Pow2k6(&t, &i110, 95, p)
	p.Mul(&tt, &t, &i110)
	p.Mul(&x194, &tt, &t1111)

	native.Pow2k6(&t, &x194, 6, p)
	p.Mul(&tt, &t, &t111)
	native.Pow2k6(&t, &tt, 3, p)
	p.Mul(&tt, &t, &t11)
	native.Pow2k6(&i225, &tt, 7, p)

	p.Mul(&t, &t1101, &i225)
	native.Pow2k6(&tt, &t, 6, p)
	p.Square(&t, &tt)
	p.Mul(&i235, &t, &t1)

	native.Pow2k6(&t, &i235, 11, p)
	p.Mul(&tt, &t, &t11111)
	native.Pow2k6(&t, &tt, 2, p)
	p.Mul(&tt, &t, &t1)
	native.Pow2k6(&i258, &tt, 8, p)

	p.Mul(&t, &t1101, &i258)
	native.Pow2k6(&tt, &t, 2, p)
	p.Mul(&t, &tt, &t11)
	native.Pow2k6(&tt, &t, 6, p)
	p.Mul(&i269, &tt, &t1011)

	native.Pow2k6(&t, &i269, 4, p)
	p.Mul(&tt, &t, &t111)
	native.Pow2k6(&t, &tt, 6, p)
	p.Mul(&tt, &t, &t11111)
	native.Pow2k6(&i286, &tt, 5, p)

	p.Mul(&t, &t1011, &i286)
	native.Pow2k6(&tt, &t, 10, p)
	p.Mul(&t, &tt, &t1101)
	native.Pow2k6(&tt, &t, 9, p)
	p.Mul(&i308, &tt, &t1101)

	native.Pow2k6(&t, &i308, 4, p)
	p.Mul(&tt, &t, &t1011)
	native.Pow2k6(&t, &tt, 6, p)
	p.Mul(&tt, &t, &t1001)
	native.Pow2k6(&i323, &tt, 3, p)

	p.Mul(&t, &t1, &i323)
	native.Pow2k6(&tt, &t, 7, p)
	p.Mul(&t, &tt, &t1011)
	native.Pow2k6(&tt, &t, 7, p)
	p.Mul(&i340, &tt, &t101)

	native.Pow2k6(&t, &i340, 5, p)
	p.Mul(&tt, &t, &t111)
	native.Pow2k6(&t, &tt, 5, p)
	p.Mul(&tt, &t, &t1111)
	native.Pow2k6(&i357, &tt, 5, p)

	p.Mul(&t, &t1011, &i357)
	native.Pow2k6(&tt, &t, 4, p)
	p.Mul(&t, &tt, &t1011)
	native.Pow2k6(&tt, &t, 5, p)
	p.Mul(&i369, &tt, &t111)

	native.Pow2k6(&t, &i369, 3, p)
	p.Mul(&tt, &t, &t11)
	native.Pow2k6(&t, &tt, 7, p)
	p.Mul(&tt, &t, &t11)
	native.Pow2k6(&i387, &tt, 6, p)

	p.Mul(&t, &t1011, &i387)
	native.Pow2k6(&tt, &t, 4, p)
	p.Mul(&t, &tt, &t101)
	native.Pow2k6(&tt, &t, 3, p)
	p.Mul(&i397, &tt, &t11)

	native.Pow2k6(&t, &i397, 4, p)
	p.Mul(&tt, &t, &t11)
	native.Pow2k6(&t, &tt, 4, p)
	p.Mul(&tt, &t, &t11)
	native.Pow2k6(&i413, &tt, 6, p)

	p.Mul(&t, &t101, &i413)
	native.Pow2k6(&tt, &t, 5, p)
	p.Mul(&t, &tt, &t101)
	native.Pow2k6(&tt, &t, 6, p)
	p.Mul(&i427, &tt, &t1011)

	native.Pow2k6(&t, &i427, 3, p)
	p.Mul(&x, &t, &t101)

	Square((*MontgomeryDomainFieldElement)(&t), (*MontgomeryDomainFieldElement)(&x))
	*wasSquare = (&native.Field6{Value: t1, Params: getP384FqParams(), Arithmetic: p}).Equal(&native.Field6{
		Value: x, Params: getP384FqParams(), Arithmetic: p,
	})
	Selectznz(out, uint1(*wasSquare), out, &x)
}

// Invert performs modular inverse.
func (p p384FqArithmetic) Invert(wasInverted *int, out, arg *[native.Field6Limbs]uint64) {
	// Implement bernstein yang invert method 2019 p.366
	const ITERATIONS = (49*384 + 57) / 17
	var a, v, r, out4, out5 [native.Field6Limbs]uint64
	var f, g, out2, out3 [native.Field6Limbs + 1]uint64
	var out1 uint64
	p.FromMontgomery(&a, arg)
	SetOne((*MontgomeryDomainFieldElement)(&r))
	Msat(&f)
	d := uint64(1)
	copy(g[:native.Field6Limbs], a[:])

	for i := 0; i < ITERATIONS-ITERATIONS%2; i += 2 {
		Divstep(&out1, &out2, &out3, &out4, &out5, d, &f, &g, &v, &r)
		Divstep(&d, &f, &g, &v, &r, out1, &out2, &out3, &out4, &out5)
	}

	Divstep(&out1, &f, &out3, &v, &out5, d, &f, &g, &v, &r)

	s := (f[len(f)-1] >> 63) & 1
	p.Neg(&a, &v)
	Selectznz(&v, uint1(s), &v, &a)
	DivstepPrecomp(&a)
	p.Mul(&v, &v, &a)
	*wasInverted = (&native.Field6{
		Value:      *arg,
		Params:     getP384FqParams(),
		Arithmetic: p,
	}).IsNonZero()
	Selectznz(out, uint1(*wasInverted), out, &v)
}

// FromBytes converts a little endian byte array into a field element.
func (p384FqArithmetic) FromBytes(out *[native.Field6Limbs]uint64, arg *[native.Field6Bytes]byte) {
	FromBytes(out, arg)
}

// ToBytes converts a field element to a little endian byte array.
func (p384FqArithmetic) ToBytes(out *[native.Field6Bytes]byte, arg *[native.Field6Limbs]uint64) {
	ToBytes(out, arg)
}

func (p384FqArithmetic) Selectznz(out, arg1, arg2 *[native.Field6Limbs]uint64, choice int) {
	Selectznz(out, uint1(choice), arg1, arg2)
}
