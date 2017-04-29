package ts

type SampleInfo struct {
	PCR uint64
	CTS uint64
	DTS uint64

	registerPCR bool
	registerCTS bool
	registerDTS bool

	isIFrameType bool

	mdatOffset int64
	mdatSize   uint32

	size uint32

	hasPCR 		bool
	hasDTS		bool
	hasCTS		bool
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

func (info SampleInfo) HasAdaptationField() (bool) {
	return info.IsRegisteringPCR()
}