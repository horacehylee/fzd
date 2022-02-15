package fzd

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/horacehylee/fzd/walker"
)

var (
	// Error where index is not opened
	ErrIndexNotOpened = errors.New("index is not opened")

	// Error where HEAD file does not exists, it should only occur when opened without previously indexed
	ErrIndexHeadDoesNotExist = fmt.Errorf("cannot open index, %v file does not exist", HeadFileName)
)

// Indexer manages file path indexes, which provides atomic reindex swapping
type Indexer struct {
	locations  map[string]LocationOption
	basePath   string
	index      bleve.Index
	indexAlias bleve.IndexAlias
	mutex      sync.RWMutex
	open       bool
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

// Open HEAD file specified index to be available for search and querying
// If HEAD file is not found, ErrIndexHeadDoesNotExist will returned
// If already opened, open again will simply check for the latest index specified in HEAD file and swap for it
func (i *Indexer) Open() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	name, err := readHead(i.basePath)
	if err != nil {
		return err
	}
	return i.openAndSwap(name)
}

// OpenAndSwap writes specified index to HEAD file, open and swap it with current index
// If same named index is already opened, only HEAD file will be overwritten again, no index will be swapped
// If no index is loaded, it will create/update HEAD file with specified index, open and load it
func (i *Indexer) OpenAndSwap(name string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.index == nil || i.index.Name() != name {
		// if same index is passed, no need to update HEAD and swap index
		err := writeHead(i.basePath, name)
		if err != nil {
			return err
		}
		err = i.openAndSwap(name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Indexer) openAndSwap(name string) error {
	prev := i.index
	if prev != nil && prev.Name() == name {
		// do nothing if index with same name is loaded
		return nil
	}

	path := filepath.Join(i.basePath, name)
	index, err := bleve.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %v specified by %v: %w", path, HeadFileName, err)
	}
	index.SetName(name)

	i.index = index
	if i.indexAlias == nil {
		i.indexAlias = bleve.NewIndexAlias()
		i.indexAlias.Add(index)
	} else {
		in := []bleve.Index{index}
		out := []bleve.Index{prev}
		i.indexAlias.Swap(in, out)

		// close previous index
		err = prev.Close()
		if err != nil {
			return fmt.Errorf("failed to close previous index: %w", err)
		}
	}
	i.open = true
	return nil
}

// Index will create new index from scratch for all file entries
// Such that files generations and deletions are not required to be tracked
// To use newly created index, use OpenAndSwap with returned index name
func (i *Indexer) Index() (string, error) {
	// no mutex locking is needed, as it will create a new index
	name := newIndexName()

	newIndexPath := filepath.Join(i.basePath, name)
	mapping, err := newIndexMapping()
	if err != nil {
		return "", err
	}
	config := make(map[string]interface{})
	builder, err := bleve.NewBuilder(newIndexPath, mapping, config)
	if err != nil {
		return "", err
	}

	for path, option := range i.locations {
		indexWalkFunc := newIndexWalkFunc(builder)

		filtersWalkFunc, err := newFiltersWalkFunc(path, option)
		// TODO: change to not fail fast
		if err != nil {
			return "", err
		}

		// combine index walkFunc last
		fn := walker.Chain(filtersWalkFunc, indexWalkFunc)

		err = walker.Walk(path, fn)
		// TODO: change to not fail fast
		if err != nil {
			return "", fmt.Errorf("failed to traverse path: %w", err)
		}
	}

	err = builder.Close()
	if err != nil {
		return "", fmt.Errorf("failed to execute index batch: %w", err)
	}
	return name, nil
}

// DocCount returns number of documents stored within the index
func (i *Indexer) DocCount() (uint64, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	if i.indexAlias == nil || !i.open {
		return 0, ErrIndexNotOpened
	}
	return i.indexAlias.DocCount()
}

// Search index with specified term and returns search result accordingly
func (i *Indexer) Search(term string) (*bleve.SearchResult, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	if i.indexAlias == nil || !i.open {
		return nil, ErrIndexNotOpened
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

// IndexName returns current loaded index name
// If index not opened, ErrIndexNotOpened is returned
func (i *Indexer) IndexName() (string, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	if i.index == nil || !i.open {
		return "", ErrIndexNotOpened
	}
	return i.index.Name(), nil
}

// Close currently opened index
// Except currently opened one (specified by HEAD file), other indexes will be clean up and removed, to keep index path clean
func (i *Indexer) Close() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if !i.open {
		return nil
	}
	var indexCloseErr error
	if i.index != nil {
		indexCloseErr = i.index.Close()
	}
	var indexAliasCloseErr error
	if i.indexAlias != nil {
		indexAliasCloseErr = i.indexAlias.Close()
	}
	if indexCloseErr != nil || indexAliasCloseErr != nil {
		return fmt.Errorf("failed index close: %v, or failed index alias close: %v", indexCloseErr, indexAliasCloseErr)
	}

	name := i.index.Name()
	err := removeIndexesExclude(i.basePath, name)
	if err != nil {
		return fmt.Errorf("failed to remove unused indexes: %w", err)
	}
	i.open = false
	return nil
}
