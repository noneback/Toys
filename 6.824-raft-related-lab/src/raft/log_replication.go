package raft

type LogEntry struct {
	Term  int
	Index int
	Data  []byte
}

type AppendEntriesArgs struct {
	// NOTICE: 2A
	Term     int
	LeaderID int

	PrevLogIndex int
	PrevLogTerm  int
	Entries      []*LogEntry
	LeaderCommit int
	//ConflictCIndex int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) getLastLog() *LogEntry {
	if len(rf.logs) != 0 {
		return rf.logs[len(rf.logs)-1]
	}
	return &LogEntry{None, None, nil}
}

func (rf *Raft) getFirstLog() *LogEntry {
	if len(rf.logs) == 0 {
		return &LogEntry{None, None, nil}
	}
	return rf.logs[0]
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer DPrintf("[append] %v %v receive AppendEntries in Term %v, req %+v resp %+v\n", rf.role, rf.me, rf.term, args, reply)
	// defer LogsInfo(rf)
	// filter all Msg from the past
	if args.Term < rf.term {
		reply.Term, reply.Success = rf.term, false
		return
	}

	if args.Term > rf.term {
		rf.newTerm(args.Term, args.LeaderID)
	}

	// if rf.role != Follower {
	// 	DPrintf("[role] %v : %v change to Follower in term %v from leader %v\n", rf.role, rf.me, rf.term, args.LeaderID)
	// }

	rf.role = Follower
	rf.resetElectionTimer()
	rf.leader = args.LeaderID

	if !rf.matchLog(args.PrevLogTerm, args.PrevLogIndex) {
		DPrintf("Node %v received unexpected appendMsg %v, prevLog %v mismatch\n", rf.me, args, rf.getLastLog())
		reply.Term, reply.Success = rf.term, false
		return
	}

	// if log matches at beginning
	newEntries := args.Entries
	if len(args.Entries) != 0 {
		// not heartbeats
		// delete conflict logs
		for i, log := range args.Entries {
			lastLog := rf.getLastLog()

			if log.Index > lastLog.Index {
				// all args.Entries is new logEntry
				break
			}
			// delete mismatched log and ignore duplicated log
			if log.Term != rf.logs[log.Index].Term {
				// mismatch
				rf.logs = rf.logs[:log.Index]
			}
			// update new Entries
			newEntries = args.Entries[i:]
		}
	}

	if len(newEntries) != 0 {
		rf.logs = append(rf.logs, newEntries...)
	}

	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = Min(args.LeaderCommit, rf.getLastLog().Index)
	}
	//TODO

	//// if the prev log matches
	//// rule 2
	//if args.PrevLogIndex > rf.getLastLog().index {
	//	DPrintf("Node %v received unexpected appendMsg %v, prevLogIndex is exceed\n", rf.me, args)
	//	reply.Term, reply.Success = rf.Term, false
	//	return
	//}
	//
	//// rule 3
	//if args.PrevLogTerm != rf.logs[args.PrevLogIndex].Term {
	//	DPrintf("Node %v received unexpected appendMsg %v, prevLogTerm mismatched\n",rf.me,args)
	//	reply.Term, reply.Success = rf.Term, false
	//	return
	//}
	//// append entries rule 4
	//rf.logs = append(rf.logs, args.Entries...)
	//// rule 5

	reply.Term, reply.Success = rf.term, true
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) genAppendEntriesArgs(peer int) *AppendEntriesArgs {
	rf.mu.RLock()
	defer rf.mu.RUnlock()

	args := &AppendEntriesArgs{
		Term:         rf.term,
		LeaderID:     rf.me,
		PrevLogIndex: rf.nextIndex[peer] - 1,
		PrevLogTerm:  -1,
		LeaderCommit: rf.commitIndex,
	}

	if len(rf.logs) >= rf.nextIndex[peer] {
		args.Entries = rf.logs[rf.nextIndex[peer]:]
	}
	if len(rf.logs) != 0 && args.PrevLogIndex >= 0 {
		args.PrevLogTerm = rf.logs[args.PrevLogIndex].Term
	}
	return args
}

func (rf *Raft) sendAppendMsg() {
	// has lock
	ackCnt := 1
	for p := range rf.peers {
		if p == rf.me {
			continue
		}
		if rf.role != Leader {
			return
		}

		go func(peer int) {
			// DPrintf("[goroutine] sendAppendMsg gen a go routine")
			args := rf.genAppendEntriesArgs(peer)

			var reply AppendEntriesReply
			if ok := rf.sendAppendEntries(peer, args, &reply); ok {
				rf.mu.Lock()
				defer rf.mu.Unlock()

				if reply.Term > rf.term {
					rf.newTerm(reply.Term, None)
					rf.role = Follower
				}

				if rf.role != Leader || reply.Term < rf.term {
					return
				}

				if reply.Success {
					ackCnt++
					rf.nextIndex[peer] += len(args.Entries)
					DPrintf("[][][] update nextIndex[%v]:%v\n", peer, rf.nextIndex[peer])
					rf.matchIndex[peer] = rf.nextIndex[peer] - 1 // 另外开一个goroutine发送消息通知follower提交记录
					// 更新提交信息
					for i := rf.commitIndex + 1; i < len(rf.logs); i++ {
						commitCnt := 1
						for pr := range rf.peers {
							if rf.matchIndex[pr] >= i {
								commitCnt++
							}
						}
						if commitCnt*2 > len(rf.peers) {
							rf.commitIndex = i
						}
					}
				} else {
					// not success
					if rf.nextIndex[peer] > InitLogIndex {
						rf.nextIndex[peer]--
						DPrintf("[drawback] node %v log conflict\n", rf.me)
					}
				}
			}
		}(p)
	}
}

func (rf *Raft) matchLog(term, index int) bool {
	if index < 0 {
		// means that init
		return true
	}

	if len(rf.logs) < index {
		return false
	}
	return len(rf.logs) == 0 || rf.logs[index].Term == term
}
