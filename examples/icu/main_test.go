package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainPrintsICUFormattingExamples(t *testing.T) {
	// Capturing os.Stdout mutates process-wide state, so this test cannot run in parallel.
	got := captureStdout(t, main)
	want := "Message\nMessages\nNo messages\n1 message\n2 messages\nNo messages\n1 message\n2 messages\nHe replied to your message.\nShe replied to your message.\nThey replied to your message.\nThe 1st floor.\nThe 2nd floor.\nThe 3rd floor.\nThe 4th floor.\n"

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
