package btree

const (
	DefaultBtreeDegree = 4 // TODO make sure

	SnapshotMagic  = 0xBADB
	SnapshotVer0   = 0x0
	SnapshotVerCur = SnapshotVer0

	U32Len = 32 / 8

	MinKvSize      = MinKvSizeVer0
	KvHeadSizeVer0 = U32Len * 2 // allLen, keyLen
	KvTailSizeVer0 = U32Len     // checksum
	MinKvSizeVer0  = KvHeadSizeVer0 + KvTailSizeVer0

	MinStoreSize      = MinStoreSizeVer0
	StoreHeadSizeVer0 = U32Len * 4 // magic, ver, kv num, kv size
	StoreTailSizeVer0 = U32Len     // checksum
	MinStoreSizeVer0  = StoreHeadSizeVer0 + StoreTailSizeVer0
)
