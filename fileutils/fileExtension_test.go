package fileutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExtensions(t *testing.T) {
	extensions, err := GetFileExtensions("../test-data/fileutils/filter", true)
	assert.Nil(t, err)
	assert.NotNil(t, extensions)
	assert.Equal(t, 7, len(extensions))

	extensions, err = GetFileExtensions("../test-data/fileutils/filter", false)
	assert.Nil(t, err)
	assert.NotNil(t, extensions)
	assert.Equal(t, 5, len(extensions))
}
