package file

import (
	"os"
	"sync"
)

type WAL struct {
	rwmtx *sync.RWMutex
	f     *os.File
}
