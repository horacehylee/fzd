package fzd

import (
	"path/filepath"

	"github.com/horacehylee/fzd/ignorer"
	"github.com/horacehylee/fzd/walker"
	"github.com/stretchr/testify/assert"
)

func (suite *FilterTestSuite) TestTopAndDirFilters() {
	t := suite.T()

	root := suite.level0Dir
	for _, filters := range [][]Filter{
		{Top, Dir},
		{Dir, Top},
	} {
		suite.visited = nil

		fn, err := newFiltersWalkFunc(root, LocationOption{
			Filters: filters,
		})
		assert.NoError(t, err)

		fn = walker.Combine(fn, suite.visitedWalkFunc)
		walker.Walk(root, fn)
		assert.Equal(t, []string{
			suite.level0Dir,
			suite.level1Dir,
		}, suite.visited)
	}
}

func (suite *FilterTestSuite) TestTopAndNotDirFilters() {
	t := suite.T()

	root := suite.level0Dir
	for _, filters := range [][]Filter{
		{Top, NotDir},
		{NotDir, Top},
	} {
		suite.visited = nil

		fn, err := newFiltersWalkFunc(root, LocationOption{
			Filters: filters,
		})
		assert.NoError(t, err)

		fn = walker.Combine(fn, suite.visitedWalkFunc)
		walker.Walk(root, fn)
		assert.Equal(t, []string{
			suite.level0Dir,
			suite.level0File,
		}, suite.visited)
	}
}

func (suite *FilterTestSuite) TestDirAndNotDirFilters() {
	t := suite.T()

	root := suite.level0Dir
	for _, filters := range [][]Filter{
		{Dir, NotDir},
		{NotDir, Dir},
	} {
		suite.visited = nil

		fn, err := newFiltersWalkFunc(root, LocationOption{
			Filters: filters,
		})
		assert.NoError(t, err)

		fn = walker.Combine(fn, suite.visitedWalkFunc)
		walker.Walk(root, fn)
		assert.Equal(t, []string{
			suite.level0Dir,
		}, suite.visited)
	}
}

func (suite *FilterTestSuite) TestDirAndTopAndIgnoreFilters() {
	t := suite.T()

	root := suite.level0Dir
	for _, filters := range [][]Filter{
		{Dir, Top},
		{Top, Dir},
	} {
		suite.visited = nil

		fn, err := newFiltersWalkFunc(root, LocationOption{
			Filters: filters,
			Ignores: []interface{}{filepath.Base(suite.level0Dir)},
		})
		assert.NoError(t, err)

		fn = walker.Combine(fn, suite.visitedWalkFunc)
		walker.Walk(root, fn)
		assert.Empty(t, suite.visited)
	}
}

func (suite *FilterTestSuite) TestFilterWalkFuncFailForIgnorerFilter() {
	t := suite.T()

	root := suite.level0Dir
	_, err := newFiltersWalkFunc(root, LocationOption{
		Filters: []Filter{Dir, Top},
		Ignores: []interface{}{123},
	})
	assert.ErrorIs(t, err, ignorer.ErrTypeNotSupported)
}

func (suite *FilterTestSuite) TestFilterWalkFuncFailForUnknownFilter() {
	t := suite.T()

	root := suite.level0Dir
	_, err := newFiltersWalkFunc(root, LocationOption{
		Filters: []Filter{Dir, Top, "xyz"},
		Ignores: []interface{}{filepath.Base(suite.level0Dir)},
	})
	assert.EqualError(t, err, "\"xyz\" filter is not supported")
}
