package tree

import (
	"encoding/hex"
	"testing"
)

func TestHash(t *testing.T) {
	emptyHash := EmptyAccountAssetNodeHash()
	println(hex.EncodeToString(emptyHash))
}
