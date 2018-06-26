package raft

import (
	"github.com/zxq-bit/playDB/pkg/proto"
)

type Raft interface {
	Run(stopCh chan struct{}) error
	Step(data []byte) (entry proto.Entry, e error)
	Config() *proto.ReplicaConfig
	// Split(splitCallback func(key []byte) (newStore store.Store, e error)) (newRaft Raft, e error) // TODO
}

type StateMachine interface {
	Apply(command []byte, index uint64) (interface{}, error)
	ApplyMemberChange(confChange *proto.ConfChange, index uint64) (interface{}, error)
	Snapshot() (proto.Snapshot, error)
	ApplySnapshot(peers []string, iter proto.Snapshot) error
	HandleFatalEvent(err error)
	HandleLeaderChange(leader uint64)
}
