package hls

import (
	"mp4"
	"fmt"
)

// Create subtitles variant list
func createMainSubtitlesDescriptor(subtitles []mp4.TrackEntry, videoId string) (s string) {
	for _, sub := range subtitles {
		s += fmt.Sprintf(`#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID="subs",DEFAULT=NO,FORCED=NO,NAME="subtitle_%s",LANGUAGE="%s",URI="%s_subtitle_%s_%d.hls"`, sub.Lang, sub.Lang, videoId, sub.Lang, sub.Bandwidth) + "\n"
	}
	return
}

// Create audio variant list
func createMainAudioDescriptor(audios []mp4.TrackEntry, videoId string) (s string) {
	for _, audio := range audios {
		s += fmt.Sprintf(`#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",LANGUAGE="%s",NAME="audio_%s",AUTOSELECT=YES,DEFAULT=YES,URI="%s_audio_%s_%d.hls"`, audio.Lang, audio.Lang, videoId, audio.Lang, audio.Bandwidth) + "\n"
	}
	return
}

// Create video quality variant list with stream
// Variant list with different video can be added
func createMainVideoDescriptor(videos []mp4.TrackEntry, videoId string) (s string) {
	for i, video := range videos {
		s += fmt.Sprintf("#EXT-X-STREAM-INF:PROGRAM-ID=1," +
			"BANDWIDTH=%d,RESOLUTION=%dx%d," +
			"CODECS=\"avc1.%.2x%.2x%.2x,mp4a.40.2\"",
			video.Bandwidth,
			video.Config.Video.Width,
			video.Config.Video.Height,
			video.Config.Video.CodecInfo[0],
			video.Config.Video.CodecInfo[1],
			video.Config.Video.CodecInfo[2])

		//s += ",AUDIO=\"audio\",SUBTITLES=\"subs\""
		if i == 0 {
			s += ",AUTOSELECT=YES,DEFAULT=YES"
		} else {
			s += ",AUTOSELECT=YES,DEFAULT=NO"
		}
		s += "\n"
		s += fmt.Sprintf("%s_video_%s_%d.hls\n", videoId, video.Lang, video.Bandwidth)
	}

	return
}

func CreateMediaDescriptor(fragmentDuration uint32, numberOfSegment uint32, videoId string, trackType string, trackLang string, trackBandwidth uint64) (s string) {
	s = "#EXTM3U\n"
	s += fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", fragmentDuration)
	s += "#EXT-X-VERSION:3\n"
	s += "#EXT-X-MEDIA-SEQUENCE:0\n"

    var i uint32
	for i = 1; i <= numberOfSegment; i++ {
		s += fmt.Sprintf("#EXTINF:%d,\n", fragmentDuration)
		s += fmt.Sprintf("%s_%s_%s_%d-%d.ts\n", videoId, trackType, trackLang, trackBandwidth, i)
	}
	s += "#EXT-X-ENDLIST"

	return
}

func CreateSubtitlesDescriptor(videoId string, trackLang string, trackBandwidth uint64) (s string) {
	s = "#EXTM3U\n"
	s += fmt.Sprintf("#EXT-X-TARGETDURATION:1\n")
	s += "#EXT-X-VERSION:3\n"
	s += "#EXT-X-MEDIA-SEQUENCE:0\n"
	s += fmt.Sprintf("#EXTINF:1,\n")
	s += fmt.Sprintf("%s_subtitle_%s_%d.vtt\n", videoId, trackLang, trackBandwidth)
	s += "#EXT-X-ENDLIST"

	return
}

func CreateMainDescriptor(jConf mp4.JsonConfig, videoId string) (s string) {
	s = "#EXTM3U\n"
	s += createMainVideoDescriptor(jConf.Tracks["video"], videoId)
	//s += createMainAudioDescriptor(jConf.Tracks["audio"], videoId)
	s += createMainSubtitlesDescriptor(jConf.Tracks["subtitle"], videoId)
	return
}

