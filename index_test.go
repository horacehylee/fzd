package fzd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type indexerCall struct {
	id   string
	data interface{}
}

type mockIndexer struct {
	calls []indexerCall
}

func (m *mockIndexer) Index(id string, data interface{}) error {
	m.calls = append(m.calls, indexerCall{
		id:   id,
		data: data,
	})
	return nil
}

func TestIndexWalkFuncIndexFileNameForBothIdandData(t *testing.T) {
	i := new(mockIndexer)
	fn := newIndexWalkFunc(i)

	path := filepath.Clean("/level0/level0.txt")
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)

	err := fn(path, fileInfo, nil)
	assert.NoError(t, err)

	assert.Equal(t, []indexerCall{
		{id: path, data: path},
	}, i.calls)
}

func TestIndexWalkFuncReturnsErrorIfPassed(t *testing.T) {
	i := new(mockIndexer)
	fn := newIndexWalkFunc(i)

	path := filepath.Clean("/level0/level0.txt")
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)
	e := errors.New("test error")

	err := fn(path, fileInfo, e)
	assert.Equal(t, e, err)
	assert.Empty(t, i.calls)
}

func TestRemoveIndexesExceptHeadAndCurrent(t *testing.T) {
	dir, err := os.MkdirTemp("", "testRemoveIndexes")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	index1 := filepath.Join(dir, "index1")
	err = os.Mkdir(index1, fileMode)
	assert.NoError(t, err)

	index1File := filepath.Join(dir, "index1", "index1.txt")
	err = os.WriteFile(index1File, []byte("content"), fileMode)
	assert.NoError(t, err)

	index2 := filepath.Join(dir, "index2")
	err = os.Mkdir(index2, fileMode)
	assert.NoError(t, err)

	index2File := filepath.Join(dir, "index2", "index2.txt")
	err = os.WriteFile(index2File, []byte("content"), fileMode)
	assert.NoError(t, err)

	head := filepath.Join(dir, HeadFileName)
	err = os.WriteFile(head, []byte("content"), fileMode)
	assert.NoError(t, err)

	err = removeIndexesExclude(dir, "index2")
	assert.NoError(t, err)

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(entries))

	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	assert.ElementsMatch(t, []string{"index2", HeadFileName}, names)
}
