package raft

import (
	"sort"
)

type LogEntry struct {
	Cmd   interface{}
	Term  int
	Index int
}

func (rf *Raft) BroadcastHeartbeatsL(isHeartbeat bool) {
	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}
		if isHeartbeat {
			// send heartbeat at once
			DPrintf("[Heartbeat] Node %v broadcast heartbeats in term %v", rf.me, rf.currentTerm)
			go rf.replicateOneRound(peer)
		} else {
			// start replicator to make agreement, and batch send
			rf.replicatorCond[peer].Signal()
		}
	}
}

// replicateOneRound send a AppendEntries rpc to peer and handle resp
func (rf *Raft) replicateOneRound(peer int) {
	rf.mu.Lock()
	if rf.state != StateLeader {
		rf.mu.Unlock()
		return
	}
	prevLogIndex := rf.nextIndex[peer] - 1
	if rf.getFirstLogL().Index > prevLogIndex { //NOTICE: why not equal
		// send Install snapshot Rpc
		req := rf.genInstallSnapshotArgsL()
		rf.mu.Unlock()
		resp := &InstallSnapshotResponse{}
		DPrintf("[InstallSnapshot] Node %v send InstallSnapshotRpc to Node %v with req %+v", rf.me, peer, req)
		if rf.sendInstallSnapshot(peer, req, resp) {
			rf.mu.Lock()
			rf.handleInstallSnapshotResponseL(peer, req, resp)
			rf.mu.Unlock()
		}
	} else {
		req := rf.genAppendEntriesArgsL(prevLogIndex)
		rf.mu.Unlock()
		DPrintf("[ReplicateOneRound] Node %v send AppendEntriesRpc to Node %v with req %+v", rf.me, peer, req)
		resp := &AppendEntriesReply{}
		if rf.sendAppendEntries(peer, req, resp) {
			rf.mu.Lock()
			rf.handleAppendEntriesRespL(peer, req, resp)
			rf.mu.Unlock()
		}
	}
}

func (rf *Raft) handleAppendEntriesRespL(peer int, req *AppendEntriesArgs, resp *AppendEntriesReply) {
	// if not leader or req is out-dated, return
	if rf.state != StateLeader || req.Term < rf.currentTerm {
		return
	}
	if resp.Success {
		// update nextIndex and matchIndex for follower
		rf.matchIndex[peer] = req.PrevLogIndex + len(req.Entries)
		rf.nextIndex[peer] = rf.matchIndex[peer] + 1

		// update leaderCommit
		newCommitIndex := rf.commitIndex
		n := len(rf.matchIndex)
		tmpMatchIndexes := make([]int, n)
		copy(tmpMatchIndexes, rf.matchIndex)
		sort.Ints(tmpMatchIndexes)
		tmpMatchIndexes = ReverseSortedIndexes(tmpMatchIndexes)

		newCommitIndex = Max(newCommitIndex, tmpMatchIndexes[n/2])

		// update commitIndex
		if newCommitIndex > rf.commitIndex {
			// whether the log's term is matched in log[index]
			if rf.matchLogL(rf.currentTerm, newCommitIndex) {
				// DPrintf("[handleAppendEntriesRespL] Node %d advance commitIndex from %d to %d with matchIndex %+v in term %d", rf.me, rf.commitIndex, newCommitIndex, rf.matchIndex, rf.currentTerm)
				// DPrintf("[HandleAppendEntriesResp] log matched {term:%v, index:%v}, update Node %v  commitIndex to %v in term %v", rf.currentTerm, newCommitIndex, rf.me, newCommitIndex, rf.currentTerm)
				DPrintf("[committedIndex] Leader %v commit index %v", rf.me, rf.commitIndex)
				rf.commitIndex = newCommitIndex
				rf.applyCond.Signal()
			} else {
				// DPrintf("[HandleAppendEntriesResp] with req %+v log not matched {term:%v, index:%v}, cannot update Node %v commitIndex to %v in term %v", req, rf.currentTerm, newCommitIndex, rf.me, newCommitIndex, rf.currentTerm)
				DPrintf("[handleAppendEntriesRespL] Node %d can not advance commitIndex from %d because the term of newCommitIndex %d is not equal to currentTerm %d", rf.me, rf.commitIndex, newCommitIndex, rf.currentTerm)
			}
		}
	} else {
		if resp.Term > rf.currentTerm {
			// out-dated term msg, turn to follower
			rf.ChangeStateL(StateFollower)
			rf.currentTerm, rf.votedFor = resp.Term, -1
			rf.persist()
		} else if resp.Term == rf.currentTerm {
			// failed because of log inconsistences
			// NOTICE:
			rf.nextIndex[peer] = resp.ConflictIndex
			if resp.ConflictTerm != -1 {
				// find the last not conflict term log, and reset nextIndex[peer]
				firstIndex := rf.getFirstLogL().Index
				for i := req.PrevLogIndex; i >= firstIndex; i-- {
					if rf.logs[i-firstIndex].Term == resp.ConflictTerm {
						rf.nextIndex[peer] = i + 1
						break
					}
				}
			}
		}
	}
	DPrintf("[after handleAppendEntriesRespL from %v] Node %+v's state is {state %+v,term %+v,commitIndex %+v,lastApplied %+v,firstLog %+v,lastLog %+v} after handling AppendEntriesResponse %+v for AppendEntriesRequest %+v", peer, rf.me, rf.state, rf.currentTerm, rf.commitIndex, rf.lastApplied, rf.getFirstLogL(), rf.getLastLogL(), resp, req)
}

// matchLogL tell log if matched
func (rf *Raft) matchLogL(term int, index int) bool {
	// a log is matched only log exists and has the same term in same index
	return index <= rf.getLastLogL().Index && rf.logs[index-rf.getFirstLogL().Index].Term == term
}

func (rf *Raft) getFirstLogL() *LogEntry {
	return &rf.logs[0]
}

func (rf *Raft) genAppendEntriesArgsL(prevLogIndex int) *AppendEntriesArgs {
	firstIndex := rf.getFirstLogL().Index
	entries := rf.logs[prevLogIndex+1-firstIndex:]

	return &AppendEntriesArgs{
		Term:         rf.currentTerm,
		LeaderID:     rf.me,
		LeaderCommit: rf.commitIndex,
		Entries:      entries,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  rf.logs[prevLogIndex-firstIndex].Term,
	}
}

// isLogUpToDateL check if log catchs up
func (rf *Raft) isLogUpToDateL(lastLogTerm int, lastLogIndex int) bool {
	lastLog := rf.getLastLogL()
	// false if rf.lastLogTerm > req.lastLogTerm, or term equals but has a smaller index
	return lastLogTerm > lastLog.Term || (lastLogTerm == lastLog.Term && lastLogIndex >= lastLog.Index)
}

// send entries related

type AppendEntriesArgs struct {
	Term         int
	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}
type AppendEntriesReply struct {
	Term          int
	Success       bool
	ConflictTerm  int
	ConflictIndex int
}

func (rf *Raft) AppendEntries(req *AppendEntriesArgs, resp *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()

	// DPrintf("[Before AppendEntries] Node %v processing, req %+v", rf.me, req)
	// defer DPrintf("[after RPC log append]Node %v Raft %+v,logs addr %p %+v in req %+v", rf.me, rf, &rf.logs, rf.logs, req)
	defer func(rf *Raft, req *AppendEntriesArgs, resp *AppendEntriesReply) {
		DPrintf("[after AppendEntries] Node %v state is {state %v,term %+v,commitIndex %+v,lastApplied %+v,firstLog %+v,lastLog %+v, logs info %v} After processing AppendEntriesRequest %+v and reply AppendEntriesResponse %+v", rf.me, rf.state, rf.currentTerm, rf.commitIndex, rf.lastApplied, rf.getFirstLogL(), rf.getLastLogL(), LogInfoToString(&rf.logs), req, resp)
	}(rf, req, resp)

	// ignore msg of earlier term
	if req.Term < rf.currentTerm {
		resp.Success, resp.Term = false, rf.currentTerm
		DPrintf("[AppendEntries not Success] Node %v handle AppendEntries not success, req.Term < rf.currentTerm", rf.me)
		return
	}
	// receive from a valid leader
	if req.Term > rf.currentTerm {
		rf.currentTerm, rf.votedFor = req.Term, -1
	}

	rf.ChangeStateL(StateFollower)
	rf.resetElectionTimerL()

	/* check log inconsistences */
	// log haven't be compacted
	if req.PrevLogIndex < rf.getFirstLogL().Index {
		resp.Term, resp.Success = 0, false
		DPrintf("[AppendEntries not Success] Node %v receives unexpected AppendEntriesRequest %+v from {Node %+v} because prevLogIndex %+v < firstLogIndex %+v", rf.me, req, req.LeaderID, req.PrevLogIndex, rf.getFirstLogL().Index)
		return
	}

	// consistences
	if !rf.matchLogL(req.PrevLogTerm, req.PrevLogIndex) {
		resp.Term, resp.Success = rf.currentTerm, false
		lastIndex := rf.getLastLogL().Index
		if lastIndex < req.PrevLogIndex {
			// log not exist and not be compact
			resp.ConflictIndex, resp.ConflictTerm = lastIndex+1, -1
		} else {
			// exist but not consistent
			firstIndex := rf.getFirstLogL().Index
			resp.ConflictTerm = rf.logs[req.PrevLogIndex-firstIndex].Term
			// reset all log entries in conflict term
			index := req.PrevLogIndex - 1
			for index >= firstIndex && rf.logs[index-firstIndex].Term == resp.ConflictTerm {
				index--
			}
			resp.ConflictIndex = index
		}
		return
	}
	// matched, append entries
	// skip duplicated
	firstIndex := rf.getFirstLogL().Index
	for index, entry := range req.Entries {
		// if req.entry is a new one, and
		if entry.Index-firstIndex >= len(rf.logs) || rf.logs[entry.Index-firstIndex].Term != entry.Term {
			rf.logs = append(rf.logs[:entry.Index-firstIndex], req.Entries[index:]...)
		}
	}

	// handle with commitIndex
	newCommitIndex := Min(req.LeaderCommit, rf.getLastLogL().Index)
	if newCommitIndex > rf.commitIndex {
		DPrintf("[AppendEntries] Node %d advance commitIndex from %d to %d with leaderCommit %d in term %d", rf.me, rf.commitIndex, newCommitIndex, req.LeaderCommit, rf.currentTerm)
		rf.commitIndex = newCommitIndex
		rf.applyCond.Signal()
	}

	resp.Term, resp.Success = rf.currentTerm, true
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) getLastLogL() *LogEntry {
	return &rf.logs[len(rf.logs)-1]
}

// replicator is a goroutine that replicate log from leader to peer continuously
func (rf *Raft) replicator(peer int) {
	rf.replicatorCond[peer].L.Lock()
	defer rf.replicatorCond[peer].L.Unlock()

	for !rf.killed() {
		for !rf.needReplicating(peer) {
			rf.replicatorCond[peer].Wait()
		}
		DPrintf("[replicator] Node %v replicator triggered by Leader %v", peer, rf.me)
		rf.replicateOneRound(peer)
	}
}

func (rf *Raft) needReplicating(peer int) bool {
	rf.mu.RLock()
	defer rf.mu.RUnlock()
	return rf.state == StateLeader && rf.matchIndex[peer] < rf.getLastLogL().Index
}

// appendEntryL append entry to the tail of logs
func (rf *Raft) appendEntryL(cmd interface{}) *LogEntry {
	lastLog := rf.getLastLogL()
	newLog := LogEntry{
		Index: lastLog.Index + 1,
		Term:  rf.currentTerm,
		Cmd:   cmd,
	}
	rf.logs = append(rf.logs, newLog)

	DPrintf("[appendEntryL] Node %+v receives a new command {%+v} to replicate in term %+v", rf.me, newLog, rf.currentTerm)
	// update matchIndex and nextIndex
	rf.matchIndex[rf.me], rf.nextIndex[rf.me] = newLog.Index, newLog.Index+1
	rf.persist()
	return &newLog
}

func (rf *Raft) applier() {
	for !rf.killed() {
		rf.mu.Lock()
		for rf.lastApplied >= rf.commitIndex {
			rf.applyCond.Wait()
		}
		// apply
		firstIndex, commitIndex, lastApplied := rf.getFirstLogL().Index, rf.commitIndex, rf.lastApplied
		entries := make([]LogEntry, commitIndex-lastApplied)
		copy(entries, rf.logs[lastApplied+1-firstIndex:commitIndex+1-firstIndex])
		rf.mu.Unlock()

		for _, e := range entries {
			rf.applyCh <- ApplyMsg{
				CommandValid: true,
				Command:      e.Cmd,
				CommandIndex: e.Index,
				CommandTerm:  e.Term,
			}
		}

		rf.mu.Lock()
		DPrintf("[applier] Node %v applies entries %v-%v in term %v", rf.me, rf.lastApplied, commitIndex, rf.currentTerm)
		rf.lastApplied = Max(rf.lastApplied, commitIndex)
		rf.mu.Unlock()
	}
}
