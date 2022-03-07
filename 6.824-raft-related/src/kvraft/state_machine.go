package kvraft

type KVStateMachine interface {
	Get(key string) (string, Err)
	Put(key, val string) Err
	Append(key, val string) Err
}

type MemoryKV struct {
	kv map[string]string
}

func NewMemoryKV() *MemoryKV {
	return &MemoryKV{make(map[string]string)}
}

func (mkv *MemoryKV) Get(key string) (string, Err) {
	if val, ok := mkv.kv[key]; ok {
		return val, OK
	}
	return "", ErrNoKey
}

func (mkv *MemoryKV) Put(key, val string) Err {
	mkv.kv[key] = val
	return OK
}

func (mkv *MemoryKV) Append(key, val string) Err {
	mkv.kv[key] += val
	return OK
}
