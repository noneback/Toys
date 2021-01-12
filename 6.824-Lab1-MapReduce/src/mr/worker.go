package mr

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"strings"
	"time"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

type worker struct {
	ID      int
	mapf    func(string, string) []KeyValue
	reducef func(string, []string) string
	nReduce int
	nMap    int
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func (w *worker) regTask() *Task {

	args := RegTaskArgs{w.ID}
	reply := RegTaskReply{}

	if ok := call("Master.RegTask", &args, &reply); !ok {
		//	log.Fatalln(w.ID, "error calling Master.RegTask,closed")
		return nil
	}

	if !reply.HasT {
		return nil
	}

	return &reply.T
}

func (w *worker) register() {
	args := RegWorkerArgs{}
	reply := RegWorkerReply{}

	call("Master.RegWorker", &args, &reply)
	w.ID = reply.ID
	w.nMap = reply.NMap
	w.nReduce = reply.NReduce

}

func (w *worker) run() {
	for {
		time.Sleep(time.Second)
		t := w.regTask()
		if t == nil {
			log.Println("All task has been finished,worker", w.ID, "closed")
			break
		}
		w.doTask(t)
	}
}

func (w *worker) report(t *Task, state ContextState) bool {
	args := ReportTaskArgs{
		WorkerID: w.ID,
		TaskID:   t.ID,
		State:    state,
	}
	reply := ReportTaskReply{}

	if ok := call("Master.ReportTask", &args, &reply); !ok {
		log.Fatalln("error reporting state")
		return false
	}

	return true

}

func (w *worker) doTask(t *Task) {

	switch t.Phase {
	case MAP:
		log.Println("Worker", w.ID, ":do map task ", t.ID)
		w.doMapTask(t)

	case REDUCE:
		log.Println("Worker", w.ID, "do reduce task ", t.ID)
		w.doReduceTask(t)
	default:
		log.Fatalln("Error in doTask , wrong phreas")
	}
}

func (w *worker) doMapTask(t *Task) {
	content, err := readFile(t.Filename)
	if err != nil {
		log.Fatalln("Failed to read file ", t.Filename)
		w.report(t, FAILED)
	}

	kvs := w.mapf(t.Filename, content)
	//partioning kvs for n reduce task

	// each for a file
	reduces := make([][]KeyValue, w.nReduce)

	for _, kv := range kvs {
		//hash the key and put into correspond reduce
		i := ihash(kv.Key) % w.nReduce
		reduces[i] = append(reduces[i], kv)
	}

	//write into files

	for index, r := range reduces {
		filename := fmt.Sprintf("./mr-%d-%d.json", t.ID, index)
		if saveKV2File(filename, r) != nil {
			log.Fatalln("Failed to save to file ", filename)
			w.report(t, FAILED)
		}
	}

	w.report(t, COMPLETE)
}

//do reduce task
func (w *worker) doReduceTask(t *Task) {
	m := make(map[string][]string)
	for i := 0; i < w.nMap; i++ {
		filename := fmt.Sprintf("./mr-%d-%d.json", i, t.ID)
		if readFile2Maps(filename, m) != nil {
			log.Fatalln("Failed to read file ", filename)
			w.report(t, FAILED)
		}
	}
	results := make([]string, 0, 100)

	for k, v := range m {
		results = append(results, fmt.Sprintf("%v %v\n", k, w.reducef(k, v)))
	}

	if ioutil.WriteFile(fmt.Sprintf("./mr-out-%d", t.ID), []byte(strings.Join(results, "")), 0600) != nil {
		log.Fatalln("write reduce file failed")
		w.report(t, FAILED)
	}
	w.report(t, COMPLETE)
}

//
// main/mrworker.go calls this function.
//

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	w := worker{
		ID:      -1,
		mapf:    mapf,
		reducef: reducef,
	}
	w.register()
	w.run()

}

//
// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	call("Master.Example", &args, &reply)

	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	//fmt.Println(err)
	return false
}
