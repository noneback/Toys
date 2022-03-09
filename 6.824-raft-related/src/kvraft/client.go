package kvraft

import (
	"crypto/rand"
	"math/big"

	"6.824/labrpc"
)

type Clerk struct {
	servers   []*labrpc.ClientEnd
	id        string
	leaderID  int
	commandID int
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	return &Clerk{
		id:        genUID(),
		leaderID:  0,
		commandID: 0,
		servers:   servers,
	}
}

func (ck *Clerk) Get(key string) string {
	return ck.SendCommand(&Command{Key: key, Op: OpGet})
}

func (ck *Clerk) Put(key string, value string) {
	ck.SendCommand(&Command{Key: key, Value: value, Op: OpPut})
}
func (ck *Clerk) Append(key string, value string) {
	ck.SendCommand(&Command{Key: key, Value: value, Op: OpAppend})
}

func (ck *Clerk) SendCommand(cmd *Command) string {

	req := CommandRequest{
		ClientID:  ck.id,
		CommandID: ck.commandID,
		Cmd:       *cmd,
	}

	DPrintf("[SnedCommand] ck send a command to server %+v\n", req)
	for {
		resp := CommandResponse{}

		if !ck.servers[ck.leaderID].Call("KVServer.HandleCommand", &req, &resp) || resp.Err == ErrTimeout || resp.Err == ErrWrongLeader {
			// not success ,continue
			DPrintf("[Command Failed] with req %+v and resp %+v", req, resp)
			ck.leaderID = (ck.leaderID + 1) % len(ck.servers)
			continue
		}
		ck.commandID++
		return resp.Value
	}
}
