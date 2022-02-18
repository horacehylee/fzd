package fzd_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/horacehylee/fzd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// fileMode is the default file mode for creating temp directory using os.MkdirTemp
const fileMode = 0700

// raceTimes is number of loops for race tests
const raceTimes = 20

type FzdTestSuite struct {
	suite.Suite
	level0Dir         string
	level0File        string
	level1Dir         string
	level1File        string
	level2Dir         string
	level2File        string
	indexesDir        string
	indexer           *fzd.Indexer
	indexAndOpenCount int
}

func TestFzdTestSuite(t *testing.T) {
	suite.Run(t, new(FzdTestSuite))
}

func (suite *FzdTestSuite) SetupSuite() {
	t := suite.T()

	indexesDir, err := os.MkdirTemp("", "testFzdSuiteIndexes")
	assert.NoError(t, err)
	suite.indexesDir = indexesDir

	dir, err := os.MkdirTemp("", "testFzdSuite")
	assert.NoError(t, err)
	suite.level0Dir = dir

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

func (suite *FzdTestSuite) TearDownSuite() {
	t := suite.T()

	if suite.level0Dir != "" {
		err := os.RemoveAll(suite.level0Dir)
		assert.NoError(t, err)
	}

	if suite.indexer != nil {
		err := suite.indexer.Close()
		assert.NoError(t, err)
	}

	if suite.indexesDir != "" {
		err := os.RemoveAll(suite.indexesDir)
		assert.NoError(t, err)
	}
}

func (suite *FzdTestSuite) SetupTest() {
	t := suite.T()

	suite.indexAndOpenCount = 0

	if suite.indexer != nil {
		err := suite.indexer.Close()
		assert.NoError(t, err)
	}

	// clear indexes directory
	entries, err := os.ReadDir(suite.indexesDir)
	assert.NoError(t, err)
	for _, e := range entries {
		err = os.RemoveAll(filepath.Join(suite.indexesDir, e.Name()))
		assert.NoError(t, err)
	}

	indexer, err := fzd.NewIndexer(suite.indexesDir, fzd.WithLocation(suite.level0Dir, fzd.LocationOption{}))
	assert.NoError(t, err)
	suite.indexer = indexer
}

func (suite *FzdTestSuite) TearDownTest() {
	t := suite.T()

	if suite.indexer != nil {
		err := suite.indexer.Close()
		assert.NoError(t, err)
	}
}

func (suite *FzdTestSuite) TestOpenFailedIfNoHeadFile() {
	t := suite.T()
	indexer := suite.indexer

	err := indexer.Open()
	assert.ErrorIs(t, err, fzd.ErrIndexHeadDoesNotExist)

	_, err = indexer.IndexName()
	assert.ErrorIs(t, err, fzd.ErrIndexNotOpened)
}

func (suite *FzdTestSuite) TestOpenIfDoneTwice() {
	t := suite.T()
	indexer := suite.indexer

	name := suite.indexAndOpen()

	err := indexer.Open()
	assert.NoError(t, err)

	indexName, err := indexer.IndexName()
	assert.NoError(t, err)
	assert.Equal(t, name, indexName)
}

func (suite *FzdTestSuite) TestIndexProducesOnlyIndex() {
	t := suite.T()
	indexer := suite.indexer

	name, err := indexer.Index()
	assert.NoError(t, err)

	dirnames := suite.readIndexesDirnames(1)
	assert.ElementsMatch(t, []string{
		name,
	}, dirnames)

	_, err = indexer.IndexName()
	assert.ErrorIs(t, err, fzd.ErrIndexNotOpened)
}

func (suite *FzdTestSuite) TestIndexProducesNewIndexAndLeavingOldOnes() {
	t := suite.T()
	indexer := suite.indexer

	name1, err := indexer.Index()
	assert.NoError(t, err)

	dirnames1 := suite.readIndexesDirnames(1)
	assert.ElementsMatch(t, []string{
		name1,
	}, dirnames1)

	name2, err := indexer.Index()
	assert.NoError(t, err)

	dirnames2 := suite.readIndexesDirnames(2)
	assert.ElementsMatch(t, []string{
		name1,
		name2,
	}, dirnames2)

	_, err = indexer.IndexName()
	assert.ErrorIs(t, err, fzd.ErrIndexNotOpened)
}

func (suite *FzdTestSuite) TestOpenAndSwapProducesHeadFile() {
	t := suite.T()
	indexer := suite.indexer

	name, err := indexer.Index()
	assert.NoError(t, err)

	err = indexer.OpenAndSwap(name)
	assert.NoError(t, err)

	headFileIndexName := suite.readHeadFile()
	assert.Equal(t, name, headFileIndexName)

	dirnames := suite.readIndexesDirnames(2)
	assert.ElementsMatch(t, []string{
		fzd.HeadFileName,
		headFileIndexName,
	}, dirnames)

	indexName, err := indexer.IndexName()
	assert.NoError(t, err)
	assert.Equal(t, name, indexName)
}

func (suite *FzdTestSuite) TestOpenAndSwapProducesHeadFileAndLeaveOtherIndexes() {
	t := suite.T()
	indexer := suite.indexer

	name1, err := indexer.Index()
	assert.NoError(t, err)

	name2, err := indexer.Index()
	assert.NoError(t, err)

	dirnames1 := suite.readIndexesDirnames(2)
	assert.ElementsMatch(t, []string{
		name1,
		name2,
	}, dirnames1)

	err = indexer.OpenAndSwap(name1)
	assert.NoError(t, err)

	headFileIndexName := suite.readHeadFile()
	assert.Equal(t, name1, headFileIndexName)

	dirnames2 := suite.readIndexesDirnames(3)
	assert.ElementsMatch(t, []string{
		fzd.HeadFileName,
		name1,
		name2,
	}, dirnames2)

	indexName, err := indexer.IndexName()
	assert.NoError(t, err)
	assert.Equal(t, name1, indexName)
}

func (suite *FzdTestSuite) TestOpenAndSwapForSameIndex() {
	t := suite.T()
	indexer := suite.indexer

	name1, err := indexer.Index()
	assert.NoError(t, err)

	err = indexer.OpenAndSwap(name1)
	assert.NoError(t, err)

	headFileIndexName1 := suite.readHeadFile()
	assert.Equal(t, name1, headFileIndexName1)

	headFileModTime1 := suite.headFileModTime()

	dirnames1 := suite.readIndexesDirnames(2)
	assert.ElementsMatch(t, []string{
		fzd.HeadFileName,
		headFileIndexName1,
	}, dirnames1)

	name2, err := indexer.Index()
	assert.NoError(t, err)

	err = indexer.OpenAndSwap(name1)
	assert.NoError(t, err)

	headFileIndexName2 := suite.readHeadFile()
	assert.Equal(t, name1, headFileIndexName2)

	headFileModTime2 := suite.headFileModTime()
	assert.Equal(t, headFileModTime1, headFileModTime2, fmt.Sprintf("%v file should not be updated if OpenAndSwap with same index", fzd.HeadFileName))

	dirnames2 := suite.readIndexesDirnames(3)
	assert.ElementsMatch(t, []string{
		fzd.HeadFileName,
		headFileIndexName2,
		name2,
	}, dirnames2, "other indexes will not be removed for OpenAndSwap")

	indexName, err := indexer.IndexName()
	assert.NoError(t, err)
	assert.Equal(t, name1, indexName)
}

func (suite *FzdTestSuite) TestDocCount() {
	t := suite.T()
	indexer := suite.indexer

	suite.indexAndOpen()

	count, err := indexer.DocCount()
	assert.NoError(t, err)
	assert.Equal(t, uint64(6), count)
}

func (suite *FzdTestSuite) TestDocCountReturnsErrorIfNotOpened() {
	t := suite.T()
	indexer := suite.indexer

	_, err := indexer.DocCount()
	assert.ErrorIs(t, err, fzd.ErrIndexNotOpened)
}

func (suite *FzdTestSuite) TestSearch() {
	t := suite.T()
	indexer := suite.indexer

	suite.indexAndOpen()

	res, err := indexer.Search("txt")
	assert.NoError(t, err)

	hits := suite.readSearchResults(res, 3)
	assert.Equal(t, []string{
		suite.level0File,
		suite.level1File,
		suite.level2File,
	}, hits)
}

func (suite *FzdTestSuite) TestSearchWith() {
	t := suite.T()
	indexer := suite.indexer

	suite.indexAndOpen()

	req := bleve.NewSearchRequest(bleve.NewWildcardQuery("*txt"))
	res, err := indexer.SearchWith(req)
	assert.NoError(t, err)

	hits := suite.readSearchResults(res, 3)
	assert.Equal(t, []string{
		suite.level0File,
		suite.level1File,
		suite.level2File,
	}, hits)
}

func (suite *FzdTestSuite) TestSearchAndDocCountAfterIndexSwapped() {
	t := suite.T()
	indexer := suite.indexer

	name1 := suite.indexAndOpen()

	indexName1, err := indexer.IndexName()
	assert.NoError(t, err)
	assert.Equal(t, name1, indexName1)

	count1, err := indexer.DocCount()
	assert.NoError(t, err)
	assert.Equal(t, uint64(6), count1)

	res1, err := indexer.Search("txt")
	assert.NoError(t, err)

	hits1 := suite.readSearchResults(res1, 3)
	assert.Equal(t, []string{
		suite.level0File,
		suite.level1File,
		suite.level2File,
	}, hits1)

	time1, err := indexer.LastIndexed()
	assert.NoError(t, err)

	extraFile := filepath.Join(suite.level0Dir, "extra.txt")
	err = os.WriteFile(extraFile, []byte("content"), fileMode)
	assert.NoError(t, err)
	defer os.Remove(extraFile)

	time.Sleep(1 * time.Second)

	name2 := suite.indexAndOpen()
	assert.NotEqual(t, name2, name1)

	indexName2, err := indexer.IndexName()
	assert.NoError(t, err)
	assert.Equal(t, name2, indexName2)

	count2, err := indexer.DocCount()
	assert.NoError(t, err)
	assert.Equal(t, uint64(7), count2)

	res2, err := indexer.Search("txt")
	assert.NoError(t, err)

	hits2 := suite.readSearchResults(res2, 4)
	assert.ElementsMatch(t, []string{
		extraFile,
		suite.level0File,
		suite.level1File,
		suite.level2File,
	}, hits2)

	time2, err := indexer.LastIndexed()
	assert.NoError(t, err)
	assert.Greater(t, time2.Unix(), time1.Unix())
}

func (suite *FzdTestSuite) TestSearchReturnsErrorIfNotOpened() {
	t := suite.T()
	indexer := suite.indexer

	_, err := indexer.Search("txt")
	assert.ErrorIs(t, err, fzd.ErrIndexNotOpened)
}

func (suite *FzdTestSuite) TestClose() {
	t := suite.T()
	indexer := suite.indexer

	name1 := suite.indexAndOpen()
	name2 := suite.indexAndOpen()

	indexName, err := indexer.IndexName()
	assert.NoError(t, err)
	assert.Equal(t, name2, indexName)

	dirnames1 := suite.readIndexesDirnames(3)
	assert.ElementsMatch(t, []string{
		fzd.HeadFileName,
		name1,
		name2,
	}, dirnames1)

	err = indexer.Close()
	assert.NoError(t, err)

	_, err = indexer.IndexName()
	assert.ErrorIs(t, err, fzd.ErrIndexNotOpened)

	dirnames2 := suite.readIndexesDirnames(2)
	assert.ElementsMatch(t, []string{
		fzd.HeadFileName,
		name2,
	}, dirnames2, "other indexes should be removed and cleaned up")
}

func (suite *FzdTestSuite) TestCloseNoErrorIfNotOpened() {
	t := suite.T()
	indexer := suite.indexer

	err := indexer.Close()
	assert.NoError(t, err)
}

func (suite *FzdTestSuite) TestIndexRace() {
	t := suite.T()
	indexer := suite.indexer

	var wg sync.WaitGroup
	wg.Add(raceTimes)
	for i := 0; i < raceTimes; i++ {
		go func() {
			_, err := indexer.Index()
			assert.NoError(t, err)
			wg.Done()
		}()
	}
	wg.Wait()
	suite.readIndexesDirnames(raceTimes)

	_, err := indexer.IndexName()
	assert.ErrorIs(t, err, fzd.ErrIndexNotOpened)
}

func (suite *FzdTestSuite) TestIndexAndOpenAndSwapRace() {
	t := suite.T()
	indexer := suite.indexer

	name := suite.indexAndOpen()

	var wg sync.WaitGroup
	wg.Add(2 * raceTimes)
	for i := 0; i < raceTimes; i++ {
		go func() {
			_, err := indexer.Index()
			assert.NoError(t, err)
			wg.Done()
		}()

		go func() {
			err := indexer.OpenAndSwap(name)
			assert.NoError(t, err)
			wg.Done()
		}()
	}
	wg.Wait()

	indexName, err := indexer.IndexName()
	assert.NoError(t, err)
	assert.Equal(t, name, indexName)
}

func (suite *FzdTestSuite) TestIndexAndDocCountRace() {
	t := suite.T()
	indexer := suite.indexer

	name := suite.indexAndOpen()

	var wg sync.WaitGroup
	wg.Add(2 * raceTimes)
	for i := 0; i < raceTimes; i++ {
		go func() {
			_, err := indexer.Index()
			assert.NoError(t, err)
			wg.Done()
		}()

		go func() {
			count, err := indexer.DocCount()
			assert.NoError(t, err)
			assert.Equal(t, uint64(6), count)
			wg.Done()
		}()
	}
	wg.Wait()
	suite.readIndexesDirnames(2 + raceTimes)

	indexName, err := indexer.IndexName()
	assert.NoError(t, err)
	assert.Equal(t, name, indexName)
}

func (suite *FzdTestSuite) TestIndexAndSearchRace() {
	t := suite.T()
	indexer := suite.indexer

	name := suite.indexAndOpen()

	var wg sync.WaitGroup
	wg.Add(2 * raceTimes)
	for i := 0; i < raceTimes; i++ {
		go func() {
			_, err := indexer.Index()
			assert.NoError(t, err)
			wg.Done()
		}()

		go func() {
			res, err := indexer.Search("txt")
			assert.NoError(t, err)
			suite.readSearchResults(res, 3)
			wg.Done()
		}()
	}
	wg.Wait()
	suite.readIndexesDirnames(2 + raceTimes)

	indexName, err := indexer.IndexName()
	assert.NoError(t, err)
	assert.Equal(t, name, indexName)
}

func (suite *FzdTestSuite) readIndexesDirnames(expectedLen int) []string {
	t := suite.T()

	entries, err := os.ReadDir(suite.indexesDir)
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(entries))

	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func (suite *FzdTestSuite) readHeadFile() string {
	path := filepath.Join(suite.indexesDir, fzd.HeadFileName)
	content, err := os.ReadFile(path)
	assert.NoError(suite.T(), err)
	return string(content)
}

func (suite *FzdTestSuite) headFileModTime() time.Time {
	path := filepath.Join(suite.indexesDir, fzd.HeadFileName)
	info, err := os.Stat(path)
	assert.NoError(suite.T(), err)
	return info.ModTime()
}

func (suite *FzdTestSuite) readSearchResults(res *bleve.SearchResult, expectedLen int) []string {
	assert.NotNil(suite.T(), res)
	assert.Equal(suite.T(), expectedLen, len(res.Hits))

	var hits []string
	for _, h := range res.Hits {
		hits = append(hits, h.ID)
	}
	return hits
}

func (suite *FzdTestSuite) indexAndOpen() string {
	name, err := suite.indexer.Index()
	assert.NoError(suite.T(), err)

	err = suite.indexer.OpenAndSwap(name)
	assert.NoError(suite.T(), err)

	dirnames := suite.readIndexesDirnames(2 + suite.indexAndOpenCount)
	assert.Contains(suite.T(), dirnames, fzd.HeadFileName)
	assert.Contains(suite.T(), dirnames, name)

	indexName, err := suite.indexer.IndexName()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), name, indexName)

	suite.indexAndOpenCount++
	return name
}
