//
// SPDX-License-Identifier: Apache-2.0
//

package fp

import (
	"math/big"
	"sync"

	"github.com/mikelodder7/curvey/native"
)

var (
	p384FpInitonce sync.Once
	p384FpParams   native.Field6Params
)

func P384FpNew() *native.Field6 {
	return &native.Field6{
		Value:      [native.Field6Limbs]uint64{},
		Params:     getP384FpParams(),
		Arithmetic: p384FpArithmetic{},
	}
}

func p384FpParamsInit() {
	// See FIPS 186-3, section D.2.3
	p384FpParams = native.Field6Params{
		R: [native.Field6Limbs]uint64{
			0xffffffff00000001,
			0x00000000ffffffff,
			0x0000000000000001,
			0x0000000000000000,
			0x0000000000000000,
			0x0000000000000000,
		},
		R2: [native.Field6Limbs]uint64{
			0xfffffffe00000001,
			0x0000000200000000,
			0xfffffffe00000000,
			0x0000000200000000,
			0x0000000000000001,
			0x0000000000000000,
		},
		R3: [native.Field6Limbs]uint64{
			0xfffffffc00000002,
			0x0000000300000002,
			0xfffffffcfffffffe,
			0x0000000300000005,
			0xfffffffdfffffffd,
			0x0000000300000002,
		},
		Modulus: [native.Field6Limbs]uint64{
			0x00000000ffffffff,
			0xffffffff00000000,
			0xfffffffffffffffe,
			0xffffffffffffffff,
			0xffffffffffffffff,
			0xffffffffffffffff,
		},
		BiModulus: new(big.Int).SetBytes([]byte{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff,
		}),
	}
}

func getP384FpParams() *native.Field6Params {
	p384FpInitonce.Do(p384FpParamsInit)
	return &p384FpParams
}

type p384FpArithmetic struct{}

// ToMontgomery converts this field to montgomery form.
func (p384FpArithmetic) ToMontgomery(out, arg *[native.Field6Limbs]uint64) {
	ToMontgomery((*MontgomeryDomainFieldElement)(out), (*NonMontgomeryDomainFieldElement)(arg))
}

// FromMontgomery converts this field from montgomery form.
func (p384FpArithmetic) FromMontgomery(out, arg *[native.Field6Limbs]uint64) {
	FromMontgomery((*NonMontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg))
}

// Neg performs modular negation.
func (p384FpArithmetic) Neg(out, arg *[native.Field6Limbs]uint64) {
	Opp((*MontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg))
}

// Square performs modular square.
func (p384FpArithmetic) Square(out, arg *[native.Field6Limbs]uint64) {
	Square((*MontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg))
}

// Mul performs modular multiplication.
func (p384FpArithmetic) Mul(out, arg1, arg2 *[native.Field6Limbs]uint64) {
	Mul((*MontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg1), (*MontgomeryDomainFieldElement)(arg2))
}

// Add performs modular addition.
func (p384FpArithmetic) Add(out, arg1, arg2 *[native.Field6Limbs]uint64) {
	Add((*MontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg1), (*MontgomeryDomainFieldElement)(arg2))
}

// Sub performs modular subtraction.
func (p384FpArithmetic) Sub(out, arg1, arg2 *[native.Field6Limbs]uint64) {
	Sub((*MontgomeryDomainFieldElement)(out), (*MontgomeryDomainFieldElement)(arg1), (*MontgomeryDomainFieldElement)(arg2))
}

// Sqrt performs modular square root.
func (p p384FpArithmetic) Sqrt(wasSquare *int, out, arg *[native.Field6Limbs]uint64) {
	// p mod 4 = 3 -> compute sqrt(x) using x^((p+1)/4) =
	// x^9850501549098619803069760025035903451269934817616361666987073351061430442874217582261816522064734500465401743278080
	var t, t1, t10, t11, t110, t111, t111000, t111111, t1111110, t1111111 [native.Field6Limbs]uint64
	var x12, x24, x31, x32, x63, x126, x252, x255, x [native.Field6Limbs]uint64

	copy(t1[:], arg[:])

	p.Square(&t10, &t1)
	p.Mul(&t11, &t1, &t10)
	p.Square(&t110, &t11)
	p.Mul(&t111, &t1, &t110)
	native.Pow2k6(&t111000, &t111, 3, p)
	p.Mul(&t111111, &t111, &t111000)
	p.Square(&t1111110, &t111111)
	p.Mul(&t1111111, &t1, &t1111110)
	native.Pow2k6(&t, &t1111110, 5, p)
	p.Mul(&x12, &t, &t111111)
	native.Pow2k6(&t, &x12, 12, p)
	p.Mul(&x24, &t, &x12)
	native.Pow2k6(&t, &x24, 7, p)
	p.Mul(&x31, &t, &t1111111)
	p.Square(&t, &x31)
	p.Mul(&x32, &t, &t1)
	native.Pow2k6(&t, &x32, 31, p)
	p.Mul(&x63, &t, &x31)
	native.Pow2k6(&t, &x63, 63, p)
	p.Mul(&x126, &t, &x63)
	native.Pow2k6(&t, &x126, 126, p)
	p.Mul(&x252, &t, &x126)
	native.Pow2k6(&t, &x252, 3, p)
	p.Mul(&x255, &t, &t111)

	native.Pow2k6(&t, &x255, 33, p)
	p.Mul(&x, &t, &x32)
	native.Pow2k6(&t, &x, 64, p)
	p.Mul(&t, &t, &t1)
	native.Pow2k6(&x, &t, 30, p)
	p.Square(&t, &x)

	*wasSquare = (&native.Field6{Value: t, Params: getP384FpParams(), Arithmetic: p}).Equal(&native.Field6{
		Value: *arg, Params: getP384FpParams(), Arithmetic: p,
	})
	Selectznz(out, uint1(*wasSquare), out, &x)
}

// Invert performs modular inverse.
func (p p384FpArithmetic) Invert(wasInverted *int, out, arg *[native.Field6Limbs]uint64) {
	// Exponentiate by p - 2
	var t [native.Field6Limbs]uint64
	f := P384FpNew()
	f.Value = *arg
	*wasInverted = f.IsNonZero()
	native.Pow6(&t, arg, &[native.Field6Limbs]uint64{
		0x00000000fffffffd,
		0xffffffff00000000,
		0xfffffffffffffffe,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
	}, f.Params, p)
	p.Selectznz(out, arg, &t, *wasInverted)
}

// FromBytes converts a little endian byte array into a field element.
func (p384FpArithmetic) FromBytes(out *[native.Field6Limbs]uint64, arg *[native.Field6Bytes]byte) {
	FromBytes(out, arg)
}

// ToBytes converts a field element to a little endian byte array.
func (p384FpArithmetic) ToBytes(out *[native.Field6Bytes]byte, arg *[native.Field6Limbs]uint64) {
	ToBytes(out, arg)
}

func (p384FpArithmetic) Selectznz(out, arg1, arg2 *[native.Field6Limbs]uint64, choice int) {
	Selectznz(out, uint1(choice), arg1, arg2)
}
