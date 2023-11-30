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
	// Use p = 3 mod 4 by Euler's criterion means
	// arg^((p+1)/4 mod p
	var t, c [native.Field6Limbs]uint64
	c1 := [native.Field6Limbs]uint64{
		0xbb3b065ab3314a5d,
		0xd606836c922c29de,
		0xf1d8d3607d0dcb77,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0x3fffffffffffffff,
	}
	native.Pow6(&t, arg, &c1, getP384FqParams(), p)
	Square((*MontgomeryDomainFieldElement)(&c), (*MontgomeryDomainFieldElement)(&t))
	*wasSquare = (&native.Field6{Value: c, Params: getP384FqParams(), Arithmetic: p}).Equal(&native.Field6{
		Value: *arg, Params: getP384FqParams(), Arithmetic: p,
	})
	Selectznz(out, uint1(*wasSquare), out, &t)
}

// Invert performs modular inverse.
func (p p384FqArithmetic) Invert(wasInverted *int, out, arg *[native.Field6Limbs]uint64) {
	// Exponentiate by p - 2
	f := P384FqNew()
	f.Value = *arg
	*wasInverted = f.IsNonZero()
	native.Pow6(out, arg, &[native.Field6Limbs]uint64{
		0xecec196accc52971,
		0x581a0db248b0a77a,
		0xc7634d81f4372ddf,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
	}, f.Params, p)
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
