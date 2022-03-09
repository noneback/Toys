package kvraft

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"6.824/labgob"
	"6.824/labrpc"
	"6.824/raft"
)

const Debug = false

func init() {
	f, err := os.OpenFile("./test/log.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)
}

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type OpsContext struct {
	Resp      *CommandResponse
	CommandID int
}

type KVServer struct {
	mu                    sync.Mutex
	me                    int
	rf                    *raft.Raft
	applyCh               chan raft.ApplyMsg
	dead                  int32 // set by Kill()
	persister             *raft.Persister
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
	// labgob.Register(&CommandRequest{})
	labgob.Register(CommandRequest{})
	labgob.Register(&SnapshotData{})
	labgob.Register(&MemoryKV{})

	applyCh := make(chan raft.ApplyMsg)

	kvs := KVServer{
		me:                    me,
		rf:                    raft.Make(servers, me, persister, applyCh),
		maxraftstate:          maxraftstate, // when to snapshot
		applyCh:               applyCh,
		dead:                  0,
		stateMachine:          NewMemoryKV(), // kv state machine interface
		persister:             persister,
		lastApplied:           0,                                   // to prevent state machine from rolling back
		lastClientAppliedOpts: make(map[string]OpsContext),         // cache last opt resp in each client, only cache the last one cuz client send request in a blocking single thread
		notifyChans:           make(map[int]chan *CommandResponse), // to communicate in  main thread and applier goroutine
	}

	kvs.restoreFromSnapshot(kvs.persister.ReadSnapshot())
	go kvs.applier()

	return &kvs
}

func (kvs *KVServer) isDuplicatedRequest(clientID string, reqID int) bool {
	if context, ok := kvs.lastClientAppliedOpts[clientID]; ok && context.CommandID == reqID {
		return true
	}
	return false
}

func (kvs *KVServer) HandleCommand(req *CommandRequest, resp *CommandResponse) {
	DPrintf("[HandleCommand] before processing req %+v\n", req)
	kvs.mu.Lock()
	// if req == nil {
	// 	panic("[HandleCommand] receive a nil req")
	// }
	if req.Cmd.Op != OpGet && kvs.isDuplicatedRequest(req.ClientID, req.CommandID) {
		lastResp := kvs.lastClientAppliedOpts[req.ClientID].Resp
		resp.Value, resp.Err = lastResp.Value, lastResp.Err
		kvs.mu.Unlock()
		return
	}
	kvs.mu.Unlock()
	index, _, isLeader := kvs.rf.Start(*req)
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
				DPrintf("[applier] receive msg from raft layer: %+v, %v\n", msg, kvs.lastApplied)
				kvs.mu.Lock()
				if msg.CommandIndex <= kvs.lastApplied {
					DPrintf("[applier] Node %v discard outdated message cuz index <= lastApplied", kvs.me)
					kvs.mu.Unlock()
					continue
				}
				// should applied to state machine
				kvs.lastApplied = msg.CommandIndex

				var resp *CommandResponse
				cmdReq, ok := msg.Command.(CommandRequest)
				if !ok {
					fmt.Printf("[type assert failed] %+v\n", msg)
					continue
				} // decode from interface
				if cmdReq.Cmd.Op != OpGet && kvs.isDuplicatedRequest(cmdReq.ClientID, cmdReq.CommandID) {
					DPrintf("[applier] Node %v received a duplicated msg", kvs.me)
					resp = kvs.lastClientAppliedOpts[cmdReq.ClientID].Resp
				} else {
					resp = kvs.applyLogToStateMachine(&cmdReq)
					if cmdReq.Cmd.Op != OpGet {
						kvs.lastClientAppliedOpts[cmdReq.ClientID] = OpsContext{Resp: resp, CommandID: cmdReq.CommandID}
					}
				}

				// why we need to check this term
				if currentTerm, isLeader := kvs.rf.GetState(); isLeader && msg.CommandTerm == currentTerm {
					ch := kvs.notifyChans[msg.CommandIndex]
					ch <- resp // send resp to handlerCommand
				}
				if kvs.needSnapshot() {
					kvs.takeSnapshot(msg.CommandIndex)
				}
				kvs.mu.Unlock()
			} else if msg.SnapshotValid {
				kvs.mu.Lock()
				if kvs.rf.CondInstallSnapshot(msg.SnapshotTerm, msg.SnapshotIndex, msg.Snapshot) {
					kvs.restoreFromSnapshot(msg.Snapshot)
					kvs.lastApplied = msg.SnapshotIndex
				}
				kvs.mu.Unlock()
			}
		case <-time.After(10 * ExecuteTimeout):

		}

	}

}

func (kvs *KVServer) applyLogToStateMachine(req *CommandRequest) *CommandResponse {
	resp := CommandResponse{
		Err:   OK,
		Value: "",
	}
	if req.Cmd.Op == OpGet {
		resp.Value, resp.Err = kvs.stateMachine.Get(req.Cmd.Key)
	} else if req.Cmd.Op == OpPut {
		resp.Err = kvs.stateMachine.Put(req.Cmd.Key, req.Cmd.Value)
	} else if req.Cmd.Op == OpAppend {
		resp.Err = kvs.stateMachine.Append(req.Cmd.Key, req.Cmd.Value)
	}
	return &resp
}

// about snapshot
type SnapshotData struct {
	StateMachine          KVStateMachine
	LastClientAppliedOpts map[string]OpsContext
}

func (kvs *KVServer) needSnapshot() bool {
	return kvs.maxraftstate != -1 && kvs.persister.RaftStateSize() >= kvs.maxraftstate
}

func (kvs *KVServer) takeSnapshot(index int) {
	w := &bytes.Buffer{}
	e := labgob.NewEncoder(w)
	if err := e.Encode(&SnapshotData{
		StateMachine:          kvs.stateMachine,
		LastClientAppliedOpts: kvs.lastClientAppliedOpts,
	}); err == nil {
		kvs.rf.Snapshot(index, w.Bytes())
	} else {
		panic(fmt.Sprintf("take snapshot failed: %+v", err))
	}

}

func (kvs *KVServer) restoreFromSnapshot(snapshot []byte) {
	if len(snapshot) == 0 {
		return
	}

	r := bytes.NewBuffer(snapshot)
	d := labgob.NewDecoder(r)
	snapshotData := &SnapshotData{}
	if err := d.Decode(&snapshotData); err == nil {
		kvs.stateMachine = snapshotData.StateMachine
		kvs.lastClientAppliedOpts = snapshotData.LastClientAppliedOpts
	} else {
		panic(fmt.Sprintf("restore form snapshot failed: %+v", err))
	}
}
