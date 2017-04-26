package ts

import "mp4"

type StreamInfo struct {
	mp4.DashConfig
	Type string
}

func (info StreamInfo) isVideo() (bool) {
	return info.Type == "video"
}

func (info StreamInfo) isAudio() (bool) {
	return info.Type == "audio"
}


/*
// Getter
func (info StreamInfo) getTimeScale() (uint32) {
	return info.dConf.Timescale
}

func (info StreamInfo) getSampleDelta() (uint32) {
	return info.dConf.SampleDelta
}

func (info StreamInfo) getMediaTime() (int64) {
	return info.dConf.MediaTime
mdatOffset
}*/