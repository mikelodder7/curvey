//
// Copyright Coinbase, Inc. All Rights Reserved.
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
	pastaFqInitonce sync.Once
	pastaFqParams   native.Field4Params
)

func PastaFqNew() *native.Field4 {
	return &native.Field4{
		Value:      [native.Field4Limbs]uint64{},
		Params:     getPastaFqParams(),
		Arithmetic: pastaFqArithmetic{},
	}
}

func pastaFqParamsInit() {
	pastaFqParams = native.Field4Params{
		R:       [native.Field4Limbs]uint64{0x5b2b3e9cfffffffd, 0x992c350be3420567, 0xffffffffffffffff, 0x3fffffffffffffff},
		R2:      [native.Field4Limbs]uint64{0xfc9678ff0000000f, 0x67bb433d891a16e3, 0x7fae231004ccf590, 0x096d41af7ccfdaa9},
		R3:      [native.Field4Limbs]uint64{0x008b421c249dae4c, 0xe13bda50dba41326, 0x88fececb8e15cb63, 0x07dd97a06e6792c8},
		Modulus: [native.Field4Limbs]uint64{0x8c46eb2100000001, 0x224698fc0994a8dd, 0x0000000000000000, 0x4000000000000000},
		BiModulus: new(big.Int).SetBytes([]byte{
			0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x22, 0x46, 0x98, 0xfc, 0x09, 0x94, 0xa8, 0xdd,
			0x8c, 0x46, 0xeb, 0x21, 0x00, 0x00, 0x00, 0x01,
		}),
	}
}

func getPastaFqParams() *native.Field4Params {
	pastaFqInitonce.Do(pastaFqParamsInit)
	return &pastaFqParams
}

type pastaFqArithmetic struct{}

func (pastaFqArithmetic) ToMontgomery(out, arg *[native.Field4Limbs]uint64) {
	fiatPastaFqToMontgomery((*fiatPastaFqMontgomeryDomainFieldElement)(out), (*fiatPastaFqNonMontgomeryDomainFieldElement)(arg))
}

func (pastaFqArithmetic) FromMontgomery(out, arg *[native.Field4Limbs]uint64) {
	fiatPastaFqFromMontgomery((*fiatPastaFqNonMontgomeryDomainFieldElement)(out), (*fiatPastaFqMontgomeryDomainFieldElement)(arg))
}

func (pastaFqArithmetic) Neg(out, arg *[native.Field4Limbs]uint64) {
	fiatPastaFqOpp((*fiatPastaFqMontgomeryDomainFieldElement)(out), (*fiatPastaFqMontgomeryDomainFieldElement)(arg))
}

func (pastaFqArithmetic) Square(out, arg *[native.Field4Limbs]uint64) {
	fiatPastaFqSquare((*fiatPastaFqMontgomeryDomainFieldElement)(out), (*fiatPastaFqMontgomeryDomainFieldElement)(arg))
}

func (pastaFqArithmetic) Mul(out, arg1, arg2 *[native.Field4Limbs]uint64) {
	fiatPastaFqMul(
		(*fiatPastaFqMontgomeryDomainFieldElement)(out),
		(*fiatPastaFqMontgomeryDomainFieldElement)(arg1),
		(*fiatPastaFqMontgomeryDomainFieldElement)(arg2),
	)
}

func (pastaFqArithmetic) Add(out, arg1, arg2 *[native.Field4Limbs]uint64) {
	fiatPastaFqAdd(
		(*fiatPastaFqMontgomeryDomainFieldElement)(out),
		(*fiatPastaFqMontgomeryDomainFieldElement)(arg1),
		(*fiatPastaFqMontgomeryDomainFieldElement)(arg2),
	)
}

func (pastaFqArithmetic) Sub(out, arg1, arg2 *[native.Field4Limbs]uint64) {
	fiatPastaFqSub(
		(*fiatPastaFqMontgomeryDomainFieldElement)(out),
		(*fiatPastaFqMontgomeryDomainFieldElement)(arg1),
		(*fiatPastaFqMontgomeryDomainFieldElement)(arg2),
	)
}

func (f pastaFqArithmetic) Sqrt(wasSquare *int, out, arg *[native.Field4Limbs]uint64) {
	// c1 := 32
	// c2 := (q - 1) / (2^c1)
	// c2 := [4]uint64{
	// 	0x0994a8dd8c46eb21,
	// 	0x00000000224698fc,
	// 	0x0000000000000000,
	// 	0x0000000040000000,
	// }
	// c3 := (c2 - 1) / 2
	c3 := [native.Field4Limbs]uint64{
		0x04ca546ec6237590,
		0x0000000011234c7e,
		0x0000000000000000,
		0x0000000020000000,
	}
	// c4 := generator
	// c5 := new(Fq).pow(&generator, c2)
	c5 := &[native.Field4Limbs]uint64{
		0x218077428c9942de,
		0xcc49578921b60494,
		0xac2e5d27b2efbee2,
		0xb79fa897f2db056,
	}
	var z, t, b, c, tv [native.Field4Limbs]uint64

	native.Pow(&z, arg, &c3, getPastaFqParams(), f)
	fiatPastaFqSquare((*fiatPastaFqMontgomeryDomainFieldElement)(&t), (*fiatPastaFqMontgomeryDomainFieldElement)(&z))
	fiatPastaFqMul((*fiatPastaFqMontgomeryDomainFieldElement)(&t), (*fiatPastaFqMontgomeryDomainFieldElement)(&t), (*fiatPastaFqMontgomeryDomainFieldElement)(arg))
	fiatPastaFqMul((*fiatPastaFqMontgomeryDomainFieldElement)(&z), (*fiatPastaFqMontgomeryDomainFieldElement)(&z), (*fiatPastaFqMontgomeryDomainFieldElement)(arg))

	copy(b[:], t[:])
	copy(c[:], c5[:])

	for i := s; i >= 2; i-- {
		for j := 1; j <= i-2; j++ {
			fiatPastaFqSquare((*fiatPastaFqMontgomeryDomainFieldElement)(&b), (*fiatPastaFqMontgomeryDomainFieldElement)(&b))
		}
		// if b == 1 flag = 0 else flag = 1
		flag := -(&native.Field4{
			Value:      b,
			Params:     getPastaFqParams(),
			Arithmetic: f,
		}).IsOne() + 1
		fiatPastaFqMul((*fiatPastaFqMontgomeryDomainFieldElement)(&tv), (*fiatPastaFqMontgomeryDomainFieldElement)(&z), (*fiatPastaFqMontgomeryDomainFieldElement)(&c))
		fiatPastaFqSelectznz(&z, fiatPastaFqUint1(flag), &z, &tv)
		fiatPastaFqSquare((*fiatPastaFqMontgomeryDomainFieldElement)(&c), (*fiatPastaFqMontgomeryDomainFieldElement)(&c))
		fiatPastaFqMul((*fiatPastaFqMontgomeryDomainFieldElement)(&tv), (*fiatPastaFqMontgomeryDomainFieldElement)(&t), (*fiatPastaFqMontgomeryDomainFieldElement)(&c))
		fiatPastaFqSelectznz(&t, fiatPastaFqUint1(flag), &t, &tv)
		copy(b[:], t[:])
	}
	fiatPastaFqSquare((*fiatPastaFqMontgomeryDomainFieldElement)(&c), (*fiatPastaFqMontgomeryDomainFieldElement)(&z))
	*wasSquare = (&native.Field4{
		Value:      c,
		Params:     getPastaFqParams(),
		Arithmetic: f,
	}).Equal(&native.Field4{
		Value:      *arg,
		Params:     getPastaFqParams(),
		Arithmetic: f,
	})
	fiatPastaFqSelectznz(out, fiatPastaFqUint1(*wasSquare), out, &z)
}

func (f pastaFqArithmetic) Invert(wasInverted *int, out, arg *[native.Field4Limbs]uint64) {
	var t [native.Field4Limbs]uint64
	// computes elem^(p - 2) mod p
	exp := [native.Field4Limbs]uint64{
		0x8c46eb20ffffffff,
		0x224698fc0994a8dd,
		0x0000000000000000,
		0x4000000000000000,
	}

	native.Pow(&t, arg, &exp, getPastaFqParams(), f)

	*wasInverted = (&native.Field4{
		Value:      *arg,
		Params:     getPastaFqParams(),
		Arithmetic: f,
	}).IsNonZero()
	fiatPastaFqSelectznz(out, fiatPastaFqUint1(*wasInverted), out, &t)
}

func (pastaFqArithmetic) FromBytes(out *[native.Field4Limbs]uint64, arg *[native.Field4Bytes]byte) {
	fiatPastaFqFromBytes(out, arg)
}

func (pastaFqArithmetic) ToBytes(out *[native.Field4Bytes]byte, arg *[native.Field4Limbs]uint64) {
	fiatPastaFqToBytes(out, arg)
}

func (pastaFqArithmetic) Selectznz(out, arg1, arg2 *[native.Field4Limbs]uint64, choice int) {
	fiatPastaFqSelectznz(out, fiatPastaFqUint1(choice), arg1, arg2)
}

// generator = 5 mod p is a generator of the `p - 1` order multiplicative
// subgroup, or in other words a primitive element of the field.
var generator = [native.Field4Limbs]uint64{0x96bc8c8cffffffed, 0x74c2a54b49f7778e, 0xfffffffffffffffd, 0x3fffffffffffffff}

// s satisfies the equation 2^s * t = q -1 with t odd.
var s = 32
