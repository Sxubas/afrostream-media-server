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

	return
}

func (section ProgramAssociationSection) ToBytes() (data Data) {
	data = *NewData(4)

	data.PushObj(section.TableID, 8)
	data.PushObj(section.SectionSyntaxIndicator, 1)
	data.PushObj(section.SectionLength, 12)
	data.PushObj(section.TransportStreamID, 16)
	data.PushObj(section.VersionNumber, 5)
	data.PushObj(section.CurrentNextIndicator, 1)
	data.PushObj(section.SectionNumber, 8)
	data.PushObj(section.LastSectionNumber, 8)


	for programIndex := 0; programIndex < len(section.Sections); programIndex++ {
		data.PushObj(section.Sections[programIndex].ProgramNumber, 16)
		data.PushObj(section.Sections[programIndex].ProgramMapID, 13)
	}

	return
}
