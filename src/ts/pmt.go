package ts

type PMT struct {
	Packet
	PointField byte // 1Byte
	Section ProgramMapSection
}

type ProgramMapSection struct {
	Bytes
	TableID                byte; // 1Byte
	SectionSyntaxIndicator byte;
	SectionLength          uint16; // 12b
	ProgramNumber          uint16; // 2Bytes
	VersionNumber          byte; // 5b
	CurrentNextIndicator   byte;
	SectionNumber          byte; // 1Byte
	LastSectionNumber      byte; // 1Byte
	PCR_PID                uint16; // 13b
	ProgramInfoLength      uint16; // 12b
	Sections               []ProgramMapSubSection;
	crc32                  uint32; // 4Bytes
}

type ProgramMapSubSection struct {
	StreamType byte; // 1Byte
	ElementaryPID uint16; // 13b
	ESInfoLength uint16; // 12b
}

// To bytes
func (pmt PMT) Bytes() (data Data) {
	data = pmt.Packet.ToBytes()
	// Set PointField
	data.PushObj(pmt.PointField, 8)

	// Program association section
	data.PushBytes(pmt.Section)

	return
}

func (section ProgramMapSection) ToBytes() (data Data) {
	data = *NewData(4)

	data.PushObj(section.TableID, 8)
	data.PushObj(section.SectionSyntaxIndicator, 1)
	data.PushObj(section.SectionLength, 12)
	data.PushObj(section.ProgramNumber, 16)
	data.PushObj(section.VersionNumber, 5)
	data.PushObj(section.CurrentNextIndicator, 1)
	data.PushObj(section.SectionNumber, 8)
	data.PushObj(section.LastSectionNumber, 8)

	for programIndex := 0; programIndex < len(section.Sections); programIndex++ {
		data.PushObj(section.Sections[programIndex].StreamType, 8)
		data.PushObj(section.Sections[programIndex].ElementaryPID, 13)
		data.PushObj(section.Sections[programIndex].ESInfoLength, 12)
	}

	return
}
