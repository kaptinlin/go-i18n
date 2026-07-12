package i18n

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestBundle(tb testing.TB, defaultLocale string, options ...Option) *I18n {
	tb.Helper()

	bundle, err := NewBundle(defaultLocale, options...)
	require.NoError(tb, err)
	return bundle
}
