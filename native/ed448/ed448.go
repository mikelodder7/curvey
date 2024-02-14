package ed448

var (
	zero          = FpNew().SetZero()
	one           = FpNew().SetOne()
	minusOne      = FpNew().Neg(one)
	neg4XTwistedD = FpNew().SetLimbs(&[7]uint64{
		0x00000000000262a8,
		0x0000000000000000,
		0x0000000000000000,
		0x0000000000000000,
		0x0000000000000000,
		0x0000000000000000,
		0x0000000000000000,
	})
	edwardsD = FpNew().SetLimbs(&[7]uint64{
		0xffffffffffff6756,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xfffffffeffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
	})
	negEdwardsD = FpNew().Neg(edwardsD)
	twistedD    = FpNew().SetLimbs(&[7]uint64{
		0xffffffffffff6755,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xfffffffeffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
	})
	twoXTwistedD = FpNew().SetLimbs(&[7]uint64{
		0xfffffffffffeceab,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xfffffffeffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
	})
	decafFactor = FpNew().SetLimbs(&[7]uint64{
		0x9642ef0f45572736,
		0x60337bf6aa20ce52,
		0x839a66f4fd6eded2,
		0x64a2d780968c14ba,
		0xa1f1a7b8a5b8d54b,
		0x3bf68d722fa26aa0,
		0x22d962fbeb24f768,
	})
	aP2Div4 = FpNew().SetLimbs(&[7]uint64{
		0x00000000000098aa,
		0x0000000000000000,
		0x0000000000000000,
		0x0000000000000000,
		0x0000000000000000,
		0x0000000000000000,
		0x0000000000000000,
	})
	j    = FpNew().SetLimbs(&[7]uint64{156326, 0, 0, 0, 0, 0, 0})
	negJ = FpNew().Neg(j)
	z    = FpNew().SetLimbs(&[7]uint64{
		0xfffffffffffffffe,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xfffffffeffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0xffffffffffffffff,
	})
	twistedBasePoint = &TwistedExtendedPoint{
		X: FpNew().SetLimbs(&[7]uint64{
			0x0000000000000000,
			0x0000000000000000,
			0x0000000000000000,
			0xfffffffe80000000,
			0xffffffffffffffff,
			0xffffffffffffffff,
			0x7fffffffffffffff,
		}),
		Y: FpNew().SetLimbs(&[7]uint64{
			0xc86079b4dfdd4a64,
			0x199b0c1e3ab470a1,
			0x14181844d73f48e5,
			0x93d5242c50452714,
			0x05264370504c74c3,
			0x8d06c13078ca2408,
			0x8508de14f04286d4,
		}),
		Z: FpNew().SetOne(),
		T: FpNew().SetLimbs(&[7]uint64{
			0x93e3c816dc198105,
			0x140362071833f4e0,
			0x19c9854dde98e342,
			0x56382384a319b575,
			0xc2b86da60f794be9,
			0xe23d5682a9ffe1dd,
			0x6d3669e173c6a450,
		}),
	}
	sqrtExp = [7]uint64{
		0x0000000000000000,
		0x0000000000000000,
		0x0000000000000000,
		0xffffffffc0000000,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0x3fffffffffffffff,
	}
	basePointOrder = FqNew().SetLimbs(&[7]uint64{
		0x2378c292ab5844f3,
		0x216cc2728dc58f55,
		0xc44edb49aed63690,
		0xffffffff7cca23e9,
		0xffffffffffffffff,
		0xffffffffffffffff,
		0x3fffffffffffffff,
	})
)
