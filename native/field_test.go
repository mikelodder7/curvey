package native

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCmpSelf(t *testing.T) {
	f := new(Field)
	// TODO: generate random field element instead of hardcode
	f.SetRaw(&[FieldLimbs]uint64{18071070103467571798, 11787850505799426140, 10631355976141928593, 4867785203635092610})

	require.Equal(t, 0, f.Cmp(f))
}
