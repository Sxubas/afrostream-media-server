package ts

import (
	"mp4"
	"time"
)

// Analyse the stream and get main information
func AnalyseStream(dConf mp4.Conf, filename string) (streamInfo *StreamInfo){

	streamInfo = new(StreamInfo)

	// Save conf
	streamInfo.Conf = dConf
	streamInfo.filename = filename

	// Retrieve needed boxes
	loadBoxes(streamInfo)

	// Get the information from the boxes
	registerInformation(streamInfo)

	return
}

func loadBoxes(info *StreamInfo) {

	// Create the container to load mp4File content
	mp4File := mp4.ReadMainBoxes(info.filename, info.Conf)

	// Get data
	// MDAT
	info.mdat = mp4File["mdat"][0].(mp4.MdatBox)

	// Get the composition time offset
	// CTTS
	if mp4File["moov.trak.mdia.minf.stbl.ctts"] != nil {
		info.ctts = mp4File["moov.trak.mdia.minf.stbl.ctts"][0].(mp4.CttsBox)
	}

	// Get the decoding time offset
	if mp4File["moov.trak.mdia.minf.stbl.stts"] != nil {
		info.stts = mp4File["moov.trak.mdia.minf.stbl.stts"][0].(mp4.SttsBox)
	}

	// Get size information
	// STSZ
	info.stsz = mp4File["moov.trak.mdia.minf.stbl.stsz"][0].(mp4.StszBox)

	// Get timeScale
	info.mdhd = mp4File["moov.trak.mdia.mdhd"][0].(mp4.MdhdBox)

	if info.isVideo() {
		// Retrieve all iFrames
		// STSS
		info.stss = mp4File["moov.trak.mdia.minf.stbl.stss"][0].(mp4.StssBox)

		// Get nal information (And SPS/PPS)
		// Avc box can be placed at two locations
		avcCBox := mp4File["moov.trak.mdia.minf.stbl.stsd.avc1.avcC"]
		if avcCBox == nil {
			avcCBox = mp4File["moov.trak.mdia.minf.stbl.stsd.encv.avcC"]
		}
		info.avcC = avcCBox[0].(mp4.AvcCBox)
	}
}

func registerInformation(streamInfo *StreamInfo) {
	streamInfo.SampleDelta = streamInfo.stts.Entries[0].SampleDelta
	streamInfo.Duration = streamInfo.mdhd.Duration

	// Get sample delta to compute PCR for each Sample
	timeScale := time.Duration(streamInfo.mdhd.Timescale)
	streamInfo.ClockScaled = float64(time.Second) / float64(timeScale*11111)

	// Check if it has composition offset (PTS/DTS)
	streamInfo.compositionTimeOffset = streamInfo.isVideo() && streamInfo.ctts.Offset != 0

	// Register PID for audio and video streams
	if streamInfo.isVideo() {
		streamInfo.PID = 258
		streamInfo.streamType = 27 //224

		// Retrieve information for nal segmentation
		// Use the avcC box: NalUnitLengthSize minus one + 1
		streamInfo.nalLengthSize = uint32(streamInfo.avcC.NalUnitSize & 0x03 + 1)
	} else {
		streamInfo.PID = 256
		streamInfo.streamType = 15 //127
	}

}
