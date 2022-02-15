package fzd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIndexer(t *testing.T) {
	basePath := "/test"
	i, err := NewIndexer(basePath)
	assert.NoError(t, err)

	assert.Equal(t, basePath, i.basePath)
	assert.Equal(t, 0, len(i.locations))
}

func TestNewIndexerEmptyBasePathReturnsError(t *testing.T) {
	_, err := NewIndexer("")
	assert.EqualError(t, err, "base path cannot be empty")
}

func TestNewIndexerWithLocation(t *testing.T) {
	basePath := "/test"
	locationPath := "/location"
	locationOption := LocationOption{}

	i, err := NewIndexer(basePath, WithLocation(locationPath, locationOption))
	assert.NoError(t, err)

	assert.Equal(t, basePath, i.basePath)
	assert.Equal(t, 1, len(i.locations))
	assert.Equal(t, map[string]LocationOption{
		locationPath: locationOption,
	}, i.locations)
}
