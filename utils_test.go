package configmapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNameArray(t *testing.T) {
	originalName, matched, index := CheckNameIsArrayAndGetIndex("USERS_IDS[1]")
	assert.True(t, matched)
	assert.Equal(t, 1, index)
	assert.Equal(t, "USERS_IDS", originalName)
	originalName, matched, index = CheckNameIsArrayAndGetIndex("USERS_IDS[1000]")
	assert.True(t, matched)
	assert.Equal(t, 1000, index)
	assert.Equal(t, "USERS_IDS", originalName)
	originalName, matched, index = CheckNameIsArrayAndGetIndex("USERS_IDS[1000]A")
	assert.False(t, matched)
	assert.Equal(t, -1, index)
	assert.Empty(t, originalName)
	originalName, matched, index = CheckNameIsArrayAndGetIndex("USERS_IDS")
	assert.False(t, matched)
	assert.Equal(t, -1, index)
	assert.Empty(t, originalName)
	originalName, matched, index = CheckNameIsArrayAndGetIndex("USERS_IDS[-1]")
	assert.False(t, matched)
	assert.Equal(t, -1, index)
	assert.Empty(t, originalName)
}
