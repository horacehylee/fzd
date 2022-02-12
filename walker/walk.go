package walker

import (
	"fmt"
	"io/fs"

	"github.com/karrick/godirwalk"
)

// FileInfo is a subset of os.FileInfo interface
type FileInfo interface {
	Name() string      // base name of the file
	Mode() fs.FileMode // file mode bits
	IsDir() bool       // abbreviation for Mode().IsDir()
}

// WalkFunc is the type of the function called by Walk to visit each file or directory, using own FileInfo interface
type WalkFunc func(path string, info FileInfo, err error) error

// entry struct that implements own FileInfo interface, it acts as wrapper for godirwalk.Dirent
type entry struct {
	*godirwalk.Dirent
}

func (e *entry) Mode() fs.FileMode {
	return e.ModeType()
}

// SkipThis is used as return value from WalkFunc to indicate skipping particular file or directory
var SkipThis = godirwalk.SkipThis

// Walk walks the file tree rooted at the specified directory
// WalkFunc parameter will be called with specified directory path and each file/directory item within it
func Walk(root string, fn WalkFunc) error {
	err := godirwalk.Walk(root, &godirwalk.Options{
		Callback: func(osPathName string, de *godirwalk.Dirent) error {
			e := &entry{Dirent: de}
			return fn(osPathName, e, nil)
		},
	})
	if err != nil {
		return fmt.Errorf("failed to traverse %v: %w", root, err)
	}
	return nil
}
