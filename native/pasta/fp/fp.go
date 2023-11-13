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
	pastaFpInitonce sync.Once
	pastaFpParams   native.Field4Params
)

func PastaFpNew() *native.Field4 {
	return &native.Field4{
		Value:      [native.Field4Limbs]uint64{},
		Params:     GetPastaFpParams(),
		Arithmetic: PastaFpArithmetic{},
	}
}

func pastaFpParamsInit() {
	pastaFpParams = native.Field4Params{
		R:       [native.Field4Limbs]uint64{0x34786d38fffffffd, 0x992c350be41914ad, 0xffffffffffffffff, 0x3fffffffffffffff},
		R2:      [native.Field4Limbs]uint64{0x8c78ecb30000000f, 0xd7d30dbd8b0de0e7, 0x7797a99bc3c95d18, 0x096d41af7b9cb714},
		R3:      [native.Field4Limbs]uint64{0xf185a5993a9e10f9, 0xf6a68f3b6ac5b1d1, 0xdf8d1014353fd42c, 0x2ae309222d2d9910},
		Modulus: [native.Field4Limbs]uint64{0x992d30ed00000001, 0x224698fc094cf91b, 0x0000000000000000, 0x4000000000000000},
		BiModulus: new(big.Int).SetBytes([]byte{
			0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x22, 0x46, 0x98, 0xfc, 0x09, 0x4c, 0xf9, 0x1b,
			0x99, 0x2d, 0x30, 0xed, 0x00, 0x00, 0x00, 0x01,
		}),
	}
}

func GetPastaFpParams() *native.Field4Params {
	pastaFpInitonce.Do(pastaFpParamsInit)
	return &pastaFpParams
}

type PastaFpArithmetic struct{}

func (PastaFpArithmetic) ToMontgomery(out, arg *[native.Field4Limbs]uint64) {
	fiatPastaFpToMontgomery((*fiatPastaFpMontgomeryDomainFieldElement)(out), (*fiatPastaFpNonMontgomeryDomainFieldElement)(arg))
}

func (PastaFpArithmetic) FromMontgomery(out, arg *[native.Field4Limbs]uint64) {
	fiatPastaFpFromMontgomery((*fiatPastaFpNonMontgomeryDomainFieldElement)(out), (*fiatPastaFpMontgomeryDomainFieldElement)(arg))
}

func (PastaFpArithmetic) Neg(out, arg *[native.Field4Limbs]uint64) {
	fiatPastaFpOpp((*fiatPastaFpMontgomeryDomainFieldElement)(out), (*fiatPastaFpMontgomeryDomainFieldElement)(arg))
}

func (PastaFpArithmetic) Square(out, arg *[native.Field4Limbs]uint64) {
	fiatPastaFpSquare((*fiatPastaFpMontgomeryDomainFieldElement)(out), (*fiatPastaFpMontgomeryDomainFieldElement)(arg))
}

func (PastaFpArithmetic) Mul(out, arg1, arg2 *[native.Field4Limbs]uint64) {
	fiatPastaPpMul(
		(*fiatPastaFpMontgomeryDomainFieldElement)(out),
		(*fiatPastaFpMontgomeryDomainFieldElement)(arg1),
		(*fiatPastaFpMontgomeryDomainFieldElement)(arg2),
	)
}

func (PastaFpArithmetic) Add(out, arg1, arg2 *[native.Field4Limbs]uint64) {
	fiatPastaFpAdd(
		(*fiatPastaFpMontgomeryDomainFieldElement)(out),
		(*fiatPastaFpMontgomeryDomainFieldElement)(arg1),
		(*fiatPastaFpMontgomeryDomainFieldElement)(arg2),
	)
}

func (PastaFpArithmetic) Sub(out, arg1, arg2 *[native.Field4Limbs]uint64) {
	fiatPastaFpSub(
		(*fiatPastaFpMontgomeryDomainFieldElement)(out),
		(*fiatPastaFpMontgomeryDomainFieldElement)(arg1),
		(*fiatPastaFpMontgomeryDomainFieldElement)(arg2),
	)
}

func (f PastaFpArithmetic) Sqrt(wasSquare *int, out, arg *[native.Field4Limbs]uint64) {
	// c1 := 32
	// c2 := (q - 1) / (2^c1)
	// c2 := [4]uint64{
	// 	0x094cf91b992d30ed,
	// 	0x00000000224698fc,
	// 	0x0000000000000000,
	// 	0x0000000040000000,
	// }
	// c3 := (c2 - 1) / 2
	c3 := [native.Field4Limbs]uint64{
		0x04a67c8dcc969876,
		0x0000000011234c7e,
		0x0000000000000000,
		0x0000000020000000,
	}
	// c4 := generator
	// c5 := new(Fp).pow(&generator, c2)
	c5 := &[native.Field4Limbs]uint64{
		0xa28db849bad6dbf0,
		0x9083cd03d3b539df,
		0xfba6b9ca9dc8448e,
		0x3ec928747b89c6da,
	}

	var z, t, b, c, tv [native.Field4Limbs]uint64

	native.Pow(&z, arg, &c3, GetPastaFpParams(), f)
	fiatPastaFpSquare((*fiatPastaFpMontgomeryDomainFieldElement)(&t), (*fiatPastaFpMontgomeryDomainFieldElement)(&z))
	fiatPastaPpMul((*fiatPastaFpMontgomeryDomainFieldElement)(&t), (*fiatPastaFpMontgomeryDomainFieldElement)(&t), (*fiatPastaFpMontgomeryDomainFieldElement)(arg))
	fiatPastaPpMul((*fiatPastaFpMontgomeryDomainFieldElement)(&z), (*fiatPastaFpMontgomeryDomainFieldElement)(&z), (*fiatPastaFpMontgomeryDomainFieldElement)(arg))

	copy(b[:], t[:])
	copy(c[:], c5[:])

	for i := s; i >= 2; i-- {
		for j := 1; j <= i-2; j++ {
			fiatPastaFpSquare((*fiatPastaFpMontgomeryDomainFieldElement)(&b), (*fiatPastaFpMontgomeryDomainFieldElement)(&b))
		}
		// if b == 1 flag = 0 else flag = 1
		flag := -(&native.Field4{
			Value:      b,
			Params:     GetPastaFpParams(),
			Arithmetic: f,
		}).IsOne() + 1
		fiatPastaPpMul((*fiatPastaFpMontgomeryDomainFieldElement)(&tv), (*fiatPastaFpMontgomeryDomainFieldElement)(&z), (*fiatPastaFpMontgomeryDomainFieldElement)(&c))
		fiatPastaFpSelectznz(&z, fiatPastaFpUint1(flag), &z, &tv)
		fiatPastaFpSquare((*fiatPastaFpMontgomeryDomainFieldElement)(&c), (*fiatPastaFpMontgomeryDomainFieldElement)(&c))
		fiatPastaPpMul((*fiatPastaFpMontgomeryDomainFieldElement)(&tv), (*fiatPastaFpMontgomeryDomainFieldElement)(&t), (*fiatPastaFpMontgomeryDomainFieldElement)(&c))
		fiatPastaFpSelectznz(&t, fiatPastaFpUint1(flag), &t, &tv)
		copy(b[:], t[:])
	}
	fiatPastaFpSquare((*fiatPastaFpMontgomeryDomainFieldElement)(&c), (*fiatPastaFpMontgomeryDomainFieldElement)(&z))
	*wasSquare = (&native.Field4{
		Value:      c,
		Params:     GetPastaFpParams(),
		Arithmetic: f,
	}).Equal(&native.Field4{
		Value:      *arg,
		Params:     GetPastaFpParams(),
		Arithmetic: f,
	})
	fiatPastaFpSelectznz(out, fiatPastaFpUint1(*wasSquare), out, &z)
}

func (f PastaFpArithmetic) Invert(wasInverted *int, out, arg *[native.Field4Limbs]uint64) {
	var t [native.Field4Limbs]uint64
	// computes elem^(p - 2) mod p
	exp := [native.Field4Limbs]uint64{
		0x992d30ecffffffff,
		0x224698fc094cf91b,
		0x0000000000000000,
		0x4000000000000000,
	}

	native.Pow(&t, arg, &exp, GetPastaFpParams(), f)

	*wasInverted = (&native.Field4{
		Value:      *arg,
		Params:     GetPastaFpParams(),
		Arithmetic: f,
	}).IsNonZero()
	fiatPastaFpSelectznz(out, fiatPastaFpUint1(*wasInverted), out, &t)
}

func (PastaFpArithmetic) FromBytes(out *[native.Field4Limbs]uint64, arg *[native.Field4Bytes]byte) {
	fiatPastaFpFromBytes(out, arg)
}

func (PastaFpArithmetic) ToBytes(out *[native.Field4Bytes]byte, arg *[native.Field4Limbs]uint64) {
	fiatPastaFpToBytes(out, arg)
}

func (PastaFpArithmetic) Selectznz(out, arg1, arg2 *[native.Field4Limbs]uint64, choice int) {
	fiatPastaFpSelectznz(out, fiatPastaFpUint1(choice), arg1, arg2)
}

type Fp fiatPastaFpMontgomeryDomainFieldElement

// generator = 5 mod p is a generator of the `p - 1` order multiplicative
// subgroup, or in other words a primitive element of the field.
var generator = [native.Field4Limbs]uint64{0xa1a55e68ffffffed, 0x74c2a54b4f4982f3, 0xfffffffffffffffd, 0x3fffffffffffffff}

// s satisfies the equation 2^s * t = q -1 with t odd.
var s = 32
