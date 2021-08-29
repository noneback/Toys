package raft

import (
	"log"
)

// Debugging
const Debug = true

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func NodesInfo(rfs []*Raft) {
	if Debug {
		for i, rf := range rfs {
			t, isLeader := rf.GetState()
			DPrintf("[info] Term %v node %v isLeader %v\n", t, i, isLeader)
			LogsInfo(rf)
		}
	}
}

func LogsInfo(rf *Raft) {
	for _, en := range rf.logs {
		DPrintf("[Log] node %v in term %v, %+v\n", rf.me, rf.term, en)
	}
}
