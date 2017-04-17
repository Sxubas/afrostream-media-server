package ts

type PMT struct {
	Packet
	PointField byte // 1Byte
	Section ProgramMapSection
}

type ProgramMapSection struct {
	TableId byte; // 1Byte
	SectionSyntaxIndicator byte;
	SectionLength uint16; // 12b
	ProgramNumber uint16; // 2Bytes
	VersionNumber byte; // 5b
	CurrentNextIndicator byte;
	SectionNumber byte; // 1Byte
	LastSectionNumber byte; // 1Byte
	PCR_PID uint16; // 13b
	ProgramInfoLength uint16; // 12b
	Sections []ProgramMapSubSection;
	crc32 uint32; // 4Bytes
}

type ProgramMapSubSection struct {
	StreamType byte; // 1Byte
	ElementaryPID uint16; // 13b
	ESInfoLength [12]byte; // 12b
}

// To bytes
func (pmt PMT) Bytes() (data Data) {
	data = *NewData(188)



	return
}

func (section ProgramMapSection) Bytes() (data []byte) {
	return
}

func (section ProgramMapSubSection) Bytes() (data []byte) {
	return
}


