package ts

// Structures
type PAT struct {
	Packet
	PointField byte // 1Byte
	Section    ProgramAssociationSection
}

type ProgramAssociationSection struct {
	Bytes
	TableID                byte // 1Byte
	SectionSyntaxIndicator byte
	SectionLength          uint16 // 12b;
	TransportStreamID      uint16   // 2Bytes
	VersionNumber          byte   // 5b
	CurrentNextIndicator   byte
	SectionNumber          byte // 1Byte
	LastSectionNumber      byte // 1Byte
	Sections               []ProgramAssociationSubSection
}

type ProgramAssociationSubSection struct {
	ProgramNumber uint16   // 2Bytes
	ProgramMapID  uint16 // 13b
}

// To Bytes
func (pat PAT) ToBytes() (data Data) {
	data = pat.Packet.ToBytes()

	// Set PointField
	data.PushObj(pat.PointField, 8)

	// Program association section
	data.PushBytes(pat.Section)

	if pat.HasPayload() {
		// Push payload
		data.PushBytes(pat.Payload)
	}

	// Fill remaining bytes with 0xff
	data.FillRemaining(0xff)

	return
}

func (section ProgramAssociationSection) ToBytes() (data Data) {
	// In general, 13 bytes after sectionLength and 3 bytes before
	sectionLength := section.GetSectionLength()
	data = *NewData(section.Size())

	data.PushObj(section.TableID, 8)
	data.PushObj(section.SectionSyntaxIndicator, 1)
	data.PushUInt(0, 1)    // Private
	data.PushUInt(0x03, 2) // Reserved
	data.PushObj(sectionLength, 12)
	data.PushObj(section.TransportStreamID, 16)
	data.PushUInt(0x03, 2) // Reserved
	data.PushObj(section.VersionNumber, 5)
	data.PushObj(section.CurrentNextIndicator, 1)
	data.PushObj(section.SectionNumber, 8)
	data.PushObj(section.LastSectionNumber, 8)

	for programIndex := 0; programIndex < len(section.Sections); programIndex++ {
		data.PushObj(section.Sections[programIndex].ProgramNumber, 16)
		data.PushUInt(0x07, 3) // Reserved
		data.PushObj(section.Sections[programIndex].ProgramMapID, 13) // Or Network_PID
	}

	crc32 := data.GenerateCRC32ToOffset()
	data.PushObj(crc32, 32)

	data.FillRemaining(0Xff)

	return
}

func (section ProgramAssociationSection) Size() (int) {
	return 	int(section.GetSectionLength() + 3) + 4 * len(section.Sections) + 4

}

// Constructor
func NewPAT() (pat *PAT) {
	pat = new(PAT)

	pat.PID = 0
	pat.PayloadUnitStartIndicator = 1
	pat.AdaptationFieldControl = 1


	pat.Section.SectionSyntaxIndicator = 1
	pat.Section.SectionLength = 13
	pat.Section.TransportStreamID = 1
	pat.Section.CurrentNextIndicator = 1

	pat.Section.Sections = make([]ProgramAssociationSubSection, 1)

	// Set PMT PID
	pat.Section.Sections[0].ProgramNumber = 1
	pat.Section.Sections[0].ProgramMapID = 4096

	return
}

func NewDebugPAT() (pat *PAT) {
	pat = new(PAT)

	pat.PID = 0
	pat.PayloadUnitStartIndicator = 1
	pat.AdaptationFieldControl = 1

	pat.Section.SectionSyntaxIndicator = 1
	pat.Section.SectionLength = 13
	pat.Section.TransportStreamID = 1
	pat.Section.CurrentNextIndicator = 1

	pat.Section.Sections = make([]ProgramAssociationSubSection, 1)

	// Set PMT PID
	pat.Section.Sections[0].ProgramNumber = 1
	pat.Section.Sections[0].ProgramMapID = 4096

	return
}

func (section *ProgramAssociationSection) GetSectionLength() (uint16) {
	// Compute the section Length
	return 9 + uint16(len(section.Sections) * 4)
}