package i18n

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleVars() {
	v := Vars{"name": "Alice", "count": 3}
	fmt.Println(v["name"])
	fmt.Println(v["count"])
	// Output:
	// Alice
	// 3
}

func TestVarsMapConversion(t *testing.T) {
	v := Vars{"name": "Alice", "count": 3}

	// Vars should be directly convertible to map[string]any.
	m := map[string]any(v)
	assert.Equal(t, "Alice", m["name"])
	assert.Equal(t, 3, m["count"])
}

func TestVarsNil(t *testing.T) {
	var v Vars
	assert.Nil(t, v)

	// A nil Vars should convert to a nil map[string]any.
	m := map[string]any(v)
	assert.Nil(t, m)
}

func TestVarsEmpty(t *testing.T) {
	v := Vars{}
	assert.NotNil(t, v)
	assert.Empty(t, v)
}

func TestVarsMixedValueTypes(t *testing.T) {
	v := Vars{
		"string": "hello",
		"int":    42,
		"float":  3.14,
		"bool":   true,
		"nil":    nil,
		"slice":  []int{1, 2, 3},
	}
	assert.Len(t, v, 6)

	tests := []struct {
		key  string
		want any
	}{
		{"string", "hello"},
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
		{"nil", nil},
		{"slice", []int{1, 2, 3}},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, v[tt.key], "key %q", tt.key)
	}
}

func TestVarsKeyLookup(t *testing.T) {
	v := Vars{"exists": "value"}

	val, ok := v["exists"]
	assert.True(t, ok)
	assert.Equal(t, "value", val)

	val, ok = v["missing"]
	assert.False(t, ok)
	assert.Nil(t, val)
}
