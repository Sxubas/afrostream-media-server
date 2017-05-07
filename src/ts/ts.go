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
	CreateProgramPackets(*streamInfo, &modifiedFragment)

	// 3) Retrieve information on the created modifiedFragment
	fragmentInfo := GetFragmentInfo(streamInfo, fragmentNumber, fragmentDuration)

	// 4) Retrieve information on all contained samples
	samplesInfo := GetSamplesInfo(*streamInfo, *fragmentInfo)

	// 5) Create PES packets
	CreateStreamPackets(*streamInfo, samplesInfo, &modifiedFragment)

	// 6) Create our fragment assembling all created packets
	fragment := FinaliseFragment(&modifiedFragment)
	return fragment
}