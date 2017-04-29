package ts

import (
	"mp4"
)

// Analyse the stream and get main information
func AnalyseStream(dConf mp4.Conf, filename string) (streamInfo StreamInfo){
	// Save conf
	info.Conf = dConf
	info.filename = filename

	// Retrieve needed boxes
	LoadBoxes(info)

	// Check if it has composition offset (PTS/DTS)
	info.compositionTimeOffset = info.isVideo() && dConf.Video.CttsBoxOffset != 0

	if info.isVideo() {
		info.PID = 257
		info.streamId = 224
	} else {
		info.PID = 256
		info.streamId = 127
	}

	return
}

func LoadBoxes(info StreamInfo) {

	// Create the container to load mp4File content
	mp4File := mp4.ReadMainBoxes(info.filename, info.Conf)

	// Get data
	// MDAT

	info.mdat = mp4File["mdat"][0].(mp4.MdatBox)

	// Get the composition time offset
	// CTTS
	info.ctts = mp4File["moov.trak.mdia.minf.stbl.ctts"][0].(mp4.CttsBox)

	// Get the decoding time offset
	info.stts = mp4File["moov.trak.mdia.minf.stbl.stts"][0].(mp4.SttsBox)

	// Retrieve all iFrames
	// STSS
	info.stss = mp4File["moov.trak.mdia.minf.stbl.stss"][0].(mp4.StssBox)

	// Get size information
	// STSZ
	info.stsz = mp4File["moov.trak.mdia.minf.stbl.stsz"][0].(mp4.StszBox)
}