package chatGPT

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestTrie_Basic(t *testing.T) {
	time.Sleep(5 * time.Second)
	fmt.Println(DefaultGPT.SendMsg("你哈", "321321", context.Background()))
}
