package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

type RegTaskArgs struct {
	WorkerID int
}

type RegTaskReply struct {
	T    Task
	HasT bool
}

type ReportTaskArgs struct {
	WorkerID int
	TaskID   int
	State    ContextState
}
type ReportTaskReply struct {
}

type RegWorkerArgs struct {
}

type RegWorkerReply struct {
	ID      int
	NReduce int
	NMap    int
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
