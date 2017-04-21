package ts

type Bytes interface {
	ToBytes() (data Data)
	Size() (int)
}