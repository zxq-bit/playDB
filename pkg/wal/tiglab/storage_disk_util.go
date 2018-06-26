package tiglab

import (
	"errors"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/zxq-bit/playDB/pkg/proto"
)

const (
	LogMagic             = 0x1EEE // uint32
	LogMaxEntryNum       = 8192
	LogMetaHeadSize      = 4 + 4 + 8 + 4 + 4 // magic + crc + shard id + cap + entry num
	LogFileMetaEntrySize = 8 + 8 + 8         // index + term + offset
	LogFileMetaSize      = LogMetaHeadSize + LogMaxEntryNum*LogFileMetaEntrySize
	LogEntryHeadSize     = 4 + 4 + 1 + 8 + 8     // crc + data size + type + term + index
	LogHsHeadSize        = 4 + 4 + 8 + 8 + 4 + 4 // magic + crc + shard id + seq id + data size + data size

	LogEntryCacheNum = 128
	LogFileCacheNum  = 5
	LogHardStateNum  = 2
)

var (
	entryHeadPool    = NewBlockBuffer(LogEntryHeadSize)
	metaBufferPool   = NewBlockBuffer(LogFileMetaSize)
	hsHeadBufferPool = NewBlockBuffer(LogHsHeadSize)

	ErrBufferSizeNotEnough   = errors.New("BufferSizeNotEnough")
	ErrLogFileFull           = errors.New("LogFileFull")
	ErrLogFileEmpty          = errors.New("LogFileEmpty")
	ErrLogEmpty              = errors.New("LogFileEmpty")
	ErrLogDirNotExist        = errors.New("ErrLogDirNotExist")
	ErrLogDirCannotAccess    = errors.New("ErrLogDirCannotAccess")
	ErrInvalidFilePos        = errors.New("InvalidFilePos")
	ErrIndexOutOfRange       = errors.New("IndexOutOfRange")
	ErrIndexOutOfFileRange   = errors.New("IndexOutOfFileRange")
	ErrLastIndexNotMatch     = errors.New("LastIndexNotMatch")
	ErrFileLastIndexNotMatch = errors.New("FileLastIndexNotMatch")
	ErrLogFileIncomplete     = errors.New("LogFileIncomplete")
	ErrLossLogFileInMiddle   = errors.New("LossLogFileInMiddle")
	ErrNilInput              = errors.New("NilInput")
	ErrEmptyHardState        = errors.New("EmptyHardState")
	ErrEmptySnapshotMeta     = errors.New("EmptySnapshotMeta")
	ErrNoHardStateFile       = errors.New("NoHardStateFile")
	ErrCrcNotMatch           = errors.New("CrcNotMatch")
	ErrBadMeta               = errors.New("BadMeta")
	ErrFirstIndexNotMatch    = errors.New("FirstIndexNotMatch")
	ErrBadMagic              = errors.New("BadMagic")
	ErrFileNameNotMatch      = errors.New("FileNameNotMatch")
)

type BlockBuffer struct {
	bufPool   *sync.Pool
	blockSize int
}

func NewBlockBuffer(blockSize int) *BlockBuffer {
	bb := &BlockBuffer{
		blockSize: blockSize,
		bufPool:   new(sync.Pool),
	}
	bb.bufPool.New = func() interface{} {
		return make([]byte, bb.blockSize)
	}
	return bb
}
func (bb *BlockBuffer) GetBuffer() []byte {
	return bb.bufPool.Get().([]byte)
}
func (bb *BlockBuffer) PutBuffer(b []byte) {
	if cap(b) >= bb.blockSize {
		bb.bufPool.Put(b[:bb.blockSize])
	}
}
func (bb *BlockBuffer) BlockSize() uint32 {
	return uint32(bb.blockSize)
}

func fileSize(f *os.File) (int64, error) {
	fi, e := f.Stat()
	if e != nil {
		return 0, e
	}
	return fi.Size(), nil
}
func getLogFilePos(index uint64) int {
	return int((index - 1) / LogMaxEntryNum)
}
func getLogFileFirstIndex(pos int) uint64 {
	return uint64(pos*LogMaxEntryNum + 1)
}
func getLogFileLastIndex(pos int) uint64 {
	return uint64(pos+1) * LogMaxEntryNum
}
func LogFileName(shardId, firstIndex, lastIndex uint64) string {
	return fmt.Sprintf("log_%d.%010x-%010x", shardId, firstIndex, lastIndex)
}
func LogFileNameByPos(shardId uint64, pos int) string {
	return LogFileName(shardId, getLogFileFirstIndex(pos), getLogFileLastIndex(pos))
}
func LogFileNameByIndex(shardId, index uint64) string {
	pos := getLogFilePos(index)
	return LogFileName(shardId, getLogFileFirstIndex(pos), getLogFileLastIndex(pos))
}
func ParseLogFileName(fileName string) (shardId, firstIndex, lastIndex uint64, e error) {
	if _, e = fmt.Sscanf(fileName, "log_%d.%010x-%010x", &shardId, &firstIndex, &lastIndex); e != nil {
		return shardId, firstIndex, lastIndex, e
	}
	if fileName != LogFileName(shardId, firstIndex, lastIndex) {
		return shardId, firstIndex, lastIndex, ErrFileNameNotMatch
	}
	return shardId, firstIndex, lastIndex, nil
}
func CurLogFileName(shardId uint64) string {
	return fmt.Sprintf("log_%d.current", shardId)
}
func LogFilePrefix(shardId uint64) string {
	return fmt.Sprintf("log_%d.", shardId)
}
func HsFileName(shardId, seqId uint64) string {
	return fmt.Sprintf("hs_%d.%x", shardId, seqId%LogHardStateNum)
}

func CutEntriesMaxSize(entries []*proto.Entry, maxSize uint64) []*proto.Entry {
	if len(entries) <= 1 || maxSize == math.MaxUint32 {
		return entries
	}
	size := entries[0].Size()
	for i := 1; i < len(entries); i++ {
		size += entries[i].Size()
		if uint64(size) > maxSize {
			return entries[:i]
		}
	}
	return entries
}
