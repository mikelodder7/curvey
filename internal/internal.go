package internal

func ReverseBytes(inBytes []byte) []byte {
	outBytes := make([]byte, len(inBytes))

	for i, j := 0, len(inBytes)-1; j >= 0; i, j = i+1, j-1 {
		outBytes[i] = inBytes[j]
	}

	return outBytes
}

func IsZeroI(a int) int {
	return ((a | -a) >> 31) + 1
}

func IsNonZeroI(a int) int {
	return -((a | -a) >> 31)
}

func CtSelect(a, b int, choice int) int {
	mask := -choice
	return a ^ ((a ^ b) & mask)
}

func CtSelectUint64(a, b uint64, choice int) uint64 {
	mask := uint64(-int64(choice))
	return a ^ ((a ^ b) & mask)
}

func IsZeroUint64I(a uint64) int {
	return int(((int64(a) | int64(-a)) >> 63) + 1)
}

func IsNotZeroUint64I(a uint64) int {
	return -int((int64(a) | int64(-a)) >> 63)
}

func IsZeroArrayI(a []uint64) int {
	t := uint64(0)
	for _, v := range a {
		t |= v
	}
	return IsZeroUint64I(t)
}

func IsNotZeroArrayI(a []uint64) int {
	t := uint64(0)
	for _, v := range a {
		t |= v
	}
	return IsNotZeroUint64I(t)
}

func EqualI(lhs, rhs []uint64) int {
	t := uint64(0)
	for i, l := range lhs {
		t |= l ^ rhs[i]
	}
	return IsZeroUint64I(t)
}

func NotEqualI(lhs, rhs []uint64) int {
	t := uint64(0)
	for i, l := range lhs {
		t |= l ^ rhs[i]
	}
	return IsNotZeroUint64I(t)
}

func CmpI(lhs, rhs []uint64) int {
	gt := uint64(0)
	lt := uint64(0)
	for i := len(lhs) - 1; i >= 0; i-- {
		// convert to two 64-bit numbers where
		// the leading bits are zeros and hold no meaning
		//  so rhs - fp actually means gt
		// and fp - rhs actually means lt.
		rhsH := rhs[i] >> 32
		rhsL := rhs[i] & 0xffffffff
		lhsH := lhs[i] >> 32
		lhsL := lhs[i] & 0xffffffff

		// Check the leading bit
		// if negative then fp > rhs
		// if positive then fp < rhs
		gt |= (rhsH - lhsH) >> 32 & 1 &^ lt
		lt |= (lhsH - rhsH) >> 32 & 1 &^ gt
		gt |= (rhsL - lhsL) >> 32 & 1 &^ lt
		lt |= (lhsL - rhsL) >> 32 & 1 &^ gt
	}
	// Make the result -1 for <, 0 for =, 1 for >
	return int(gt) - int(lt)
}

func CmpBytesI(lhs, rhs []byte) int {
	gt := 0
	lt := 0
	for i := len(lhs) - 1; i >= 0; i-- {
		// convert to two numbers where
		// the leading bits are zeros and hold no meaning
		//  so rhs - fp actually means gt
		// and fp - rhs actually means lt.
		r := int(rhs[i])
		l := int(lhs[i])

		// Check the leading bit
		// if negative then fp > rhs
		// if positive then fp < rhs
		gt |= (r - l) >> 32 & 1 &^ lt
		lt |= (l - r) >> 32 & 1 &^ gt
	}
	// Make the result -1 for <, 0 for =, 1 for >
	return gt - lt
}
