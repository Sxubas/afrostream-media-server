package ts

// Create Elementary stream packets containing our stream src
func RegisterStreamPackets(samplesInfo []SampleInfo, fragment FragmentData) {

	// For each sample
	for true {
		// Create the elementary stream

		// Create packets stream

		// Fill packets payload
	}
}

func CreateElementaryStream(info SampleInfo, data FragmentData) (bytes []byte) {
	// If DTS needed

	// If CTS needed

	return
}

func CreateFragments(info SampleInfo, fragment FragmentData) (packets []Packet){
	// Create the first fragment with adaptation field
	// Check adaptation field
	// IF pts
	// IF isIFrame set RAP

	// Compute the number of fragments needed

	// Create fragments
	for true {
		// Insert in the payload the part + stuffing bytes if needed
	}

	return
}

func FillFragments(packets []Packet, bytes []byte) {
	// While there is data left
	for true {
		// Get the corresponding part of the packet
		// Register the payload
	}
}