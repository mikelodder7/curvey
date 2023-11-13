//
// Copyright Coinbase, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
//

package fq

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mikelodder7/curvey/native"
)

func randUnit64(t *testing.T) uint64 {
	t.Helper()
	b := [8]byte{}
	_, err := rand.Read(b[:])
	require.NoError(t, err)
	return binary.BigEndian.Uint64(b[:])
}

func randUnit32(t *testing.T) uint32 {
	t.Helper()
	b := [4]byte{}
	_, err := rand.Read(b[:])
	require.NoError(t, err)
	return binary.BigEndian.Uint32(b[:])
}

func TestFqSetOne(t *testing.T) {
	fq := PastaFqNew().SetOne()
	require.NotNil(t, fq)
	require.Equal(t, 1, fq.IsOne())
}

func TestFqSetUint64(t *testing.T) {
	act := PastaFqNew().SetUint64(1 << 60)
	require.NotNil(t, act)
	// Remember it will be in montgomery form
	require.Equal(t, int(act.Value[0]), 0x4c46eb2100000001)
}

func TestFqAdd(t *testing.T) {
	lhs := PastaFqNew().SetOne()
	rhs := PastaFqNew().SetOne()
	exp := PastaFqNew().SetUint64(2)
	res := PastaFqNew().Add(lhs, rhs)
	require.NotNil(t, res)
	require.Equal(t, 1, res.Equal(exp))

	// Fuzz test
	for i := 0; i < 25; i++ {
		// Divide by 4 to prevent overflow false errors
		l := randUnit64(t) >> 2
		r := randUnit64(t) >> 2
		e := l + r
		lhs.SetUint64(l)
		rhs.SetUint64(r)
		exp.SetUint64(e)

		a := PastaFqNew().Add(lhs, rhs)
		require.NotNil(t, a)
		require.Equal(t, exp, a)
	}
}

func TestFqSub(t *testing.T) {
	lhs := PastaFqNew().SetOne()
	rhs := PastaFqNew().SetOne()
	exp := PastaFqNew().SetZero()
	res := PastaFqNew().Sub(lhs, rhs)
	require.NotNil(t, res)
	require.Equal(t, 1, res.Equal(exp))

	// Fuzz test
	for i := 0; i < 25; i++ {
		// Divide by 4 to prevent overflow false errors
		l := randUnit64(t) >> 2
		r := randUnit64(t) >> 2
		if l < r {
			l, r = r, l
		}
		e := l - r
		lhs.SetUint64(l)
		rhs.SetUint64(r)
		exp.SetUint64(e)

		a := PastaFqNew().Sub(lhs, rhs)
		require.NotNil(t, a)
		require.Equal(t, exp, a)
	}
}

func TestFqMul(t *testing.T) {
	lhs := PastaFqNew().SetOne()
	rhs := PastaFqNew().SetOne()
	exp := PastaFqNew().SetOne()
	res := PastaFqNew().Mul(lhs, rhs)
	require.NotNil(t, res)
	require.Equal(t, 1, res.Equal(exp))

	// Fuzz test
	for i := 0; i < 25; i++ {
		// Divide by 4 to prevent overflow false errors
		l := randUnit32(t)
		r := randUnit32(t)
		e := uint64(l) * uint64(r)
		lhs.SetUint64(uint64(l))
		rhs.SetUint64(uint64(r))
		exp.SetUint64(e)

		a := PastaFqNew().Mul(lhs, rhs)
		require.NotNil(t, a)
		require.Equal(t, exp, a)
	}
}

func TestFqDouble(t *testing.T) {
	a := PastaFqNew().SetUint64(2)
	e := PastaFqNew().SetUint64(4)
	require.Equal(t, e, PastaFqNew().Double(a))

	for i := 0; i < 25; i++ {
		tv := randUnit32(t)
		ttv := uint64(tv) * 2
		a = PastaFqNew().SetUint64(uint64(tv))
		e = PastaFqNew().SetUint64(ttv)
		require.Equal(t, e, PastaFqNew().Double(a))
	}
}

func TestFqSquare(t *testing.T) {
	a := PastaFqNew().SetUint64(4)
	e := PastaFqNew().SetUint64(16)
	require.Equal(t, e, a.Square(a))

	for i := 0; i < 25; i++ {
		j := randUnit32(t)
		exp := uint64(j) * uint64(j)
		e.SetUint64(exp)
		a.SetUint64(uint64(j))
		require.Equal(t, e, a.Square(a))
	}
}

func TestFqNeg(t *testing.T) {
	g := PastaFqNew().SetRaw(&generator)
	a := PastaFqNew().SetOne()
	a.Neg(a)
	e := PastaFqNew().SetRaw(&[native.Field4Limbs]uint64{0x311bac8400000004, 0x891a63f02652a376, 0, 0})
	require.Equal(t, e, a)
	a.Neg(g)
	e = PastaFqNew().SetRaw(&[native.Field4Limbs]uint64{0xf58a5e9400000014, 0xad83f3b0bf9d314e, 0x2, 0x0})
	require.Equal(t, e, a)
}

func TestFqExp(t *testing.T) {
	e := PastaFqNew().SetUint64(8)
	a := PastaFqNew().SetUint64(2)
	by := PastaFqNew().SetUint64(3)
	require.Equal(t, e, a.Exp(a, by))
}

func TestFqSqrt(t *testing.T) {
	t1 := PastaFqNew().SetUint64(2)
	t2 := PastaFqNew().Neg(t1)
	t3 := PastaFqNew().Square(t1)
	_, wasSquare := t3.Sqrt(t3)
	require.True(t, wasSquare)
	require.Equal(t, 1, t1.Equal(t3)|t2.Equal(t3))
	t1.SetUint64(5)
	_, wasSquare = PastaFqNew().Sqrt(t1)
	require.False(t, wasSquare)
}

func TestFqInvert(t *testing.T) {
	twoInv := PastaFqNew().SetLimbs(&[native.Field4Limbs]uint64{0xc623759080000001, 0x11234c7e04ca546e, 0x0000000000000000, 0x2000000000000000})
	two := PastaFqNew().SetUint64(2)
	a, inverted := PastaFqNew().Invert(two)
	require.True(t, inverted)
	require.Equal(t, a, twoInv)

	rootOfUnity := PastaFqNew().SetLimbs(&[native.Field4Limbs]uint64{0xa70e2c1102b6d05f, 0x9bb97ea3c106f049, 0x9e5c4dfd492ae26e, 0x2de6a9b8746d3f58})
	rootOfUnityInv := PastaFqNew().SetLimbs(&[native.Field4Limbs]uint64{0x57eecda0a84b6836, 0x4ad38b9084b8a80c, 0xf4c8f353124086c1, 0x2235e1a7415bf936})
	a, inverted = PastaFqNew().Invert(rootOfUnity)
	require.True(t, inverted)
	require.Equal(t, a, rootOfUnityInv)

	lhs := PastaFqNew().SetUint64(9)
	rhs := PastaFqNew().SetUint64(3)
	rhsInv, inverted := PastaFqNew().Invert(rhs)
	require.True(t, inverted)
	require.Equal(t, rhs, PastaFqNew().Mul(lhs, rhsInv))

	rhs.SetZero()
	_, inverted = PastaFqNew().Invert(rhs)
	require.False(t, inverted)
}

func TestFqCMove(t *testing.T) {
	t1 := PastaFqNew().SetUint64(5)
	t2 := PastaFqNew().SetUint64(10)
	require.Equal(t, t1, PastaFqNew().CMove(t1, t2, 0))
	require.Equal(t, t2, PastaFqNew().CMove(t1, t2, 1))
}

func TestFqBytes(t *testing.T) {
	t1 := PastaFqNew().SetUint64(99)
	seq := t1.Bytes()
	t2, err := PastaFqNew().SetBytes(&seq)
	require.NoError(t, err)
	require.Equal(t, t1, t2)

	for i := 0; i < 25; i++ {
		t1.SetUint64(randUnit64(t))
		seq = t1.Bytes()
		_, err = t2.SetBytes(&seq)
		require.NoError(t, err)
		require.Equal(t, t1, t2)
	}
}

func TestFqBigInt(t *testing.T) {
	t1 := PastaFqNew().SetBigInt(big.NewInt(9999))
	t2 := PastaFqNew().SetBigInt(t1.BigInt())
	require.Equal(t, t1, t2)

	e := PastaFqNew().SetRaw(&[native.Field4Limbs]uint64{0x7bb1416dea3d6ae3, 0x62f9108a340aa525, 0x303b3f30fcaa477f, 0x11c9ef5422d80a4d})
	b := new(big.Int).SetBytes([]byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9})
	t1.SetBigInt(b)
	require.Equal(t, e, t1)
	e.Value[0] = 0x1095a9b315c2951e
	e.Value[1] = 0xbf4d8871d58a03b8
	e.Value[2] = 0xcfc4c0cf0355b880
	e.Value[3] = 0x2e3610abdd27f5b2
	b.Neg(b)
	t1.SetBigInt(b)
	require.Equal(t, e, t1)
}

func TestFqSetBytesWide(t *testing.T) {
	e := PastaFqNew().SetLimbs(&[native.Field4Limbs]uint64{0xe22bd0d1b22cc43e, 0x6b84e5b52490a7c8, 0x264262941ac9e229, 0x27dcfdf361ce4254})
	a := PastaFqNew().SetBytesWide(&[64]byte{
		0x69, 0x23, 0x5a, 0x0b, 0xce, 0x0c, 0xa8, 0x64,
		0x3c, 0x78, 0xbc, 0x01, 0x05, 0xef, 0xf2, 0x84,
		0xde, 0xbb, 0x6b, 0xc8, 0x63, 0x5e, 0x6e, 0x69,
		0x62, 0xcc, 0xc6, 0x2d, 0xf5, 0x72, 0x40, 0x92,
		0x28, 0x11, 0xd6, 0xc8, 0x07, 0xa5, 0x88, 0x82,
		0xfe, 0xe3, 0x97, 0xf6, 0x1e, 0xfb, 0x2e, 0x3b,
		0x27, 0x5f, 0x85, 0x06, 0x8d, 0x99, 0xa4, 0x75,
		0xc0, 0x2c, 0x71, 0x69, 0x9e, 0x58, 0xea, 0x52,
	})
	require.Equal(t, e, a)
}
