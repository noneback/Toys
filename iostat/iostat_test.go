package iostat

import (
	"testing"
	"time"
)

func TestIoStatCollector(t *testing.T) {
	ic := NewIostatCollector(Config{
		internal:   1,
		dev:        "nvme0n1",
		sectorSize: 512,
	})
	ic.Init()
	ic.Start()
	time.Sleep(100 * time.Second)
	ic.Stop()
}
