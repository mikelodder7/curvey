package ed448

import (
	"bytes"
	"encoding/hex"
	"github.com/mikelodder7/curvey/internal"
	"github.com/mikelodder7/curvey/native"
	"github.com/mikelodder7/curvey/native/ed448/fp"
	"github.com/mikelodder7/curvey/native/ed448/fq"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func getGen() *EdwardsPoint {
	xBytes, _ := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa955555555555555555555555555555555555555555555555555555555")
	yBytes, _ := hex.DecodeString("ae05e9634ad7048db359d6205086c2b0036ed7a035884dd7b7e36d728ad8c4b80d6565833a2a3098bbbcb2bed1cda06bdaeafbcdea9386ed")
	x, _ := fp.FpNew().SetBytes(internal.ReverseBytes(xBytes))
	y, _ := fp.FpNew().SetBytes(internal.ReverseBytes(yBytes))

	return (&AffinePoint{x, y}).ToEdwards()
}

func TestEdwardsPoint_Mul(t *testing.T) {
	g := EdwardsPointNew().SetGenerator()
	b := EdwardsPointNew().Mul(g, fq.FqNew().SetOne())
	require.Equal(t, 1, g.EqualI(b))

	b.Mul(g, fq.FqNew().SetUint64(3))
	g3 := EdwardsPointNew().Double(g)
	g3.Add(g3, g)
	require.Equal(t, 1, g3.EqualI(b))

	s := fq.FqNew().SetRaw(&[7]uint64{
		0x9448501F5F760C7F,
		0xB1D62D289FF632A5,
		0xFB7259020A4BF8C6,
		0x8D644FABF13585C7,
		0x423A2889BDAA49B2,
		0xA3D48A3A88A74625,
		0x0421191999A3072D,
	})
	b.Mul(g, s)
	s.Invert(s)
	b.Mul(b, s)
	a := EdwardsPointNew().SetGenerator()
	require.Equal(t, 1, a.EqualI(b))

	var f [7]uint64

	for i := 0; i < 20; i++ {
		for j := 0; j < s.Value.Params.Limbs; j++ {
			f[j] = rand.Uint64()
		}
		s.SetLimbs(&f)

		b.Mul(a, s)

		s.Invert(s)

		b.Mul(b, s)
		require.Equal(t, 1, a.EqualI(b))
	}
}

func TestEdwardsPoint_IsOnCurve(t *testing.T) {
	gen := getGen()
	require.Equal(t, 1, gen.IsOnCurve())
}

func TestEdwardsPoint_Compress(t *testing.T) {
	gen := getGen()
	decompressed, err := gen.Compress().Decompress()
	require.NoError(t, err)
	require.Equal(t, 1, decompressed.EqualI(gen))
}

func TestEdwardsPoint_Isogeny(t *testing.T) {
	gen := getGen()
	a := gen.ToTwisted().ToUntwisted()
	gen.Double(gen)
	gen.Double(gen)
	require.Equal(t, 1, a.EqualI(gen))
}

func TestEdwardsPoint_IsTorsionFree(t *testing.T) {
	require.Equal(t, 1, EdwardsPointNew().SetGenerator().IsTorsionFree())
	require.Equal(t, 1, EdwardsPointNew().SetIdentity().IsTorsionFree())

	bytes, _ := hex.DecodeString("13b6714c7a5f53101bbec88f2f17cd30f42e37fae363a5474efb4197ed6005df5861ae178a0c2c16ad378b7befed0d0904b7ced35e9f674180")
	compressed := CompressedEdwardsY(bytes)
	_, err := compressed.Decompress()
	require.Error(t, err)
}

func TestEdwardsPoint_Hash(t *testing.T) {
	dst := []byte("QUUX-V01-CS02-with-edwards448_XOF:SHAKE256_ELL2_RO_")
	msgs := [][]byte{
		[]byte(""), hexit("73036d4a88949c032f01507005c133884e2f0d81f9a950826245dda9e844fc78186c39daaa7147ead3e462cff60e9c6340b58134480b4d17"), hexit("94c1d61b43728e5d784ef4fcb1f38e1075f3aef5e99866911de5a234f1aafdc26b554344742e6ba0420b71b298671bbeb2b7736618634610"),
		[]byte("abc"), hexit("4e0158acacffa545adb818a6ed8e0b870e6abc24dfc1dc45cf9a052e98469275d9ff0c168d6a5ac7ec05b742412ee090581f12aa398f9f8c"), hexit("894d3fa437b2d2e28cdc3bfaade035430f350ec5239b6b406b5501da6f6d6210ff26719cad83b63e97ab26a12df6dec851d6bf38e294af9a"),
		[]byte("abcdef0123456789"), hexit("2c25b4503fadc94b27391933b557abdecc601c13ed51c5de68389484f93dbd6c22e5f962d9babf7a39f39f994312f8ca23344847e1fbf176"), hexit("d5e6f5350f430e53a110f5ac7fcc82a96cb865aeca982029522d32601e41c042a9dfbdfbefa2b0bdcdc3bc58cca8a7cd546803083d3a8548"),
		[]byte("q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"), hexit("a1861a9464ae31249a0e60bf38791f3663049a3f5378998499a83292e159a2fecff838eb9bc6939e5c6ae76eb074ad4aae39b55b72ca0b9a"), hexit("580a2798c5b904f8adfec5bd29fb49b4633cd9f8c2935eb4a0f12e5dfa0285680880296bb729c6405337525fb5ed3dff930c137314f60401"),
		[]byte("a512_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), hexit("987c5ac19dd4b47835466a50b2d9feba7c8491b8885a04edf577e15a9f2c98b203ec2cd3e5390b3d20bba0fa6fc3eecefb5029a317234401"), hexit("5e273fcfff6b007bb6771e90509275a71ff1480c459ded26fc7b10664db0a68aaa98bc7ecb07e49cf05b80ae5ac653fbdd14276bbd35ccbc"),
	}

	hasher := native.EllipticPointHasherShake256()
	for i := 0; i < len(msgs); i += 3 {
		p := EdwardsPointNew().Hash(hasher, msgs[i], dst)
		require.Equal(t, 1, p.IsOnCurve())

		pp := p.ToAffine()
		x := pp.X.Bytes()
		xx := internal.ReverseBytes(msgs[i+1])
		y := pp.Y.Bytes()
		yy := internal.ReverseBytes(msgs[i+2])
		require.Equal(t, 0, bytes.Compare(xx, x[:]))
		require.Equal(t, 0, bytes.Compare(yy, y[:]))
	}
}

func hexit(s string) []byte {
	bb, _ := hex.DecodeString(s)
	return bb
}
