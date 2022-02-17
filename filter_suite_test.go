package fzd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/horacehylee/fzd/walker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	fileMode = 0700
)

type FilterTestSuite struct {
	suite.Suite
	level0Dir       string
	level0File      string
	level1Dir       string
	level1File      string
	level2Dir       string
	level2File      string
	visited         []string
	visitedWalkFunc walker.WalkFunc
}

func TestFilterTestSuite(t *testing.T) {
	suite.Run(t, new(FilterTestSuite))
}

func (suite *FilterTestSuite) SetupSuite() {
	t := suite.T()

	level0Dir, err := os.MkdirTemp("", "level0")
	assert.NoError(t, err)
	suite.level0Dir = level0Dir

	suite.level0File = filepath.Join(suite.level0Dir, "level0.txt")
	err = os.WriteFile(suite.level0File, []byte("content"), fileMode)
	assert.NoError(t, err)

	suite.level1Dir = filepath.Join(suite.level0Dir, "level1")
	err = os.Mkdir(suite.level1Dir, fileMode)
	assert.NoError(t, err)

	suite.level1File = filepath.Join(suite.level1Dir, "level1.txt")
	err = os.WriteFile(suite.level1File, []byte("content"), fileMode)
	assert.NoError(t, err)

	suite.level2Dir = filepath.Join(suite.level1Dir, "level2")
	err = os.Mkdir(suite.level2Dir, fileMode)
	assert.NoError(t, err)

	suite.level2File = filepath.Join(suite.level2Dir, "level2.txt")
	err = os.WriteFile(suite.level2File, []byte("content"), fileMode)
	assert.NoError(t, err)
}

func (suite *FilterTestSuite) SetupTest() {
	suite.visited = nil
	suite.visitedWalkFunc = func(path string, info walker.FileInfo, err error) error {
		suite.visited = append(suite.visited, path)
		return nil
	}
}

func (suite *FilterTestSuite) TearDownSuite() {
	t := suite.T()

	if suite.level0Dir != "" {
		err := os.RemoveAll(suite.level0Dir)
		assert.NoError(t, err)
	}
}

func (suite *FilterTestSuite) TestTopFilter() {
	t := suite.T()

	root := suite.level0Dir
	fn := walker.Chain(
		withTopFilter(root),
		suite.visitedWalkFunc,
	)
	walker.Walk(root, fn)
	assert.Equal(t, []string{
		suite.level0Dir,
		suite.level0File,
		suite.level1Dir,
	}, suite.visited)
}

func (suite *FilterTestSuite) TestDirFilter() {
	t := suite.T()

	root := suite.level0Dir
	fn := walker.Chain(
		withDirFilter(),
		suite.visitedWalkFunc,
	)
	walker.Walk(root, fn)
	assert.Equal(t, []string{
		suite.level0Dir,
		suite.level1Dir,
		suite.level2Dir,
	}, suite.visited)
}

func (suite *FilterTestSuite) TestNotDirFilter() {
	t := suite.T()

	root := suite.level0Dir
	fn := walker.Chain(
		withNotDirFilter(root),
		suite.visitedWalkFunc,
	)
	walker.Walk(root, fn)
	assert.Equal(t, []string{
		suite.level0Dir,
		suite.level0File,
	}, suite.visited)
}

func (suite *FilterTestSuite) TestIgnoreFilter() {
	t := suite.T()

	root := suite.level0Dir

	ignorerWalkFunc, err := withIgnoreFilter("level1")
	assert.NoError(t, err)

	fn := walker.Chain(
		ignorerWalkFunc,
		suite.visitedWalkFunc,
	)
	walker.Walk(root, fn)
	assert.Equal(t, []string{
		suite.level0Dir,
		suite.level0File,
	}, suite.visited)
}
