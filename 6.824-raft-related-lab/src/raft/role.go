package raft

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type RoleType int

type Msg struct{}

const (
	Leader    RoleType = 1
	Follower  RoleType = 2
	Candidate RoleType = 3
)

func (r RoleType) String() string {
	switch r {
	case Leader:
		return "Leader"
	case Candidate:
		return "Candidate"
	case Follower:
		return "Follower"
	}
	return ""
}

// func (rf *Raft) changeRole(role RoleType, term int) {
// 	log.Printf("[-]raft %v change to %v and term %v\n", rf.me, role, term)
// 	rf.role = role
// 	rf.term = term

// 	switch role {
// 	case Leader:
// 		rf.heartsbeatsTicker.Reset(HeartBeatsTimeout)
// 		rf.electionTicker.Stop()
// 	case Follower:
// 		rf.heartsbeatsTicker.Stop()
// 		rf.electionTicker.Reset(rf.randomElectionTimeout)
// 	case Candidate:
// 		rf.heartsbeatsTicker.Stop()
// 		rf.electionTicker.Reset(rf.randomElectionTimeout)
// 	}
// }

func (rf *Raft) changeRole2(role RoleType) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	log.Printf("[change role]raft %v change to %v\n", rf.me, role)
	rf.role = role
	if role == Candidate {
		rf.term++
	}
}

func (rf *Raft) LeaderProcess() {

	for {
		log.Printf("[+]raft %v heartbeats time out,role %v in term %v\n", rf.me, rf.role, rf.term)
		// send appendEntries
		rf.mu.Lock()
		args := &AppendEntriesArgs{
			Term:     rf.term,
			LeaderID: rf.me,
		}
		rf.mu.Unlock()

		for i := range rf.peers {
			if i == rf.me {
				continue
			}
			reply := &AppendEntriesReply{}
			go func(server int) {
				if ok := rf.sendAppendEntries(server, args, reply); !ok {
					if reply.Success {
						//
					} else {
						rf.roleChan <- Follower
					}
				}
			}(i)
		}

		select {
		case <-time.After(HeartBeatsTimeout):
			log.Printf("[heartbeat]raft %v role %v in term %v", rf.me, rf.role, rf.term)
		case args := <-rf.appChan:
			log.Printf("[heartbeat] raft %v receive heatbeats from %v in term %v\n", rf.me, args.LeaderID, args.Term)
			rf.changeRole2(Follower)
			rf.resetElectionTicker()
			return
		case args := <-rf.voteChan:
			log.Printf("[vote] raft %v vote for %v in term %v\n", rf.me, args.CandidateID, args.Term)
			rf.changeRole2(Follower)
		case <-rf.electionTicker.C:
			log.Println("leader election out")
		}
	}

}

func (rf *Raft) CandidateProcess() {
	// change role to follow -> request vote -> result
	// becomeCandidate
	rf.mu.Lock()
	args := &RequestVoteArgs{
		Term:        rf.term,
		CandidateID: rf.me,
	}
	rf.VotedFor = rf.me
	rf.mu.Unlock()

	voted := 1
	isVoted := false
	votedMu := &sync.Mutex{}
	voteChan := make(chan bool)
	rf.resetElectionTicker()

	for i := range rf.peers {
		if rf.me == i {
			continue
		}
		reply := &RequestVoteReply{}

		go func(server int) {

			defer votedMu.Unlock()

			if ok := rf.sendRequestVote(server, args, reply); ok {
				log.Println("init")
				if reply.VoteGranted {
					log.Printf("[gain vote]candidate %v gain vote from %v\n", rf.me, server)
					votedMu.Lock()
					voted++
					if !isVoted && voted > len(rf.peers)/2 {
						isVoted = true
						votedMu.Unlock()
						voteChan <- true
					}
				} else if reply.Term != 0 {
					log.Println("turn to follower reply:", reply)
					// become follower
					// term == 0 means that the node has voted
					voteChan <- false
				}
				votedMu.Unlock()
			}
		}(i)

		select {
		case <-rf.electionTicker.C:
			rf.resetElectionTicker()
			rf.changeRole2(Candidate)
			return
		case args := <-rf.appChan:
			log.Printf("[heartbeat] raft %v receive heatbeats from %v in term %v\n", rf.me, args.LeaderID, args.Term)
			rf.changeRole2(Follower)
			return
		case args := <-rf.voteChan:
			log.Printf("[vote] raft %v vote for %v in term %v\n", rf.me, args.CandidateID, args.Term)
			rf.changeRole2(Follower)
		case flag := <-voteChan:
			log.Println("-----")
			if flag {
				rf.changeRole2(Leader)
			} else {
				rf.changeRole2(Follower)
			}
			log.Println("+++++")
			return
		}
	}
}

func (rf *Raft) FollowerProcess() {
	for {
		rf.resetElectionTicker()
		select {
		case <-rf.electionTicker.C:
			log.Printf("[lection timeout]raft %v election time out,role %v in term %v\n", rf.me, rf.role, rf.term)
			rf.changeRole2(Candidate)
			return
		case args := <-rf.appChan:
			log.Printf("[heartbeat] raft %v receive heatbeats from %v in term %v\n", rf.me, args.LeaderID, args.Term)
		case args := <-rf.voteChan:
			log.Printf("[vote] raft %v vote for %v in term %v\n", rf.me, args.CandidateID, args.Term)

		}
	}

}

func (rf *Raft) resetElectionTicker() {
	rand.Seed(time.Now().UnixNano())
	randomElectionTimeout := time.Duration(rand.Int63n(600)+400) * time.Microsecond
	rf.electionTicker.Reset(randomElectionTimeout)
}
