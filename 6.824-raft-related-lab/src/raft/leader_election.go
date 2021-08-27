package raft

type RequestVoteArgs struct {
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer DPrintf("[vote] node %v, candidate %v, req %+v reply %+v\n", rf.me, args.CandidateID, args, reply)
	// vote false
	if args.Term < rf.term || (args.Term == rf.term && rf.votedFor != None && rf.votedFor != args.CandidateID) {
		reply.Term, reply.VoteGranted = rf.term, false
		return
	}

	if args.Term > rf.term {
		rf.newTerm(args.Term, None)
		rf.role = Follower
	}

	// TODO 2B
	// if the candidate's log is up-to-date
	if !rf.isLogUpToDate(args.LastLogTerm, args.LastLogIndex) {
		reply.Term, reply.VoteGranted = rf.term, false
		return
	}

	rf.votedFor = args.CandidateID
	rf.resetElectionTimer()
	reply.Term, reply.VoteGranted = rf.term, true
}

// isLogUpToDate tells whether candidate's log is fresher than follower
func (rf *Raft) isLogUpToDate(term, index int) bool {
	llog := rf.getLastLog()
	if term > llog.Term {
		return true
	}
	if term == llog.Term && index >= llog.Index {
		return true
	}
	return false
}

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) startElection() {
	// has lock
	rf.newTerm(rf.term+1, None)
	rf.role = Candidate
	DPrintf("[role] %v : %v become Candidate in term %v", rf.role, rf.me, rf.term)
	rf.votedFor = rf.me
	// RequestVote

	args := &RequestVoteArgs{
		Term:         rf.term,
		CandidateID:  rf.me,
		LastLogIndex: rf.getLastLog().Index,
		LastLogTerm:  rf.getLastLog().Term,
	}
	voteCnt := 1
	for p := range rf.peers {
		if p == rf.me {
			continue
		}

		if rf.role != Candidate {
			return
		}

		go func(peer int) {
			DPrintf("[goroutine] sendRequestVote gen a go routine")
			var reply RequestVoteReply

			if ok := rf.sendRequestVote(peer, args, &reply); ok {
				rf.mu.Lock()
				defer rf.mu.Unlock()

				if rf.role != Candidate {
					return
				}

				if reply.VoteGranted {
					voteCnt++
					if voteCnt*2 > len(rf.peers) {
						rf.role = Leader
						DPrintf("[leader] %v: become leader in term %v", rf.me, rf.term)
						rf.leader = rf.me
						rf.resetHeartbeatTimer()
						rf.sendAppendMsg()
					}
				} else if reply.Term > rf.term {
					rf.newTerm(reply.Term, None)
					DPrintf("%v: become follower in term %v", rf.me, rf.term)
					rf.role = Follower
				}
			}
		}(p)
	}

}
