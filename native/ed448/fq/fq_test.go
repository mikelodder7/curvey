package fq

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFq_Add(t *testing.T) {
	onehalf := FqNew().SetLimbs(&[7]uint64{
		2, 0, 0, 0, 0, 0, 0,
	})
	_, wasInverted := onehalf.Invert(onehalf)
	require.Equal(t, wasInverted, 1)
	fmt.Printf("%v", onehalf)
}
