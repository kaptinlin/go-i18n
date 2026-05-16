package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainPrintsEmbeddedMessages(t *testing.T) {
	// Capturing os.Stdout mutates process-wide state, so this test cannot run in parallel.
	got := captureStdout(t, main)
	want := "Hello, world\nhello, John\n"

	assert.Equal(t, want, got)
}

func TestRunPrintsLoadErrorAndFallbackKeys(t *testing.T) {
	// Capturing os.Stdout mutates process-wide state, so this test cannot run in parallel.
	got := captureStdout(t, func() {
		run(fstest.MapFS{}, "[")
	})
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")

	require.Len(t, lines, 3)
	assert.NotEmpty(t, lines[0])
	assert.Equal(t, "hello_world", lines[1])
	assert.Equal(t, "hello_name", lines[2])
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
