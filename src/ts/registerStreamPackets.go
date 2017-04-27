package ts

// Create Elementary stream packets containing our stream src
func RegisterStreamPackets(stream StreamInfo, samplesInfo []SampleInfo, fragment FragmentData) {

	// For each sample
	for _, sample := range samplesInfo {
		// Create the elementary stream
		stream := CreateElementaryStream(sample)

		// Create packets stream
		packets := createPackets(sample, stream)

		// Fill packets payload
		fillPackets(packets, stream)
	}
}

func CreateElementaryStream(stream StreamInfo, sample SampleInfo) (bytes []byte) {

	size := 24 + 8 + 16 // startCode + stream id + ps packet length

	// Optional PES Header
	size += 16 + 8 // Optional + PES Header length + PTS + DTS

	flag7 := uint32(0)

	// If CTS needed
	if sample.hasCTS {
		size += 33
		flag7 |= 1 << 7
	}

	// If DTS needed
	if sample.hasDTS {
		size += 33
		flag7 |= 1 << 6
	}

	headerLength := uint32(size + 4)/8
	packetLength := headerLength + sample.size

	data := NewData(packetLength) // Superior byte
	data.PushUInt(1, 24)
	data.PushUInt(127, 8)
	data.PushUInt(8, (size + 4)/8)
	data.PushUInt(1, 1)
	data.PushUInt(0, 7)
	data.PushUInt(flag7, 8)
	data.PushUInt(headerLength, 8)

	// If CTS needed
	if sample.hasCTS {
		/*
		bs_write( s, 3, (timestamp >> 30) & 0x07 ); // timestamp [32..30]
		bs_write1( s, 1 );                          // marker_bit
		bs_write( s, 8, (timestamp >> 22) & 0xff ); // timestamp [29..15]
		bs_write( s, 7, (timestamp >> 15) & 0x7f ); // timestamp [29..15]
		bs_write1( s, 1 );                          // marker_bit
		bs_write( s, 8, (timestamp >> 7) & 0xff );  // timestamp [14..0]
		bs_write( s, 7, timestamp & 0x7f );         // timestamp [14..0]
		bs_write1( s, 1 );*/
	}

	// If DTS needed
	if sample.hasDTS {

	}
	return
}

func createPackets(info SampleInfo, stream []byte) (packets []Packet){
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

func fillPackets(packets []Packet, bytes []byte) {
	// While there is data left
	for true {
		// Get the corresponding part of the packet
		// Register the payload
	}
}