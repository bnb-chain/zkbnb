package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTxTypeArray(t *testing.T) {
	array, err := logic.GetTxTypeArray(uint(5))
	fmt.Printf("%v", array)
	assert.NotNil(t, array)
	assert.Nil(t, err)
}
