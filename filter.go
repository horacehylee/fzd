package fzd

import (
	"fmt"
	"path/filepath"

	"github.com/horacehylee/fzd/ignorer"
	"github.com/horacehylee/fzd/walker"
)

// Filter for filtering file entries while traversing through file trees
type Filter string

const (
	// Top filters only immediate descendant of the path
	Top Filter = "top"

	// Dir filters only directories
	Dir Filter = "dir"

	// NotDir filters only non directories
	NotDir Filter = "not_dir"
)

func withTopFilter(root string) walker.WalkFunc {
	cleanedRoot := filepath.Clean(root)
	return func(path string, info walker.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dirname := filepath.Dir(path)
		// continue if root is current path, as root's parent directory will not equal to root
		if cleanedRoot != path && cleanedRoot != dirname {
			return walker.SkipThis
		}
		return nil
	}
}

func withDirFilter() walker.WalkFunc {
	return func(path string, info walker.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return walker.SkipThis
		}
		return nil
	}
}

func withNotDirFilter(root string) walker.WalkFunc {
	cleanedRoot := filepath.Clean(root)
	return func(path string, info walker.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// if root is current path, continue to walk to get entries inside
		if cleanedRoot == path {
			return nil
		}
		if info.IsDir() {
			return walker.SkipThis
		}
		return nil
	}
}

func withIgnoreFilter(ignores ...interface{}) (walker.WalkFunc, error) {
	ignorer, err := ignorer.NewIgnorer(ignores...)
	if err != nil {
		return nil, err
	}
	f := func(path string, info walker.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if ignorer.MatchesPath(path) {
			return walker.SkipThis
		}
		return nil
	}
	return f, nil
}

func newFiltersWalkFunc(root string, option LocationOption) (walker.WalkFunc, error) {
	var walkFuncs []walker.WalkFunc
	for _, f := range option.Filters {
		switch f {
		case Top:
			walkFuncs = append(walkFuncs, withTopFilter(root))
		case Dir:
			walkFuncs = append(walkFuncs, withDirFilter())
		case NotDir:
			walkFuncs = append(walkFuncs, withNotDirFilter(root))
		default:
			return nil, fmt.Errorf("\"%v\" filter is not supported", f)
		}
	}

	// add ignoreFilter's walkFunc last, as it requires more computational effort
	if len(option.Ignores) != 0 {
		ignoreWalkFunc, err := withIgnoreFilter(option.Ignores...)
		if err != nil {
			return nil, err
		}
		walkFuncs = append(walkFuncs, ignoreWalkFunc)
	}
	return walker.Combine(walkFuncs...), nil
}
