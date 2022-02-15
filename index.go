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
