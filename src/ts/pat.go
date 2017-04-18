package ts

// Structures
type PAT struct {
	Packet
	PointField byte // 1Byte
	Section    ProgramAssociationSection

}

type ProgramAssociationSection struct {
	Bytes
	TableID byte; // 1Byte
	SectionSyntaxIndicator byte;
	SectionLength uint16; // 12b;
	TransportStreamID byte; // 2b
	VersionNumber byte; // 5b
	CurrentNextIndicator byte;
	SectionNumber byte; // 1Byte
	LastSectionNumber byte; // 1Byte
	Sections []ProgramAssociationSubSection;
}

type ProgramAssociationSubSection struct {
	ProgramNumber byte; // 2b
	ProgramMapID uint16; // 13b
}

// To Bytes
func (pat PAT) Bytes() (data Data) {
	data = pat.Packet.ToBytes()

	// Set PointField
	data.PushObj(pat.PointField, 8)

	// Program association section
	data.PushBytes(pat.Section)

	if (pat.HasPayload()) {
		// Push payload
		data.PushBytes(pat.Payload)
	}

	// Fill remaining bytes with 0xff
	data.FillRemaining(0xff)

	return
}

func (section ProgramAssociationSection) ToBytes() (data Data) {
	data = *NewData(4)

	data.PushObj(section.TableID, 8)
	data.PushObj(section.SectionSyntaxIndicator, 1)
	data.PushObj(0, 1) // Private
	data.PushObj(0x03, 2) // Reserved
	data.PushObj(section.SectionLength, 12)
	data.PushObj(section.TransportStreamID, 16)
	data.PushObj(0x03, 2) // Reserved
	data.PushObj(section.VersionNumber, 5)
	data.PushObj(section.CurrentNextIndicator, 1)
	data.PushObj(section.SectionNumber, 8)
	data.PushObj(section.LastSectionNumber, 8)


	for programIndex := 0; programIndex < len(section.Sections); programIndex++ {
		data.PushObj(section.Sections[programIndex].ProgramNumber, 16)
		data.PushObj(0x07, 3) // Reserved
		data.PushObj(section.Sections[programIndex].ProgramMapID, 13) // Or Network_PID
	}

	data.PushObj(data.GenerateCRC32(), 32)

	return
}

// Constructor
func NewPAT() (pat *PAT) {
	pat = new(PAT)

	pat.PID = 0
	pat.PayloadUnitStartIndicator = 1
	pat.AdaptationFieldControl = 1

	pat.Section.SectionSyntaxIndicator = 1
	pat.Section.SectionLength = 13
	pat.Section.CurrentNextIndicator = 1

	pat.Section.Sections = make([]ProgramAssociationSubSection, 1)

	// Set PMT PID
	pat.Section.Sections[0].ProgramNumber = 1
	pat.Section.Sections[0].ProgramMapID = 4096

	return
}