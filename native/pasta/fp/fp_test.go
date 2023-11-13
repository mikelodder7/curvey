//
// Copyright Coinbase, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
//

package fp

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

func TestFpSetOne(t *testing.T) {
	fp := PastaFpNew().SetOne()
	require.NotNil(t, fp)
	require.Equal(t, 1, fp.IsOne())
}

func TestFpSetUint64(t *testing.T) {
	act := PastaFpNew().SetUint64(1 << 60)
	require.NotNil(t, act)
	// Remember it will be in montgomery form
	require.Equal(t, int(act.Value[0]), 0x592d30ed00000001)
}

func TestFpAdd(t *testing.T) {
	lhs := PastaFpNew().SetOne()
	rhs := PastaFpNew().SetOne()
	exp := PastaFpNew().SetUint64(2)
	res := PastaFpNew().Add(lhs, rhs)
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

		a := PastaFpNew().Add(lhs, rhs)
		require.NotNil(t, a)
		require.Equal(t, exp, a)
	}
}

func TestFpSub(t *testing.T) {
	lhs := PastaFpNew().SetOne()
	rhs := PastaFpNew().SetOne()
	exp := PastaFpNew().SetZero()
	res := PastaFpNew().Sub(lhs, rhs)
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

		a := PastaFpNew().Sub(lhs, rhs)
		require.NotNil(t, a)
		require.Equal(t, exp, a)
	}
}

func TestFpMul(t *testing.T) {
	lhs := PastaFpNew().SetOne()
	rhs := PastaFpNew().SetOne()
	exp := PastaFpNew().SetOne()
	res := PastaFpNew().Mul(lhs, rhs)
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

		a := PastaFpNew().Mul(lhs, rhs)
		require.NotNil(t, a)
		require.Equal(t, exp, a)
	}
}

func TestFpDouble(t *testing.T) {
	a := PastaFpNew().SetUint64(2)
	e := PastaFpNew().SetUint64(4)
	require.Equal(t, e, PastaFpNew().Double(a))

	for i := 0; i < 25; i++ {
		tv := randUnit32(t)
		ttv := uint64(tv) * 2
		a = PastaFpNew().SetUint64(uint64(tv))
		e = PastaFpNew().SetUint64(ttv)
		require.Equal(t, e, PastaFpNew().Double(a))
	}
}

func TestFpSquare(t *testing.T) {
	a := PastaFpNew().SetUint64(4)
	e := PastaFpNew().SetUint64(16)
	require.Equal(t, e, a.Square(a))

	for i := 0; i < 25; i++ {
		j := randUnit32(t)
		exp := uint64(j) * uint64(j)
		e.SetUint64(exp)
		a.SetUint64(uint64(j))
		require.Equal(t, e, a.Square(a))
	}
}

func TestFpNeg(t *testing.T) {
	g := PastaFpNew().SetRaw(&generator)
	a := PastaFpNew().SetOne()
	a.Neg(a)
	e := [native.Field4Limbs]uint64{7256640077462241284, 9879318615658062958, 0, 0}
	require.Equal(t, e, a.Value)
	a.Neg(g)
	e = [native.Field4Limbs]uint64{0xf787d28400000014, 0xad83f3b0ba037627, 0x2, 0x0}
	require.Equal(t, e, a.Value)
}

func TestFpExp(t *testing.T) {
	e := PastaFpNew().SetUint64(8)
	a := PastaFpNew().SetUint64(2)
	by := PastaFpNew().SetUint64(3)
	require.Equal(t, e, a.Exp(a, by))
}

func TestFpSqrt(t *testing.T) {
	t1 := PastaFpNew().SetUint64(2)
	t2 := PastaFpNew().Neg(t1)
	t3 := PastaFpNew().Square(t1)
	_, wasSquare := t3.Sqrt(t3)
	require.True(t, wasSquare)
	require.Equal(t, 1, t1.Equal(t3)|t2.Equal(t3))
	t1.SetUint64(5)
	_, wasSquare = PastaFpNew().Sqrt(t1)
	require.False(t, wasSquare)
}

func TestFpInvert(t *testing.T) {
	twoInv := PastaFpNew().SetLimbs(&[native.Field4Limbs]uint64{0xcc96987680000001, 0x11234c7e04a67c8d, 0x0000000000000000, 0x2000000000000000})
	two := PastaFpNew().SetUint64(2)
	a, inverted := PastaFpNew().Invert(two)
	require.True(t, inverted)
	require.Equal(t, a, twoInv)

	rootOfUnity := PastaFpNew().SetLimbs(&[native.Field4Limbs]uint64{0xbdad6fabd87ea32f, 0xea322bf2b7bb7584, 0x362120830561f81a, 0x2bce74deac30ebda})
	rootOfUnityInv := PastaFpNew().SetLimbs(&[native.Field4Limbs]uint64{0xf0b87c7db2ce91f6, 0x84a0a1d8859f066f, 0xb4ed8e647196dad1, 0x2cd5282c53116b5c})
	a, inverted = PastaFpNew().Invert(rootOfUnity)
	require.True(t, inverted)
	require.Equal(t, a, rootOfUnityInv)

	lhs := PastaFpNew().SetUint64(9)
	rhs := PastaFpNew().SetUint64(3)
	rhsInv, inverted := PastaFpNew().Invert(rhs)
	require.True(t, inverted)
	require.Equal(t, rhs, PastaFpNew().Mul(lhs, rhsInv))

	rhs.SetZero()
	_, inverted = PastaFpNew().Invert(rhs)
	require.False(t, inverted)
}

func TestFpCMove(t *testing.T) {
	t1 := PastaFpNew().SetUint64(5)
	t2 := PastaFpNew().SetUint64(10)
	require.Equal(t, t1, PastaFpNew().CMove(t1, t2, 0))
	require.Equal(t, t2, PastaFpNew().CMove(t1, t2, 1))
}

func TestFpBytes(t *testing.T) {
	t1 := PastaFpNew().SetUint64(99)
	seq := t1.Bytes()
	t2, err := PastaFpNew().SetBytes(&seq)
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

func TestFpBigInt(t *testing.T) {
	t1 := PastaFpNew().SetBigInt(big.NewInt(9999))
	t2 := PastaFpNew().SetBigInt(t1.BigInt())
	require.Equal(t, t1, t2)

	e := PastaFpNew().SetRaw(&[native.Field4Limbs]uint64{0x8c6bc70550c87761, 0xce2c6c48e7063731, 0xf1275fd1e4607cd6, 0x3e6762e63501edbd})
	b := new(big.Int).SetBytes([]byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9})
	t1.SetBigInt(b)
	require.Equal(t, e, t1)
	e.Value[0] = 0xcc169e7af3788a0
	e.Value[1] = 0x541a2cb32246c1ea
	e.Value[2] = 0xed8a02e1b9f8329
	e.Value[3] = 0x1989d19cafe1242
	b.Neg(b)
	t1.SetBigInt(b)
	require.Equal(t, e, t1)
}

func TestFpSetBytesWide(t *testing.T) {
	e := PastaFpNew().SetLimbs(&[native.Field4Limbs]uint64{0x3daec14d565241d9, 0x0b7af45b6073944b, 0xea5b8bd611a5bd4c, 0x150160330625db3d})
	a := PastaFpNew().SetBytesWide(&[64]byte{
		0xa1, 0x78, 0x76, 0x29, 0x41, 0x56, 0x15, 0xee,
		0x65, 0xbe, 0xfd, 0xdb, 0x6b, 0x15, 0x3e, 0xd8,
		0xb5, 0xa0, 0x8b, 0xc6, 0x34, 0xd8, 0xcc, 0xd9,
		0x58, 0x27, 0x27, 0x12, 0xe3, 0xed, 0x08, 0xf5,
		0x89, 0x8e, 0x22, 0xf8, 0xcb, 0xf7, 0x8d, 0x03,
		0x41, 0x4b, 0xc7, 0xa3, 0xe4, 0xa1, 0x05, 0x35,
		0xb3, 0x2d, 0xb8, 0x5e, 0x77, 0x6f, 0xa4, 0xbf,
		0x1d, 0x47, 0x2f, 0x26, 0x7e, 0xe2, 0xeb, 0x26,
	})
	require.Equal(t, e, a)
}
