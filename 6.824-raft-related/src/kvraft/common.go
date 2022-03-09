package kvraft

import (
	"fmt"
	"time"
)

const (
	OK             = "OK"
	ErrNoKey       = "ErrNoKey"
	ErrWrongLeader = "ErrWrongLeader"
	ErrTimeout     = "ErrTimeout"
	ExecuteTimeout = time.Second

	OpPut    OpType = "Put"
	OpAppend OpType = "Append"
	OpGet    OpType = "Get"
)

type OpType string

type Err string

// Put or Append
// type PutAppendArgs struct {
// 	Key   string
// 	Value string
// 	Op    string // "Put" or "Append"
// 	UID   string
// 	// You'll have to add definitions here.
// 	// Field names must start with capital letters,
// 	// otherwise RPC will break.
// }

// type PutAppendReply struct {
// 	Err Err
// }

// type GetArgs struct {
// 	Key string
// 	UID string
// 	// You'll have to add definitions here.
// }

// type GetReply struct {
// 	Err   Err
// 	Value string
// }

type Command struct {
	Key   string
	Value string
	Op    OpType
}

type CommandRequest struct {
	ClientID  string
	CommandID int
	Cmd       Command
}

type CommandResponse struct {
	Err   Err
	Value string
}

func genUID() string {
	return fmt.Sprintf("%v%v%v", time.Now().UnixNano(), nrand(), nrand())
}
