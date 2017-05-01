package ts

import (
	"mp4"
)


func CreateHLSFragmentWithConf(dConf mp4.Conf, filename string, fragmentNumber uint32, fragmentDuration uint32) ([]byte) {

	if fragmentNumber == 0 {
		panic("Fragment number incorrect")
	}

	// Variables data used to create our modifiedFragment
	modifiedFragment := FragmentData{}

	// 1) analyse the stream and found get main information
	streamInfo := AnalyseStream(dConf, filename)

	// 2) Create program packets
	RegisterProgramPackets(*streamInfo, &modifiedFragment)

	// 3) Retrieve information on the created modifiedFragment
	fragmentInfo := GetFragmentInfo(streamInfo, fragmentNumber, fragmentDuration)

	// 4) Retrieve information on all contained samples
	samplesInfo := GetSamplesInfo(*streamInfo, *fragmentInfo)

	// 5) Create PES packets
	RegisterStreamPackets(*streamInfo, samplesInfo, &modifiedFragment)

	// 6) Create our fragment assembling all created packets
	fragment := FinaliseFragment(&modifiedFragment)
	return fragment
}
/*
sample.CTS = 7200
sample.DTS = 0

sameTimeStamps := sample.DTS == sample.CTS

// If CTS needed
headerLength := uint32(5)
if !sameTimeStamps {
headerLength += 5
}

streamSize := 9 + headerLength + sample.size

data := NewData(int(streamSize))
data.PushUInt(1, 24) 			// Packet start id code
data.PushUInt(sample.pesStream, 8) 	// Pes stream
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

if sameTimeStamps {
data.PushUInt(0x02, 2)
} else {
data.PushUInt(0x03, 2)
}

data.PushUInt(0, 1)				// ESCR flag
data.PushUInt(0, 1)				// ESCR Rate flag
data.PushUInt(0, 1)				// DSM_trick_mode_flag
data.PushUInt(0, 1)				// additional_copy_info_flag
data.PushUInt(0, 1)				// PES_CRC_flag
data.PushUInt(0, 1)				// PES_extension_flag
data.PushUInt(headerLength, 8)			// Header length



pushTimestamp(sample.CTS, data)

if !sameTimeStamps {
data.PushUInt(1, 4)
pushTimestamp(sample.DTS, data)*/