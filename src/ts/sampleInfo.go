package ts

type SampleInfo struct {
	PCR uint32
	CTS uint32
	DTS uint32

	RegisterPCR bool
	RegisterCTS bool
	RegisterDTS bool

	IsIFrameType bool

	MdatOffset uint32
	MdatSize uint32

	Size uint32
}

func (info SampleInfo) IsRegisteringPCR() (bool) {
	return info.RegisterPCR
}
func (info SampleInfo) IsRegisteringDTS() (bool) {
	return info.RegisterDTS
}
func (info SampleInfo) IsRegisteringCTS() (bool) {
	return info.RegisterCTS
}
func (info SampleInfo) IsIframe() (bool) {
	return info.IsIFrameType
}
