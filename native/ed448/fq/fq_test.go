package fq

import (
	"fmt"
	"github.com/mikelodder7/curvey/internal"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestFq_Add(t *testing.T) {
	a := FqNew().SetOne()
	b := FqNew().SetOne()

	sum := FqNew().Add(a, b)
	exp := &Fq{Value: &internal.Field{Value: []uint64{0xe439eb6aa53dd868, 0xf499ec6b91d38556, 0xdd8925b2894e4b7e, 0x419aee0b1, 0, 0, 0}}}
	require.Equal(t, 1, exp.EqualI(sum))

	var f [7]uint64
	for i := 0; i < 20; i++ {
		for j := 0; j < a.Value.Params.Limbs; j++ {
			f[j] = rand.Uint64()
		}
		a.SetLimbs(&f)
		for j := 0; j < a.Value.Params.Limbs; j++ {
			f[j] = rand.Uint64()
		}
		b.SetLimbs(&f)
		sum.Add(a, b)

		require.Equal(t, 1, a.EqualI(sum.Sub(sum, b)))
	}
}

func TestFq_Neg(t *testing.T) {
	a := FqNew().SetOne()
	b := FqNew().Neg(a)

	require.Equal(t, 1, b.Add(b, a).IsZero())

	var f [7]uint64
	for i := 0; i < 20; i++ {
		for j := 0; j < a.Value.Params.Limbs; j++ {
			f[j] = rand.Uint64()
		}
		a.SetLimbs(&f)
		b.Neg(a)

		require.Equal(t, 1, a.Add(a, b).IsZero())
	}
}

func TestFq_Invert(t *testing.T) {
	o := FqNew().SetLimbs(&[7]uint64{
		7, 0, 0, 0, 0, 0, 0,
	})
	oneseventh, wasInverted := FqNew().Invert(o)
	require.Equal(t, wasInverted, 1)

	oneover := FqNew()
	a := FqNew().Mul(o, oneseventh)
	e := FqNew().SetOne()
	require.Equal(t, 1, e.EqualI(a))

	var f [7]uint64
	for i := 0; i < 20; i++ {
		for j := 0; j < o.Value.Params.Limbs; j++ {
			f[j] = rand.Uint64()
		}

		o.SetLimbs(&f)
		_, wasInverted = oneover.Invert(o)
		a.Mul(oneover, o)

		require.Equal(t, 1, wasInverted)
		require.Equal(t, 1, e.EqualI(a))
	}
}

func TestFq_Sqrt(t *testing.T) {
	a := FqNew().SetLimbs(&[7]uint64{16, 0, 0, 0, 0, 0, 0})
	sqrt, wasSquare := FqNew().Sqrt(a)
	require.Equal(t, 1, wasSquare)
	o := FqNew().Square(sqrt)
	require.Equal(t, 1, a.EqualI(o))

	var f [7]uint64
	for i := 0; i < 20; i++ {
		for j := 0; j < o.Value.Params.Limbs; j++ {
			f[j] = rand.Uint64()
		}

		a.SetLimbs(&f)
		_, wasSquare = sqrt.Sqrt(a)
		o.Square(sqrt)

		require.Equal(t, wasSquare, a.EqualI(o))
	}
}

func TestFq_Constants(t *testing.T) {
	onehalf := FqNew().SetUint64(2)
	onehalf.Invert(onehalf)
	onefourth := FqNew().SetUint64(4)
	onefourth.Invert(onefourth)

	fmt.Printf("%v\n%v\n", onehalf, onefourth)
	fmt.Printf("%v\n%v\n", oneHalf, one4th)
}
