package ts

type SampleInfo struct {
	PCR uint64
	CTS uint32
	DTS uint32

	registerPCR bool
	registerCTS bool
	registerDTS bool

	isIFrameType bool

	mdatOffset uint32
	mdatSize   uint32

	size uint32
}

func (info SampleInfo) IsRegisteringPCR() (bool) {
	return info.registerPCR
}
func (info SampleInfo) IsRegisteringDTS() (bool) {
	return info.registerDTS
}
func (info SampleInfo) IsRegisteringCTS() (bool) {
	return info.registerCTS
}
func (info SampleInfo) IsIframe() (bool) {
	return info.isIFrameType
}
