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
