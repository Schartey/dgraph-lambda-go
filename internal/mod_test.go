package internal

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_findGoModFile_Fail(t *testing.T) {
	root, err := filepath.Abs(".")
	assert.NoError(t, err)

	_, err = findGoModFile(root, 1)
	assert.Error(t, err, "could not find go.mod file")
}

func Test_findGoModFile_Success(t *testing.T) {
	root, err := filepath.Abs(".")
	assert.NoError(t, err)

	path, err := findGoModFile(root, 2)

	assert.NoError(t, err)
	assert.Contains(t, path, "dgraph-lambda-go")

	root, err = filepath.Abs(".")
	assert.NoError(t, err)

	path, err = findGoModFile(root, 3)

	assert.NoError(t, err)
	assert.Contains(t, path, "dgraph-lambda-go")
}
