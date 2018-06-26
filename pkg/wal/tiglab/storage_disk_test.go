package tiglab

import (
	"math"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/zxq-bit/playDB/pkg/proto"
	"github.com/zxq-bit/playDB/pkg/wal"
)

const (
	testBaseDir = "/tmp/shardDB_storage"
)

var (
	testShardId          uint64
	testRaftLog          *DiskRotateStorage
	testEntries          []*proto.Entry
	testEntryNum         = LogMaxEntryNum + LogEntryCacheNum
	testEntryDataMaxSize = 65536
)

func TestInit(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	// rand.Seed(0)
	testShardId = uint64(rand.Int63())
	e := os.MkdirAll(testBaseDir, 0755)
	if e != nil {
		t.Fatalf("MkdirAll %v failed, %v", testBaseDir, e)
	}
	testEntries = makeEntries(1, testEntryNum, testEntryDataMaxSize)
}

func Test_Interface(t *testing.T) {
	var (
		si wal.WAL
		e  error
	)
	testRaftLog, e = NewDiskRotateStorage(testBaseDir, testShardId)
	if e != nil {
		t.Fatalf("NewDiskRotateStorage %v %v failed, %v", testBaseDir, testShardId, e)
	}
	si = testRaftLog
	t.Logf("NewDiskRotateStorage return %v, %v", si, e)
}

func Test_Write(t *testing.T) {
	raftLog := testRaftLog
	if raftLog == nil {
		t.Fatalf("NewDiskRotateStorage %v %v failed, exit.", testBaseDir, testShardId)
	}

	e := raftLog.StoreEntries(testEntries)
	if e != nil {
		t.Fatalf("StoreEntries testEntries(%d) failed, %v", len(testEntries), e)
	}

	lhs := [][2]uint64{
		{1, LogEntryCacheNum},
		{1, LogMaxEntryNum + LogEntryCacheNum},
		{1, LogMaxEntryNum},
		{LogMaxEntryNum + 1, LogMaxEntryNum + LogEntryCacheNum},
	}
	for i := range lhs {
		lo, hi := lhs[i][0], lhs[i][1]
		entries := testEntries[lo-1 : hi]
		e := raftLog.StoreEntries(entries)
		if e != nil {
			t.Fatalf("StoreEntries testEntries[%d:%d] len=%d failed, %v", lo, hi, len(entries), e)
		}
		t.Logf("StoreEntries[%d, %d] len=%d", lo, hi, len(entries))
	}
}

func Test_Read(t *testing.T) {
	raftLog := testRaftLog
	if raftLog == nil {
		t.Fatalf("NewDiskRotateStorage %v %v failed, exit.", testBaseDir, testShardId)
	}

	e := raftLog.StoreEntries(testEntries)
	if e != nil {
		t.Fatalf("StoreEntries testEntries(%d) failed, %v", len(testEntries), e)
	}

	lhs := [][2]uint64{
		{1, LogEntryCacheNum},
		{1, LogMaxEntryNum},
		{1, LogMaxEntryNum + LogEntryCacheNum},
		{LogMaxEntryNum + 1, LogMaxEntryNum + LogEntryCacheNum},
	}
	for i := range lhs {
		lo, hi := lhs[i][0], lhs[i][1]+1
		es, isCompact, e := raftLog.Entries(lo, hi, math.MaxUint32)
		if e != nil {
			t.Fatalf("Entries[%d, %d) testEntries(%d) failed, %v", lo, hi, len(testEntries), e)
		}
		t.Logf("Entries[%d, %d), isCompact=%v, entryNum=%d", lo, hi, isCompact, len(es))
	}
}

func TestCleanup(t *testing.T) {
	if testRaftLog != nil {
		testRaftLog.Close()
	}
	os.RemoveAll(testBaseDir)
}

func makeEntries(startIndex uint64, entriesNum, maxDataSize int) []*proto.Entry {
	data := make([]byte, maxDataSize)
	entries := make([]*proto.Entry, entriesNum)
	for i := range entries {
		entries[i] = &proto.Entry{
			Index: uint64(i) + startIndex,
			Term:  1,
			Type:  proto.EntryNormal,
			Data:  data[:rand.Intn(maxDataSize)],
		}
	}
	return entries
}
