package p384_test

import (
	"encoding/hex"
	"fmt"
	"github.com/mikelodder7/curvey/internal"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mikelodder7/curvey/native"
	"github.com/mikelodder7/curvey/native/p384"
	"github.com/mikelodder7/curvey/native/p384/fp"
)

func TestP384PointArithmetic_Double(t *testing.T) {
	g := p384.PointNew().Generator()
	pt1 := p384.PointNew().Double(g)
	pt2 := p384.PointNew().Add(g, g)
	pt3 := p384.PointNew().Mul(g, fp.P384FpNew().SetUint64(2))

	e1 := pt1.Equal(pt2)
	e2 := pt1.Equal(pt3)
	e3 := pt2.Equal(pt3)
	require.Equal(t, 1, e1)
	require.Equal(t, 1, e2)
	require.Equal(t, 1, e3)
}

func TestP384PointArithmetic_Hash(t *testing.T) {
	var b [32]byte
	sc, err := p384.PointNew().Hash(b[:], native.EllipticPointHasherSha256())

	ss := p384.PointNew()
	ss.ToAffine(sc)

	x := ss.X.Bytes()
	y := ss.Y.Bytes()
	fmt.Printf("x = %s\n", hex.EncodeToString(internal.ReverseScalarBytes(x[:])))
	fmt.Printf("y = %s\n", hex.EncodeToString(internal.ReverseScalarBytes(y[:])))

	//sc1 := curvey.P384().NewIdentityPoint().Hash(b[:])
	//fmt.Printf("%v\n", sc1)
	//
	require.NoError(t, err)
	require.True(t, !sc.IsIdentity())
	require.True(t, sc.IsOnCurve())
}
