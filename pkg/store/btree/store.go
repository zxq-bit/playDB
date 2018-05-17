package btree

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"sync"
	"sync/atomic"

	"github.com/google/btree"

	"github.com/zxq-bit/playDB/pkg/errors"
	"github.com/zxq-bit/playDB/pkg/store"
)

type Store struct {
	pSize uint32 // pack size // should less than (2 << 32)
	root  *btree.BTree
}

func NewStore() *Store {
	return &Store{
		root: newRoot(),
	}
}

func newRoot() *btree.BTree {
	return btree.New(DefaultBtreeDegree)
}

func (s *Store) Set(k, v []byte) (e error) {
	var (
		prev *KV
		next = NewKV(k, v)
	)
	item := s.root.ReplaceOrInsert(next)
	if item != nil {
		prev = item.(*KV)
	}
	if prev != nil {
		atomic.AddUint32(&s.pSize, next.pSize+^(prev.pSize-1))
	} else {
		atomic.AddUint32(&s.pSize, next.pSize)
	}
	return
}
func (s *Store) Get(k []byte) (v []byte, exist bool, err error) {
	item := s.root.Get(NewK(k))
	if item == nil {
		return nil, false, nil
	}
	kv := item.(*KV)
	if kv == nil {
		return nil, false, errors.New().ErrBadValue(k)
	}
	return kv.Value, true, nil
}
func (s *Store) Del(k []byte) (exist bool, err error) {
	var prev *KV
	item := s.root.Delete(NewK(k))
	if item != nil {
		prev = item.(*KV)
	}
	if prev != nil {
		atomic.AddUint32(&s.pSize, ^(prev.pSize - 1))
	}
	return prev != nil, nil
}

func (s *Store) Range(a, b []byte) (kvs [][2][]byte, err error) {
	iter := func(item btree.Item) bool {
		if item == nil {
			return true
		}
		kv := item.(*KV)
		if kv == nil {
			return true
		}
		kvs = append(kvs, [2][]byte{kv.Key, kv.Value})
		return true
	}
	s.root.AscendRange(NewK(a), NewK(b), iter)
	return
}

func (s *Store) Split(k []byte) (a, b store.Store) {
	as, bs := NewStore(), NewStore()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	splitKey := NewK(k)
	// TODO more efficient
	go func() {
		s.root.AscendLessThan(splitKey, func(item btree.Item) bool {
			as.root.ReplaceOrInsert(item)
			return true
		})
		wg.Done()
	}()
	go func() {
		s.root.AscendGreaterOrEqual(splitKey, func(item btree.Item) bool {
			bs.root.ReplaceOrInsert(item)
			return true
		})
		wg.Done()
	}()
	wg.Wait()
	return as, bs
}
func (s *Store) Snapshot() (b []byte, e error) {
	return s.snapshotVer0()
}
func (s *Store) Recover(b []byte) (e error) {
	// check buffer
	lb := uint32(len(b))
	if lb < MinStoreSize {
		return errors.New().ErrDataTooSmall("store", MinKvSize, lb)
	}
	// check magic
	if magic := binary.LittleEndian.Uint32(b); magic != SnapshotMagic {
		return errors.New().ErrBadSnapshotMagic(SnapshotMagic, magic)
	}
	// check ver // no choice now
	ver := binary.LittleEndian.Uint32(b[U32Len:])
	switch ver {
	case SnapshotVer0:
		return s.recoverVer0(b)
	default:
		return errors.New().ErrUnknownSnapshotVer(ver)
	}
}
func (s *Store) recoverVer0(b []byte) (e error) {
	// TODO more efficient
	// check buffer
	lb := uint32(len(b))
	// check size & num
	kvNum := binary.LittleEndian.Uint32(b[U32Len*2:])
	pSize := binary.LittleEndian.Uint32(b[U32Len*3:])
	if clb := storePackSizeVer0(pSize); clb != lb {
		return errors.New().ErrSnapshotSizeMismatch(clb, lb)
	}
	if kvNum*MinKvSize > pSize {
		return errors.New().ErrSnapshotKvNumTooMany(pSize/MinKvSize, kvNum)
	}
	// checksum
	cr := binary.LittleEndian.Uint32(b[lb-StoreTailSizeVer0:])
	cc := crc32.ChecksumIEEE(b[:lb-StoreTailSizeVer0])
	if cc != cr {
		return errors.New().ErrChecksumMismatch("store", cr, cc)
	}
	// read data
	kvData := b[StoreHeadSizeVer0 : StoreHeadSizeVer0+pSize]
	pos := uint32(0)
	num := uint32(0)
	root := newRoot()
	for pos < pSize {
		var kv KV
		n, e := kv.Parse(kvData[pos:])
		if e != nil {
			return e
		}
		num++
		pos += n
		root.ReplaceOrInsert(NewKV(kv.Key, kv.Value))
	}
	if pos != pSize { // should error on kv.Parse
		return errors.New().ErrSnapshotKvSizeMismatch(pSize, pos)
	}
	if num != kvNum { // should error on kv.Parse
		return errors.New().ErrSnapshotKvSizeMismatch(kvNum, num)
	}
	// update
	s.pSize = pSize
	s.root = root
	return nil
}
func (s *Store) snapshotVer0() (b []byte, e error) {
	// TODO more efficient
	// buffer
	pSize := s.getPackSize()
	bSize := storePackSizeVer0(pSize)
	b = make([]byte, bSize)
	// header
	kvNum := uint32(s.root.Len())
	binary.LittleEndian.PutUint32(b, SnapshotMagic)
	binary.LittleEndian.PutUint32(b[U32Len:], SnapshotVerCur)
	binary.LittleEndian.PutUint32(b[U32Len*2:], kvNum)
	binary.LittleEndian.PutUint32(b[U32Len*3:], pSize)
	// data
	pos := uint32(StoreHeadSizeVer0)
	num := 0
	s.root.Ascend(func(item btree.Item) bool {
		var n uint32
		if item == nil {
			return true
		}
		kv := item.(*KV)
		if kv == nil {
			return true
		}
		n, _, e = kv.Pack(b[pos:])
		if e != nil {
			return false
		}
		num++
		pos += n
		return true
	})
	if pos-StoreHeadSizeVer0 != pSize {
		return nil, errors.New().ErrStoreKvSizeMismatch(pSize, pos-StoreHeadSizeVer0)
	}
	// checksum
	checksum := crc32.ChecksumIEEE(b[:pos])
	binary.LittleEndian.PutUint32(b[pos:], checksum)
	return b, nil
}

func (s *Store) getPackSize() uint32 { return atomic.LoadUint32(&s.pSize) }

type KV struct { // no modify
	pSize uint32 // pack size
	Key   []byte
	Value []byte
}

func NewK(k []byte) *KV {
	return &KV{Key: k}
}
func NewKV(k, v []byte) *KV {
	// calc
	kLen := len(k)
	kvLen := kLen + len(v)
	// copy
	buf := make([]byte, kvLen)
	copy(buf[:kLen], k)
	copy(buf[kLen:], v)
	// set
	kv := &KV{Key: buf[:kLen], Value: buf[kLen:]}
	kv.pSize = kvPackSize(kvLen)
	return kv
}
func (kv *KV) Less(item btree.Item) bool {
	if item == nil {
		return false
	}
	other := item.(*KV)
	if other == nil {
		return false
	}
	return bytes.Compare(kv.Key, other.Key) < 0
}

func (kv *KV) Pack(b []byte) (n, checksum uint32, e error) {
	if len(b) < int(kv.pSize) {
		return 0, 0, errors.New().ErrOutOfBuffer(kv.pSize, len(b))
	}
	keyLen := uint32(len(kv.Key))
	// header
	binary.LittleEndian.PutUint32(b[0:], kv.pSize)
	binary.LittleEndian.PutUint32(b[U32Len:], keyLen)
	// data
	copy(b[KvHeadSizeVer0:], kv.Key)
	copy(b[KvHeadSizeVer0+keyLen:], kv.Value)
	// checksum
	checksum = crc32.ChecksumIEEE(b[:kv.pSize-KvTailSizeVer0])
	binary.LittleEndian.PutUint32(b[kv.pSize-KvTailSizeVer0:], checksum)
	// n
	n = kv.pSize
	return
}
func (kv *KV) Parse(b []byte) (n uint32, e error) {
	// check buffer
	lb := uint32(len(b))
	if lb < MinKvSize {
		return 0, errors.New().ErrDataTooSmall("kv", MinKvSize, lb)
	}
	// check size
	kv.pSize = binary.LittleEndian.Uint32(b)
	if kv.pSize > lb {
		return 0, errors.New().ErrKVSizeOut(lb, kv.pSize)
	}
	// check key len
	keyLen := binary.LittleEndian.Uint32(b[U32Len:])
	kvLen := kv.pSize - MinKvSizeVer0
	if keyLen > kvLen {
		return 0, errors.New().ErrKeySizeOut(kvLen, keyLen)
	}
	// checksum
	checkLen := kv.pSize - KvTailSizeVer0
	cr := binary.LittleEndian.Uint32(b[checkLen:]) // checksum read
	cc := crc32.ChecksumIEEE(b[:checkLen])         // checksum calc
	if cr != cc {
		return 0, errors.New().ErrChecksumMismatch("kv", cr, cc)
	}
	// copy
	buf := make([]byte, kvLen)
	copy(buf, b[KvHeadSizeVer0:])
	kv.Key = buf[:keyLen]
	kv.Value = buf[keyLen:]
	return kv.pSize, nil
}

func kvPackSize(kvLen int) uint32 {
	return kvPackSizeVer0(kvLen)
}

// ver0

func kvPackSizeVer0(kvLen int) uint32 {
	return MinKvSize + uint32(kvLen)
}

func storePackSizeVer0(kvSize uint32) uint32 {
	return MinStoreSize + kvSize
}
