package ts

type FragmentInfo struct {
	Type string
	Mp4 map[string][]interface{}
}

func (info FragmentInfo) isVideo() (bool) {
	return info.Type == "video"
}

func (info FragmentInfo) isAudio() (bool) {
	return info.Type == "audio"
}
