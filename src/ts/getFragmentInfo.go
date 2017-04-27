package ts

// Get information on the fragment. Start and end of the samples list
func GetFragmentInfo(streamInfo StreamInfo, fragmentNumber uint32, fragmentDuration uint32) (fragmentInfo FragmentInfo) {

	// Set number and duration
	fragmentInfo.number = fragmentNumber
	fragmentInfo.duration = fragmentDuration

	// Get start and end of all samples
	// Save I-Frames indices to use it as RAP (Random access point)
	registerStartAndEndOfFragment(streamInfo, fragmentInfo)

	// Retrieve MDAT offset and size
	registerMdatOffset(streamInfo, fragmentInfo)

	// Retrieve CTTS start Offset
	registerCTTSStartOffset(streamInfo, fragmentInfo)

	return
}

func registerStartAndEndOfFragment(stream StreamInfo, frag FragmentInfo) {
	// Init data
	sampleStart := uint32((((float64(frag.number) - 1) * float64(frag.duration)) * float64(stream.Timescale)) / float64(stream.SampleDelta))
	sampleEnd := uint32(((float64(frag.number) * float64(frag.duration)) * float64(stream.Timescale)) / float64(stream.SampleDelta))

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

func registerMdatOffset(stream StreamInfo, frag FragmentInfo) {
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

func registerCTTSStartOffset(stream StreamInfo, frag FragmentInfo) {

	var cttsSampleCount uint32
	var i uint32

	for i = 0; i < frag.sampleStart; i++ {
		// Getting the number of ctts offset
		if stream.compositionTimeOffset {
			if cttsSampleCount > 0 {
				cttsSampleCount--
				if cttsSampleCount == 0 {
					frag.cttsOffset++
				}
			} else {
				cttsSampleCount = stream.ctts.Entries[frag.cttsOffset].SampleCount - 1
				if cttsSampleCount == 0 {
					frag.cttsOffset++
				}
			}
		}
	}
}
