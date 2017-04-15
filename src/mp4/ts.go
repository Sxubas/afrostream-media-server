package ts

import (
	"mp4"
)


func CreateHLSInit();

func CreateHLSFragment() {
	// Create SDT 17
	// Create PAT 0
	// Create PMT 4096
	// Continue with mdat
		// Video stream 256
		// Audio stream 255
}

func CreateHLSInitWithConf();

func CreateHLSFragmentWithConf();
