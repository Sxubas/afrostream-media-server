package ts

import (
	"encoding/binary"
)


// Get information on all samples in the fragment
func GetSamplesInfo(stream StreamInfo, fragmentInfo FragmentInfo) (sampleInfo []SampleInfo) {

	sampleInfo = make([]SampleInfo, fragmentInfo.getSampleCount())

	// Get size And offset of the sample in MDat
	registerSamplesSizes(stream, fragmentInfo, &sampleInfo)

	// Retrieve all NAL unit length in this sample
	registerNalUnits(stream, &sampleInfo)

	if stream.isVideo() {

		// Registers all iFrames
		registerISamples(fragmentInfo, &sampleInfo)

		// Compute the pcr
		registerPCRSamples(stream, fragmentInfo, &sampleInfo)

		// Retrieve the compositionTimeOffset and decodingTimeOffset
		registerCTSAndDTSSamples(stream, fragmentInfo, &sampleInfo)
	}

	return
}

func registerNalUnits(info StreamInfo, sampleInfo *[]SampleInfo) {

	// Set the size of mdat to read NAL unit length only
	info.mdat.Size = info.nalLengthSize

	for i := 0; i < len(*sampleInfo); i++ {
		sample := &(*sampleInfo)[i]

		// Get the start and end offset in the mdat
		offset := sample.mdatOffset
		endOffset := sample.mdatOffset + int64(sample.mdatSize - info.nalLengthSize)

		// Parse all possibles NAL unit
		for offset < endOffset {

			// Create the NAL Unit to keep information
			nalUnit := NALUnit{}

			// Read the bytes representing the NAL unit length
			info.mdat.Offset = offset

			// Set start and size of the NAL
			// start = Offset + NAL length size
			nalUnit.mdatOffset = offset + int64(info.nalLengthSize)
			bytes := info.mdat.ToBytes()
			nalUnit.mdatSize = binary.BigEndian.Uint32(bytes)

			// Add Unit to other saved NAL Units
			sample.NALUnits = append(sample.NALUnits, nalUnit)
			offset += int64(nalUnit.mdatSize) + 4
		}
	}
}

func registerSamplesSizes(stream StreamInfo, info FragmentInfo, sampleInfo *[]SampleInfo) {

	offset := stream.mdat.Offset

	if stream.stsz.SampleSize == 0 {
		for i := 0; i < len(*sampleInfo); i++ {
			(*sampleInfo)[i].mdatSize = stream.stsz.EntrySize[uint32(i) + info.sampleStart]
			(*sampleInfo)[i].size = (*sampleInfo)[i].mdatSize
			(*sampleInfo)[i].mdatOffset = offset
			offset += int64((*sampleInfo)[i].mdatSize)
		}
	} else {
		for i := 0; i < len(*sampleInfo); i++ {
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


	for i := 0; i < len(*sampleInfo); i++ {

		(*sampleInfo)[i].PCR = uint64((fragmentInfo.sampleStart) + uint32(i)* stream.SampleDelta)
	}
}

func registerCTSAndDTSSamples(stream StreamInfo, fragmentInfo FragmentInfo, sampleInfo *[]SampleInfo) {

	if len(*sampleInfo) == 0 {
		return
	}

	emitter := IEmitter{}
	emitter.Min_emit = 0

	cttsOffset := fragmentInfo.cttsOffset
	cttsSampleCount := fragmentInfo.cttsSampleCount
	sttsOffset := fragmentInfo.sttsOffset
	sttsSampleCount := fragmentInfo.sttsSampleCount
	dts := fragmentInfo.dts

	// Register first sample
	sample := &(*sampleInfo)[0]
	sample.DTS = dts
	if stream.compositionTimeOffset {
		sample.CTS = uint64(stream.ctts.Entries[cttsOffset].SampleOffset) + dts // - dConf.MediaTime
	} else {
		sample.CTS = dts
	}

	for i := 1; i < len(*sampleInfo); i++ {

		// Update Composition time offset
		if stream.compositionTimeOffset {
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
		}

		// Update Decoding time offset
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
		dts += uint64(stream.stts.Entries[sttsOffset].SampleDelta)

		// Register DTS and CTS
		sample = &(*sampleInfo)[i]
		sample.DTS = dts
		if stream.compositionTimeOffset {
			sample.CTS = uint64(stream.ctts.Entries[cttsOffset].SampleOffset) + dts // - dConf.MediaTime
		} else {
			sample.CTS = dts
		}
	}
}