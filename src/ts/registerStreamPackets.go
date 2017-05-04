package ts

// Create Elementary stream packets containing our stream src
func RegisterStreamPackets(streamInfo StreamInfo, samplesInfo []SampleInfo, fragment *FragmentData) {

	// For each sample
	for _, sample := range samplesInfo {
		// Create the elementary stream
		elementaryStream := CreateElementaryStreamSrc(streamInfo, sample)

		// Create packets stream
		pes := createPackets(streamInfo, sample, uint32(len(elementaryStream)))

		// Fill packets payload
		fillPackets(&pes, elementaryStream)

		// Append fragment to PES
		fragment.pes = append(fragment.pes, pes...)
	}
}


func CreateElementaryStreamSrc(stream StreamInfo, sample SampleInfo) ([]byte) {

	sameTimeStamps := sample.DTS == sample.CTS

	// Create data holding the elementary stream
	streamSize, headerLength := getStreamSizeAndHeaderLength(stream, sample, sameTimeStamps)
	data := NewData(streamSize)

	pushSampleHeader(stream, sample, sameTimeStamps, streamSize, headerLength, data)
	// For each NAL Units
	for _, unit := range sample.NALUnits {

		// Packet start id code
		data.PushUInt(1, 24)

		// Add the corresponding data
		stream.mdat.Offset = unit.mdatOffset
		stream.mdat.Size = unit.mdatSize
		data.PushAll(stream.mdat.ToBytes())
	}

	return data.Data
}

func getStreamSizeAndHeaderLength(stream StreamInfo, sample SampleInfo, sameTimeStamps bool) (streamSize int, headerLength int) {
	headerLength = 5
	if !sameTimeStamps {
		headerLength += 5
	}

	// Stream size with header
	streamSize = 9 + headerLength + int(sample.size)

	// Replace start code prefix with NALLength size
	totalNALLengthSize := uint32(len(sample.NALUnits)) * stream.nalLengthSize
	startCodePrefixSize := uint32(len(sample.NALUnits)) * 3
	streamSize += int(startCodePrefixSize) - int(totalNALLengthSize)

	return
}

func pushSampleHeader(stream StreamInfo, sample SampleInfo, sameTimeStamps bool, streamSize int, headerLength int, data *Data) {

	// If CTS needed
	flagPTSCode := uint32(0x02)
	if !sameTimeStamps {
		flagPTSCode = 0x03
	}

	data.PushUInt(1, 24) 			// Packet start id code
	data.PushUInt(224, 8) 			// Pes stream
	if stream.isVideo() {
		data.PushUInt(0, 16) 		// Stream size
	} else {
		data.PushUInt(uint32(streamSize - 5), 16)  // Stream size
	}

	data.PushUInt(0x2, 2) 			// '10'
	data.PushUInt(0, 2) 				// PES_Scrambling_control
	data.PushUInt(0, 1) 				// PES_Priority
	data.PushUInt(1, 1)				// data alignment indicator
	data.PushUInt(0, 1)				// copyright
	data.PushUInt(0, 1)				// original or copy

	data.PushUInt(flagPTSCode, 2) 			// PTS and DTS flag

	data.PushUInt(0, 1)				// ESCR flag
	data.PushUInt(0, 1)				// ESCR Rate flag
	data.PushUInt(0, 1)				// DSM_trick_mode_flag
	data.PushUInt(0, 1)				// additional_copy_info_flag
	data.PushUInt(0, 1)				// PES_CRC_flag
	data.PushUInt(0, 1)				// PES_extension_flag
	data.PushUInt(uint32(headerLength), 8)	// Header length

	data.PushUInt(flagPTSCode,4) 			// PTS and DTS flag

	pushTimestamp(sample.CTS, data)

	if !sameTimeStamps {
		data.PushUInt(1, 4)
		pushTimestamp(sample.DTS, data)
	}
}

func CreateSampleHeader(stream StreamInfo, sample SampleInfo) (data *Data) {

	sameTimeStamps := sample.DTS == sample.CTS

	// If CTS needed
	headerLength := uint32(5)
	flagPTSCode := uint32(0x02)
	if !sameTimeStamps {
		headerLength += 5
		flagPTSCode = 0x03
	}

	streamSize := 9 + headerLength + sample.size

	data = NewData(int(streamSize))
	data.PushUInt(1, 24) 			// Packet start id code
	data.PushUInt(224, 8) 			// Pes stream
	if stream.isVideo() {
		data.PushUInt(0, 16) 		// Stream size
	} else {
		data.PushUInt(streamSize - 5, 16)  // Stream size
	}


	data.PushUInt(0x2, 2) 			// '10'
	data.PushUInt(0, 2) 				// PES_Scrambling_control
	data.PushUInt(0, 1) 				// PES_Priority
	data.PushUInt(1, 1)				// data alignment indicator
	data.PushUInt(0, 1)				// copyright
	data.PushUInt(0, 1)				// original or copy

	data.PushUInt(flagPTSCode, 2) 			// PTS and DTS flag

	data.PushUInt(0, 1)				// ESCR flag
	data.PushUInt(0, 1)				// ESCR Rate flag
	data.PushUInt(0, 1)				// DSM_trick_mode_flag
	data.PushUInt(0, 1)				// additional_copy_info_flag
	data.PushUInt(0, 1)				// PES_CRC_flag
	data.PushUInt(0, 1)				// PES_extension_flag
	data.PushUInt(headerLength, 8)			// Header length

	data.PushUInt(flagPTSCode,4) 			// PTS and DTS flag

	pushTimestamp(sample.CTS, data)

	if !sameTimeStamps {
		data.PushUInt(1, 4)
		pushTimestamp(sample.DTS, data)
	}

	return data
}

func pushTimestamp(timestamp uint64, data *Data) {
	data.PushUInt64(timestamp >> 30, 3) 	// timestamp [32..30]
	data.PushUInt(1, 1) 			// marker_bit
	data.PushUInt64(timestamp >> 15, 15) 	// timestamp [29..15]
	data.PushUInt(1, 1) 			// marker_bit
	data.PushUInt64(timestamp, 15) 		// timestamp [14..0]
	data.PushUInt(1, 1) 			// marker_bit
}

func createPackets(info StreamInfo, sample SampleInfo, elementaryStreamSize uint32) (packets []PES){
	// Create the first fragment with adaptation field
	pid := info.PID
	streamId := info.streamType

	// Create the first packet
	var firstPacket PES = *NewStartStream(pid, streamId)

	pcr := PCR{}
	pcr.BaseMediaDecodeTime = sample.PCR
	firstPacket.setPCR(pcr)

	// IF isIFrame set RAP
	if sample.IsIframe() {
		firstPacket.RandomAccessIndicator = 1
	}
	firstPacket.setAdaptationControl(true, true)

	// Compute the number of fragments needed
	restingSize := uint32(188 - firstPacket.HeaderAndAdaptationSize())
	firstPacket.Payload.EmptySize = restingSize

	neededPackets := uint32(0)

	var sizeToStuff uint32
	var lastPacketPESSize uint32
	field := AdaptationField{}

	stuffingCase := 0

	// If there are only one packet needed
	if restingSize >= elementaryStreamSize {
		stuffingCase = 1
		lastPacketPESSize = elementaryStreamSize
		sizeToStuff = 188 - 4 - elementaryStreamSize // + First byte in header
	} else {
		// Compute number of needed packets to be filled with elementary stream
		neededPackets = RoundDivision32(elementaryStreamSize - restingSize, 184)

		// Get stream size in the last packet
		lastPacketPESSize = (elementaryStreamSize - restingSize) % 184

		sizeToStuff = 184 - lastPacketPESSize
		// If there are still something to write
		if sizeToStuff != 0 {
			stuffingCase = 1
			// If there is enough space to write payload and adaptation field in this last packet
			if sizeToStuff < uint32(field.Size()) {
				// We need another packet
				neededPackets++
				stuffingCase = 2
			}
		}
	}

	// Create packets
	packets = make([]PES, neededPackets + 1)
	packets[0] = firstPacket
	for i := uint32(1); i < neededPackets + 1; i++ {
		packets[i] = *NewStream(pid)
		packets[i].setAdaptationControl(false, true)
		packets[i].Payload.EmptySize = 184
	}

	// If we have not enough space in the last packet with the adpation field
	switch stuffingCase {
	case 1:
		// Fill last packet adaptation field
		packets[len(packets) - 1].setAdaptationControl(true, true)
		packets[len(packets) - 1].setTotalAdaptationSize(byte(sizeToStuff))
		packets[len(packets) - 1].Payload.EmptySize = lastPacketPESSize
		break;
	case 2:
		// Add adaptation field for last two packets
		packets[len(packets) - 2].setAdaptationControl(true, true)
		packets[len(packets) - 2].EmptySize = packets[len(packets) - 2].EmptySize - uint32(field.Size())
		restingPES := lastPacketPESSize - packets[len(packets) - 2].EmptySize

		packets[len(packets) - 1].setAdaptationControl(true, true)
		packets[len(packets) - 1].setTotalAdaptationSize(byte(packets[len(packets) - 1].EmptySize - restingPES))
		packets[len(packets) - 1].Payload.EmptySize = restingPES
		break;
	}


	return
}

func fillPackets(packets *[]PES, elementaryStream []byte) {

	offset := uint32(0)
	finalSize := uint32(len(elementaryStream))
	packetId := uint32(0)

	var extractedSize uint32
	var payloadSize uint32

	// While there is data left
	for offset != finalSize {
		// Get the corresponding part of the packet
		payloadSize = uint32((*packets)[packetId].Payload.EmptySize)
		extractedSize = Min32(payloadSize, finalSize - offset)

		// Register the payload
		(*packets)[packetId].Data = elementaryStream[offset:offset+extractedSize]

		// Go to next packet
		offset += extractedSize
		packetId++
	}

}