package fzd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getIgnorer(t *testing.T, values ...interface{}) *ignorer {
	ignorer, err := newIgnorer(values)
	assert.NoError(t, err)
	return ignorer
}

func TestMatchFilesInGitignoreStyle(t *testing.T) {
	ignorer := getIgnorer(t,
		"testing",
		"!important",
		"*[Bb]ackup*",
	)

	assert.True(t, ignorer.matchesPath("./testing"))
	assert.True(t, ignorer.matchesPath("./testing/"))
	assert.True(t, ignorer.matchesPath("./testing/more_files"))
	assert.True(t, ignorer.matchesPath("/testing"))

	assert.False(t, ignorer.matchesPath("./abc/testing/important"), "")
	assert.False(t, ignorer.matchesPath("./abc/testing/important/some_files"))

	assert.True(t, ignorer.matchesPath("./HelloBackup"))
	assert.True(t, ignorer.matchesPath("./backup"))
	assert.True(t, ignorer.matchesPath("./backups"))
	assert.True(t, ignorer.matchesPath("./backup_temp"))
}

func TestNewIgnorerCanCombineSlices(t *testing.T) {
	ignorer := getIgnorer(t,
		"testing",
		[]string{
			"!important",
		},
	)

	assert.False(t, ignorer.matchesPath("./abc/testing/important"), "")
	assert.False(t, ignorer.matchesPath("./abc/testing/important/some_files"))
}

func TestReturnErrorIfTypeNotStringRelated(t *testing.T) {
	var err error
	_, err = newIgnorer([]interface{}{123})
	assert.EqualError(t, err, "int type is not supported, only string or []string")

	_, err = newIgnorer([]interface{}{true})
	assert.EqualError(t, err, "bool type is not supported, only string or []string")

	_, err = newIgnorer([]interface{}{[]int{123}})
	assert.EqualError(t, err, "[]int type is not supported, only string or []string")
}
