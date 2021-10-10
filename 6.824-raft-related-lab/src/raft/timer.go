package raft

import (
	"math/rand"
	"time"
)

const (
	HeartBeatsTimeout = time.Duration(100) * time.Microsecond
)

func RandomizedElectionTimeout() time.Duration {
	rand.Seed(time.Now().UnixNano())
	ms := 450 + rand.Intn(300)
	return time.Duration(ms) * time.Millisecond
}

func (rf *Raft) resetElectionTimer() {
	// defer DPrintf("[timer] node %v reset Election timer\n", rf.me)
	if rf.role == Leader {
		return
	}
	rf.electionTimer.Reset(RandomizedElectionTimeout())
}

func (rf *Raft) resetHeartbeatTimer() {
	// defer DPrintf("[timer] node %v reset Heartbeats timer\n", rf.me)
	rf.heartbeatTimer.Reset(HeartBeatsTimeout)
}
