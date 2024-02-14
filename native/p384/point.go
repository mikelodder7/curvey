package p384

import (
	"sync"

	"github.com/mikelodder7/curvey/internal"
	"github.com/mikelodder7/curvey/native"
	"github.com/mikelodder7/curvey/native/p384/fp"
)

var (
	p384PointInitonce     sync.Once
	p384PointParams       native.EllipticPoint6Params
	p384PointSswuInitOnce sync.Once
	p384PointSswuParams   native.Sswu6Params
)

func PointNew() *native.EllipticPoint6 {
	return &native.EllipticPoint6{
		X:          fp.P384FpNew(),
		Y:          fp.P384FpNew(),
		Z:          fp.P384FpNew(),
		Params:     getPointParams(),
		Arithmetic: &pointArithmetic{},
	}
}

func pointParamsInit() {
	// How these values were derived
	// left for informational purposes
	//params := elliptic.P384().Params()
	//a := big.NewInt(-3)
	//a.Mod(a, params.P)
	//capA := fp.P384FpNew().SetBigInt(a)
	//capB := fp.P384FpNew().SetBigInt(params.B)
	//gx := fp.P384FpNew().SetBigInt(params.Gx)
	//gy := fp.P384FpNew().SetBigInt(params.Gy)

	p384PointParams = native.EllipticPoint6Params{
		A: fp.P384FpNew().SetRaw(&[native.Field6Limbs]uint64{
			0x00000003fffffffc,
			0xfffffffc00000000,
			0xfffffffffffffffb,
			0xffffffffffffffff,
			0xffffffffffffffff,
			0xffffffffffffffff,
		}),
		B: fp.P384FpNew().SetRaw(&[native.Field6Limbs]uint64{
			0x081188719d412dcc,
			0xf729add87a4c32ec,
			0x77f2209b1920022e,
			0xe3374bee94938ae2,
			0xb62b21f41f022094,
			0xcd08114b604fbff9,
		}),
		Gx: fp.P384FpNew().SetRaw(&[native.Field6Limbs]uint64{
			0x3dd0756649c0b528,
			0x20e378e2a0d6ce38,
			0x879c3afc541b4d6e,
			0x6454868459a30eff,
			0x812ff723614ede2b,
			0x4d3aadc2299e1513,
		}),
		Gy: fp.P384FpNew().SetRaw(&[native.Field6Limbs]uint64{
			0x23043dad4b03a4fe,
			0xa1bfa8bf7bb4a9ac,
			0x8bade7562e83b050,
			0xc6c3521968f4ffd9,
			0xdd8002263969a840,
			0x2b78abc25a15c5e9,
		}),
		BitSize: 384,
		Name:    "P384",
	}
}

func getPointParams() *native.EllipticPoint6Params {
	p384PointInitonce.Do(pointParamsInit)
	return &p384PointParams
}

func getPointSswuParams() *native.Sswu6Params {
	p384PointSswuInitOnce.Do(pointSswuParamsInit)
	return &p384PointSswuParams
}

func pointSswuParamsInit() {
	// How these values were derived
	// left for informational purposes
	//params := elliptic.P384().Params()
	//
	//// c1 = (q - 3) / 4
	//c1 := new(big.Int).Set(params.P)
	//c1.Sub(c1, big.NewInt(3))
	//c1.Rsh(c1, 2)
	//
	//a := big.NewInt(-3)
	//a.Mod(a, params.P)
	//b := new(big.Int).Set(params.B)
	//z := big.NewInt(-12)
	//z.Mod(z, params.P)
	//// sqrt(-Z^3)
	//zTmp := new(big.Int).Exp(z, big.NewInt(3), nil)
	//zTmp = zTmp.Neg(zTmp)
	//zTmp.Mod(zTmp, params.P)
	//c2 := new(big.Int).ModSqrt(zTmp, params.P)
	//
	//var capC1Bytes [48]byte
	//c1.FillBytes(capC1Bytes[:])
	//capC1 := fp.P384FpNew().SetRaw(&[native.Field6Limbs]uint64{
	//	binary.BigEndian.Uint64(capC1Bytes[40:48]),
	//	binary.BigEndian.Uint64(capC1Bytes[32:40]),
	//	binary.BigEndian.Uint64(capC1Bytes[24:32]),
	//	binary.BigEndian.Uint64(capC1Bytes[16:24]),
	//	binary.BigEndian.Uint64(capC1Bytes[8:16]),
	//	binary.BigEndian.Uint64(capC1Bytes[:8]),
	//})
	//capC2 := fp.P384FpNew().SetBigInt(c2)
	//capA := fp.P384FpNew().SetBigInt(a)
	//capB := fp.P384FpNew().SetBigInt(b)
	//capZ := fp.P384FpNew().SetBigInt(z)

	p384PointSswuParams = native.Sswu6Params{
		C1: [native.Field6Limbs]uint64{
			0x000000003fffffff,
			0xbfffffffc0000000,
			0xffffffffffffffff,
			0xffffffffffffffff,
			0xffffffffffffffff,
			0x3fffffffffffffff,
		},
		C2: [native.Field6Limbs]uint64{
			0x5a79354f07af57aa,
			0xe75a4ed1906b8b55,
			0x7588d991f122de6d,
			0x186bd88f5904d48e,
			0xb1e7aa2d06efdb00,
			0x1abba906c2b08b2a,
		},
		A: [native.Field6Limbs]uint64{
			0x00000003fffffffc,
			0xfffffffc00000000,
			0xfffffffffffffffb,
			0xffffffffffffffff,
			0xffffffffffffffff,
			0xffffffffffffffff,
		},
		B: [native.Field6Limbs]uint64{
			0x81188719d412dcc,
			0xf729add87a4c32ec,
			0x77f2209b1920022e,
			0xe3374bee94938ae2,
			0xb62b21f41f022094,
			0xcd08114b604fbff9,
		},
		Z: [native.Field6Limbs]uint64{
			0x0000000cfffffff3,
			0xfffffff300000000,
			0xfffffffffffffff2,
			0xffffffffffffffff,
			0xffffffffffffffff,
			0xffffffffffffffff,
		},
	}
}

type pointArithmetic struct{}

func (p pointArithmetic) Hash(out *native.EllipticPoint6, hash *native.EllipticPointHasher, msg, dst []byte) error {
	var u []byte
	sswuParams := getPointSswuParams()

	switch hash.Type() {
	case native.XMD:
		u = native.ExpandMsgXmd(hash, msg, dst, 144)
	case native.XOF:
		u = native.ExpandMsgXof(hash, msg, dst, 144)
	}
	var buf [96]byte
	copy(buf[:72], internal.ReverseBytes(u[:72]))
	u0 := fp.P384FpNew().SetBytesWide(&buf)
	copy(buf[:72], internal.ReverseBytes(u[72:]))
	u1 := fp.P384FpNew().SetBytesWide(&buf)

	q0x, q0y := sswuParams.Osswu3mod4(u0)
	q1x, q1y := sswuParams.Osswu3mod4(u1)

	out.X = q0x
	out.Y = q0y
	out.Z.SetOne()
	tv := &native.EllipticPoint6{
		X:          q1x,
		Y:          q1y,
		Z:          fp.P384FpNew().SetOne(),
		Params:     getPointParams(),
		Arithmetic: out.Arithmetic,
	}
	p.Add(out, out, tv)
	return nil
}

func (p *pointArithmetic) Double(out, arg *native.EllipticPoint6) {
	//curve := elliptic.P384()
	//affine := PointNew()
	//affine.ToAffine(arg)
	//x, y := affine.X.BigInt(), affine.Y.BigInt()
	//
	//x, y = curve.Double(x, y)
	//
	//out.X.SetBigInt(x)
	//out.Y.SetBigInt(y)
	//out.Z.SetOne()

	xx := fp.P384FpNew()
	yy := fp.P384FpNew()
	zz := fp.P384FpNew()
	t := fp.P384FpNew()
	xy2 := fp.P384FpNew()
	yz2 := fp.P384FpNew()
	xz2 := fp.P384FpNew()
	bzzPart := fp.P384FpNew()
	bzz3Part := fp.P384FpNew()
	yyMBzz3 := fp.P384FpNew()
	yyPBzz3 := fp.P384FpNew()
	yFrag := fp.P384FpNew()
	xFrag := fp.P384FpNew()
	zz3 := fp.P384FpNew()
	bxz2Part := fp.P384FpNew()
	bxz6Part := fp.P384FpNew()
	xx3Mzz3 := fp.P384FpNew()

	xx.Square(arg.X)
	yy.Square(arg.Y)
	zz.Square(arg.Z)

	xy2.Mul(arg.X, arg.Y)
	xy2.Double(xy2)

	yz2.Mul(arg.Y, arg.Z)
	yz2.Double(yz2)

	xz2.Mul(arg.X, arg.Z)
	xz2.Double(xz2)

	bzzPart.Mul(arg.Params.B, zz)
	bzzPart.Sub(bzzPart, xz2)

	bzz3Part.Double(bzzPart)
	bzz3Part.Add(bzz3Part, bzzPart)

	yyMBzz3.Sub(yy, bzz3Part)
	yyPBzz3.Add(yy, bzz3Part)

	yFrag.Mul(yyPBzz3, yyMBzz3)
	xFrag.Mul(yyMBzz3, xy2)

	zz3.Double(zz)
	zz3.Add(zz3, zz)

	t.Add(zz3, xx)
	bxz2Part.Mul(arg.Params.B, xz2)
	bxz2Part.Sub(bxz2Part, t)

	bxz6Part.Double(bxz2Part)
	bxz6Part.Add(bxz6Part, bxz2Part)

	xx3Mzz3.Double(xx)
	xx3Mzz3.Add(xx3Mzz3, xx)
	xx3Mzz3.Sub(xx3Mzz3, zz3)

	t.Mul(bxz6Part, yz2)
	out.X.Sub(xFrag, t)

	t.Mul(xx3Mzz3, bxz6Part)
	out.Y.Add(yFrag, t)

	out.Z.Mul(yz2, yy)
	out.Z.Double(out.Z)
	out.Z.Double(out.Z)
}

func (p *pointArithmetic) Add(out, arg1, arg2 *native.EllipticPoint6) {
	//curve := elliptic.P384()
	//affine1 := PointNew()
	//affine1.ToAffine(arg1)
	//x1, y1 := affine1.X.BigInt(), affine1.Y.BigInt()
	//
	//affine2 := PointNew()
	//affine2.ToAffine(arg2)
	//x2, y2 := affine2.X.BigInt(), affine2.Y.BigInt()
	//
	//x, y := curve.Add(x1, y1, x2, y2)
	//
	//out.X.SetBigInt(x)
	//out.Y.SetBigInt(y)
	//out.Z.SetOne()

	xx := fp.P384FpNew()
	yy := fp.P384FpNew()
	zz := fp.P384FpNew()
	t := fp.P384FpNew()
	tt := fp.P384FpNew()
	xyPairs := fp.P384FpNew()
	yzPairs := fp.P384FpNew()
	xzPairs := fp.P384FpNew()
	bzzPart := fp.P384FpNew()
	bzz3Part := fp.P384FpNew()
	yyMBzz3 := fp.P384FpNew()
	yyPBzz3 := fp.P384FpNew()
	zz3 := fp.P384FpNew()
	bxzPart := fp.P384FpNew()
	bxz3Part := fp.P384FpNew()
	xx3Mzz3 := fp.P384FpNew()

	xx.Mul(arg1.X, arg2.X)
	yy.Mul(arg1.Y, arg2.Y)
	zz.Mul(arg1.Z, arg2.Z)

	t.Add(xx, yy)
	tt.Add(arg2.X, arg2.Y)
	xyPairs.Add(arg1.X, arg1.Y)
	xyPairs.Mul(xyPairs, tt)
	xyPairs.Sub(xyPairs, t)

	t.Add(yy, zz)
	tt.Add(arg2.Y, arg2.Z)
	yzPairs.Add(arg1.Y, arg1.Z)
	yzPairs.Mul(yzPairs, tt)
	yzPairs.Sub(yzPairs, t)

	t.Add(xx, zz)
	tt.Add(arg2.X, arg2.Z)
	xzPairs.Add(arg1.X, arg1.Z)
	xzPairs.Mul(xzPairs, tt)
	xzPairs.Sub(xzPairs, t)

	t.Mul(arg1.Params.B, zz)
	bzzPart.Sub(xzPairs, t)

	bzz3Part.Double(bzzPart)
	bzz3Part.Add(bzz3Part, bzzPart)

	yyMBzz3.Sub(yy, bzz3Part)
	yyPBzz3.Add(yy, bzz3Part)

	zz3.Double(zz)
	zz3.Add(zz3, zz)

	t.Add(zz3, xx)
	bxzPart.Mul(arg1.Params.B, xzPairs)
	bxzPart.Sub(bxzPart, t)

	bxz3Part.Double(bxzPart)
	bxz3Part.Add(bxz3Part, bxzPart)

	xx3Mzz3.Double(xx)
	xx3Mzz3.Add(xx3Mzz3, xx)
	xx3Mzz3.Sub(xx3Mzz3, zz3)

	t.Mul(yzPairs, bxz3Part)
	out.X.Mul(yyPBzz3, xyPairs)
	out.X.Sub(out.X, t)

	t.Mul(xx3Mzz3, bxz3Part)
	out.Y.Mul(yyPBzz3, yyMBzz3)
	out.Y.Add(out.Y, t)

	t.Mul(xyPairs, xx3Mzz3)
	out.Z.Mul(yyMBzz3, yzPairs)
	out.Z.Add(out.Z, t)
}

func (p pointArithmetic) IsOnCurve(arg *native.EllipticPoint6) bool {
	affine := PointNew()
	p.ToAffine(affine, arg)

	lhs := fp.P384FpNew().Square(affine.Y)
	rhs := fp.P384FpNew()
	p.RhsEquation(rhs, affine.X)

	return lhs.Equal(rhs) == 1
}

func (pointArithmetic) ToAffine(out, arg *native.EllipticPoint6) {
	var wasInverted int
	var zero, x, y, z [native.Field6Limbs]uint64
	f := arg.X.Arithmetic

	f.Invert(&wasInverted, &z, &arg.Z.Value)
	f.Mul(&x, &arg.X.Value, &z)
	f.Mul(&y, &arg.Y.Value, &z)

	out.Z.SetOne()
	// If point at infinity this does nothing
	f.Selectznz(&x, &zero, &x, wasInverted)
	f.Selectznz(&y, &zero, &y, wasInverted)
	f.Selectznz(&z, &zero, &out.Z.Value, wasInverted)

	out.X.Value = x
	out.Y.Value = y
	out.Z.Value = z
	out.Params = arg.Params
	out.Arithmetic = arg.Arithmetic
}

func (pointArithmetic) RhsEquation(out, x *native.Field6) {
	// Elliptic curve equation for p256 is: y^2 = x^3 + ax + b
	out.Square(x)
	out.Add(out, getPointParams().A)
	out.Mul(out, x)
	out.Add(out, getPointParams().B)
}
