package raft

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
