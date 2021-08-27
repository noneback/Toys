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
		}
	}
}
