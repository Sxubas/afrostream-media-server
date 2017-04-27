package ts


// Get information on all samples in the fragment
func GetSamplesInfo(stream StreamInfo, fragmentInfo FragmentInfo, fragment FragmentData) (sampleInfo []SampleInfo) {

	sampleInfo = make([]SampleInfo, fragmentInfo.getSampleCount())

	// Get size And offset of the sample in MDat
	registerSamplesSizes(stream, fragmentInfo, sampleInfo)

	if stream.isVideo() {

		// Registers all iFrames
		registerISamples(fragmentInfo, sampleInfo);

		// IF > PCR EMiT
		// Compute the pcr
		// (tfdt.baseMediaDecodeTime)

		// IF > CTS_EMIT > MAXEmit
		// Retrieve the compositionTimeOffset // StreamSize +=

		// IF DTS
		// Retrieve the DecodingTimeOffset // StreamSize +=

	}

	return
}

func registerSamplesSizes(stream StreamInfo, info FragmentInfo, sampleInfo []SampleInfo) {
	var i uint32
	offset := uint32(stream.mdat.Offset)

	if stream.stsz.SampleSize == 0 {
		for i = 0; i <= uint32(len(sampleInfo)); i++ {
			sampleInfo[i].size = stream.stsz.EntrySize[i+info.sampleStart]
			sampleInfo[i].mdatOffset = offset
			offset += sampleInfo[i].size
		}
	} else {
		for i = 0; i <= uint32(len(sampleInfo)); i++ {
			sampleInfo[i].size = stream.stsz.SampleSize
			sampleInfo[i].mdatOffset = offset
			offset += sampleInfo[i].size
		}
	}
}

func registerISamples(info FragmentInfo, sampleInfo []SampleInfo) {

	var i uint32
	for i = 0; i <= uint32(len(info.iFramesIndices)); i++ {

		if info.isFrameInFragment(info.iFramesIndices[i]) {
			sampleInfo[info.iFramesIndices[i]-info.sampleStart].isIFrameType = true
		}
	}
}

func registerPCRSamples(stream StreamInfo, sampleInfo []SampleInfo) {

}

func registerPTSSamples(stream StreamInfo, sampleInfo []SampleInfo) {

}

func registerDTSSamples(stream StreamInfo, sampleInfo []SampleInfo) {

}
