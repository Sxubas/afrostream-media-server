package ts

type Payload struct {
	Bytes
	PointerField byte
	Data         []byte
	EmptySize	 uint32 // Used to create payload
}

func (payload Payload) ToBytes() (data Data) {
	data = *NewData(len(payload.Data))
	data.Data = payload.Data
	data.Offset = len(payload.Data) * 8
	return
}

func (payload Payload) Size() (int) {
	if payload.EmptySize != 0 {
		return int(payload.EmptySize)
	}
	return len(payload.Data)
}
