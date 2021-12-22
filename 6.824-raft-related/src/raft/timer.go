package raft

import (
	"math/rand"
	"time"
)

const (
	HeartBeatsTimeout = time.Duration(100) * time.Millisecond
)

func GetHeartbeatTimeout() time.Duration {
	return HeartBeatsTimeout
}

func GetRandomElectionTimeout() time.Duration {
	rand.Seed(time.Now().UnixNano())
	ms := 350 + rand.Intn(300)
	return time.Duration(ms) * time.Millisecond
}

func (rf *Raft) resetElectionTimerL() {
	if rf.state == StateLeader {
		return
	}
	rf.electionTimer.Reset(GetRandomElectionTimeout())
}

func (rf *Raft) resetHeartbeatTimerL() {
	rf.heartbeatTimer.Reset(HeartBeatsTimeout)
}
