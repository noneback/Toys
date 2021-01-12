package mr

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

const (
	RUNNING ContextState = iota
	FAILED
	READY
	IDEL
	COMPLETE
)

const (
	MAX_PROCESSING_TIME = time.Second * 5
	SCHEDULE_INTERVAL   = time.Second
)

const (
	MAP PhaseKind = iota
	REDUCE
)

type ContextState int

type PhaseKind int

type Task struct {
	ID       int
	Filename string
	Phase    PhaseKind
}

type TaskContext struct {
	t         *Task
	state     ContextState
	workerID  int
	startTime time.Time
}

func readFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func saveKV2File(filename string, kvs []KeyValue) error {

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)

	for _, kv := range kvs {
		err := enc.Encode(&kv)
		if err != nil {
			return err
		}
	}
	return nil
}

func readFile2Maps(filename string, m map[string][]string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(file)

	for {
		var kv KeyValue
		if err := dec.Decode(&kv); err != nil {
			break
		}

		if _, ok := m[kv.Key]; !ok {
			m[kv.Key] = make([]string, 0, 100)
		}

		m[kv.Key] = append(m[kv.Key], kv.Value)
	}
	return nil
}
