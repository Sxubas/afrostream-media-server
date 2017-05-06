package ts

import (
	"mp4"
	"fmt"
)

func CreateMainDescriptor(jConf mp4.JsonConfig) (s string) {
	s = "#EXTM3U"
	s += CreateMainAudioDescriptor(jConf.Tracks["audio"])
	s += CreateMainSubtitlesDescriptor(jConf.Tracks["subtitle"])
	s += CreateMainVideoDescriptor(jConf.Tracks["video"])
	return
}

// Create audio variant list
func CreateMainAudioDescriptor(audios []mp4.TrackEntry) (s string) {
	for _, audio := range audios {
		s += fmt.Sprintf("#EXT-X-MEDIA:" +
			"TYPE=AUDIO,GROUP-ID=\"audio\"," +
			"LANGUAGE=\"%s\"," +
			"NAME=\"%s\"," +
			"AUTOSELECT=YES," +
			"DEFAULT=YES," +
			"URI=\"audio/%s/index.m3u8\"\n",
			audio.Lang,
			audio.Lang,
			audio.Lang)
	}
	return
}

// Create subtitles variant list
func CreateMainSubtitlesDescriptor(subtitles []mp4.TrackEntry) (s string) {
	for _, sub := range subtitles {
		s += fmt.Sprintf("#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subs\"," +
			"DEFAULT=NO,FORCED=NO," +
			"NAME=\"%s\"," +
			"LANGUAGE=\"%s\"," +
			"URI=\"%s\"\n",
			sub.Lang,
			sub.Lang,
			sub.File)
	}
	return
}

// Create video quality variant list with stream
// Variant list with different video can be added
func CreateMainVideoDescriptor(videos []mp4.TrackEntry) (s string) {

	for i, video := range videos {
		s += fmt.Sprintf("#EXT-X-STREAM-INF:PROGRAM-ID=1," +
			"BANDWIDTH=%d,RESOLUTION=%dx%d," +
			"CODECS=\"avc1.%.2X%.2X%.2X,mp4a.40.2\"," +
			"AUDIO=\"audio\",SUBTITLES=\"subs\"\n",
			video.Bandwidth,
			video.Config.Video.Width,
			video.Config.Video.Height,
			video.Config.Video.CodecInfo[0],
			video.Config.Video.CodecInfo[1],
			video.Config.Video.CodecInfo[2])
		s += fmt.Sprintf("video/%d/index.m3u8\n", i)
	}
	return
}


func CreateMediaDescriptor(param string, extension string, fragmentDuration uint32, numberOfSegment int) (s string) {

	s = "#EXTM3U\n"
	s += "#EXT-X-PLAYLIST-TYPE:VOD\n"
	s += fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", fragmentDuration)
	s += "#EXT-X-VERSION:3\n"
	s += "#EXT-X-MEDIA-SEQUENCE:0\n"

	for i := 1; i <= numberOfSegment; i++ {
		s += fmt.Sprintf("EXT-INF:%d,\n", fragmentDuration)
		s += fmt.Sprintf("%s%d.%s\n", param, i, extension)
	}

	s += "#EXT-X-ENDLIST"

	return
}
