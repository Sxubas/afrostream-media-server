package ts

import (
	"mp4"
	"fmt"
	"strconv"
)

func CreateVideoAdaptationSetHLS(tracks []mp4.TrackEntry, videoId string, sDuration uint32) (s string, err error) {
	/*s = "EXTM3U\n"
	s += "#EXT-X-VERSION:3"
	s += "#EXT-X-MEDIA-SEQUENCE:0"
	s += "#EXT-X-TARGETDURATION:" + string(sDuration);*/

	for i, t := range tracks {
		s += "#EXTINF:" + fmt.Sprintf("%d", t.Config.Duration/uint64(t.Config.Timescale))  + "\n";
		s += "./" + t.Name + "-" + strconv.Itoa(i) + "-" + "0" + ".ts\n";
	}

	return
}

func CreateAudioAdaptationSetHLS(tracks []mp4.TrackEntry, videoId string, sDuration uint32) (s string, err error) {
	/*s = "EXTM3U\n"
	s += "#EXT-X-VERSION:3"
	s += "#EXT-X-MEDIA-SEQUENCE:0"
	s += "#EXT-X-TARGETDURATION:" + string(sDuration);*/

	for i, t := range tracks {
		s += "#EXTINF:" + fmt.Sprintf("%d", t.Config.Duration/uint64(t.Config.Timescale)) + "\n";
		s += "./" + t.Name + "-" + strconv.Itoa(i) + "-" + "0" + ".ts\n";
	}

	return
}

func CreateHLSPlaylist(jConf mp4.JsonConfig, videoId string) (playlist string) {
	playlist = ""
	playlist += "#EXTM3U\n"
	playlist += "#EXT-X-TARGETDURATION:" + fmt.Sprintf("%d", jConf.SegmentDuration) + "\n";

	a, _ := CreateVideoAdaptationSetHLS(jConf.Tracks["video"], videoId, jConf.SegmentDuration)
	playlist += a
	a, _ = CreateAudioAdaptationSetHLS(jConf.Tracks["audio"], videoId, jConf.SegmentDuration)
	playlist += a
	fmt.Printf("%s", playlist)
	return
}