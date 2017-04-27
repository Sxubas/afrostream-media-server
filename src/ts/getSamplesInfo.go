package ts


// Get information on all samples in the fragment
func GetSamplesInfo(stream StreamInfo, fragmentInfo FragmentInfo, fragment FragmentData) (sampleInfo []SampleInfo) {

	sampleInfo = make([]SampleInfo, fragmentInfo.getSampleCount())

	// Get size And offset of the sample in MDat
	registerSamplesSizes(stream, fragmentInfo, sampleInfo)

	if stream.isVideo() {

		// Registers all iFrames
		registerISamples(fragmentInfo, sampleInfo)

		// Compute the pcr
		registerPCRSamples(stream, fragmentInfo, sampleInfo)

		// Retrieve the compositionTimeOffset and decodingTimeOffset
		registerCTSAndDTSSamples(stream, fragmentInfo, sampleInfo)
	}

	return
}

func registerSamplesSizes(stream StreamInfo, info FragmentInfo, sampleInfo []SampleInfo) {

	offset := uint32(stream.mdat.Offset)

	if stream.stsz.SampleSize == 0 {
		for i, sample := range sampleInfo {
			sample.size = stream.stsz.EntrySize[uint32(i) + info.sampleStart]
			sample.mdatOffset = offset
			offset += sample.size
		}
	} else {
		for _, sample := range sampleInfo {
			sample.size = stream.stsz.SampleSize
			sample.mdatOffset = offset
			offset += sample.size
		}
	}
}

func registerISamples(info FragmentInfo, sampleInfo []SampleInfo) {

	for _, iFrameId := range info.iFramesIndices{

		if info.isFrameInFragment(iFrameId) {
			sampleInfo[iFrameId-info.sampleStart].isIFrameType = true
		}
	}
}

func registerPCRSamples(stream StreamInfo, fragmentInfo FragmentInfo, sampleInfo []SampleInfo) {

	pcrEmitter := IEmitter{}
	pcrEmitter.Min_emit = 40

	for sampleID, sample := range sampleInfo {

		if pcrEmitter.Emit() {
			sample.PCR = uint64((fragmentInfo.sampleStart + uint32(sampleID))* stream.SampleDelta)
			pcrEmitter.Reset()
		}
	}
}

func registerCTSAndDTSSamples(stream StreamInfo, fragmentInfo FragmentInfo, sampleInfo []SampleInfo) {

	emitter := IEmitter{}
	emitter.Min_emit = 1

	cttsOffset := fragmentInfo.cttsOffset
	cttsSampleCount := fragmentInfo.cttsSampleCount
	sttsOffset := fragmentInfo.sttsOffset
	sttsSampleCount := fragmentInfo.sttsSampleCount
	dts := fragmentInfo.dts

	for _, sample := range sampleInfo {

		// Update Composition time offset
		if cttsSampleCount > 0 {
			cttsSampleCount--
			if cttsSampleCount == 0 {
				cttsOffset++
			}
		} else {
			cttsSampleCount = stream.ctts.Entries[cttsOffset].SampleCount - 1
			if cttsSampleCount == 0 {
				cttsOffset++
			}
		}

		// Update Decoding time offset
		dts += stream.stts.Entries[sttsOffset].SampleDelta
		if sttsSampleCount > 0 {
			sttsSampleCount--
			if sttsSampleCount == 0 {
				sttsOffset++
			}
		} else {
			sttsSampleCount = stream.stts.Entries[sttsOffset].SampleCount - 1
			if cttsSampleCount == 0 {
				cttsOffset++
			}
		}

		if emitter.Emit() {
			sample.DTS = dts
			sample.CTS = stream.ctts.Entries[cttsOffset].SampleOffset + dts // - dConf.MediaTime
			emitter.Reset()
		}
	}
}