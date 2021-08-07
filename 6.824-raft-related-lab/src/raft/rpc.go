package raft

import "log"

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

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// log.Printf("[vote] candidate %v request vote\n", args.CandidateID)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// if caller.term > cur.term
	if args.Term > rf.term {
		rf.newTerm(args.Term)
		rf.role = Follower
	}
	reply.VoteGranted = false
	// log.Printf("[test] %v and rf.term %v and vote %v", args, rf.term, rf.VotedFor)
	if rf.VotedFor == None || rf.VotedFor == args.CandidateID {
		// log.Printf("[init] args %+v, rf.term %v,rf.vote %v\n", args, rf.term, rf.VotedFor)
		reply.Term = rf.term
		reply.VoteGranted = true
		rf.VotedFor = args.CandidateID
		rf.setNextElectionTime()
		return
	}
	reply.Term = rf.term
}

type AppendEntriesArgs struct {
	// NOTICE: 2A
	Term     int
	LeaderID int

	PrevLogIndex int
	Entries      []string
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	log.Printf("[heartbeat] %v receive heartbeat in term %v from %v\n", rf.me, rf.term, args.Term)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// if caller.term < cur.term
	reply.Success = false
	if args.Term >= rf.term {
		rf.newTerm(args.Term)
		rf.role = Follower
		rf.leader = args.LeaderID
		reply.Success = true
		rf.setNextElectionTime()
	}

	reply.Term = rf.term
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) newTerm(term int) {
	rf.term = term
	rf.VotedFor = None
}
