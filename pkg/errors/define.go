package errors

import "fmt"

var (
	// get
	ReasonBadValue = "BadValue"
	// kv pack
	ReasonOutOfBuffer = "OutOfBuffer"
	// kv parse
	ReasonDataTooSmall     = "DataTooSmall"
	ReasonKVSizeOut        = "KeyValueSizeOut"
	ReasonKeySizeOut       = "KeySizeOut"
	ReasonChecksumMismatch = "ChecksumMismatch"
	// snapshot
	ReasonStoreKvSizeMismatch = "StoreKvSizeMismatch"
	// recover
	ReasonBadSnapshotMagic            = "BadSnapshotMagic"
	ReasonUnknownSnapshotVer          = "UnknownSnapshotVer"
	ReasonSnapshotPackageSizeMismatch = "SnapshotPackageSizeMismatch"
	ReasonSnapshotKvNumTooMany        = "SnapshotKvNumTooMany"
	ReasonSnapshotKvSizeMismatch      = "SnapshotKvSizeMismatch"
	ReasonSnapshotKvNumMismatch       = "SnapshotKvNumMismatch"
)

func bytesPrint(b []byte) string {
	return fmt.Sprintf("[%d]'%s'", len(b), string(b))
}

func (e *ErrBuilder) ErrBadValue(k []byte) error {
	return e.SetReason(ReasonBadValue).SetFormatMsg("kv pair of %s store in bad format", bytesPrint(k))
}
func (e *ErrBuilder) ErrOutOfBuffer(need uint32, got int) error {
	return e.SetReason(ReasonOutOfBuffer).SetFormatMsg("need %d got %d", need, got)
}
func (e *ErrBuilder) ErrDataTooSmall(name string, min, got uint32) error {
	return e.SetReason(ReasonDataTooSmall).SetFormatMsg("%s need %d at least got %d", name, min, got)
}
func (e *ErrBuilder) ErrKVSizeOut(bufSize, kvSize uint32) error {
	return e.SetReason(ReasonKVSizeOut).SetFormatMsg("buffer size is %d, kv size is %d", bufSize, kvSize)
}
func (e *ErrBuilder) ErrKeySizeOut(kvLen, keyLen uint32) error {
	return e.SetReason(ReasonKeySizeOut).SetFormatMsg("total kv size is %d, key size is %d", kvLen, keyLen)
}
func (e *ErrBuilder) ErrChecksumMismatch(name string, read, calc uint32) error {
	return e.SetReason(ReasonChecksumMismatch).SetFormatMsg("%s checksum mismatch, read %X, calc %X", name, read, calc)
}
func (e *ErrBuilder) ErrStoreKvSizeMismatch(calc, got uint32) error {
	return e.SetReason(ReasonStoreKvSizeMismatch).SetFormatMsg("store kv size mismatch, calc %X, got %X", calc, got)
}
func (e *ErrBuilder) ErrBadSnapshotMagic(want, got uint32) error {
	return e.SetReason(ReasonBadSnapshotMagic).SetFormatMsg("bad snapshot magic, want %X, got %X", want, got)
}
func (e *ErrBuilder) ErrUnknownSnapshotVer(ver uint32) error {
	return e.SetReason(ReasonUnknownSnapshotVer).SetFormatMsg("unknown snapshot version %d", ver)
}
func (e *ErrBuilder) ErrSnapshotSizeMismatch(read, got uint32) error {
	return e.SetReason(ReasonSnapshotPackageSizeMismatch).SetFormatMsg("read store pack size mismatch, read %X, got %X", read, got)
}
func (e *ErrBuilder) ErrSnapshotKvNumTooMany(least, got uint32) error {
	return e.SetReason(ReasonSnapshotKvNumTooMany).SetFormatMsg("read store kv num too many, least %X, got %X", least, got)
}
func (e *ErrBuilder) ErrSnapshotKvSizeMismatch(want, got uint32) error {
	return e.SetReason(ReasonSnapshotKvSizeMismatch).SetFormatMsg("read store kv size mismatch, want %X, got %X", want, got)
}
func (e *ErrBuilder) ErrSnapshotKvNumMismatch(want, got uint32) error {
	return e.SetReason(ReasonSnapshotKvNumMismatch).SetFormatMsg("read store kv num mismatch, want %X, got %X", want, got)
}
