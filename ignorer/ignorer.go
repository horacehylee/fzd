package ignorer

import (
	"errors"
	"fmt"

	gitignore "github.com/sabhiram/go-gitignore"
)

type Ignorer struct {
	matcher *gitignore.GitIgnore
}

var ErrTypeNotSupported = errors.New("type is not supported")

func NewIgnorer(elements []interface{}) (*Ignorer, error) {
	lines := make([]string, len(elements))
	for _, element := range elements {
		fmt.Printf("%v", element)
		switch elem := element.(type) {
		case string:
			lines = append(lines, elem)
		case []string:
			lines = append(lines, elem...)
		case []interface{}:
			strings := make([]string, len(elem))
			for i, e := range elem {
				strings[i] = fmt.Sprint(e)
			}
			lines = append(lines, strings...)
		default:
			return nil, fmt.Errorf("%T %w, only string, []string or []interface{}", element, ErrTypeNotSupported)
		}
	}
	matcher := gitignore.CompileIgnoreLines(lines...)
	return &Ignorer{
		matcher: matcher,
	}, nil
}

func (i *Ignorer) MatchesPath(path string) bool {
	return i.matcher.MatchesPath(path)
}
