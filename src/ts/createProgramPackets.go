package ts

// Create main packets program: PAT and PMT
func CreateProgramPackets(info StreamInfo, fragment *FragmentData) {
	// Create PAT
	fragment.pat = *NewPAT()

	// Create program stream
	fragment.pmt = *NewPMT(info.PID)
	fragment.pmt.Section.Sections = make([]ProgramMapSubSection, 1)

	// Register stream
	fragment.pmt.Section.Sections[0].StreamType = byte(info.streamType)
	fragment.pmt.Section.Sections[0].ElementaryPID = info.PID
	fragment.pmt.Section.Sections[0].ESInfoLength = 0
}
