package internal

import "math/big"

var jubjubQ = new(big.Int).SetBytes([]byte{
	0x0e, 0x7d, 0xb4, 0xea, 0x65, 0x33, 0xaf, 0xa9, 0x06, 0x67, 0x3b, 0x01, 0x01, 0x34, 0x3b, 0x00, 0xa6, 0x68, 0x20, 0x93, 0xcc, 0xc8, 0x10, 0x82, 0xd0, 0x97, 0x0e, 0x5e, 0xd6, 0xf7, 0x2c, 0xb7,
})

var jubjubP = new(big.Int).SetBytes([]byte{
	0x73, 0xed, 0xa7, 0x53, 0x29, 0x9d, 0x7d, 0x48, 0x33, 0x39, 0xd8, 0x08, 0x09, 0xa1, 0xd8, 0x05, 0x53, 0xbd, 0xa4, 0x02, 0xff, 0xfe, 0x5b, 0xfe, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x01,
})
