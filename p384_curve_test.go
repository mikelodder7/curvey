//
// SPDX-License-Identifier: Apache-2.0
//

package curvey

import (
	"crypto/elliptic"
	crand "crypto/rand"
	"encoding/hex"
	"github.com/mikelodder7/curvey/internal"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScalarP384Random(t *testing.T) {
	curve := P384()
	sc := curve.Scalar.Random(testRng())
	s, ok := sc.(*ScalarP384)
	require.True(t, ok)
	expected := bhex("95329d16399498b31af5c9e51f73f0170f99b14680ddb8bc1f72bd3fc5eb66682ca72b3e815f694d035337db04e2f015")
	require.Equal(t, s.value.BigInt(), expected)
	// Try 10 random values
	for i := 0; i < 10; i++ {
		sc := p384.Scalar.Random(crand.Reader)
		_, ok := sc.(*ScalarP384)
		require.True(t, ok)
		require.True(t, !sc.IsZero())
	}
}

func TestScalarP384Hash(t *testing.T) {
	var b [32]byte
	p384 := P384()
	sc := p384.Scalar.Hash(b[:])
	s, ok := sc.(*ScalarP384)
	require.True(t, ok)
	expected := bhex("351657d8c32a8c72a126865eb4e103cbefe4ccf072111bcc34abddbf45d169897cef74c988e6e40caa23748a79cd8238")
	require.Equal(t, s.value.BigInt(), expected)
}

func TestScalarP384Zero(t *testing.T) {
	p384 := P384()
	sc := p384.Scalar.Zero()
	require.True(t, sc.IsZero())
	require.True(t, sc.IsEven())
}

func TestScalarP384One(t *testing.T) {
	p384 := P384()
	sc := p384.Scalar.One()
	require.True(t, sc.IsOne())
	require.True(t, sc.IsOdd())
}

func TestScalarP384New(t *testing.T) {
	p384 := P384()
	three := p384.Scalar.New(3)
	require.True(t, three.IsOdd())
	four := p384.Scalar.New(4)
	require.True(t, four.IsEven())
	neg1 := p384.Scalar.New(-1)
	require.True(t, neg1.IsEven())
	neg2 := p384.Scalar.New(-2)
	require.True(t, neg2.IsOdd())
}

func TestScalarP384Square(t *testing.T) {
	p384 := P384()
	three := p384.Scalar.New(3)
	nine := p384.Scalar.New(9)
	require.Equal(t, three.Square().Cmp(nine), 0)
}

func TestScalarP384Cube(t *testing.T) {
	p384 := P384()
	three := p384.Scalar.New(3)
	twentySeven := p384.Scalar.New(27)
	require.Equal(t, three.Cube().Cmp(twentySeven), 0)
}

func TestScalarP384Double(t *testing.T) {
	p384 := P384()
	three := p384.Scalar.New(3)
	six := p384.Scalar.New(6)
	require.Equal(t, three.Double().Cmp(six), 0)
}

func TestScalarP384Neg(t *testing.T) {
	p384 := P384()
	one := p384.Scalar.One()
	neg1 := p384.Scalar.New(-1)
	require.Equal(t, one.Neg().Cmp(neg1), 0)
	lotsOfThrees := p384.Scalar.New(333333)
	expected := p384.Scalar.New(-333333)
	require.Equal(t, lotsOfThrees.Neg().Cmp(expected), 0)
}

func TestScalarP384Invert(t *testing.T) {
	p384 := P384()
	nine := p384.Scalar.New(9)
	actual, _ := nine.Invert()
	sa, _ := actual.(*ScalarP384)
	curve := elliptic.P384()

	bn := new(big.Int).SetInt64(9)
	bn.ModInverse(bn, curve.Params().N)

	expected, err := p384.Scalar.SetBigInt(bn)
	require.NoError(t, err)
	require.Equal(t, sa.Cmp(expected), 0)
}

func TestScalarP384Sqrt(t *testing.T) {
	p384 := P384()
	nine := p384.Scalar.New(9)
	actual, err := nine.Sqrt()
	sa, _ := actual.(*ScalarP384)
	expected, _ := p384.Scalar.SetBytes([]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 199, 99, 77, 129, 244, 55, 45, 223, 88, 26, 13, 178, 72, 176, 167, 122, 236, 236, 25, 106, 204, 197, 41, 112})
	require.NoError(t, err)
	require.Equal(t, sa.Cmp(expected), 0)
}

func TestScalarP384Add(t *testing.T) {
	p384 := P384()
	nine := p384.Scalar.New(9)
	six := p384.Scalar.New(6)
	fifteen := nine.Add(six)
	require.NotNil(t, fifteen)
	expected := p384.Scalar.New(15)
	require.Equal(t, expected.Cmp(fifteen), 0)
	n := new(big.Int).Set(elliptic.P384().Params().N)
	n.Sub(n, big.NewInt(3))

	upper, err := p384.Scalar.SetBigInt(n)
	require.NoError(t, err)
	actual := upper.Add(nine)
	require.NotNil(t, actual)
	require.Equal(t, actual.Cmp(six), 0)
}

func TestScalarP384Sub(t *testing.T) {
	p384 := P384()
	nine := p384.Scalar.New(9)
	six := p384.Scalar.New(6)
	n := new(big.Int).Set(elliptic.P384().Params().N)
	n.Sub(n, big.NewInt(3))

	expected, err := p384.Scalar.SetBigInt(n)
	require.NoError(t, err)
	actual := six.Sub(nine)
	require.Equal(t, expected.Cmp(actual), 0)

	actual = nine.Sub(six)
	require.Equal(t, actual.Cmp(p384.Scalar.New(3)), 0)
}

func TestScalarP384Mul(t *testing.T) {
	p384 := P384()
	nine := p384.Scalar.New(9)
	six := p384.Scalar.New(6)
	actual := nine.Mul(six)
	require.Equal(t, actual.Cmp(p384.Scalar.New(54)), 0)
	n := new(big.Int).Set(elliptic.P384().Params().N)
	n.Sub(n, big.NewInt(1))
	upper, err := p384.Scalar.SetBigInt(n)
	require.NoError(t, err)
	require.Equal(t, upper.Mul(upper).Cmp(p384.Scalar.New(1)), 0)
}

func TestScalarP384Div(t *testing.T) {
	p384 := P384()
	nine := p384.Scalar.New(9)
	actual := nine.Div(nine)
	require.Equal(t, actual.Cmp(p384.Scalar.New(1)), 0)
	require.Equal(t, p384.Scalar.New(54).Div(nine).Cmp(p384.Scalar.New(6)), 0)
}

func TestScalarP384Serialize(t *testing.T) {
	p384 := P384()
	sc := p384.Scalar.New(255)
	sequence := sc.Bytes()
	require.Equal(t, len(sequence), 48)
	require.Equal(t, sequence, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff})
	ret, err := p384.Scalar.SetBytes(sequence)
	require.NoError(t, err)
	require.Equal(t, ret.Cmp(sc), 0)

	// Try 10 random values
	for i := 0; i < 10; i++ {
		sc = p384.Scalar.Random(crand.Reader)
		sequence = sc.Bytes()
		require.Equal(t, len(sequence), 48)
		ret, err = p384.Scalar.SetBytes(sequence)
		require.NoError(t, err)
		require.Equal(t, ret.Cmp(sc), 0)
	}
}

func TestScalarP384Nil(t *testing.T) {
	p384 := P384()
	one := p384.Scalar.New(1)
	require.Nil(t, one.Add(nil))
	require.Nil(t, one.Sub(nil))
	require.Nil(t, one.Mul(nil))
	require.Nil(t, one.Div(nil))
	require.Nil(t, p384.Scalar.Random(nil))
	require.Equal(t, one.Cmp(nil), -2)
	_, err := p384.Scalar.SetBigInt(nil)
	require.Error(t, err)
}

func TestPointP384Random(t *testing.T) {
	p384 := P384()
	sc := p384.Point.Random(testRng())
	s, ok := sc.(*PointP384)
	require.True(t, ok)
	expectedX, _ := new(big.Int).SetString("49b27651d5520340b66fdccffc9dcf3edcebbd0e9599ba1df6df218a637d193d3da35317ee34858109f6bce30bbffcf1", 16)
	expectedY, _ := new(big.Int).SetString("452195e873427e05f57db477ce66f7b623d87fbb25f0b34b682f6b4a61ac3d7ef547d2f1e2c6748de2c0e6fd1c692049", 16)
	require.Equal(t, s.X().BigInt(), expectedX)
	require.Equal(t, s.Y().BigInt(), expectedY)
	// Try 10 random values
	for i := 0; i < 10; i++ {
		sc := p384.Point.Random(crand.Reader)
		_, ok := sc.(*PointP384)
		require.True(t, ok)
		require.True(t, !sc.IsIdentity())
	}
}

func TestPointP384Hash(t *testing.T) {
	var b [32]byte
	p384 := P384()
	sc := p384.Point.Hash(b[:])
	s, ok := sc.(*PointP384)
	require.True(t, ok)

	expectedX, _ := new(big.Int).SetString("23a66d02546d8587c64a44510ad044069312e79f68b563d9c5ba1b390f108c171a61f71607e80ccf0ff1cc863ff3ff0d", 16)
	expectedY, _ := new(big.Int).SetString("886d7320a86f149c9154db6c252e050d7f2f12156a0fc1c60053107c7414395e061a4c8e6a22be857703e2568f46ee34", 16)
	require.Equal(t, s.X().BigInt(), expectedX)
	require.Equal(t, s.Y().BigInt(), expectedY)
}

func TestPointP384Identity(t *testing.T) {
	p384 := P384()
	sc := p384.Point.Identity()
	require.True(t, sc.IsIdentity())
	require.Equal(t, sc.ToAffineCompressed(), []byte{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
}

func TestPointP384Generator(t *testing.T) {
	p384 := P384()
	sc := p384.Point.Generator()
	s, ok := sc.(*PointP384)
	require.True(t, ok)
	require.Equal(t, s.X().BigInt(), elliptic.P384().Params().Gx)
	require.Equal(t, s.Y().BigInt(), elliptic.P384().Params().Gy)
}

func TestPointP384Set(t *testing.T) {
	p384 := P384()
	iden, err := p384.Point.Set(big.NewInt(0), big.NewInt(0))
	require.NoError(t, err)
	require.True(t, iden.IsIdentity())
	_, err = p384.Point.Set(elliptic.P384().Params().Gx, elliptic.P384().Params().Gy)
	require.NoError(t, err)
}

func TestPointP384Double(t *testing.T) {
	p384 := P384()
	g := p384.Point.Generator()
	g2 := g.Double()
	require.True(t, g2.Equal(g.Mul(p384.Scalar.New(2))))
	i := p384.Point.Identity()
	require.True(t, i.Double().Equal(i))
}

func TestPointP384Neg(t *testing.T) {
	p384 := P384()
	g := p384.Point.Generator().Neg()
	require.True(t, g.Neg().Equal(p384.Point.Generator()))
	require.True(t, p384.Point.Identity().Neg().Equal(p384.Point.Identity()))
}

func TestPointP384Add(t *testing.T) {
	p384 := P384()
	pt := p384.Point.Generator()
	require.True(t, pt.Add(pt).Equal(pt.Double()))
	require.True(t, pt.Mul(p384.Scalar.New(3)).Equal(pt.Add(pt).Add(pt)))
}

func TestPointP384Sub(t *testing.T) {
	p384 := P384()
	g := p384.Point.Generator()
	pt := p384.Point.Generator().Mul(p384.Scalar.New(4))
	require.True(t, pt.Sub(g).Sub(g).Sub(g).Equal(g))
	require.True(t, pt.Sub(g).Sub(g).Sub(g).Sub(g).IsIdentity())
}

func TestPointP384Mul(t *testing.T) {
	p384 := P384()
	g := p384.Point.Generator()
	pt := p384.Point.Generator().Mul(p384.Scalar.New(4))
	require.True(t, g.Double().Double().Equal(pt))

	u1Bytes, _ := hex.DecodeString("f6f388f3fa75856ca10b1b8e978115322942d4b3f8e1d72c97911a29144283e396f511e55eb918cbd1b72d2e0bd015d3")
	u1, _ := p384.NewScalar().SetBytes(u1Bytes)
	expectedX, _ := hex.DecodeString("adeab2a5fc286e5254f573988e9098b734442fff35a42c9c5c0c1920155841086278878b32fba4d2afd54f3bb8f67188")
	lhs := p384.ScalarBaseMult(u1)
	p := lhs.(*PointP384)
	x := p.X().Bytes()
	require.Equal(t, x[:], internal.ReverseScalarBytes(expectedX))
}

func TestPointP384Serialize(t *testing.T) {
	p384 := P384()
	ss := p384.Scalar.Random(testRng())
	g := p384.Point.Generator()

	ppt := g.Mul(ss)

	require.Equal(t, ppt.ToAffineCompressed(), []byte{0x2, 0xed, 0x9d, 0x68, 0x1a, 0xf3, 0xfd, 0x33, 0xbe, 0x4f, 0xe7, 0xcd, 0xd, 0xfa, 0x20, 0x39, 0x86, 0xa9, 0x95, 0x3, 0xb8, 0xf1, 0x21, 0x65, 0xea, 0xdb, 0xc3, 0xd4, 0x1, 0x4b, 0x78, 0xda, 0x66, 0xac, 0x4c, 0x15, 0x9a, 0xe5, 0x4b, 0xaf, 0x93, 0x90, 0x8a, 0x21, 0x43, 0x84, 0x63, 0x46, 0xd5})
	require.Equal(t, ppt.ToAffineUncompressed(), []byte{0x4, 0xed, 0x9d, 0x68, 0x1a, 0xf3, 0xfd, 0x33, 0xbe, 0x4f, 0xe7, 0xcd, 0xd, 0xfa, 0x20, 0x39, 0x86, 0xa9, 0x95, 0x3, 0xb8, 0xf1, 0x21, 0x65, 0xea, 0xdb, 0xc3, 0xd4, 0x1, 0x4b, 0x78, 0xda, 0x66, 0xac, 0x4c, 0x15, 0x9a, 0xe5, 0x4b, 0xaf, 0x93, 0x90, 0x8a, 0x21, 0x43, 0x84, 0x63, 0x46, 0xd5, 0xdb, 0xd6, 0xf9, 0x8e, 0xaa, 0xa6, 0x22, 0x0, 0x49, 0xc2, 0x47, 0xea, 0x75, 0x6e, 0x60, 0x4c, 0xdc, 0x1d, 0xad, 0xe5, 0x5d, 0xb8, 0x99, 0x9a, 0xa0, 0xcd, 0x28, 0xf6, 0x59, 0xa7, 0xb2, 0x52, 0xdd, 0x36, 0x3e, 0x91, 0xa5, 0x71, 0xd8, 0x23, 0x69, 0x4b, 0x92, 0xbd, 0x8b, 0x20, 0x65, 0x3c})
	retP, err := ppt.FromAffineCompressed(ppt.ToAffineCompressed())
	require.NoError(t, err)
	require.True(t, ppt.Equal(retP))
	retP, err = ppt.FromAffineUncompressed(ppt.ToAffineUncompressed())
	require.NoError(t, err)
	require.True(t, ppt.Equal(retP))

	// smoke test
	for i := 0; i < 25; i++ {
		s := p384.Scalar.Random(crand.Reader)
		pt := g.Mul(s)
		cmprs := pt.ToAffineCompressed()
		require.Equal(t, len(cmprs), 49)
		retC, err := pt.FromAffineCompressed(cmprs)
		require.NoError(t, err)
		require.True(t, pt.Equal(retC))

		un := pt.ToAffineUncompressed()
		require.Equal(t, len(un), 97)
		retU, err := pt.FromAffineUncompressed(un)
		require.NoError(t, err)
		require.True(t, pt.Equal(retU))
	}
}

func TestPointP384Nil(t *testing.T) {
	p384 := P384()
	one := p384.Point.Generator()
	require.Nil(t, one.Add(nil))
	require.Nil(t, one.Sub(nil))
	require.Nil(t, one.Mul(nil))
	require.Nil(t, p384.Scalar.Random(nil))
	require.False(t, one.Equal(nil))
	_, err := p384.Scalar.SetBigInt(nil)
	require.Error(t, err)
}

func TestPointP384SumOfProducts(t *testing.T) {
	lhs := new(PointP384).Generator().Mul(new(ScalarP384).New(50))
	points := make([]Point, 5)
	for i := range points {
		points[i] = new(PointP384).Generator()
	}
	scalars := []Scalar{
		new(ScalarP384).New(8),
		new(ScalarP384).New(9),
		new(ScalarP384).New(10),
		new(ScalarP384).New(11),
		new(ScalarP384).New(12),
	}
	rhs := lhs.SumOfProducts(points, scalars)
	require.NotNil(t, rhs)
	require.True(t, lhs.Equal(rhs))
}
