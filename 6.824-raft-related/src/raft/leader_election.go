package raft

func (rf *Raft) startElectionL() {
	// send request to all peers
	req := rf.genRequestVoteArgsL()
	grantedVotes := 1
	rf.votedFor = rf.me

	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}
		// start a goroutine to call RequestVote
		go func(p int) {
			resp := &RequestVoteReply{}
			if rf.sendRequestVoteL(p, req, resp) {
				// handle response
				rf.mu.Lock()
				defer rf.mu.Unlock()
				DPrintf("[startElectionL] Node %v receives resp %+v from Node %v after send req %+v in term %v", rf.me, resp, p, req, rf.currentTerm)
				if rf.currentTerm == req.Term && rf.state == StateCandidate {
					// only when candidate term equal to req.term, the resp need to handle, otherwise ignore it.
					if resp.VoteGranted {
						grantedVotes++
						if grantedVotes*2 > len(rf.peers) {
							DPrintf("[startElectionL] Node %v grant majority votes in term %v", rf.me, rf.currentTerm)
							rf.ChangeStateL(StateLeader)
							rf.BroadcastHeartbeatsL(true)
						}
					} else if rf.currentTerm < resp.Term {
						DPrintf("[startElectionL] Node %v find a new leader %v with term %v and steps down in term %v", rf.me, p, resp.Term, rf.currentTerm)
						rf.ChangeStateL(StateFollower)
						rf.currentTerm, rf.votedFor = resp.Term, -1
					}
				}
			}
		}(peer)
	}
}

func (rf *Raft) genRequestVoteArgsL() *RequestVoteArgs {
	lastLog := rf.getLastLogL()
	return &RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateID:  rf.me,
		LastLogIndex: lastLog.Index,
		LastLogTerm:  lastLog.Index,
	}
}

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

func (rf *Raft) RequestVote(req *RequestVoteArgs, resp *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// first of all, refuse all out-dated req, or has voted to other candidate
	if req.Term < rf.currentTerm || (req.Term == rf.currentTerm && rf.votedFor != -1 && rf.votedFor != req.CandidateID) {
		resp.Term = rf.currentTerm
		resp.VoteGranted = false
		return
	}
	// if meet a bigger term req, turn it into follower, update term, reset voteFor
	if req.Term > rf.currentTerm {
		rf.ChangeStateL(StateFollower)
		rf.votedFor, rf.currentTerm = -1, req.Term
	}

	// check if candidate log catches up
	// if !rf.isLogUpToDateL(req.LastLogTerm, req.LastLogIndex) {
	// 	resp.Term, resp.VoteGranted = rf.currentTerm, false
	// 	return
	// }
	// vote for candidate and reset election timer
	rf.votedFor = req.CandidateID
	rf.resetElectionTimerL()
	resp.Term, resp.VoteGranted = rf.currentTerm, true
}

func (rf *Raft) sendRequestVoteL(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}
