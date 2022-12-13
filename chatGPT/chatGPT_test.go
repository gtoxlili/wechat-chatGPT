package chatGPT

import (
	"context"
	"fmt"
	"testing"
)

func TestTrie_Basic(t *testing.T) {
	fmt.Println(DefaultGPT().SendMsg("健康检查", "healthCheck", context.Background()))
}
