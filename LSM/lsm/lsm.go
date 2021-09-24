package lsm

type LSM struct {
	memtable   *memtable
	immutables []*memtable
	levels     *levelManager
	opt        *Option
}

type Option struct {
	WorkDir        string
	MemtableSize   uint64
	SSTableMaxSize uint64
}

func NewLSM()           {}
func (l *LSM) Get()     {}
func (l *LSM) Set()     {}
func (l *LSM) Compact() {}
func (l *LSM) Close()   {}
