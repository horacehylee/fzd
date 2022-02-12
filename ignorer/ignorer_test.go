package ignorer_test

import (
	"errors"
	"testing"

	"github.com/horacehylee/fzd/ignorer"
	"github.com/stretchr/testify/assert"
)

func newIgnorer(t *testing.T, values ...interface{}) *ignorer.Ignorer {
	ignorer, err := ignorer.NewIgnorer(values...)
	assert.NoError(t, err)
	return ignorer
}

func TestMatchFilesInGitignoreStyle(t *testing.T) {
	ignorer := newIgnorer(t,
		"testing",
		"!important",
		"*[Bb]ackup*",
	)

	assert.True(t, ignorer.MatchesPath("./testing"))
	assert.True(t, ignorer.MatchesPath("./testing/"))
	assert.True(t, ignorer.MatchesPath("./testing/more_files"))
	assert.True(t, ignorer.MatchesPath("/testing"))

	assert.False(t, ignorer.MatchesPath("./abc/testing/important"), "")
	assert.False(t, ignorer.MatchesPath("./abc/testing/important/some_files"))

	assert.True(t, ignorer.MatchesPath("./HelloBackup"))
	assert.True(t, ignorer.MatchesPath("./backup"))
	assert.True(t, ignorer.MatchesPath("./backups"))
	assert.True(t, ignorer.MatchesPath("./backup_temp"))
}

func TestNewIgnorerCanCombineSlices(t *testing.T) {
	ignorer := newIgnorer(t,
		"testing",
		[]string{
			"!important",
		},
		[]interface{}{
			"!moreImportant",
		},
	)

	assert.False(t, ignorer.MatchesPath("./abc/testing/important"), "")
	assert.False(t, ignorer.MatchesPath("./abc/testing/important/some_files"))
	assert.False(t, ignorer.MatchesPath("./abc/testing/moreImportant/some_files"))
}

func TestNewIgnorerWithDeeplyNestedSlices(t *testing.T) {
	ignorer := newIgnorer(t,
		"testing",
		[]interface{}{
			"!moreImportant",
			[]interface{}{
				"!important",
			},
		},
	)

	assert.False(t, ignorer.MatchesPath("./abc/testing/important"), "")
	assert.False(t, ignorer.MatchesPath("./abc/testing/important/some_files"))
	assert.False(t, ignorer.MatchesPath("./abc/testing/moreImportant/some_files"))
}

func TestReturnErrorIfTypeNotStringRelated(t *testing.T) {
	var err error
	_, err = ignorer.NewIgnorer(123)
	assert.EqualError(t, err, "int type is not supported, only string, []string or []interface{}")
	assert.True(t, errors.Is(err, ignorer.ErrTypeNotSupported))

	_, err = ignorer.NewIgnorer(true)
	assert.EqualError(t, err, "bool type is not supported, only string, []string or []interface{}")
	assert.True(t, errors.Is(err, ignorer.ErrTypeNotSupported))

	_, err = ignorer.NewIgnorer([]int{123})
	assert.EqualError(t, err, "[]int type is not supported, only string, []string or []interface{}")
	assert.True(t, errors.Is(err, ignorer.ErrTypeNotSupported))

	_, err = ignorer.NewIgnorer(
		[]interface{}{
			"123",
			[]interface{}{
				123,
			},
		},
	)
	assert.EqualError(t, err, "int type is not supported, only string, []string or []interface{}")
	assert.True(t, errors.Is(err, ignorer.ErrTypeNotSupported))

	_, err = ignorer.NewIgnorer(
		[]interface{}{
			123,
			[]interface{}{
				"123",
			},
		},
	)
	assert.EqualError(t, err, "int type is not supported, only string, []string or []interface{}")
	assert.True(t, errors.Is(err, ignorer.ErrTypeNotSupported))
}
