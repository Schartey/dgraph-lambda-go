package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Packages(t *testing.T) {
	packages := &Packages{}

	pkg, err := packages.Load("github.com/schartey/dgraph-lambda-go/internal")
	assert.NoError(t, err)
	assert.Equal(t, "internal", pkg.Name)
	assert.Equal(t, "github.com/schartey/dgraph-lambda-go/internal", pkg.PkgPath)

	cached, err := packages.PackageFromPath("github.com/schartey/dgraph-lambda-go/internal")
	assert.NoError(t, err)
	assert.Equal(t, pkg, cached)
}

func Test_Packages_Fail(t *testing.T) {
	packages := &Packages{}

	_, err := packages.Load("invalid/package")
	assert.Error(t, err)

	_, err = packages.PackageFromPath("invalid/package")
	assert.Error(t, err)
}
