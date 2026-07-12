package i18n

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatCompiledRejectsNonStringResult(t *testing.T) {
	t.Parallel()

	_, err := formatCompiled(func(any) (any, error) {
		return []string{"one", "two"}, nil
	}, nil)

	require.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrUnsupported)
}
