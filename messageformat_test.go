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

func TestMessageFormatPreservesRegionalPluralRules(t *testing.T) {
	t.Parallel()

	const message = "{count, plural, one {one} other {other}}"
	tests := []struct {
		locale string
		want   string
	}{
		{locale: "pt", want: "one"},
		{locale: "pt-PT", want: "other"},
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			t.Parallel()

			bundle := newTestBundle(t, tt.locale)
			localizer := bundle.NewLocalizer(tt.locale)

			dynamic, err := localizer.Format(message, Vars{"count": 0})
			require.NoError(t, err)
			assert.Equal(t, tt.want, dynamic)

			require.NoError(t, bundle.LoadMessages(map[string]map[string]string{
				tt.locale: {"items": message},
			}))
			assert.Equal(t, tt.want, localizer.Get("items", Vars{"count": 0}))
		})
	}
}
