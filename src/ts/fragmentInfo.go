package ts

type FragmentInfo struct {
	sampleStart    	uint32
	sampleEnd      	uint32
	iFramesIndices 	[]uint32
	number         	uint32
	duration       	uint32
	cttsOffset	   	uint32
	cttsSampleCount uint32
	sttsOffset		uint32
	sttsSampleCount uint32
	dts				uint32
}

func (info FragmentInfo) getSampleCount() (uint32) {
	return info.sampleEnd - info.sampleStart + 1
}

func (info FragmentInfo) isFrameInFragment(i uint32) (bool) {
	return i >= info.sampleStart && i <= info.sampleEnd
}

