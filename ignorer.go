package fzd

import (
	"fmt"

	gitignore "github.com/sabhiram/go-gitignore"
)

type ignorer struct {
	matcher *gitignore.GitIgnore
}

func newIgnorer(elements []interface{}) (*ignorer, error) {
	lines := make([]string, len(elements))
	for _, element := range elements {
		switch elem := element.(type) {
		case string:
			lines = append(lines, elem)
		case []string:
			lines = append(lines, elem...)
		default:
			return nil, fmt.Errorf("%T type is not supported, only string or []string", element)
		}
	}
	matcher := gitignore.CompileIgnoreLines(lines...)
	return &ignorer{
		matcher: matcher,
	}, nil
}

func (i *ignorer) matchesPath(path string) bool {
	return i.matcher.MatchesPath(path)
}
