package bloomfilter

import (
	"fmt"
	"testing"
)

func TestDefaultHasher(t *testing.T) {
	fmt.Println(Hash([]byte("[haser]")))
}
