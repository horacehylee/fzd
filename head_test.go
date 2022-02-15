package fzd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteAndReadHead(t *testing.T) {
	dir, err := os.MkdirTemp("", "testHead")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	indexName := "some index"
	err = writeHead(dir, indexName)
	assert.NoError(t, err)

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, HeadFileName, entries[0].Name())

	name, err := readHead(dir)
	assert.NoError(t, err)
	assert.Equal(t, indexName, name)
}

func TestWriteAndReadHeadForInvalidLocation(t *testing.T) {
	dir, err := os.MkdirTemp("", "testHead")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "invalid")

	_, err = readHead(path)
	assert.ErrorIs(t, err, ErrIndexHeadDoesNotExist)

	err = writeHead(path, "test")
	assert.Error(t, err)
}
