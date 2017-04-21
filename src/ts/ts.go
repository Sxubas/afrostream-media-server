package ts

import (
	"mp4"
	"fmt"
)

func Test() {
	var header Header
	header.TransportErrorIndicator = 1
	header.PayloadUnitStartIndicator = 0
	header.PID = 513
	header.AdaptationFieldControl = 3
	header.ContinuityCounter = 3

	data := header.ToBytes()
	data.PrintBinary()
	data.PrintHex()
}


func CreateHLSFragment(mp4m map[string][]interface{}, fragmentNumber uint32, fragmentDuration uint32) (bytes []Bytes) {
	var isVideo bool
	var pmt PMT
	//var tfhdSize uint32

	// Check if the mp4m is audio or video
	if mp4m["moov.trak.mdia.minf.smhd"] != nil {
		isVideo = false
	}
	if mp4m["moov.trak.mdia.minf.vmhd"] != nil {
		isVideo = true
	}

	//if isVideo {
	//	pmt = *NewPMT(256)
	//} else {
	//	pmt = *NewPMT(255)
	//}

	pmt = *NewDebugPMT()

	// Retrieve start and end to write
	sampleStart, sampleEnd, baseMediaDecodeTime := GetSampleInfo(mp4m, fragmentNumber, fragmentDuration, isVideo)
	_ = baseMediaDecodeTime
	// Compute offset
	stsz := mp4m["moov.trak.mdia.minf.stbl.stsz"][0].(mp4.StszBox)
	mdat := mp4m["mdat"][0].(mp4.MdatBox)
	var i uint32
	for i = 0; i < sampleStart; i++ {
		mdat.Offset += int64(stsz.EntrySize[i])
	}

	// Get the PCR for this element
	// Create first element with no payload
	startStream := NewStartStream(baseMediaDecodeTime)

	numberOfStreamPackets := int(sampleEnd - sampleStart)
	numberOfPackets := numberOfStreamPackets + 3
	bytes = make([]Bytes, numberOfPackets)

	bytes[0] = NewPAT()
	bytes[1] = pmt
	bytes[2] = startStream

	remainingBytes := mdat.Size

	isStartingPayload := true
	// For i and next ... Write bytes
	for i := 0; i < numberOfStreamPackets; i++ {
		pes := NewPes()
		pes.PID = 48
		if isStartingPayload {
			pes.PayloadUnitStartIndicator = 1
			isStartingPayload = false
		}
		pes.AdaptationFieldControl = 0x01 // No Adaptation field, payload only
		pes.ContinuityCounter = byte(i + 1 % 16)


		// Get max to write
		mdat.Size = Min32(184, remainingBytes)

		if remainingBytes < 0 {
			panic("error no remaining Bytes to write")
		}

		// Write 184 bytes
		pes.Payload.Data = mdat.ToBytes()

		mdat.Offset += 184
		remainingBytes -= 184

		bytes[i+3] = pes
	}

	return
}

func CreateHLSFragmentDebug(mp4m map[string][]interface{}, fragmentNumber uint32, fragmentDuration uint32) (bytes []Bytes) {

	// Debug Init
	baseMediaDecodeTime := uint64(63000)
	numberOfStreamPackets := 298
	mdat := mp4m["mdat"][0].(mp4.MdatBox)
	mdat.Offset = 172
	numberOfPackets := numberOfStreamPackets + 2
	bytes = make([]Bytes, numberOfPackets)

	// Start debugging
	bytes[0] = NewDebugPAT()
	bytes[1] = NewDebugPMT()

	streamBytes := createDebugStream(numberOfStreamPackets, baseMediaDecodeTime)
	fillDebugStream(mdat, streamBytes)

	// Copying packets
	for i,packet := range streamBytes {
		bytes[i + 2] = packet
	}

	return
}

func createDebugStream(numberOfStreamPackets int, baseMediaDecodeTime uint64) (bytes []Bytes) {
	bytes = make([]Bytes, numberOfStreamPackets)

	bytes[0] = NewStartStreamDebug(baseMediaDecodeTime)

	for i := 1; i < int(numberOfStreamPackets); i++ {
		pes := NewPes()
		pes.PID = 256
		pes.AdaptationFieldControl = 0x01 // No Adaptation field, payload only
		pes.ContinuityCounter = byte(i % 16)
		bytes[i] = pes
	}

	return
}
func fillDebugStream(mdat mp4.MdatBox, bytes []Bytes) {

	remainingBytes := mdat.Size

	// First stream packet
	startBytes := []byte{0x00, 0x00, 0x01, 0xe0, 0x00, 0x00, 0x80, 0xc0, 0x0a, 0x31, 0x00, 0x09, 0x07, 0x4d, 0x11, 0x00, 0x07, 0xd8, 0x61, 0x00, 0x00, 0x00, 0x01, 0x09, 0xf0, 0x00, 0x00, 0x00, 0x01}
	lenToWrite := uint32(176 - len(startBytes))
	mdat.Size = Min32(lenToWrite, remainingBytes)
	startStream := bytes[0].(*PES)

	data := NewData(176)
	data.PushAll(startBytes)
	toBytes := mdat.ToBytes()
	data.PushAll(toBytes)
	data.PrintHexFull()
	startStream.Payload.Data = data.Data
	remainingBytes -= lenToWrite

	startStream.ToBytes()

	for i := 1; i < len(bytes); i++ {
		// Get max to write
		mdat.Size = Min32(184, remainingBytes)
		if remainingBytes < 0 {
			panic("error no remaining Bytes to write")
		}

		// Write 184 bytes
		pes := bytes[i].(*PES)
		pes.Payload.Data = mdat.ToBytes()

		mdat.Offset += 184
		remainingBytes -= 184
	}
}

func GetSampleInfo(mp4m map[string][]interface{},
	fragmentNumber uint32,
	fragmentDuration uint32,
	isVideo bool) (sampleStart, sampleEnd uint32, baseMediaDecodeTime uint64) {

	mdhd := mp4m["moov.trak.mdia.mdhd"][0].(mp4.MdhdBox)
	stts := mp4m["moov.trak.mdia.minf.stbl.stts"][0].(mp4.SttsBox)
	defaultSampleDuration := stts.Entries[0].SampleDelta

	// Compute the start, end of the stream
	sampleStart = (((uint32(fragmentNumber) - 1) * uint32(fragmentDuration)) * uint32(mdhd.Timescale)) / uint32(defaultSampleDuration)
	sampleEnd = (((uint32(fragmentNumber) * uint32(fragmentDuration)) * uint32(mdhd.Timescale)) / uint32(defaultSampleDuration)) - 1


	if isVideo {
		var stss mp4.StssBox
		stss = mp4m["moov.trak.mdia.minf.stbl.stss"][0].(mp4.StssBox)
		var i uint32
		sampleStartSet := false
		for i = 0; (i < stss.EntryCount) && ((stss.SampleNumber[i] - 1) < sampleEnd); i++ {
			if (stss.SampleNumber[i] - 1) >= sampleStart {
				if sampleStartSet == false {
					sampleStart = stss.SampleNumber[i] - 1
					sampleStartSet = true
				}
				//iFramesToSet = append(iFramesToSet, stss.SampleNumber[i]-1-sampleStart)
			}
		}
		if i < stss.EntryCount {
			sampleEnd = stss.SampleNumber[i] - 1
		} else {
			//lastSegment = true
		}
	}
	sampleEnd--

	baseMediaDecodeTime = uint64(sampleStart) * uint64(defaultSampleDuration)

	return
}

func CreateHLSFragmentWithConf(mp4 map[string][]interface{}, fragmentNumber uint32, fragmentDuration uint32) (ts []Packet) {
	return
}

func PrintFragments(fragments *[]Bytes, n int) {
	for i := 0; i < n && i < len(*fragments); i++ {
		fmt.Printf("% 8X\n", (*fragments)[i].ToBytes())
	}
}
