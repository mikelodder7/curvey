//
// SPDX-License-Identifier: Apache-2.0
//

package curvey

import (
	crand "crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mikelodder7/curvey/native"
	"github.com/mikelodder7/curvey/native/pasta/fp"
)

func TestPointPallasAddDoubleMul(t *testing.T) {
	curve := PALLAS()
	g := curve.NewGeneratorPoint()
	id := curve.NewIdentityPoint()
	require.True(t, g.Add(id).Equal(g))

	g2 := g.Add(g)
	require.True(t, g.Double().Equal(g2))
	g3 := g.Add(g2)
	s := curve.NewScalar().New(3).(*ScalarPallas)
	require.True(t, g3.Equal(g.Mul(s)))

	s.Value.SetUint64(4)
	g4 := g3.Add(g)
	require.True(t, g4.Equal(g2.Double()))
	require.True(t, g4.Equal(g.Mul(s)))
}

func TestPointPallasHash(t *testing.T) {
	curve := PALLAS()
	h0 := curve.Point.Hash(nil)
	require.True(t, h0.IsOnCurve())
	h1 := curve.Point.Hash([]byte{})
	require.True(t, h1.IsOnCurve())
	require.True(t, h0.Equal(h1))
	h2 := curve.Point.Hash([]byte{1})
	require.True(t, h2.IsOnCurve())
}

func TestPointPallasNeg(t *testing.T) {
	curve := PALLAS()
	g := curve.NewGeneratorPoint().Neg()
	require.True(t, g.Neg().Equal(curve.NewGeneratorPoint()))
	id := curve.NewIdentityPoint()
	require.True(t, id.Neg().Equal(id))
}

func TestPointPallasRandom(t *testing.T) {
	curve := PALLAS()
	a := curve.Point.Random(testRng()).(*PointPallas)
	require.NotNil(t, a.EllipticPoint.X)
	require.NotNil(t, a.EllipticPoint.Y)
	require.NotNil(t, a.EllipticPoint.Z)
	require.True(t, a.IsOnCurve())
	e := native.EllipticPoint{
		X: fp.PastaFpNew().SetRaw(&[native.Field4Limbs]uint64{
			0x7263083d01d4859c,
			0x65a03323b5a3d204,
			0xe71d73222b136668,
			0x1d1b1bcf1256b539,
		}),
		Y: fp.PastaFpNew().SetRaw(&[native.Field4Limbs]uint64{
			0x8cc2516ffe23e1bb,
			0x5418f941eeaca812,
			0x16c9af658a846f29,
			0x11c572091c418668,
		}),
		Z: fp.PastaFpNew().SetRaw(&[native.Field4Limbs]uint64{
			0xa879589adb77a88e,
			0x5444a531a19f2406,
			0x637ff77c51dda524,
			0x0369e90d219ce821,
		}),
	}
	require.Equal(t, 1, a.EllipticPoint.Equal(&e))
}

func TestPointPallasSerialize(t *testing.T) {
	curve := PALLAS()
	ss := curve.Scalar.Random(testRng()).(*ScalarPallas)
	g := curve.NewGeneratorPoint()

	ppt := g.Mul(ss)
	require.Equal(t, ppt.ToAffineCompressed(), []byte{0x1c, 0x6d, 0x47, 0x1f, 0x4a, 0x81, 0xcd, 0x8, 0x4e, 0xb3, 0x17, 0x9a, 0xcd, 0x17, 0xe2, 0x9a, 0x24, 0x69, 0xb, 0x4e, 0x69, 0x5f, 0x35, 0x1a, 0x92, 0x12, 0x95, 0xc9, 0xe6, 0xd3, 0x7a, 0x0})
	require.Equal(t, ppt.ToAffineUncompressed(), []byte{0x1c, 0x6d, 0x47, 0x1f, 0x4a, 0x81, 0xcd, 0x8, 0x4e, 0xb3, 0x17, 0x9a, 0xcd, 0x17, 0xe2, 0x9a, 0x24, 0x69, 0xb, 0x4e, 0x69, 0x5f, 0x35, 0x1a, 0x92, 0x12, 0x95, 0xc9, 0xe6, 0xd3, 0x7a, 0x0, 0x80, 0x5c, 0xa1, 0x56, 0x6d, 0x1b, 0x87, 0x5f, 0xb0, 0x2e, 0xae, 0x85, 0x4e, 0x86, 0xa9, 0xcd, 0xde, 0x37, 0x6a, 0xc8, 0x4a, 0x80, 0xf6, 0x43, 0xaa, 0xe6, 0x2c, 0x2d, 0x15, 0xdb, 0xda, 0x29})
	retP, err := curve.Point.FromAffineCompressed(ppt.ToAffineCompressed())
	require.NoError(t, err)
	require.True(t, ppt.Equal(retP))
	retP, err = curve.Point.FromAffineUncompressed(ppt.ToAffineUncompressed())
	require.NoError(t, err)
	require.True(t, ppt.Equal(retP))

	// smoke test
	for i := 0; i < 25; i++ {
		s := curve.Scalar.Random(crand.Reader).(*ScalarPallas)
		pt := g.Mul(s)
		cmprs := pt.ToAffineCompressed()
		require.Equal(t, len(cmprs), 32)
		retC, err := curve.Point.FromAffineCompressed(cmprs)
		require.NoError(t, err)
		require.True(t, pt.Equal(retC))

		un := pt.ToAffineUncompressed()
		require.Equal(t, len(un), 64)
		retU, err := curve.Point.FromAffineUncompressed(un)
		require.NoError(t, err)
		require.True(t, pt.Equal(retU))
	}
}

func TestPointPallasCMove(t *testing.T) {
	curve := PALLAS()
	a := curve.Point.Random(crand.Reader).(*PointPallas)
	b := curve.Point.Random(crand.Reader).(*PointPallas)
	c := curve.NewIdentityPoint().(*PointPallas)
	require.Equal(t, 1, c.EllipticPoint.CMove(a.EllipticPoint, b.EllipticPoint, 1).Equal(b.EllipticPoint))
	require.Equal(t, 1, c.EllipticPoint.CMove(a.EllipticPoint, b.EllipticPoint, 0).Equal(a.EllipticPoint))
}

func TestPointPallasSumOfProducts(t *testing.T) {
	curve := PALLAS()
	s := curve.NewScalar().New(50)
	lhs := curve.ScalarBaseMult(s)
	points := make([]Point, 5)
	for i := range points {
		points[i] = curve.NewGeneratorPoint().(*PointPallas)
	}
	scalars := []Scalar{
		new(ScalarPallas).New(8),
		new(ScalarPallas).New(9),
		new(ScalarPallas).New(10),
		new(ScalarPallas).New(11),
		new(ScalarPallas).New(12),
	}
	rhs := lhs.SumOfProducts(points, scalars)
	require.NotNil(t, rhs)
	require.True(t, lhs.Equal(rhs))
}
