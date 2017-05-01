package ts


// Get information on all samples in the fragment
func GetSamplesInfo(stream StreamInfo, fragmentInfo FragmentInfo) (sampleInfo []SampleInfo) {

	sampleInfo = make([]SampleInfo, fragmentInfo.getSampleCount())

	// Get size And offset of the sample in MDat
	registerSamplesSizes(stream, fragmentInfo, &sampleInfo)

	if stream.isVideo() {

		// Registers all iFrames
		// registerISamples(fragmentInfo, &sampleInfo)

		// Compute the pcr
		registerPCRSamples(stream, fragmentInfo, &sampleInfo)

		// Retrieve the compositionTimeOffset and decodingTimeOffset
		if stream.compositionTimeOffset {
			registerCTSAndDTSSamples(stream, fragmentInfo, &sampleInfo)
		}
	}

	return
}

func registerSamplesSizes(stream StreamInfo, info FragmentInfo, sampleInfo *[]SampleInfo) {

	offset := stream.mdat.Offset

	if stream.stsz.SampleSize == 0 {
		for i := 0; i < len(*sampleInfo); i++ {
			(*sampleInfo)[i].pesStream = 224

			(*sampleInfo)[i].mdatSize = stream.stsz.EntrySize[uint32(i) + info.sampleStart]
			(*sampleInfo)[i].size = (*sampleInfo)[i].mdatSize
			(*sampleInfo)[i].mdatOffset = offset
			offset += int64((*sampleInfo)[i].mdatSize)
		}
	} else {
		for i := 0; i < len(*sampleInfo); i++ {
			(*sampleInfo)[i].pesStream = 224

			(*sampleInfo)[i].mdatSize = stream.stsz.SampleSize
			(*sampleInfo)[i].size = (*sampleInfo)[i].mdatSize
			(*sampleInfo)[i].mdatOffset = offset
			offset += int64((*sampleInfo)[i].mdatSize)
		}
	}
}

func registerISamples(info FragmentInfo, sampleInfo *[]SampleInfo) {

	for i := 0; i < len(info.iFramesIndices); i++ {
		iFrameId := info.iFramesIndices[i]
		if info.isFrameInFragment(iFrameId) {
			(*sampleInfo)[iFrameId-info.sampleStart].isIFrameType = true
		}
	}
}

func registerPCRSamples(stream StreamInfo, fragmentInfo FragmentInfo, sampleInfo *[]SampleInfo) {

	pcrEmitter := IEmitter{}
	pcrEmitter.Min_emit = 0

	for i := 0; i < len(*sampleInfo); i++ {

		if pcrEmitter.Emit() {
			(*sampleInfo)[i].PCR = uint64((fragmentInfo.sampleStart + uint32(i))* stream.SampleDelta)
			(*sampleInfo)[i].hasPCR = true
			pcrEmitter.Reset()
		}
	}
}

func registerCTSAndDTSSamples(stream StreamInfo, fragmentInfo FragmentInfo, sampleInfo *[]SampleInfo) {

	emitter := IEmitter{}
	emitter.Min_emit = 1

	cttsOffset := fragmentInfo.cttsOffset
	cttsSampleCount := fragmentInfo.cttsSampleCount
	sttsOffset := fragmentInfo.sttsOffset
	sttsSampleCount := fragmentInfo.sttsSampleCount
	dts := fragmentInfo.dts

	for i := 0; i < len(*sampleInfo); i++ {

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
		dts += uint64(stream.stts.Entries[sttsOffset].SampleDelta)
		if sttsSampleCount > 0 {
			sttsSampleCount--
			if sttsSampleCount == 0 {
				sttsOffset++
			}
		} else {
			sttsSampleCount = stream.stts.Entries[sttsOffset].SampleCount - 1
			if sttsSampleCount == 0 {
				sttsOffset++
			}
		}

		(*sampleInfo)[i].hasDTS = emitter.Emit()
		(*sampleInfo)[i].hasCTS = (*sampleInfo)[i].hasDTS
		if (*sampleInfo)[i].hasDTS {
			(*sampleInfo)[i].DTS = dts
			(*sampleInfo)[i].CTS = uint64(stream.ctts.Entries[cttsOffset].SampleOffset) + dts // - dConf.MediaTime
			emitter.Reset()
		}
	}
}