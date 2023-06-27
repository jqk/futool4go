package futool4go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSizeString(t *testing.T) {
	assert.Equal(t, "0 bytes", ToSizeString(0))
	assert.Equal(t, "100 bytes", ToSizeString(100))
	assert.Equal(t, "1.309 KB", ToSizeString(1340))
	assert.Equal(t, "1.31 KB", ToSizeString(1340, 2))
	assert.Equal(t, "1.309 MB", ToSizeString(1340*1024))
	assert.Equal(t, "1.3086 GB", ToSizeString(1340*1024*1024, 4))
}
