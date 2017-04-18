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
	fragmentDuration uint32) (ts []Packet){

	ts = make([]Packet, 0);
	var isVideo bool

	// Check if the mp4 is audio or video
	if mp4["moov.trak.mdia.minf.smhd"] != nil {
		isVideo = false
	}
	if mp4["moov.trak.mdia.minf.vmhd"] != nil {
		isVideo = true
	}

	// Retrieve main information

	// Create default PAT
	pat := NewPAT()

	// Create Default PMT
	pmt := NewPMT()

	// Fill with Stream


	// Create PAT 0
	// Create PMT 4096
	// Continue with mdat
		// Video stream 256
		// Audio stream 255
	return
}

func CreateHLSInitWithConf(mp4 map[string][]interface{}, fragmentNumber uint32, fragmentDuration uint32) (ts []Packet) {



	return
}

func CreateHLSFragmentWithConf() {

}


