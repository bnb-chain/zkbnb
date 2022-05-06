package test

import (
	"fmt"
	"testing"

	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/logic"
	"github.com/stretchr/testify/assert"
)

func TestGetTxTypeArray(t *testing.T) {
	array, err := logic.GetTxTypeArray(uint(5))
	fmt.Printf("%v", array)
	assert.NotNil(t, array)
	assert.Nil(t, err)
}
