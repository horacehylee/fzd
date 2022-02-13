package walker

import "errors"

// Chain to combine multiple WalkFuncs to one WalkFunc
// If one of the WalkFunc returned SkipThis, then WalkFunc chain will terminate early
func Chain(walkFuncs ...WalkFunc) WalkFunc {
	return func(path string, info FileInfo, err error) error {
		for _, f := range walkFuncs {
			err = f(path, info, err)
			if errors.Is(err, SkipThis) {
				// terminate early if error returned is SkipThis
				return err
			}
		}
		return err
	}
}
