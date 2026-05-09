package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainPrintsTranslationSources(t *testing.T) {
	// Capturing os.Stdout mutates process-wide state, so this test cannot run in parallel.
	got := captureStdout(t, main)
	want := "=== Lookup Examples ===\nKey: \"hello\"\n  Text:   \"你好，World！\"\n  Locale: \"zh-Hans\"\n  Source: \"direct\"\n\nKey: \"goodbye\"\n  Text:   \"Goodbye, World!\"\n  Locale: \"en\"\n  Source: \"fallback\"\n\nKey: \"unknown_key\"\n  Text:   \"unknown_key\"\n  Locale: \"en\"\n  Source: \"missing\"\n\n\n=== Detecting Fallback vs Direct Hit ===\n  \"hello\": direct hit in zh-Hans\n  \"goodbye\": fallback from en\n\n=== Context Disambiguation ===\n  GetX verb: 发布文章\n  GetX noun: 文章\nKey: \"Post <verb>\"\n  Text:   \"发布文章\"\n  Locale: \"zh-Hans\"\n  Source: \"direct\"\n\n"

	assert.Equal(t, want, got)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old })

	fn()
	require.NoError(t, w.Close())

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	require.NoError(t, r.Close())
	return buf.String()
}
