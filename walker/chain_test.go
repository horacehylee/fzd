package walker_test

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

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

func TestChain(t *testing.T) {
	called := make([]string, 0)

	w1 := func(path string, info walker.FileInfo, err error) error {
		called = append(called, path)
		assert.Equal(t, len(called), 1)
		return nil
	}
	w2 := func(path string, info walker.FileInfo, err error) error {
		called = append(called, path)
		assert.Equal(t, len(called), 2)
		return nil
	}
	fn := walker.Chain(w1, w2)

	path := "/test"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)
	err := fn(path, fileInfo, nil)

	assert.NoError(t, err)
	assert.Equal(t, []string{path, path}, called)
}

func TestChainSecondIsSkippedIfFirstReturnedSkipThis(t *testing.T) {
	called := make([]string, 0)

	w1 := func(path string, info walker.FileInfo, err error) error {
		called = append(called, path)
		assert.Equal(t, len(called), 1)
		return walker.SkipThis
	}
	w2 := func(path string, info walker.FileInfo, err error) error {
		assert.Fail(t, "should not be called")
		return nil
	}
	fn := walker.Chain(w1, w2)

	path := "/test"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)
	err := fn(path, fileInfo, nil)

	assert.Equal(t, walker.SkipThis, err)
	assert.Equal(t, []string{path}, called)
}

func TestChainSecondIsNotSkippedIfFirstReturnedOtherError(t *testing.T) {
	e := errors.New("test error")
	called := make([]string, 0)

	w1 := func(path string, info walker.FileInfo, err error) error {
		called = append(called, path)
		assert.Equal(t, len(called), 1)
		assert.Nil(t, err)
		return e
	}
	w2 := func(path string, info walker.FileInfo, err error) error {
		called = append(called, path)
		assert.Equal(t, len(called), 2)
		assert.Equal(t, e, err)
		return err
	}
	fn := walker.Chain(w1, w2)

	path := "/test"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)
	err := fn(path, fileInfo, nil)

	assert.Equal(t, e, err)
	assert.Equal(t, []string{path, path}, called)
}

func TestChainSecondErrorIsReturnedInsteadOfFirstError(t *testing.T) {
	e1 := errors.New("test error 1")
	e2 := errors.New("test error 2")
	called := make([]string, 0)

	w1 := func(path string, info walker.FileInfo, err error) error {
		called = append(called, path)
		assert.Equal(t, len(called), 1)
		assert.Nil(t, err)
		return e1
	}
	w2 := func(path string, info walker.FileInfo, err error) error {
		called = append(called, path)
		assert.Equal(t, len(called), 2)
		assert.Equal(t, e1, err)
		return e2
	}
	fn := walker.Chain(w1, w2)

	path := "/test"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)
	err := fn(path, fileInfo, nil)

	assert.Equal(t, e2, err)
	assert.Equal(t, []string{path, path}, called)
}

func TestChainReturnSameErrorIfCombinedDifferently(t *testing.T) {
	e1 := errors.New("test error 1")
	e2 := errors.New("test error 2")
	e3 := errors.New("test error 3")

	type called struct {
		calls []string
	}

	w1 := func(c *called) walker.WalkFunc {
		return func(path string, info walker.FileInfo, err error) error {
			c.calls = append(c.calls, path)
			assert.Equal(t, len(c.calls), 1)
			assert.Nil(t, err)
			return e1
		}
	}
	w2 := func(c *called) walker.WalkFunc {
		return func(path string, info walker.FileInfo, err error) error {
			c.calls = append(c.calls, path)
			assert.Equal(t, len(c.calls), 2)
			assert.Equal(t, e1, err)
			return e2
		}
	}
	w3 := func(c *called) walker.WalkFunc {
		return func(path string, info walker.FileInfo, err error) error {
			c.calls = append(c.calls, path)
			assert.Equal(t, len(c.calls), 3)
			assert.Equal(t, e2, err)
			return e3
		}
	}

	called1 := &called{calls: make([]string, 0)}
	called2 := &called{calls: make([]string, 0)}
	called3 := &called{calls: make([]string, 0)}
	fn1 := walker.Chain(w1(called1), w2(called1), w3(called1))
	fn2 := walker.Chain(w1(called2), walker.Chain(w2(called2), w3(called2)))
	fn3 := walker.Chain(walker.Chain(w1(called3), w2(called3)), w3(called3))

	path := "/test"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)
	err1 := fn1(path, fileInfo, nil)
	err2 := fn2(path, fileInfo, nil)
	err3 := fn3(path, fileInfo, nil)

	assert.Equal(t, e3, err1)
	assert.Equal(t, err1, err2)
	assert.Equal(t, err2, err3)

	assert.Equal(t, []string{path, path, path}, called1.calls)
	assert.Equal(t, called1, called2)
	assert.Equal(t, called2, called3)
}

func TestChainSkipEquallyIfCombinedDifferentlyForFirst(t *testing.T) {
	type called struct {
		calls []string
	}

	w1 := func(c *called) walker.WalkFunc {
		return func(path string, info walker.FileInfo, err error) error {
			c.calls = append(c.calls, path)
			assert.Equal(t, len(c.calls), 1)
			assert.Nil(t, err)
			return walker.SkipThis
		}
	}
	w2 := func(c *called) walker.WalkFunc {
		return func(path string, info walker.FileInfo, err error) error {
			assert.Fail(t, "w2 should not be called")
			return nil
		}
	}
	w3 := func(c *called) walker.WalkFunc {
		return func(path string, info walker.FileInfo, err error) error {
			assert.Fail(t, "w3 should not be called")
			return nil
		}
	}

	called1 := &called{calls: make([]string, 0)}
	called2 := &called{calls: make([]string, 0)}
	called3 := &called{calls: make([]string, 0)}
	fn1 := walker.Chain(w1(called1), w2(called1), w3(called1))
	fn2 := walker.Chain(w1(called2), walker.Chain(w2(called2), w3(called2)))
	fn3 := walker.Chain(walker.Chain(w1(called3), w2(called3)), w3(called3))

	path := "/test"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)
	err1 := fn1(path, fileInfo, nil)
	err2 := fn2(path, fileInfo, nil)
	err3 := fn3(path, fileInfo, nil)

	assert.Equal(t, walker.SkipThis, err1)
	assert.Equal(t, err1, err2)
	assert.Equal(t, err2, err3)

	assert.Equal(t, []string{path}, called1.calls)
	assert.Equal(t, called1, called2)
	assert.Equal(t, called2, called3)
}

func TestChainSkipEquallyIfCombinedDifferentlyForSecond(t *testing.T) {
	e1 := errors.New("test error 1")

	type called struct {
		calls []string
	}

	w1 := func(c *called) walker.WalkFunc {
		return func(path string, info walker.FileInfo, err error) error {
			c.calls = append(c.calls, path)
			assert.Equal(t, len(c.calls), 1)
			assert.Nil(t, err)
			return e1
		}
	}
	w2 := func(c *called) walker.WalkFunc {
		return func(path string, info walker.FileInfo, err error) error {
			c.calls = append(c.calls, path)
			assert.Equal(t, len(c.calls), 2)
			assert.Equal(t, e1, err)
			return walker.SkipThis
		}
	}
	w3 := func(c *called) walker.WalkFunc {
		return func(path string, info walker.FileInfo, err error) error {
			assert.Fail(t, "w3 should not be called")
			return nil
		}
	}

	called1 := &called{calls: make([]string, 0)}
	called2 := &called{calls: make([]string, 0)}
	called3 := &called{calls: make([]string, 0)}
	fn1 := walker.Chain(w1(called1), w2(called1), w3(called1))
	fn2 := walker.Chain(w1(called2), walker.Chain(w2(called2), w3(called2)))
	fn3 := walker.Chain(walker.Chain(w1(called3), w2(called3)), w3(called3))

	path := "/test"
	fileInfo := newMockFileInfo(filepath.Base(path), fileMode, false)
	err1 := fn1(path, fileInfo, nil)
	err2 := fn2(path, fileInfo, nil)
	err3 := fn3(path, fileInfo, nil)

	assert.Equal(t, walker.SkipThis, err1)
	assert.Equal(t, err1, err2)
	assert.Equal(t, err2, err3)

	assert.Equal(t, []string{path, path}, called1.calls)
	assert.Equal(t, called1, called2)
	assert.Equal(t, called2, called3)
}
