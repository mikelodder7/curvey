//
// SPDX-License-Identifier: Apache-2.0
//

package curvey

import (
	crand "crypto/rand"
	"encoding/hex"
	"github.com/mikelodder7/curvey/native/ed448/fq"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScalarEd448Random(t *testing.T) {
	ed448 := ED448()
	// Try 10 random values
	for i := 0; i < 10; i++ {
		sc := ed448.Scalar.Random(crand.Reader)
		_, ok := sc.(*ScalarEd448)
		require.True(t, ok)
		require.True(t, !sc.IsZero())
	}
}

func TestScalarEd448Hash(t *testing.T) {
	var b []byte
	sc := ED448().Scalar.Hash(b[:])
	s, ok := sc.(*ScalarEd448)
	require.True(t, ok)
	expected, _ := hex.DecodeString("000295c173bdce27f6c92ccaa741e4e0c83a7ecf6508c271e490d85f4d09cb7d62e44246a664eaeff78f3413a427b17b1a0e07e116baac0ce3")
	require.Equal(t, s.Bytes(), expected)
}

func TestScalarEd448Zero(t *testing.T) {
	sc := ED448().Scalar.Zero()
	require.True(t, sc.IsZero())
	require.True(t, sc.IsEven())
}

func TestScalarEd448One(t *testing.T) {
	sc := ED448().Scalar.One()
	require.True(t, sc.IsOne())
	require.True(t, sc.IsOdd())
}

func TestScalarEd448New(t *testing.T) {
	three := ED448().Scalar.New(3)
	require.True(t, three.IsOdd())
	four := ED448().Scalar.New(4)
	require.True(t, four.IsEven())
	neg1 := ED448().Scalar.New(-1)
	require.True(t, neg1.IsEven())
	neg2 := ED448().Scalar.New(-2)
	require.True(t, neg2.IsOdd())
}

func TestScalarEd448Square(t *testing.T) {
	three := ED448().Scalar.New(3)
	nine := ED448().Scalar.New(9)
	require.Equal(t, three.Square().Cmp(nine), 0)
}

func TestScalarEd448Cube(t *testing.T) {
	three := ED448().Scalar.New(3)
	twentySeven := ED448().Scalar.New(27)
	require.Equal(t, three.Cube().Cmp(twentySeven), 0)
}

func TestScalarEd448Double(t *testing.T) {
	three := ED448().Scalar.New(3)
	six := ED448().Scalar.New(6)
	require.Equal(t, three.Double().Cmp(six), 0)
}

func TestScalarEd448Neg(t *testing.T) {
	one := ED448().Scalar.One()
	neg1 := ED448().Scalar.New(-1)
	require.Equal(t, one.Neg().Cmp(neg1), 0)
	lotsOfThrees := ED448().Scalar.New(333333)
	expected := ED448().Scalar.New(-333333)
	require.Equal(t, lotsOfThrees.Neg().Cmp(expected), 0)
}

func TestScalarEd448Invert(t *testing.T) {
	nine := ED448().Scalar.New(9)
	actual, _ := nine.Invert()
	require.Equal(t, nine.Mul(actual).IsOne(), true)
}

func TestScalarEd448Sqrt(t *testing.T) {
	nine := ED448().Scalar.New(9)
	actual, _ := nine.Sqrt()
	require.Equal(t, actual.Square().Cmp(nine), 0)
}

func TestScalarEd448Add(t *testing.T) {
	nine := ED448().Scalar.New(9)
	six := ED448().Scalar.New(6)
	fifteen := nine.Add(six)
	require.NotNil(t, fifteen)
	expected := ED448().Scalar.New(15)
	require.Equal(t, expected.Cmp(fifteen), 0)
}

func TestScalarEd448Sub(t *testing.T) {
	nine := ED448().Scalar.New(9)
	six := ED448().Scalar.New(6)
	n := new(big.Int).Set(fq.FqNew().Value.Params.BiModulus)
	n.Sub(n, big.NewInt(3))

	expected, err := ED448().Scalar.SetBigInt(n)
	require.NoError(t, err)
	actual := six.Sub(nine)
	require.Equal(t, expected.Cmp(actual), 0)

	actual = nine.Sub(six)
	require.Equal(t, actual.Cmp(ED448().Scalar.New(3)), 0)
}

func TestScalarEd448Mul(t *testing.T) {
	nine := ED448().Scalar.New(9)
	six := ED448().Scalar.New(6)
	actual := nine.Mul(six)
	require.Equal(t, actual.Cmp(ED448().Scalar.New(54)), 0)
	n := new(big.Int).Set(fq.FqNew().Value.Params.BiModulus)
	n.Sub(n, big.NewInt(1))
	upper, err := ED448().Scalar.SetBigInt(n)
	require.NoError(t, err)
	require.Equal(t, upper.Mul(upper).Cmp(ED448().Scalar.New(1)), 0)
}

func TestScalarEd448Div(t *testing.T) {
	nine := ED448().Scalar.New(9)
	actual := nine.Div(nine)
	require.Equal(t, actual.Cmp(ED448().Scalar.New(1)), 0)
	require.Equal(t, ED448().Scalar.New(54).Div(nine).Cmp(ED448().Scalar.New(6)), 0)
}

func TestScalarEd448Serialize(t *testing.T) {
	sc := ED448().Scalar.New(255)
	sequence := sc.Bytes()
	require.Equal(t, len(sequence), 57)
	require.Equal(t, sequence, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff})
	ret, err := ED448().Scalar.SetBytes(sequence)
	require.NoError(t, err)
	require.Equal(t, ret.Cmp(sc), 0)

	// Try 10 random values
	for i := 0; i < 10; i++ {
		sc = ED448().Scalar.Random(crand.Reader)
		sequence = sc.Bytes()
		require.Equal(t, len(sequence), 57)
		ret, err = ED448().Scalar.SetBytes(sequence)
		require.NoError(t, err)
		require.Equal(t, ret.Cmp(sc), 0)
	}
}

func TestScalarEd448Nil(t *testing.T) {
	one := ED448().Scalar.New(1)
	require.Nil(t, one.Add(nil))
	require.Nil(t, one.Sub(nil))
	require.Nil(t, one.Mul(nil))
	require.Nil(t, one.Div(nil))
	require.Nil(t, ED448().Scalar.Random(nil))
	require.Equal(t, one.Cmp(nil), -2)
	_, err := ED448().Scalar.SetBigInt(nil)
	require.Error(t, err)
}

func TestPointEd448Random(t *testing.T) {
	// Try 10 random values
	for i := 0; i < 25; i++ {
		sc := ED448().Point.Random(crand.Reader)
		_, ok := sc.(*PointEd448)
		require.True(t, ok)
		require.True(t, !sc.IsIdentity())
	}
}

func TestPointEd448Hash(t *testing.T) {
	var b [57]byte
	sc := ED448().Point.Hash(b[:])
	s, ok := sc.(*PointEd448)
	require.True(t, ok)
	expected, _ := hex.DecodeString("1f3842fe9f6456b899c934711c03d756d9065e7d026a29a430f691d7ee952a36d122ee8fb8a34f77c6532a28af437c77679eac8031cf17b180")
	require.Equal(t, s.ToAffineCompressed(), expected)
}

func TestPointEd448Identity(t *testing.T) {
	sc := ED448().Point.Identity()
	require.True(t, sc.IsIdentity())
	require.Equal(t, sc.ToAffineCompressed(), []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

func TestPointEd448Generator(t *testing.T) {
	sc := ED448().Point.Generator()
	s, ok := sc.(*PointEd448)
	require.True(t, ok)
	require.Equal(t, s.X(), Edwards448Curve().Params().Gx)
	require.Equal(t, s.Y(), Edwards448Curve().Params().Gy)
}

func TestPointEd448Set(t *testing.T) {
	iden, err := ED448().Point.Set(big.NewInt(0), big.NewInt(1))
	require.NoError(t, err)
	require.True(t, iden.IsIdentity())
	_, err = ED448().Point.Set(Edwards448Curve().Params().Gx, Edwards448Curve().Params().Gy)
	require.NoError(t, err)
}

func TestPointEd448Double(t *testing.T) {
	g := ED448().Point.Generator()
	g2 := g.Double()
	require.True(t, g2.Equal(g.Mul(ED448().Scalar.New(2))))
	i := ED448().Point.Identity()
	require.True(t, i.Double().Equal(i))
}

func TestPointEd448Neg(t *testing.T) {
	g := ED448().Point.Generator().Neg()
	require.True(t, g.Neg().Equal(ED448().Point.Generator()))
	require.True(t, ED448().Point.Identity().Neg().Equal(ED448().Point.Identity()))
}

func TestPointEd448Add(t *testing.T) {
	pt := ED448().Point.Generator()
	require.True(t, pt.Add(pt).Equal(pt.Double()))
	require.True(t, pt.Mul(ED448().Scalar.New(3)).Equal(pt.Add(pt).Add(pt)))
}

func TestPointEd448Sub(t *testing.T) {
	g := ED448().Point.Generator()
	pt := ED448().Point.Generator().Mul(ED448().Scalar.New(4))
	require.True(t, pt.Sub(g).Sub(g).Sub(g).Equal(g))
	require.True(t, pt.Sub(g).Sub(g).Sub(g).Sub(g).IsIdentity())
}

func TestPointEd448Mul(t *testing.T) {
	g := ED448().Point.Generator()
	pt := ED448().Point.Generator().Mul(ED448().Scalar.New(4))
	require.True(t, g.Double().Double().Equal(pt))
}

func TestPointEd448Serialize(t *testing.T) {
	g := ED448().Point.Generator()

	// smoke test
	for i := 0; i < 25; i++ {
		s := ED448().Scalar.Random(crand.Reader)
		pt := g.Mul(s)
		cmprs := pt.ToAffineCompressed()
		require.Equal(t, len(cmprs), 57)
		retC, err := pt.FromAffineCompressed(cmprs)
		require.NoError(t, err)
		require.True(t, pt.Equal(retC))

		un := pt.ToAffineUncompressed()
		require.Equal(t, len(un), 112)
		retU, err := pt.FromAffineUncompressed(un)
		require.NoError(t, err)
		require.True(t, pt.Equal(retU))
	}
}

func TestPointEd448Nil(t *testing.T) {
	one := ED448().Point.Generator()
	require.Nil(t, one.Add(nil))
	require.Nil(t, one.Sub(nil))
	require.Nil(t, one.Mul(nil))
	require.Nil(t, ED448().Scalar.Random(nil))
	require.False(t, one.Equal(nil))
	_, err := ED448().Scalar.SetBigInt(nil)
	require.Error(t, err)
}

func TestPointEd448SumOfProducts(t *testing.T) {
	lhs := new(PointEd448).Generator().Mul(new(ScalarEd448).New(50))
	points := make([]Point, 5)
	for i := range points {
		points[i] = new(PointEd448).Generator()
	}
	scalars := []Scalar{
		new(ScalarEd448).New(8),
		new(ScalarEd448).New(9),
		new(ScalarEd448).New(10),
		new(ScalarEd448).New(11),
		new(ScalarEd448).New(12),
	}
	rhs := lhs.SumOfProducts(points, scalars)
	require.NotNil(t, rhs)
	require.True(t, lhs.Equal(rhs))
}
