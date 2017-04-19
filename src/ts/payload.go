package ts

type Payload struct {
	Bytes
	PointerField byte
	Data         []byte
}

func (payload Payload) ToBytes() (data Data) {
	data = *NewData(len(payload.Data))
	data.Data = payload.Data
	data.Offset = len(payload.Data) * 8
	return
}
