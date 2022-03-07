package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"bytes"
	"sync"
	"sync/atomic"
	"time"

	"6.824/labgob"
	"6.824/labrpc"
)

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
	CommandTerm  int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type NodeState string

const (
	StateLeader    NodeState = "Leader"
	StateFollower  NodeState = "Follower"
	StateCandidate NodeState = "Candidate"
)

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.RWMutex        // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()
	// state a Raft server must maintain.
	applyCond      *sync.Cond
	applyCh        chan ApplyMsg
	state          NodeState
	replicatorCond []*sync.Cond

	currentTerm int
	votedFor    int
	logs        []LogEntry // leave the first entry a dummy entry in order to save snapshot info

	commitIndex int // index of highest log entry known to be committed (initialized to 0, increases monotonically
	lastApplied int // index of highest log entry applied to state machine (initialized to 0 increases monotonically

	nextIndex  []int // index of the next log entry to send to that server (init to leader last log index + 1)
	matchIndex []int // index of highest log entry known to be replicated on server (init to 0 , incs)

	// timer
	electionTimer  *time.Timer
	heartbeatTimer *time.Timer
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.RLock()
	defer rf.mu.RUnlock()
	return rf.currentTerm, StateLeader == rf.state
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	rf.persister.SaveRaftState(rf.encodeStateL())
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}

	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	var currentTerm, votedFor int
	var logs []LogEntry

	if d.Decode(&currentTerm) != nil || d.Decode(&votedFor) != nil || d.Decode(&logs) != nil {
		DPrintf("[Read Persist] Node %v decode failed", rf.me)
		return
	}
	rf.currentTerm, rf.votedFor, rf.logs = currentTerm, votedFor, logs
	rf.lastApplied, rf.commitIndex = rf.logs[0].Index, rf.logs[0].Index // in snapshot
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// if is leader
	if rf.state != StateLeader {
		DPrintf("[Start] Not Leader %v\n", rf.me)
		return -1, -1, false
	}
	DPrintf("[Start] Node %v\n", rf.me)
	// append to log
	newLog := rf.appendEntryL(command)
	// make an agreement
	rf.BroadcastHeartbeatsL(false)
	return newLog.Index, newLog.Term, true
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
func (rf *Raft) ticker() {
	for !rf.killed() {
		select {
		case <-rf.electionTimer.C:
			rf.mu.Lock()
			// state transfer, incs term, start election and reset timer
			rf.ChangeStateL(StateCandidate)
			rf.currentTerm += 1
			rf.startElectionL()
			rf.resetElectionTimerL()
			rf.mu.Unlock()

		case <-rf.heartbeatTimer.C:
			rf.mu.Lock()
			if rf.state == StateLeader {
				// send append entries and reset timer
				rf.BroadcastHeartbeatsL(true)
				rf.resetHeartbeatTimerL()
			}
			rf.mu.Unlock()
		}
	}
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{
		peers:          peers,
		persister:      persister,
		me:             me,
		dead:           0,
		applyCh:        applyCh,
		replicatorCond: make([]*sync.Cond, len(peers)),
		state:          StateFollower,
		currentTerm:    0,
		votedFor:       -1,
		logs:           make([]LogEntry, 1, MaxLogSize),
		nextIndex:      make([]int, len(peers)),
		matchIndex:     make([]int, len(peers)),
		heartbeatTimer: time.NewTimer(GetHeartbeatTimeout()),
		electionTimer:  time.NewTimer(GetRandomElectionTimeout()),
	}
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	lastLog := rf.getLastLogL()
	rf.applyCond = sync.NewCond(&rf.mu)
	for i := 0; i < len(peers); i++ {
		rf.matchIndex[i] = 0
		rf.nextIndex[i] = lastLog.Index + 1

		if i != rf.me {
			rf.replicatorCond[i] = sync.NewCond(&sync.Mutex{})
			go rf.replicator(i)
		}
	}
	// start ticker goroutine to start elections
	go rf.ticker()
	go rf.applier()

	return rf
}

// ChangeStateL only handle with state related attr
func (rf *Raft) ChangeStateL(state NodeState) {
	// change state, voteFor
	if rf.state == state {
		return
	}
	DPrintf("[ChangeState] Node %+v changes state from %+v to %+v in term %+v", rf.me, rf.state, state, rf.currentTerm)
	rf.state = state

	switch rf.state {
	case StateCandidate:
	case StateFollower:
		rf.heartbeatTimer.Stop()
		rf.resetElectionTimerL()
	case StateLeader:
		// A new Leader doesn't know peer's matchIndex, need to reset
		lastLog := rf.getLastLogL()
		for i := 0; i < len(rf.peers); i++ {
			rf.matchIndex[i], rf.nextIndex[i] = 0, lastLog.Index+1
		}
		rf.electionTimer.Stop()
		rf.resetHeartbeatTimerL()
	}
}

// some utils about raft node
func (rf *Raft) encodeStateL() []byte {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.logs)
	return w.Bytes()
}
