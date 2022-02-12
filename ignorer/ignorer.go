package ignorer

import (
	"errors"
	"fmt"

	gitignore "github.com/sabhiram/go-gitignore"
)

type Ignorer struct {
	lines   []string
	matcher *gitignore.GitIgnore
}

var ErrTypeNotSupported = errors.New("type is not supported")

func NewIgnorer(elements ...interface{}) (*Ignorer, error) {
	i := &Ignorer{
		lines: make([]string, len(elements)),
	}
	for _, elem := range elements {
		err := i.resolveElement(elem)
		if err != nil {
			return nil, err
		}
	}
	i.matcher = gitignore.CompileIgnoreLines(i.lines...)
	return i, nil
}

func (i *Ignorer) resolveElement(v interface{}) error {
	switch t := v.(type) {
	case string:
		i.lines = append(i.lines, t)
	case []string:
		i.lines = append(i.lines, t...)
	case []interface{}:
		for _, tt := range t {
			err := i.resolveElement(tt)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("%T %w, only string, []string or []interface{}", t, ErrTypeNotSupported)
	}
	return nil
}

func (i *Ignorer) MatchesPath(path string) bool {
	return i.matcher.MatchesPath(path)
}
