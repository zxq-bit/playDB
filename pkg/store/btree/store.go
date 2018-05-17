package btree

import (
	"bytes"
	"github.com/google/btree"
)

type Store struct {
	root *btree.BTree
}

func NewStore() *Store {
	return &Store{
		root: btree.New(DefaultBtreeDegree),
	}
}

func (s *Store) Set(k, v []byte) (e error) {
	s.root.ReplaceOrInsert(NewKV(k, v))
	return
}
func (s *Store) Get(k []byte) (v []byte, exist bool, err error) {
	s.root.Get(k)
	return
}
func (s *Store) Range(a, b []byte) (kv [][2][]byte, err error) { return }
func (s *Store) Split(k []byte) (a, b Store)                   { return }
func (s *Store) Snapshot() (b []byte, e error)                 { return }

type KV struct {
	Key   []byte
	Value []byte
}

func NewKV(k, v []byte) *KV { return &KV{Key: k, Value: v} }
func (kv *KV) Less(item btree.Item) bool {
	other := item.(*KV)
	if other == nil {
		return false
	}
	return bytes.Compare(kv.Key, other.Key) < 0
}
