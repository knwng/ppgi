package runtime

type Conn interface {
	Write(val []byte) (int, error)
	Read(buf []byte) (int, error)
	Close()
}
