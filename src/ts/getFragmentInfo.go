package ts

// Get information on the fragment. Start and end of the samples list
func GetFragmentInfo(streamInfo *StreamInfo, fragmentNumber uint32, fragmentDuration uint32) (fragmentInfo *FragmentInfo) {

	fragmentInfo = new(FragmentInfo)

	// Set number and duration
	fragmentInfo.number = fragmentNumber
	fragmentInfo.duration = fragmentDuration

	// Get start and end of all samples
	// Save I-Frames indices to use it as RAP (Random access point)
	registerStartAndEndOfFragment(*streamInfo, fragmentInfo)

	// Retrieve MDAT offset and size
	registerMdatOffset(streamInfo, fragmentInfo)

	// Retrieve CTTS start Offset for CTS
	registerCTSStart(*streamInfo, fragmentInfo)

	// Retrieve STTS start offset for DTS
	registerDTSStart(*streamInfo, fragmentInfo)

	return
}

func registerStartAndEndOfFragment(stream StreamInfo, frag *FragmentInfo) {
	// Init data
	sampleStart := uint32((((float64(frag.number) - 1) * float64(frag.duration)) * float64(stream.Timescale)) / float64(stream.SampleDelta))
	sampleEnd := uint32(((float64(frag.number) * float64(frag.duration)) * float64(stream.Timescale)) / float64(stream.SampleDelta))


	//fmt.Printf("%d %d", sampleStart, sampleEnd)
	stss := stream.stss
	stsz := stream.stsz
	iFramesIndices := make([]uint32, 0)

	// Retrieve
	if stream.isVideo() {
		// STSS sample number of I-Frames
		var i uint32
		sampleStartSet := false
		for i = 0; (i < stss.EntryCount) && ((stss.SampleNumber[i] - 1) < sampleEnd); i++ {
			if (stss.SampleNumber[i] - 1) >= sampleStart {
				if sampleStartSet == false {
					sampleStart = stss.SampleNumber[i] - 1
					sampleStartSet = true
				}
				iFramesIndices = append(iFramesIndices, stss.SampleNumber[i]-1-sampleStart)
			}
		}

		// If we have not reach the last entry count
		if i < stss.EntryCount {
			// Retrieve the sampleEnd
			sampleEnd = stss.SampleNumber[i] - 1
		}
	}
	sampleEnd--

	if sampleEnd > (stsz.SampleCount - 1) {
		sampleEnd = stsz.SampleCount - 1
	}

	frag.sampleStart = sampleStart
	frag.sampleEnd = sampleEnd
	frag.iFramesIndices = iFramesIndices

}

func registerMdatOffset(stream *StreamInfo, frag *FragmentInfo) {
	// Get to the first element using stsz table referecing size
	var i uint32
	for i = 0; i < frag.sampleStart; i++ {
		if stream.stsz.SampleSize == 0 {
			stream.mdat.Offset += int64(stream.stsz.EntrySize[i])
		} else {
			stream.mdat.Offset += int64(stream.stsz.SampleSize)
		}
	}
}

func registerCTSStart(stream StreamInfo, frag *FragmentInfo) {

	var i uint32

	for i = 0; i < frag.sampleStart; i++ {
		// Getting the number of ctts offset
		if stream.compositionTimeOffset {
			if frag.cttsSampleCount > 0 {
				frag.cttsSampleCount--
				if frag.cttsSampleCount == 0 {
					frag.cttsOffset++
				}
			} else {
				frag.cttsSampleCount = stream.ctts.Entries[frag.cttsOffset].SampleCount - 1
				if frag.cttsSampleCount == 0 {
					frag.cttsOffset++
				}
			}
		}
	}
}

func registerDTSStart(stream StreamInfo, frag *FragmentInfo) {

	var i uint32
	frag.dts = 0
	for i = 0; i < frag.sampleStart; i++ {
		// Update Decoding time offset
		if frag.sttsSampleCount > 0 {
			frag.sttsSampleCount--
			if frag.sttsSampleCount == 0 {
				frag.sttsOffset++
			}
		} else {
			frag.sttsSampleCount = stream.stts.Entries[frag.sttsOffset].SampleCount - 1
			if frag.sttsSampleCount == 0 {
				frag.sttsOffset++
			}
		}
		frag.dts += uint64(stream.stts.Entries[frag.sttsOffset].SampleDelta)
	}
}
