package fzd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/google/uuid"
	"github.com/horacehylee/fzd/walker"
)

const (
	headFile = "HEAD"
)

var (
	errIndexNotOpened        = errors.New("index is not opened")
	ErrIndexHeadDoesNotExist = fmt.Errorf("cannot open index, %v file does not exist", headFile)
)

// Indexer manages file path indexes, which provides atomic reindex swapping
type Indexer struct {
	locations  map[string]LocationOption
	basePath   string
	index      bleve.Index
	indexAlias bleve.IndexAlias
}

// LocationOption of options on traversing the specified directory location tree
type LocationOption struct {

	// Filters for files and directories from the location path
	Filters []Filter

	// Ignores is list of gitignore patterns for ignoring files and directories
	// It allows nested string structures
	Ignores []interface{}
}

// IndexerOption for options on indexing setup
type IndexerOption func(*Indexer)

// WithLocation allows specificing directory location and options for traversing it
func WithLocation(path string, option LocationOption) IndexerOption {
	return func(i *Indexer) {
		i.locations[path] = option
	}
}

// NewIndexer with specified base path and list of IndexerOptions
func NewIndexer(basePath string, options ...IndexerOption) (*Indexer, error) {
	if basePath == "" {
		return nil, fmt.Errorf("base path cannot be empty")
	}
	i := &Indexer{
		locations: make(map[string]LocationOption),
		basePath:  basePath,
	}
	for _, option := range options {
		option(i)
	}
	return i, nil
}

func (i *Indexer) Open() error {
	headPath := filepath.Join(i.basePath, headFile)
	if _, err := os.Stat(headPath); errors.Is(err, os.ErrNotExist) {
		return ErrIndexHeadDoesNotExist
	}
	// read HEAD file for index name
	content, err := os.ReadFile(headPath)
	if err != nil {
		return fmt.Errorf("failed to read %v: %w", headPath, err)
	}
	name := string(content)
	return i.open(name)
}

func (i *Indexer) open(name string) error {
	path := filepath.Join(i.basePath, name)
	index, err := bleve.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %v specified by %v: %w", path, headFile, err)
	}

	if i.indexAlias == nil {
		i.indexAlias = bleve.NewIndexAlias()
		i.indexAlias.Add(index)
	} else {
		in := []bleve.Index{index}
		out := []bleve.Index{i.index}
		i.indexAlias.Swap(in, out)

		// close old index
		// return this closing error after index pointer is updated
		err = i.index.Close()
	}
	i.index = index
	return err
}

// Index will atomic swap with index from scratch for all file entries if one is already opened
// Such that files generations and deletions are not required to be tracked
func (i *Indexer) Index() error {
	name := uuid.NewString()
	newIndexPath := filepath.Join(i.basePath, name)
	mapping, err := newIndexMapping()
	if err != nil {
		return err
	}
	config := make(map[string]interface{})
	b, err := bleve.NewBuilder(newIndexPath, mapping, config)
	if err != nil {
		return err
	}

	for path, option := range i.locations {
		indexWalkFunc := newIndexWalkFunc(b)

		filtersWalkFunc, err := newFiltersWalkFunc(path, option)
		// TODO: change to not fail fast
		if err != nil {
			return err
		}

		// combine index walkFunc last
		fn := walker.Chain(filtersWalkFunc, indexWalkFunc)

		err = walker.Walk(path, fn)
		// TODO: change to not fail fast
		if err != nil {
			return fmt.Errorf("failed to traverse path: %w", err)
		}
	}

	err = b.Close()
	// err := i.index.Batch(b)
	// TODO: change to not fail fast
	if err != nil {
		return fmt.Errorf("failed to execute index batch: %w", err)
	}

	// HEAD file update to new index
	headPath := filepath.Join(i.basePath, headFile)
	f, err := os.OpenFile(headPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open %v file: %w", headFile, err)
	}
	defer f.Close()

	_, err = f.WriteString(name)
	if err != nil {
		return fmt.Errorf("failed to write to %v file: %w", headFile, err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("failed to close %v file: %w", headFile, err)
	}

	// If other index is opened, swap to the new index instead atomically and close it
	err = i.open(name)
	if err != nil {
		return err
	}

	// Clean up old indexes
	entries, err := os.ReadDir(i.basePath)
	if err != nil {
		return fmt.Errorf("failed to read %v: %w", i.basePath, err)
	}
	for _, e := range entries {
		if e.Name() == headFile || e.Name() == name {
			continue
		}
		p := filepath.Join(i.basePath, e.Name())
		err = os.RemoveAll(p)
		if err != nil {
			return fmt.Errorf("failed to clean up %v: %w", p, err)
		}
	}
	return nil
}

func (i *Indexer) DocCount() (uint64, error) {
	if i.indexAlias == nil {
		return 0, errIndexNotOpened
	}
	return i.indexAlias.DocCount()
}

func (i *Indexer) Close() error {
	if i.indexAlias == nil {
		// index not initialized, no need close
		return nil
	}
	return i.indexAlias.Close()
}

func (i *Indexer) Search(term string) (*bleve.SearchResult, error) {
	if i.indexAlias == nil {
		return nil, errIndexNotOpened
	}
	// TODO: may have configuration to allow tweak of these settings
	queryString := bleve.NewQueryStringQuery(term)

	fuzzy := bleve.NewFuzzyQuery(term)
	fuzzy.SetBoost(2)

	wildcard := bleve.NewWildcardQuery(term)
	wildcard.SetBoost(2)

	prefix := bleve.NewPrefixQuery(term)
	prefix.SetBoost(2)

	match := bleve.NewMatchQuery(term)
	match.SetBoost(5)

	union := bleve.NewDisjunctionQuery(fuzzy, prefix, queryString, wildcard, match)
	req := bleve.NewSearchRequest(union)
	return i.indexAlias.Search(req)
}
