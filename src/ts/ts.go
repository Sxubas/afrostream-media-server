package ts

import (
	"mp4"
)


func CreateHLSFragmentWithConf(dConf mp4.DashConfig, fragmentNumber uint32, fragmentDuration uint32) ([]byte) {

	// Variables data used to create our modifiedFragment
	modifiedFragment := FragmentData{}


	// 1) analyse the stream and found get main information
	streamInfo := AnalyseStream(mp4)

	// 2) Create program packets
	RegisterProgramPackets(streamInfo, modifiedFragment)

	// 3) Retrieve information on the created modifiedFragment
	fragmentInfo := GetFragmentInfo(streamInfo, fragmentNumber, fragmentDuration)

	// 4) Retrieve information on all contained samples
	samplesInfo := GetSamplesInfo(fragmentInfo, modifiedFragment)

	// 5) Create PES packets
	RegisterStreamPackets(samplesInfo, modifiedFragment)

	// 6) Create our fragment assembling all created packets
	return FinaliseFragment(modifiedFragment)
}

