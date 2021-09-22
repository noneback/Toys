package bloomfilter

import (
	"fmt"
	"testing"
)

func TestDefaultHasher(t *testing.T) {
	fmt.Println("[haser]")
	fmt.Println(Hash("test"))
	fmt.Println(Hash("tese2"))
	fmt.Println(Hash("test"))
}
