package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateRange(t *testing.T) {
	assert.NoError(t, ValidateRangeNumbers[int](2, "2..4"))
	assert.Error(t, ValidateRangeNumbers[int](2, "3..7"))
	assert.NoError(t, ValidateRangeNumbers[int](-2, "-2..4"))
	assert.NoError(t, ValidateRangeNumbers[float64](-2.000007, "-2.001..-1"))
}

func TestValidateNumbersSet(t *testing.T) {
	assert.NoError(t, ValidateNumbersSet[int](590, "100,200,300,500,590,600,620"))
	assert.Error(t, ValidateNumbersSet[int](500.000001, "500,600"))
	assert.NoError(t, ValidateNumbersSet[float64](500.000001, "500.000001,600"))
	assert.NoError(t, ValidateNumbersSet[int](-700000, "-700000,600"))
	assert.NoError(t, ValidateNumbersSet[uint64](uint64(200), "600,200,0,-1"))
}

func TestValidateStringSet(t *testing.T) {
	assert.Error(t, ValidateStringSet("a", "john,bob, bryan ,stephan"))
	assert.NoError(t, ValidateStringSet("a", "john,bob,a,stephan"))
}
