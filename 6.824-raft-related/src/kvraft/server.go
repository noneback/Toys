package kvraft

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"6.824/labgob"
	"6.824/labrpc"
	"6.824/raft"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type Op struct {
	// Your definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
}

type OpsContext struct {
	resp      *CommandResponse
	commandID int
}

type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg
	dead    int32 // set by Kill()

	maxraftstate          int // snapshot if log grows this big
	stateMachine          KVStateMachine
	lastApplied           int
	lastClientAppliedOpts map[string]OpsContext
	notifyChans           map[int]chan *CommandResponse
}

//
// the tester calls Kill() when a KVServer instance won't
// be needed again. for your convenience, we supply
// code to set rf.dead (without needing a lock),
// and a killed() method to test rf.dead in
// long-running loops. you can also add your own
// code to Kill(). you're not required to do anything
// about this, but it may be convenient (for example)
// to suppress debug output from a Kill()ed instance.
func (kv *KVServer) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
}

func (kv *KVServer) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	labgob.Register(&CommandRequest{})
	applyCh := make(chan raft.ApplyMsg)

	kvs := KVServer{
		me:                    me,
		rf:                    raft.Make(servers, me, persister, applyCh),
		maxraftstate:          maxraftstate, // when to snapshot
		applyCh:               applyCh,
		dead:                  0,
		stateMachine:          NewMemoryKV(),                       // kv state machine interface
		lastApplied:           -1,                                  // to prevent state machine from rolling back
		lastClientAppliedOpts: make(map[string]OpsContext),         // cache last opt resp in each client, only cache the last one cuz client send request in a blocking single thread
		notifyChans:           make(map[int]chan *CommandResponse), // to communicate in  main thread and applier goroutine
	}
	go kvs.applier()

	return &kvs
}

func (kvs *KVServer) isDuplicatedRequest(clientID string, reqID int) bool {
	if context, ok := kvs.lastClientAppliedOpts[clientID]; ok && context.commandID == reqID {
		return true
	}
	return false
}

func (kvs *KVServer) HandleCommand(req *CommandRequest, resp *CommandResponse) {
	kvs.mu.Lock()
	if req.Cmd.Op != OpGet && kvs.isDuplicatedRequest(req.ClientID, req.CommandID) {
		lastResp := kvs.lastClientAppliedOpts[req.ClientID].resp
		resp.Value, resp.Err = lastResp.Value, lastResp.Err
		kvs.mu.Unlock()
		return
	}
	kvs.mu.Unlock()
	index, _, isLeader := kvs.rf.Start(req)
	if !isLeader {
		resp.Value, resp.Err = "", ErrWrongLeader
		return
	}

	kvs.mu.Lock()
	ch := kvs.genNotifyChan(index) // for applier
	kvs.mu.Unlock()

	select {
	case result := <-ch:
		resp.Value, resp.Err = result.Value, result.Err
	case <-time.After(ExecuteTimeout):
		resp.Value, resp.Err = "", ErrTimeout
	}

	go func(index int) {
		kvs.mu.Lock()
		defer kvs.mu.Unlock()
		kvs.removeOutdatedNotifyChan(index)
	}(index)
}

func (kvs *KVServer) genNotifyChan(index int) chan *CommandResponse {
	ch := make(chan *CommandResponse)
	kvs.notifyChans[index] = ch
	return ch
}

func (kvs *KVServer) removeOutdatedNotifyChan(index int) {
	if ch, ok := kvs.notifyChans[index]; ok {
		close(ch)
		delete(kvs.notifyChans, index)
	}
}

func (kvs *KVServer) applier() {

	for !kvs.killed() {
		select {
		case msg := <-kvs.applyCh: // blocking opt

			if msg.CommandValid {
				kvs.mu.Lock()
				if msg.CommandIndex <= kvs.lastApplied {
					DPrintf("[applier] Node %v discard outdated message cuz index <= lastApplied", kvs.me)
					kvs.mu.Unlock()
					continue
				}
				// should applied to state machine
				kvs.lastApplied = msg.CommandIndex

				var resp *CommandResponse
				cmdReq := msg.Command.(*CommandRequest) // decode from interface
				if cmdReq.Cmd.Op != OpGet && kvs.isDuplicatedRequest(cmdReq.ClientID, cmdReq.CommandID) {
					DPrintf("[applier] Node %v received a duplicated msg", kvs.me)
					resp = kvs.lastClientAppliedOpts[cmdReq.ClientID].resp
				} else {
					resp = kvs.applyLogToStateMachine(cmdReq)
					if cmdReq.Cmd.Op != OpGet {
						kvs.lastClientAppliedOpts[cmdReq.ClientID] = OpsContext{resp: resp, commandID: cmdReq.CommandID}
					}
				}

				// why we need to check this term
				if currentTerm, isLeader := kvs.rf.GetState(); isLeader && msg.CommandTerm == currentTerm {
					ch := kvs.notifyChans[msg.CommandIndex]
					ch <- resp // send resp to handlerCommand
				}

				kvs.mu.Unlock()

			} else if msg.SnapshotValid {
				kvs.mu.Lock()
				if kvs.rf.CondInstallSnapshot(msg.SnapshotTerm, msg.SnapshotIndex, msg.Snapshot) {
				}
				kvs.mu.Unlock()
			} else {
				panic(fmt.Sprintf("unexpected message %+v", msg))
			}
		case <-time.After(10 * ExecuteTimeout):

		}

	}

}

func (kvs *KVServer) restoreSnapshot() {}

func (kvs *KVServer) applyLogToStateMachine(req *CommandRequest) *CommandResponse {
	resp := CommandResponse{}
	if req.Cmd.Op == OpGet {
		resp.Value, resp.Err = kvs.stateMachine.Get(req.Cmd.Key)
	} else if req.Cmd.Op == OpPut {
		resp.Err = kvs.stateMachine.Put(req.Cmd.Key, req.Cmd.Value)
	} else if req.Cmd.Op == OpAppend {
		resp.Err = kvs.stateMachine.Append(req.Cmd.Key, req.Cmd.Value)
	}
	return &resp
}
