package ts

// Create main packets program: PAT and PMT
func RegisterProgramPackets(info StreamInfo, fragment FragmentData) {
	// Create PAT
	fragment.pat = *NewPAT()

	// If the stream is video
	if info.isVideo() {
		// Create program video stream
		fragment.pmt = *NewPMT(257)
	} else {
		// Create program audio stream
		fragment.pmt = *NewPMT(256)
	}
}
