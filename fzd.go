package fzd

import (
	"errors"
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/regexp"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/horacehylee/fzd/walker"
)

var (
	errIndexNotInitialized = errors.New("index is not initilized")
)

type Indexer struct {
	locations map[string]LocationOption
	path      string
	index     bleve.Index
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

func NewIndexer(path string, options ...IndexerOption) (*Indexer, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	i := &Indexer{
		locations: make(map[string]LocationOption),
		path:      path,
	}
	for _, option := range options {
		option(i)
	}
	index, err := newIndex(i.path)
	if err != nil {
		return nil, err
	}
	i.index = index
	return i, nil
}

func newIndex(path string) (bleve.Index, error) {
	index, err := bleve.Open(path)
	if err != nil {
		if errors.Is(err, bleve.ErrorIndexPathDoesNotExist) {
			mapping, err := newIndexMapping()
			if err != nil {
				return nil, err
			}
			return bleve.New(path, mapping)
		} else {
			return nil, err
		}
	}
	return index, nil
}

func newIndexMapping() (*mapping.IndexMappingImpl, error) {
	mapping := bleve.NewIndexMapping()

	// TODO: may have config for custom tokenizer regexp (not limiting for undescore one)
	tokenizerName := "custom_tokenizer"
	analyzerName := "custom_analyzer"
	err := mapping.AddCustomTokenizer(tokenizerName,
		map[string]interface{}{
			"type":   regexp.Name,
			"regexp": `[^\W_]+`,
		},
	)
	if err != nil {
		return nil, err
	}

	err = mapping.AddCustomAnalyzer(analyzerName,
		map[string]interface{}{
			"type":      custom.Name,
			"tokenizer": tokenizerName,
		})
	if err != nil {
		return nil, err
	}
	mapping.DefaultAnalyzer = analyzerName
	return mapping, nil
}

// TODO: Index should be atomic, and refresh from scratch for all files
// Such that files generations and deletions are not required to be tracked
// May use bleve.Builder instead (https://github.com/blevesearch/bleve/blob/v2.3.0/index/scorch/builder.go#L45)
func (i *Indexer) Index() error {
	for path, option := range i.locations {
		b := i.index.NewBatch()
		indexWalkFunc, err := i.newIndexWalkFunc(b)
		if err != nil {
			return err
		}

		filtersWalkFunc, err := newFiltersWalkFunc(path, option)
		// TODO: change to not fail fast
		if err != nil {
			return err
		}

		// combine indexing walkFunc last
		fn := walker.Combine(filtersWalkFunc, indexWalkFunc)

		err = walker.Walk(path, fn)
		// TODO: change to not fail fast
		if err != nil {
			return fmt.Errorf("failed to traverse path: %w", err)
		}

		err = i.index.Batch(b)
		// TODO: change to not fail fast
		if err != nil {
			return fmt.Errorf("failed to execute index batch: %w", err)
		}
	}
	return nil
}

func (i *Indexer) newIndexWalkFunc(b *bleve.Batch) (walker.WalkFunc, error) {
	if i.index == nil {
		return nil, errIndexNotInitialized
	}
	fn := func(path string, info walker.FileInfo, err error) error {
		// fmt.Printf("%s %s\n", info.Mode(), path)
		// return nil

		// return i.index.Index(path, path)
		return b.Index(path, path)
	}
	return fn, nil
}

func (i *Indexer) DocCount() (uint64, error) {
	if i.index == nil {
		return 0, errIndexNotInitialized
	}
	return i.index.DocCount()
}

func (i *Indexer) Close() error {
	if i.index == nil {
		// index not initialized, no need close
		return nil
	}
	return i.index.Close()
}

func (i *Indexer) Search(term string) (*bleve.SearchResult, error) {
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
	return i.index.Search(req)
}
