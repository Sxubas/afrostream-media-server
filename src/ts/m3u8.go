package ts

import (
	"mp4"
	"fmt"
)

func CreateMainDescriptor(jConf mp4.JsonConfig, videoId string) (s string) {
	s = "#EXTM3U"
	s += CreateMainAudioDescriptor(jConf.Tracks["audio"], videoId)
	s += CreateMainSubtitlesDescriptor(jConf.Tracks["subtitle"])
	s += CreateMainVideoDescriptor(jConf.Tracks["video"], videoId)
	return
}

// Create audio variant list
func CreateMainAudioDescriptor(audios []mp4.TrackEntry, videoId string) (s string) {
	for _, audio := range audios {
		s += fmt.Sprintf("#EXT-X-MEDIA:" +
			"TYPE=AUDIO,GROUP-ID=\"audio\"," +
			"LANGUAGE=\"%s\"," +
			"NAME=\"%s\"," +
			"AUTOSELECT=YES," +
			"DEFAULT=YES," +
			"URI=\"%s/hls/audio/%s/index.m3u8\"\n",
			audio.Lang,
			audio.Lang,
			videoId,
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
func CreateMainVideoDescriptor(videos []mp4.TrackEntry, videoId string) (s string) {

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
		s += fmt.Sprintf("%s/hls/video/%d/index.m3u8\n", videoId, i)
	}
	return
}


func CreateMediaDescriptor(videoId string, param string, extension string, fragmentDuration uint32, numberOfSegment int) (s string) {

	s = "#EXTM3U\n"
	s += "#EXT-X-PLAYLIST-TYPE:VOD\n"
	s += fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", fragmentDuration)
	s += "#EXT-X-VERSION:3\n"
	s += "#EXT-X-MEDIA-SEQUENCE:0"

	for i := 0; i < numberOfSegment; i++ {
		s += fmt.Sprintf("EXT-INF:%d,\n", fragmentDuration)
		s += fmt.Sprintf("%s/hls/%s-%d.%s", videoId, param, i, extension)
	}

	s += "#EXT-X-ENDLIST"

	return
}
