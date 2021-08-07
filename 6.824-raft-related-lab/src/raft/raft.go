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
	//	"bytes"

	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
)

const (
	HeartBeatsTimeout = time.Duration(100) * time.Microsecond
	None              = -1
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

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()
	// NOTICE: 2A
	role                  RoleType
	term                  int
	VotedFor              int
	leader                int
	randomElectionTimeout time.Duration
	nextElectionTime      time.Time
	electionTicker        *time.Timer
	roleChan              chan RoleType
	appChan               chan *AppendEntriesArgs
	heartbeatsChan        chan Msg
	voteChan              chan *RequestVoteArgs

	roleMutex sync.Mutex

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	log         []string
	commitIndex int
	lastApplied int
	nextIndex   []int
	matchIndex  []int
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return rf.term, rf.leader == rf.me
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

//
// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
//
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

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
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
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

func (rf *Raft) tick() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	log.Printf("%v: tick state %v\n", rf.me, rf.role)

	if rf.role == Leader {
		rf.setNextElectionTime()
		rf.sendAppendMsg()
	}

	if time.Now().After(rf.nextElectionTime) {
		log.Printf("%v election timeout\n", rf.me)
		rf.setNextElectionTime()
		rf.startElection()
	}

}

func (rf *Raft) ticker() {
	for rf.killed() == false {
		rf.tick()
		time.Sleep(50 * time.Millisecond)
	}
}

func (rf *Raft) setNextElectionTime() {
	rand.Seed(time.Now().UnixNano())
	ms := 400 + rand.Intn(350)
	rf.nextElectionTime = time.Now().Add(time.Duration(ms) * time.Microsecond)
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	// NOTICE: 2A
	rf.mu = sync.Mutex{}
	rf.roleMutex = sync.Mutex{}
	rf.term = 1
	rf.role = Follower
	rf.VotedFor = None
	rf.leader = None
	rf.setNextElectionTime()

	rf.appChan = make(chan *AppendEntriesArgs)
	rf.voteChan = make(chan *RequestVoteArgs)

	rand.Seed(time.Now().UnixNano())

	// rf.randomElectionTimeout = time.Duration(rand.Int63n(200)+400) * time.Microsecond
	// rf.electionTicker = time.NewTimer(rf.randomElectionTimeout)
	// Your initialization code here (2A, 2B, 2C).

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()

	return rf
}

func (rf *Raft) startElection() {
	// has lock
	rf.newTerm(rf.term + 1)
	rf.role = Candidate
	log.Printf("%v become candidate in term %v", rf.me, rf.term)
	rf.VotedFor = rf.me
	// RequestVote
	args := &RequestVoteArgs{
		Term:        rf.term,
		CandidateID: rf.me,
	}
	voteCnt := 1
	for p := range rf.peers {
		if p == rf.me {
			continue
		}

		go func(peer int) {
			var reply RequestVoteReply

			if ok := rf.sendRequestVote(peer, args, &reply); ok {
				rf.mu.Lock()
				defer rf.mu.Unlock()
				// log.Printf("%v: Request reply in its term %v from %v: %v ", rf.me, rf.term, peer, &reply)
				if reply.Term > rf.term {
					rf.newTerm(reply.Term)
					DPrintf("%v: become follower in term %v", rf.me, rf.term)
					rf.role = Follower
				}

				if reply.VoteGranted {
					voteCnt++
					if voteCnt*2 > len(rf.peers) {
						rf.role = Leader
						log.Printf("[leader] %v: become leader in term %v", rf.me, rf.term)
						rf.leader = rf.me
						rf.sendAppendMsg()
					}
				}

			}
		}(p)
	}

}

func (rf *Raft) sendAppendMsg() {
	// has lock
	args := &AppendEntriesArgs{
		Term:     rf.term,
		LeaderID: rf.me,
	}

	for p := range rf.peers {
		if p == rf.me {
			continue
		}

		go func(peer int) {
			var reply AppendEntriesReply
			if ok := rf.sendAppendEntries(peer, args, &reply); ok {
				rf.mu.Lock()
				defer rf.mu.Unlock()

				if reply.Term > rf.term {
					rf.newTerm(reply.Term)
					rf.role = Follower
				}
			}
		}(p)
	}

}
