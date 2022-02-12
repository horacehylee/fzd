package walker_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/horacehylee/fzd/walker"
	"github.com/karrick/godirwalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// fileMode is the default file mode for creating temp directory using os.MkdirTemp
const fileMode = 0700

type item struct {
	path  string
	name  string
	mode  fs.FileMode
	isDir bool
}

type WalkTestSuite struct {
	suite.Suite
	level0Dir       string
	level0File      string
	level1Dir       string
	level1File      string
	level2Dir       string
	level2File      string
	visited         []item
	visitedWalkFunc walker.WalkFunc
}

func TestWalkerTestSuite(t *testing.T) {
	suite.Run(t, new(WalkTestSuite))
}

func (suite *WalkTestSuite) SetupSuite() {
	t := suite.T()

	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	suite.level0Dir = dir

	suite.level0File = filepath.Join(dir, "level0.txt")
	err = os.WriteFile(suite.level0File, []byte("content"), fileMode)
	assert.NoError(t, err)

	suite.level1Dir = filepath.Join(dir, "level1")
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

func (suite *WalkTestSuite) SetupTest() {
	suite.visited = make([]item, 0)
	suite.visitedWalkFunc = func(path string, info walker.FileInfo, err error) error {
		suite.visited = append(suite.visited, item{
			path:  path,
			name:  info.Name(),
			mode:  info.Mode(),
			isDir: info.IsDir(),
		})
		return nil
	}
}

func (suite *WalkTestSuite) TearDownSuite() {
	t := suite.T()

	if suite.level0Dir != "" {
		err := os.RemoveAll(suite.level0Dir)
		assert.NoError(t, err)
	}
}

func newItem(t *testing.T, path string, isDir bool) item {
	de, err := godirwalk.NewDirent(path)
	assert.NoError(t, err)

	return item{
		path:  path,
		name:  filepath.Base(path),
		mode:  de.ModeType(), // godirwalk returnes slightly different fs.FileMode than os.Stat
		isDir: isDir,
	}
}

func (suite *WalkTestSuite) TestWalkFromDir() {
	t := suite.T()

	root := suite.level0Dir
	fn := walker.Combine([]walker.WalkFunc{
		suite.visitedWalkFunc,
	})
	walker.Walk(root, fn)
	assert.Equal(t, []item{
		newItem(t, suite.level0Dir, true),
		newItem(t, suite.level0File, false),
		newItem(t, suite.level1Dir, true),
		newItem(t, suite.level1File, false),
		newItem(t, suite.level2Dir, true),
		newItem(t, suite.level2File, false),
	}, suite.visited)
}

func (suite *WalkTestSuite) TestWalkFromDirWithExtraSeparatorAtEnd() {
	t := suite.T()

	root := suite.level0Dir + string(os.PathSeparator)
	fn := walker.Combine([]walker.WalkFunc{
		suite.visitedWalkFunc,
	})
	walker.Walk(root, fn)
	assert.Equal(t, []item{
		// WalkFunc should be called with clean path
		newItem(t, filepath.Clean(root), true),
		newItem(t, suite.level0File, false),
		newItem(t, suite.level1Dir, true),
		newItem(t, suite.level1File, false),
		newItem(t, suite.level2Dir, true),
		newItem(t, suite.level2File, false),
	}, suite.visited)
}

func (suite *WalkTestSuite) TestWalkFromFile() {
	t := suite.T()

	root := suite.level0File
	fn := walker.Combine([]walker.WalkFunc{
		suite.visitedWalkFunc,
	})
	walker.Walk(root, fn)
	assert.Empty(t, suite.visited)
}
