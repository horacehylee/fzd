package fzd

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/regexp"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/horacehylee/fzd/walker"
)

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
		e := i.Index(path, path)
		if e != nil {
			return e
		}
		return i.Index(path, path)
	}
}
