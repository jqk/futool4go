package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareVersions(t *testing.T) {
	// same length, only compare numbers.
	assert.Equal(t, -1, CompareVersions("1.1.0.20", "1.1.1.5"))
	assert.Equal(t, 1, CompareVersions("1.1.1.20", "1.1.1.5"))
	// case insensitive, numbers are different.
	assert.Equal(t, -1, CompareVersions("1.1a.0", "1.1A.1"))
	// case insensitive, numbers are same.
	assert.Equal(t, 0, CompareVersions("1.1a.0", "1.1A.0"))
	// alphatbit order if numbers are same.
	assert.Equal(t, -1, CompareVersions("1.1-a.1", "1.1-b.1"))
	// 'a' equals to '0a'.
	assert.Equal(t, 1, CompareVersions("1.1.a", "1.1.0"))
	assert.Equal(t, -1, CompareVersions("1.1.a", "1.1.1"))
	// no suffix in subverison is newer.
	assert.Equal(t, -1, CompareVersions("1.1.0", "1.1b.1"))
	// different version length.
	assert.Equal(t, -1, CompareVersions("1.1", "1.1.1"))
	assert.Equal(t, 0, CompareVersions("1.1", "1.1.0"))
	// different version length with newer subversion.
	assert.Equal(t, 1, CompareVersions("1.2", "1.1.1"))
	// prefix and suffix dots are trimed.
	// spaces in subversion are ignored.
	assert.Equal(t, 0, CompareVersions(" 1. 2 .", ".1. 2"))
	// prefix is treated as a string.
	// 'v1' is a string, and it's subversion number is 0,
	// compared with '1' in second parameter, which subversion number is 1.
	assert.Equal(t, -1, CompareVersions("v1.1", "1.1"))
	// ' -234' is trimed and treated as string '-234'.
	assert.Equal(t, 0, CompareVersions("1.1 -234", "1.1-234"))
}
