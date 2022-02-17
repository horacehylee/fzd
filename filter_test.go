package fzd

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/horacehylee/fzd/ignorer"
	"github.com/horacehylee/fzd/walker"
	"github.com/stretchr/testify/assert"
)

type mockFileinfo struct {
	mode  fs.FileMode
	name  string
	isDir bool
}

func (m *mockFileinfo) Mode() fs.FileMode {
	return m.mode
}

func (m *mockFileinfo) Name() string {
	return m.name
}

func (m *mockFileinfo) IsDir() bool {
	return m.isDir
}

func newMockFileInfo(name string, mode fs.FileMode, isDir bool) *mockFileinfo {
	return &mockFileinfo{
		mode:  mode,
		name:  name,
		isDir: isDir,
	}
}

func TestTopFilterPassForTopLevelFile(t *testing.T) {
	root := "/level0"
	fn := withTopFilter(root)

	path := "/level0/level0.txt"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)

	err := fn(filepath.Clean(path), fileInfo, nil)
	assert.NoError(t, err)
}

func TestTopFilterPassForSamePath(t *testing.T) {
	root := "/level0"
	fn := withTopFilter(root)

	path := root
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, true)

	err := fn(filepath.Clean(path), fileInfo, nil)
	assert.NoError(t, err)
}

func TestTopFilterPassForTopLevelDirectory(t *testing.T) {
	root := "/level0"
	fn := withTopFilter(root)

	path := "/level0/level1"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, true)

	err := fn(filepath.Clean(path), fileInfo, nil)
	assert.NoError(t, err)
}

func TestTopFilterSkipIfNonTopLevelFile(t *testing.T) {
	root := "/level0"
	fn := withTopFilter(root)

	path := "/level0/level1/level1.txt"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)

	err := fn(filepath.Clean(path), fileInfo, nil)
	assert.Equal(t, walker.SkipThis, err)
}

func TestTopFilterReturnsErrorIfPassed(t *testing.T) {
	root := "/level0"
	fn := withTopFilter(root)

	path := "/level0/level0.txt"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)
	e := errors.New("test error")

	err := fn(filepath.Clean(path), fileInfo, e)
	assert.Equal(t, e, err)
}

func TestTopFilterReturnsErrorIfPassedForNonTopLevelFile(t *testing.T) {
	root := "/level0"
	fn := withTopFilter(root)

	path := "/level0/level1/level1.txt"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)

	e := errors.New("test error")
	err := fn(filepath.Clean(path), fileInfo, e)
	assert.Equal(t, e, err)
}

func TestDirFilterPassForDir(t *testing.T) {
	fn := withDirFilter()

	path := "/level0"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, true)

	err := fn(filepath.Clean(path), fileInfo, nil)
	assert.NoError(t, err)
}

func TestDirFilterSkipForFile(t *testing.T) {
	fn := withDirFilter()

	path := "/level0/level0.txt"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)

	err := fn(filepath.Clean(path), fileInfo, nil)
	assert.Equal(t, walker.SkipThis, err)
}

func TestDirFilterReturnsErrorIfPassed(t *testing.T) {
	fn := withDirFilter()

	path := "/level0"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, true)

	e := errors.New("test error")
	err := fn(filepath.Clean(path), fileInfo, e)
	assert.Equal(t, e, err)
}

func TestDirFilterReturnsErrorIfPassedForFile(t *testing.T) {
	fn := withDirFilter()

	path := "/level0/level0.txt"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)

	e := errors.New("test error")
	err := fn(filepath.Clean(path), fileInfo, e)
	assert.Equal(t, e, err)
}

func TestNotDirFilterForFile(t *testing.T) {
	root := "/level0"
	fn := withNotDirFilter(root)

	path := "/level0/level0.txt"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)

	err := fn(filepath.Clean(path), fileInfo, nil)
	assert.NoError(t, err)
}

func TestNotDirFilterForSamePath(t *testing.T) {
	root := "/level0"
	fn := withNotDirFilter(root)

	path := root
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, true)

	err := fn(filepath.Clean(path), fileInfo, nil)
	assert.NoError(t, err)
}

func TestNotDirFilterSkipForDir(t *testing.T) {
	root := "/level0"
	fn := withNotDirFilter(root)

	path := "/level0/level1"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, true)

	err := fn(filepath.Clean(path), fileInfo, nil)
	assert.Equal(t, walker.SkipThis, err)
}

func TestNotDirFilterReturnsErrorIfPassedForFile(t *testing.T) {
	root := "/level0"
	fn := withNotDirFilter(root)

	path := "/level0/level0.txt"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)

	e := errors.New("test error")
	err := fn(filepath.Clean(path), fileInfo, e)
	assert.Equal(t, e, err)
}

func TestNotDirFilterReturnsErrorIfPassedForDir(t *testing.T) {
	root := "/level0"
	fn := withNotDirFilter(root)

	path := "/level0/level1"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, true)

	e := errors.New("test error")
	err := fn(filepath.Clean(path), fileInfo, e)
	assert.Equal(t, e, err)
}

func TestIgnoreFilter(t *testing.T) {
	fn, err := withIgnoreFilter("[Ll]evel*.txt")
	assert.NoError(t, err)

	for _, path := range []string{
		"/level0/level0.txt",
		"/Level1.txt",
	} {
		fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)

		err = fn(filepath.Clean(path), fileInfo, nil)
		assert.Equal(t, walker.SkipThis, err)
	}
}

func TestIgnoreFilterPassed(t *testing.T) {
	fn, err := withIgnoreFilter("[Ll]evel")
	assert.NoError(t, err)

	for _, path := range []string{
		"/level0/level0.txt",
		"/Level1.txt",
	} {
		fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)

		err = fn(filepath.Clean(path), fileInfo, nil)
		assert.NoError(t, err)
	}
}

func TestIgnoreFilterReturnsErrorIfPassed(t *testing.T) {
	fn, err := withIgnoreFilter("[Ll]evel*.txt")
	assert.NoError(t, err)

	path := "/level0/level1"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, true)

	e := errors.New("test error")
	err = fn(filepath.Clean(path), fileInfo, e)
	assert.Equal(t, e, err)
}

func TestWithIgnoreFilterReturnsErrorIfNotStringRelated(t *testing.T) {
	_, err := withIgnoreFilter(123)
	assert.ErrorIs(t, err, ignorer.ErrTypeNotSupported)
}
