package exchanges

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecimalPrecision(t *testing.T) {
	assert.Equal(t, 1, int(decimalPrecision(0.1)))
	assert.Equal(t, 2, int(decimalPrecision(0.01)))
	assert.Equal(t, 6, int(decimalPrecision(0.000001)))
	assert.Equal(t, 8, int(decimalPrecision(0.00000001)))

	assert.Equal(t, 2, int(decimalPrecision(0.08)))
	assert.Equal(t, 0, int(decimalPrecision(2)))
}
