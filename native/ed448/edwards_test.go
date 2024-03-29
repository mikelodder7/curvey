package ed448

import (
	"encoding/hex"
	"github.com/mikelodder7/curvey/internal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEdwardsPoint_IsOnCurve(t *testing.T) {
	xBytes, _ := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa955555555555555555555555555555555555555555555555555555555")
	yBytes, _ := hex.DecodeString("ae05e9634ad7048db359d6205086c2b0036ed7a035884dd7b7e36d728ad8c4b80d6565833a2a3098bbbcb2bed1cda06bdaeafbcdea9386ed")

	x, _ := FpNew().SetBytes(internal.ReverseBytes(xBytes))
	y, _ := FpNew().SetBytes(internal.ReverseBytes(yBytes))

	gen := (&AffinePoint{x, y}).ToEdwards()
	require.Equal(t, 1, gen.IsOnCurve())
}
