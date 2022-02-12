package fzd

import (
	"fmt"

	"github.com/horacehylee/fzd/walker"
)

type Indexer struct {
	locations map[string]LocationOption
}

type LocationOption struct {
	// Filters for files and directories from the location path
	Filters []Filter

	// Ignores is list of gitignore patterns for ignoring files and directories
	// It allows nested string structures
	Ignores []interface{}
}

type IndexerOption func(*Indexer)

func WithLocation(path string, option LocationOption) IndexerOption {
	return func(i *Indexer) {
		i.locations[path] = option
	}
}

func NewIndexer(options ...IndexerOption) *Indexer {
	i := &Indexer{
		locations: make(map[string]LocationOption),
	}
	for _, option := range options {
		option(i)
	}
	return i
}

func (i *Indexer) Index() error {
	w := func(path string, info walker.FileInfo, err error) error {
		fmt.Printf("%s %s\n", info.Mode(), path)
		return nil
	}
	for path, option := range i.locations {
		fn, err := newWalkFunc(w, path, option)
		// TODO: change to not fail fast
		if err != nil {
			return err
		}

		err = walker.Walk(path, fn)
		// TODO: change to not fail fast
		if err != nil {
			return fmt.Errorf("failed to traverse path: %w", err)
		}
	}
	return nil
}

// TODO: may remove the fn walker.WalkFunc parameter (that should be the common indexing WalkFunc)
func newWalkFunc(fn walker.WalkFunc, root string, option LocationOption) (walker.WalkFunc, error) {
	filtersWalkFunc, err := newFiltersWalkFunc(root, option)
	if err != nil {
		return nil, err
	}
	// combine user's walkFunc at last
	return walker.Combine(filtersWalkFunc, fn), nil
}
