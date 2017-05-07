package ts

import "mp4"

type StreamInfo struct {
	mp4.Conf
	filename string

	mdat mp4.MdatBox
	ctts mp4.CttsBox
	stss mp4.StssBox
	stsz mp4.StszBox
	stts mp4.SttsBox
	avcC mp4.AvcCBox
	mdhd mp4.MdhdBox

	compositionTimeOffset bool
	PID                   uint16
	streamType            uint32

	// Sample delta scaled with timeScale
	ClockScaled float64

	// Number of bytes describing the nal length (e.g. 4)
	nalLengthSize		  uint32
}

func (info StreamInfo) isVideo() (bool) {
	return info.Type == "video"
}

func (info StreamInfo) isAudio() (bool) {
	return info.Type == "audio"
}