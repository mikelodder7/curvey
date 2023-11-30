package pasta

import (
	"sync"

	"github.com/mikelodder7/curvey/native"
	"github.com/mikelodder7/curvey/native/pasta/fp"
)

var (
	pallasPointInitonce sync.Once
	pallasPointParams   native.EllipticPoint4Params
)

type PallasEllipticPoint struct {
	native.EllipticPoint4
}

func PointNew() *native.EllipticPoint4 {
	return &native.EllipticPoint4{
		X:          fp.PastaFpNew(),
		Y:          fp.PastaFpNew(),
		Z:          fp.PastaFpNew(),
		Params:     getPallasPointParams(),
		Arithmetic: &pallasPointArithmetic{},
	}
}

func pallasPointParamsInit() {
	pallasPointParams = native.EllipticPoint4Params{
		A:       fp.PastaFpNew().SetZero(),
		B:       fp.PastaFpNew().SetUint64(5),
		Gx:      fp.PastaFpNew().SetOne(),
		Gy:      fp.PastaFpNew().SetRaw(&[native.Field4Limbs]uint64{0x2f474795455d409d, 0xb443b9b74b8255d9, 0x270c412f2c9a5d66, 0x8e00f71ba43dd6b}),
		BitSize: 255,
		Name:    "pallas",
	}
}

func getPallasPointParams() *native.EllipticPoint4Params {
	pallasPointInitonce.Do(pallasPointParamsInit)
	return &pallasPointParams
}

type pallasPointArithmetic struct{}

func (pallasPointArithmetic) Hash(out *native.EllipticPoint4, hash *native.EllipticPointHasher, msg, dst []byte) error {
	var u []byte

	switch hash.Type() {
	case native.XMD:
		u = native.ExpandMsgXmd(hash, msg, dst, 128)
	case native.XOF:
		u = native.ExpandMsgXof(hash, msg, dst, 128)
	}
	q0 := PointNew()
	q1 := PointNew()
	r0 := PointNew()
	r1 := PointNew()
	var buf [64]byte
	copy(buf[:], u[:64])
	u0 := fp.PastaFpNew().SetBytesWide(&buf)
	copy(buf[:], u[64:])
	u1 := fp.PastaFpNew().SetBytesWide(&buf)

	mapSswu(q0, u0)
	mapSswu(q1, u1)
	isoMap(r0, q0)
	isoMap(r1, q1)
	out.Add(r0, r1)
	return nil
}

func (pallasPointArithmetic) Double(out, arg *native.EllipticPoint4) {
	var a, b, c, d, e, f, x, y, z [native.Field4Limbs]uint64
	u := arg.X.Arithmetic

	// essentially paraphrased https://github.com/MinaProtocol/c-reference-signer/blob/master/crypto.c#L306-L337
	u.Square(&a, &arg.X.Value)
	u.Square(&b, &arg.Y.Value)
	u.Square(&c, &b)
	u.Add(&x, &arg.X.Value, &b)
	u.Square(&y, &x)
	u.Sub(&z, &y, &a)
	u.Sub(&x, &z, &c)
	u.Add(&d, &x, &x)
	u.Add(&e, &a, &a)
	u.Add(&e, &e, &a)
	u.Square(&f, &e)
	u.Add(&y, &d, &d)
	u.Sub(&x, &f, &y)
	u.Sub(&y, &d, &x)
	u.Add(&f, &c, &c)
	u.Add(&f, &f, &c)
	u.Add(&f, &f, &c)
	u.Add(&f, &f, &c)
	u.Add(&f, &f, &c)
	u.Add(&f, &f, &c)
	u.Add(&f, &f, &c)
	u.Mul(&z, &e, &y)
	u.Sub(&y, &z, &f)
	u.Mul(&f, &arg.Y.Value, &arg.Z.Value)
	u.Add(&z, &f, &f)

	// If identity
	e1 := arg.Z.IsZero()
	u.Selectznz(&out.X.Value, &x, &arg.X.Value, e1)
	u.Selectznz(&out.Y.Value, &y, &arg.Y.Value, e1)
	u.Selectznz(&out.Z.Value, &z, &arg.Z.Value, e1)
}

func (p pallasPointArithmetic) Add(out, arg1, arg2 *native.EllipticPoint4) {
	e1 := arg1.Z.IsZero()
	e2 := arg2.Z.IsZero()

	var z1z1, z2z2, u1, u2, s1, s2, zero [native.Field4Limbs]uint64
	var h, i, j, r, v, x3, y3, z3, t1 [native.Field4Limbs]uint64
	darg1 := PointNew()
	a := arg1.X.Arithmetic

	a.Square(&z1z1, &arg1.Z.Value)
	a.Square(&z2z2, &arg2.Z.Value)
	a.Mul(&u1, &arg1.X.Value, &z2z2)
	a.Mul(&u2, &arg2.X.Value, &z1z1)
	a.Mul(&s1, &arg1.Y.Value, &z2z2)
	a.Mul(&s1, &s1, &arg2.Z.Value)
	a.Mul(&s2, &arg2.Y.Value, &z1z1)
	a.Mul(&s2, &s2, &arg1.Z.Value)

	t := (u1[0] ^ u2[0]) | (u1[1] ^ u2[1]) | (u1[2] ^ u2[2]) | (u1[3] ^ u2[3])
	// if u1 == u2
	e3 := int(((int64(t) | int64(-t)) >> 63) + 1)
	t = (s1[0] ^ s2[0]) | (s1[1] ^ s2[1]) | (s1[2] ^ s2[2]) | (s1[3] ^ s2[3])
	// if s1 == s2
	e4 := int(((int64(t) | int64(-t)) >> 63) + 1)
	p.Double(darg1, arg1)

	a.Sub(&h, &u2, &u1)
	a.Add(&i, &h, &h)
	a.Square(&i, &i)
	a.Mul(&j, &i, &h)
	a.Sub(&r, &s2, &s1)
	a.Add(&r, &r, &r)
	a.Mul(&v, &u1, &i)
	a.Square(&x3, &r)
	a.Sub(&x3, &x3, &j)
	a.Add(&t1, &v, &v)
	a.Sub(&x3, &x3, &t1)
	a.Mul(&s1, &s1, &j)
	a.Add(&s1, &s1, &s1)
	a.Sub(&t1, &v, &x3)
	a.Mul(&y3, &r, &t1)
	a.Sub(&y3, &y3, &s1)
	a.Add(&z3, &arg1.Z.Value, &arg2.Z.Value)
	a.Square(&z3, &z3)
	a.Sub(&z3, &z3, &z1z1)
	a.Sub(&z3, &z3, &z2z2)
	a.Mul(&z3, &z3, &h)

	// if arg1 == inf
	out.X.CMove(out.X, arg2.X, e1)
	out.Y.CMove(out.Y, arg2.Y, e1)
	out.Z.CMove(out.Z, arg2.Z, e1)
	// if arg2 == inf
	out.X.CMove(out.X, arg1.X, e2)
	out.Y.CMove(out.Y, arg1.Y, e2)
	out.Z.CMove(out.Z, arg1.Z, e2)
	e1 ^= 1
	e2 ^= 1
	// if u1 == u2 && s1 == s2
	out.X.CMove(out.X, darg1.X, e1&e2&e3&e4)
	out.Y.CMove(out.Y, darg1.Y, e1&e2&e3&e4)
	out.Z.CMove(out.Z, darg1.Z, e1&e2&e3&e4)
	// if u1 == u2 && s1 != s2
	e4 ^= 1
	a.Selectznz(&out.X.Value, &out.X.Value, &zero, e1&e2&e3&e4)
	a.Selectznz(&out.Y.Value, &out.Y.Value, &zero, e1&e2&e3&e4)
	a.Selectznz(&out.Z.Value, &out.Z.Value, &zero, e1&e2&e3&e4)
	// if u1 != u2
	e3 ^= 1
	a.Selectznz(&out.X.Value, &out.X.Value, &x3, e1&e2&e3)
	a.Selectznz(&out.Y.Value, &out.Y.Value, &y3, e1&e2&e3)
	a.Selectznz(&out.Z.Value, &out.Z.Value, &z3, e1&e2&e3)
}

func (pallasPointArithmetic) IsOnCurve(arg *native.EllipticPoint4) bool {
	var z2, z4, z6, x2, x3, lhs, rhs [native.Field4Limbs]uint64

	u := arg.X.Arithmetic
	u.Square(&z2, &arg.Z.Value)
	u.Square(&z4, &z2)
	u.Mul(&z6, &z4, &z2)
	u.Square(&x2, &arg.X.Value)
	u.Mul(&x3, &x2, &arg.X.Value)

	u.Square(&lhs, &arg.Y.Value)
	copy(rhs[:], arg.Params.B.Value[:])
	u.Mul(&rhs, &rhs, &z6)
	u.Add(&rhs, &rhs, &x3)
	t := (lhs[0] ^ rhs[0]) |
		(lhs[1] ^ rhs[1]) |
		(lhs[2] ^ rhs[2]) |
		(lhs[3] ^ rhs[3])
	e := int(((int64(t) | int64(-t)) >> 63) + 1)
	return arg.Z.IsZero()|e == 1
}

func (pallasPointArithmetic) ToAffine(out, arg *native.EllipticPoint4) {
	var wasInverted int
	var zero, x, y, z, zinv [native.Field4Limbs]uint64
	f := arg.X.Arithmetic

	f.Invert(&wasInverted, &zinv, &arg.Z.Value)
	f.Square(&z, &zinv)
	f.Mul(&zinv, &z, &zinv)
	f.Mul(&x, &arg.X.Value, &z)
	f.Mul(&y, &arg.Y.Value, &zinv)

	// If point at infinity this does nothing
	f.Selectznz(&x, &zero, &x, wasInverted)
	f.Selectznz(&y, &zero, &y, wasInverted)
	f.Selectznz(&z, &zero, &out.Z.Params.R, wasInverted)

	out.X.Value = x
	out.Y.Value = y
	out.Z.Value = z
}

func (pallasPointArithmetic) RhsEquation(out, x *native.Field4) {
	// Elliptic curve equation for pallas is: y^2 = x^3 + b
	out.Square(x)
	out.Mul(out, x)
	out.Add(out, getPallasPointParams().B)
}

func mapSswu(p *native.EllipticPoint4, u *native.Field4) {
	isoa := [native.Field4Limbs]uint64{0x7fc5d29077bb08de, 0x93090252cf122108, 0x49f63ff5da1145bb, 0x1c6d4f087137f0dc}
	isob := [native.Field4Limbs]uint64{0xf7f22478ffffec3d, 0xa6dec35433e1339b, 0xfffffffffffffd5a, 0x3fffffffffffffff}
	z := [native.Field4Limbs]uint64{0x1d2df02400000034, 0xf6571331e3a2999b, 0x0000000000000006, 0x0000000000000000}
	// c1 := new(fp.Fp).Neg(isoa)
	// c1.Invert(c1)
	// c1.Mul(isob, c1)
	c1 := [native.Field4Limbs]uint64{
		0x1ee770ce078456ec,
		0x48cfd64c2ce76be0,
		0x43d5774c0ab79e2f,
		0x23368d2bdce28cf3,
	}
	// c2 := new(fp.Fp).Neg(z)
	// c2.Invert(c2)
	c2 := [native.Field4Limbs]uint64{
		0x03df915f89d89d8a,
		0x8f1e8db09ef82653,
		0xd89d89d89d89d89d,
		0x1d89d89d89d89d89,
	}

	var u2, tv1, tv2, x1, x2, gx1, gx2, x, y [native.Field4Limbs]uint64
	var wasInverted, wasSquare int
	a := u.Arithmetic

	a.Square(&u2, &u.Value)
	a.Mul(&tv1, &z, &u2)
	a.Square(&tv2, &tv1)
	a.Add(&x1, &tv1, &tv2)
	a.Invert(&wasInverted, &x1, &x1)
	t := x1[0] ^ x1[1] ^ x1[2] ^ x1[3]
	e1 := int(((int64(t) | int64(-t)) >> 63) + 1)
	a.Add(&x1, &x1, &u.Params.R)
	a.Selectznz(&x1, &x1, &c2, e1)
	a.Mul(&x1, &x1, &c1)
	a.Square(&gx1, &x1)
	a.Add(&gx1, &gx1, &isoa)
	a.Mul(&gx1, &gx1, &x1)
	a.Add(&gx1, &gx1, &isob)
	a.Mul(&x2, &tv1, &x1)
	a.Mul(&tv2, &tv1, &tv2)
	a.Mul(&gx2, &gx1, &tv2)
	a.Sqrt(&wasSquare, &gx1, &gx1)
	a.Selectznz(&x, &x2, &x1, wasSquare)
	a.Sqrt(&wasInverted, &gx2, &gx2)
	a.Selectznz(&y, &gx2, &gx1, wasSquare)
	a.FromMontgomery(&tv1, &y)
	// if signs are the same t == 0, otherwise 1
	t = uint64(u.Bytes()[0]&1) ^ tv1[0]&1
	e3 := int(t)
	a.Neg(&tv2, &y)
	// negate if signs are not equal
	a.Selectznz(&y, &y, &tv2, e3)

	copy(p.X.Value[:], x[:])
	copy(p.Y.Value[:], y[:])
	p.Z.SetOne()
}

var isomapper = [13][native.Field4Limbs]uint64{
	{0xc6e037a01c71c71d, 0x130ac6c4e8b8fc2b, 0x0000000000000000, 0x4000000000000000},
	{0x4c6e64f2323d5cee, 0x501f41cfd25ec1f0, 0x05dee76e883f5ca7, 0x33183c981332cc59},
	{0x6a3ee7799df56376, 0x126b79ab78c7152f, 0x3260d1c7394f73d9, 0x3faf24198196224d},
	{0x6eeb22cb38e38d91, 0x857a8f27ada1851f, 0xffffffffffffffe9, 0x3fffffffffffffff},
	{0x7fa53608c4284457, 0xe12b216a243a01b1, 0x34d622e2ca3a41e0, 0xbda2158acc92f21},
	{0xa60b71c1a8e17d58, 0x762a4b7ccb2def32, 0x503620030b6137e1, 0x3778ceb0aad59e24},
	{0xcaaf22d9b425ed0a, 0xbc70754050aca717, 0xaaaaaaaaaaaaaaaa, 0x2aaaaaaaaaaaaaaa},
	{0x26373279191eae77, 0xa80fa0e7e92f60f8, 0x82ef73b7441fae53, 0x198c1e4c0999662c},
	{0x64b468a19883c203, 0x8760f32b499db7b0, 0xe0dabda7ccdfe1b, 0x29822f08307bd31b},
	{0xee9add1584bda1bb, 0x5fa29228d90933b, 0x5555555555555568, 0x1555555555555555},
	{0xc0e6983a63c6683, 0x62e3fe9d3afd7f18, 0xcf4134542f5762d1, 0x31c73205032dc6b1},
	{0xbfc7f36afaa47806, 0x1df1b07e4eefdb60, 0xf0a260092223a7a4, 0x266a6c120080da6c},
	{0x6d4ccfb000000870, 0x33aace8e7975d8dc, 0x0000000000000121, 0x0000000000000000},
}

// Implements a degree 3 isogeny map.
// The input and output are in Jacobian coordinates, using the method
// in "Avoiding inversions" [WB2019, section 4.3].
func isoMap(out, arg *native.EllipticPoint4) {
	var z [4][native.Field4Limbs]uint64
	var numX, divX, numY, divY, t, z0, x, y [native.Field4Limbs]uint64
	a := arg.X.Arithmetic
	a.Square(&z[0], &arg.Z.Value)     // z^2
	a.Mul(&z[1], &z[0], &arg.Z.Value) // z^3
	a.Square(&z[2], &z[0])            // z^4
	a.Square(&z[3], &z[1])            // z^6

	// ((iso[0] * x + iso[1] * z^2) * x + iso[2] * z^4) * x + iso[3] * z^6
	copy(numX[:], isomapper[0][:])
	a.Mul(&numX, &numX, &arg.X.Value)
	a.Mul(&t, &isomapper[1], &z[0])
	a.Add(&numX, &numX, &t)
	a.Mul(&numX, &numX, &arg.X.Value)
	a.Mul(&t, &isomapper[2], &z[2])
	a.Add(&numX, &numX, &t)
	a.Mul(&numX, &numX, &arg.X.Value)
	a.Mul(&t, &isomapper[3], &z[3])
	a.Add(&numX, &numX, &t)

	// (z^2 * x + iso[4] * z^4) * x + iso[5] * z^6
	copy(divX[:], z[0][:])
	a.Mul(&divX, &divX, &arg.X.Value)
	a.Mul(&t, &isomapper[4], &z[2])
	a.Add(&divX, &divX, &t)
	a.Mul(&divX, &divX, &arg.X.Value)
	a.Mul(&t, &isomapper[5], &z[3])
	a.Add(&divX, &divX, &t)

	// (((iso[6] * x + iso[7] * z2) * x + iso[8] * z4) * x + iso[9] * z6) * y
	copy(numY[:], isomapper[6][:])
	a.Mul(&numY, &numY, &arg.X.Value)
	a.Mul(&t, &isomapper[7], &z[0])
	a.Add(&numY, &numY, &t)
	a.Mul(&numY, &numY, &arg.X.Value)
	a.Mul(&t, &isomapper[8], &z[2])
	a.Add(&numY, &numY, &t)
	a.Mul(&numY, &numY, &arg.X.Value)
	a.Mul(&t, &isomapper[9], &z[3])
	a.Add(&numY, &numY, &t)
	a.Mul(&numY, &numY, &arg.Y.Value)

	// (((x + iso[10] * z2) * x + iso[11] * z4) * x + iso[12] * z6) * z3
	copy(divY[:], arg.X.Value[:])
	a.Mul(&t, &isomapper[10], &z[0])
	a.Add(&divY, &divY, &t)
	a.Mul(&divY, &divY, &arg.X.Value)
	a.Mul(&t, &isomapper[11], &z[2])
	a.Add(&divY, &divY, &t)
	a.Mul(&divY, &divY, &arg.X.Value)
	a.Mul(&t, &isomapper[12], &z[3])
	a.Add(&divY, &divY, &t)
	a.Mul(&divY, &divY, &z[1])

	a.Mul(&z0, &divX, &divY)
	a.Mul(&x, &numX, &divY)
	a.Mul(&x, &x, &z0)
	a.Mul(&y, &numY, &divX)
	a.Square(&t, &z0)
	a.Mul(&y, &y, &t)

	copy(out.X.Value[:], x[:])
	copy(out.Y.Value[:], y[:])
	copy(out.Z.Value[:], z0[:])
}
