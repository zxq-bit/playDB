package proto

import (
	"github.com/zxq-bit/playDB/pkg/raft/etcd/raftpb"
)

type ConfChange = raftpb.ConfChange
type ConfChangeType = raftpb.ConfChangeType
type ConfState = raftpb.ConfState
type Entry = raftpb.Entry
type EntryType = raftpb.EntryType
type HardState = raftpb.HardState
type Message = raftpb.Message
type MessageType = raftpb.MessageType
type Snapshot = raftpb.Snapshot
type SnapshotMetadata = raftpb.SnapshotMetadata

type ReplicaConfig struct {
	ID    int
	Peers []string
}

type SnapIterator interface {
	// if error=io.EOF represent snapshot terminated.
	Next() ([]byte, error)
}

type SnapshotMeta struct {
	Index uint64
	Term  uint64
	Peers []string
}

const (
	ConfAddNode    ConfChangeType = 0
	ConfRemoveNode ConfChangeType = 1
	ConfUpdateNode ConfChangeType = 2

	EntryNormal     EntryType = 0
	EntryConfChange EntryType = 1
)
