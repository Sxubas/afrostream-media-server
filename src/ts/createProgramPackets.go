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
	/*if info.isAudio() {
		fragment.pmt.Section.Sections[0].ESInfoLength = 6
		fragment.pmt.Section.Sections[0].Descriptor = DataB(StrToBytes("0a 04 65 6e 67 00"))
	}*/
}
