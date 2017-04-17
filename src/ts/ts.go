package ts

import (
)

func Test() {
	var header Header;
	header.TransportErrorIndicator = 1
	header.PayloadUnitStartIndicator = 0;
	header.PID = 513;
	header.AdaptationFieldControl = 3;
	header.ContinuityCounter = 3

	data := header.ToBytes();
	data.PrintBinary()
	data.PrintHex()
}

func CreateHLSInit() {

}

func CreateHLSFragment(mp4 map[string][]interface{},
	fragmentNumber uint32,
	fragmentDuration uint32,
	PMTSpacing uint32) (ts []interface{}){

	ts = make([]interface{}, 0);

	// Retrieve main information

	// Create default PAT

	// Create Default PMT

	// Fill with Stream


	// Create PAT 0
	// Create PMT 4096
	// Continue with mdat
		// Video stream 256
		// Audio stream 255
	return
}

func CreateHLSInitWithConf() {

}

func CreateHLSFragmentWithConf() {

}


