package ts

// Create Elementary stream packets containing our stream src
func RegisterStreamPackets(streamInfo StreamInfo, samplesInfo []SampleInfo, fragment FragmentData) {

	// For each sample
	for _, sample := range samplesInfo {
		// Create the elementary stream
		elementaryStream := CreateElementaryStream(streamInfo, sample)

		// Create packets stream
		packets := createPackets(streamInfo, samplesInfo)

		// Fill packets payload
		fillPackets(packets, elementaryStream)
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

	headerLength := uint32(size + 4)/8  // Superior byte
	packetLength := headerLength + sample.size

	data := NewData(int(packetLength)*8)
	data.PushUInt(1, 24)
	data.PushUInt(stream.streamId, 8)
	data.PushUInt(8, (size + 4)/8)
	data.PushUInt(1, 1)
	data.PushUInt(0, 7)
	data.PushUInt(flag7, 8)
	data.PushUInt(headerLength, 8)

	// If CTS needed
	if sample.hasCTS {
		pushTimestamp(sample.CTS, data)
	}

	// If DTS needed
	if sample.hasDTS {
		pushTimestamp(sample.DTS, data)
	}

	// Fill to the end of header
	data.FillTo(0xff, int(headerLength * 8))

	// Fill with the sample data
	savedOffset := stream.mdat.Offset
	savedSize	:= stream.mdat.Size
	data.PushAll(stream.mdat.ToBytes())
	stream.mdat.Offset = savedOffset
	stream.mdat.Size = savedSize
	return
}

func pushTimestamp(timestamp uint64, data *Data) {
	data.PushUInt64(timestamp >> 30, 3) 	// timestamp [32..30]
	data.PushUInt(1, 1) 			// marker_bit
	data.PushUInt64(timestamp >> 15, 15) 	// timestamp [29..15]
	data.PushUInt(1, 1) 			// marker_bit
	data.PushUInt64(timestamp, 15) 		// timestamp [14..0]
	data.PushUInt(1, 1) 			// marker_bit
}

func createPackets(info StreamInfo, sampleinfo []SampleInfo) (packets []PES){
	// Create the first fragment with adaptation field
	pid := info.PID
	streamId := info.streamId

	for _, sample := range sampleinfo {

		// Create the first packet
		var firstPacket PES = *NewStartStream(pid, streamId)
		if sample.HasAdaptationField() {
			firstPacket.AdaptationFieldControl = 0x03
		} else {
			firstPacket = *NewStream(pid)
		}

		if sample.hasPCR {
			pcr := PCR{}
			pcr.BaseMediaDecodeTime = sample.PCR
			firstPacket.setPCR(pcr)
		}

		// IF isIFrame set RAP
		firstPacket.RandomAccessIndicator = byte(sample.IsIframe())
		firstPacket.setAdaptationControl(true, true)

		// Compute the number of fragments needed
		restingSize := uint32(188 - firstPacket.AdaptationField.Size())
		firstPacket.Payload.EmptySize = restingSize
		neededPackets := (sample.size - restingSize)/184

		// Create packets
		packets = make([]PES, neededPackets + 1)
		packets[0] = firstPacket

		for i := uint32(1); i < neededPackets; i++ {
			packets[i] = *NewStream(pid)
			packets[i].setAdaptationControl(false, true)
			packets[i].Payload.EmptySize = 184
		}
	}



	return
}

func fillPackets(packets []PES, elementaryStream []byte) {

	offset := uint32(0)
	finalSize := uint32(len(elementaryStream))
	packetId := uint32(0)

	var extractedSize uint32
	var payloadSize uint32

	// While there is data left
	for offset != finalSize {
		// Get the corresponding part of the packet
		payloadSize = uint32(packets[packetId].Payload.Size())
		extractedSize = Min32(payloadSize, finalSize - offset)

		// Register the payload
		packets[packetId].Data = elementaryStream[offset:offset+extractedSize]

		// Go to next packet
		offset += extractedSize
		packetId++
	}

	// Fill the end of the last packet
	packetId--
	lastRestingSize := payloadSize - extractedSize
	if lastRestingSize > 0  {
		fillingBytes := NewData(int(lastRestingSize*8))
		fillingBytes.FillRemaining(0xff)
		packets[packetId].Data = append(packets[packetId].Data, fillingBytes.Data...)
	}

}