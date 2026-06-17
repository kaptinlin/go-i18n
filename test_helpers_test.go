package i18n

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestBundle(tb testing.TB, options ...Option) *I18n {
	tb.Helper()

	bundle, err := NewBundle(options...)
	require.NoError(tb, err)
	return bundle
}
