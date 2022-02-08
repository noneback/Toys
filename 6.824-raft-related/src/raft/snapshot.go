package raft

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	DPrintf("[snapshot] Node %v make a snapshot at index %v\n", rf.me, index)

	snapshotIndex := rf.getFirstLogL().Index
	if snapshotIndex >= index {
		DPrintf("[snapshot] Node %v snapshot request is lagged, no need to snapshot", rf.me)
		return
	}
	// trim log and snapshot
	rf.logs = rf.logs[index-snapshotIndex:]
	rf.logs[0].Cmd = nil
	rf.persister.SaveStateAndSnapshot(rf.encodeStateL(), snapshot)
}

//
// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
//
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {
	// fmt.Println("[CondInstallSnapshot]...")
	rf.mu.Lock()
	defer rf.mu.Unlock()
	DPrintf("[CondInstallSnapshot] Node %v with lastIncludedTerm %v and lastIncludedIndex %v", rf.me, lastIncludedTerm, lastIncludedIndex)

	// out-dated
	if lastIncludedIndex <= rf.commitIndex {
		DPrintf("[CondInstallSnapshot] Node %v no need to CondInstallSnapshot, cuz snapshot lagged", rf.me)
		return false
	}
	// trim log
	if lastIncludedIndex > rf.getLastLogL().Index {
		rf.logs = make([]LogEntry, 1)
	} else {
		rf.logs = rf.logs[lastIncludedIndex-rf.getFirstLogL().Index:]
		rf.logs[0].Cmd = nil
	}
	// update dummy entry at logs[0] and committedIndex, appliedIndex,
	rf.logs[0].Term, rf.logs[0].Index = lastIncludedTerm, lastIncludedIndex
	rf.lastApplied, rf.commitIndex = lastIncludedIndex, lastIncludedIndex
	rf.persister.SaveStateAndSnapshot(rf.encodeStateL(), snapshot)
	return true
}

type InstallSnapshotRequest struct {
	Term              int
	LeaderID          int
	LastIncludedIndex int
	LastIncludedTerm  int
	Data              []byte
}

type InstallSnapshotResponse struct {
	Term int
}

func (rf *Raft) InstallSnapshot(req *InstallSnapshotRequest, resp *InstallSnapshotResponse) {
	// fmt.Println("[InstallSnapshot] rpc call")
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// out-dated cmd
	if req.Term < rf.currentTerm {
		resp.Term = rf.currentTerm
		return
	}

	if req.Term > rf.currentTerm {
		rf.votedFor, rf.currentTerm = -1, req.Term // may in a leader election stage
		rf.persist()
	}
	rf.ChangeStateL(StateFollower)
	rf.resetElectionTimerL()
	resp.Term = rf.currentTerm
	// out-dated snapshot
	if req.LastIncludedIndex <= rf.commitIndex {
		return
	}

	go func() {
		rf.applyCh <- ApplyMsg{
			SnapshotValid: true,
			Snapshot:      req.Data,
			SnapshotIndex: req.LastIncludedIndex,
			SnapshotTerm:  req.LastIncludedTerm,
		}
	}()
}

func (rf *Raft) sendInstallSnapshot(peer int, req *InstallSnapshotRequest, resp *InstallSnapshotResponse) bool {
	return rf.peers[peer].Call("Raft.InstallSnapshot", req, resp)
}

func (rf *Raft) genInstallSnapshotArgsL() *InstallSnapshotRequest {
	firstLog := rf.getFirstLogL()
	return &InstallSnapshotRequest{
		Term:              rf.currentTerm,
		LeaderID:          rf.me,
		LastIncludedIndex: firstLog.Index,
		LastIncludedTerm:  firstLog.Term,
		Data:              rf.persister.ReadSnapshot(),
	}
}

func (rf *Raft) handleInstallSnapshotResponseL(peer int, req *InstallSnapshotRequest, resp *InstallSnapshotResponse) {
	// if leader election between snapshot rpc send and receive, then no need to handle
	if rf.state == StateLeader && rf.currentTerm == req.Term {
		if resp.Term > rf.currentTerm {
			// rpc failed cuz lower term
			rf.ChangeStateL(StateFollower)
			rf.currentTerm, rf.votedFor = resp.Term, -1
			rf.persist()
		} else {
			rf.matchIndex[peer], rf.nextIndex[peer] = req.LastIncludedIndex, req.LastIncludedIndex+1
		}
	}
}
