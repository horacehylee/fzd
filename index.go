package fzd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/regexp"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/google/uuid"
	"github.com/horacehylee/fzd/walker"
)

func newIndexName() string {
	return uuid.NewString()
}

func newIndexMapping() (*mapping.IndexMappingImpl, error) {
	mapping := bleve.NewIndexMapping()

	// TODO: may have config for custom tokenizer regexp (not limiting for undescore one)
	customTokenizerName := "custom_tokenizer"
	customAnalyzerName := "custom_analyzer"
	err := mapping.AddCustomTokenizer(customTokenizerName,
		map[string]interface{}{
			"type":   regexp.Name,
			"regexp": `[^\W_]+`,
		},
	)
	if err != nil {
		return nil, err
	}

	err = mapping.AddCustomAnalyzer(customAnalyzerName,
		map[string]interface{}{
			"type":      custom.Name,
			"tokenizer": customTokenizerName,
		})
	if err != nil {
		return nil, err
	}
	mapping.DefaultAnalyzer = customAnalyzerName
	return mapping, nil
}

type indexer interface {
	Index(id string, data interface{}) error
}

func newIndexWalkFunc(i indexer) walker.WalkFunc {
	return func(path string, info walker.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return i.Index(path, path)
	}
}

func removeIndexesExclude(basePath string, name string) error {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read %v: %w", basePath, err)
	}
	for _, e := range entries {
		if e.Name() == HeadFileName || e.Name() == name {
			continue
		}
		path := filepath.Join(basePath, e.Name())
		err = os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("failed to clean up %v: %w", path, err)
		}
	}
	return nil
}

// Wrapper of bleve.IndexAlias for single index alias
type singleIndexAlias struct {
	alias bleve.IndexAlias
	index bleve.Index
}

func newSingleIndexAlias(index bleve.Index) *singleIndexAlias {
	return &singleIndexAlias{
		alias: bleve.NewIndexAlias(index),
		index: index,
	}
}

func (s *singleIndexAlias) name() string {
	if s.index == nil {
		return ""
	}
	return s.index.Name()
}

func (s *singleIndexAlias) close() error {
	var indexCloseErr error
	if s.index != nil {
		indexCloseErr = s.index.Close()
	}
	var indexAliasCloseErr error
	if s.alias != nil {
		indexAliasCloseErr = s.alias.Close()
	}
	if indexCloseErr != nil || indexAliasCloseErr != nil {
		return fmt.Errorf("failed index close: %v, or failed index alias close: %v", indexCloseErr, indexAliasCloseErr)
	}
	return nil
}

func (s *singleIndexAlias) swap(in bleve.Index) {
	ins := []bleve.Index{in}
	outs := []bleve.Index{s.index}
	s.index = in
	s.alias.Swap(ins, outs)
}

func (s *singleIndexAlias) get() bleve.Index {
	return s.index
}

func (s *singleIndexAlias) docCount() (uint64, error) {
	return s.alias.DocCount()
}

func (s *singleIndexAlias) search(req *bleve.SearchRequest) (*bleve.SearchResult, error) {
	return s.alias.Search(req)
}
