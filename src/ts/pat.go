package ts

// Structures
type PAT struct {
	Packet
	PointField byte // 1Byte
	Sections ProgramAssociationSection
}

type ProgramAssociationSection struct {
	TableID byte; // 1Byte
	SectionSyntaxIndicator byte;
	SectionLength uint16; // 12b;
	TransportStreamID byte; // 2b
	VersionNumber byte; // 5b
	CurrentNextIndicator byte;
	SectionNumber byte; // 1Byte
	LastSectionNumber byte; // 1Byte
	Sections []ProgramAssociationSubSection;
	crc32 byte; // 4b
}

type ProgramAssociationSubSection struct {
	ProgramNumber byte; // 2b
	ProgramMapID uint16; // 13b
}

// To Bytes
func (pat PAT) Bytes() (data Data) {
	data = pat.Packet.Bytes()

	// Set PointField
	data[offset] = pat.PointField
	offset++

	// Program association section


	return
}

func (section ProgramAssociationSection) Bytes() (data []byte) {
	data = make([]byte, 4)

	data[0] = section.TableID

	data[1]  = section.SectionSyntaxIndicator << 7
	data[1] |= byte(section.SectionLength >> 5)

	data[2]  = byte(section.SectionLength) << 5
	//data[2] |= byte(section.TransportStreamID >> 6) <<


	return
}

func (section ProgramAssociationSubSection) Bytes() (data []byte) {


	return
}