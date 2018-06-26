package bucket

type Bucket interface {
	Set(k, v []byte) error
	Get(k []byte) (v []byte, exist bool, err error)
	Del(k []byte) (exist bool, err error)
	Range(a, b []byte) (kv [][2][]byte, err error)
	Split(k []byte) (a, b Bucket, e error)

	Start() error
	Stop()
	Remove()
}
