// Copyright (c) 2015
//      Sebastien Petit & Afrostream - www.afrostream.tv - spebsd@gmail.com.
//      All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
//    may be used to endorse or promote products derived from this software
//    without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
// OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
// LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
// OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
// SUCH DAMAGE.

package main

import (
	"log"
	"mp4"
	"net/http"
	"os"
	//"io"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"ts"
	"syscall"
)

func readFile(filename string) (data []byte, r error) {
	f, err := os.Open(filename)
	if err != nil {
		r = err
		return
	}
	fi, err := f.Stat()
	if err != nil {
		r = err
		return
	}
	size := fi.Size()
	data = make([]byte, size)
	offset := 0
	for size > 0 {
		count, err := f.Read(data[offset:])
		if err != nil {
			r = err
			return
		}
		size -= int64(count)
		offset += count
	}

	return
}

func createExternalSubtitlesAdaptationSet(tracks []mp4.TrackEntry) (s string, err error) {
	s = ""
	for _, t := range tracks {
		s += fmt.Sprintf(`    <AdaptationSet mimeType="text/vtt" lang="%s">`, t.Lang) + "\n"
		s += fmt.Sprintf(`      <Representation id="%s" bandwidth="%d">`, t.Name, t.Bandwidth) + "\n"
		s += fmt.Sprintf(`        <BaseURL>../../%s</BaseURL>`, t.File) + "\n"
		s += `      </Representation>` + "\n"
		s += `    </AdaptationSet>` + "\n"
	}

	return
}

func createAudioAdaptationSet(tracks []mp4.TrackEntry, videoId string, sDuration uint32) (s string, err error) {
	var minBandwidth uint64
	var maxBandwidth uint64

	minBandwidth = 0
	maxBandwidth = 0
	for _, t := range tracks {
		if minBandwidth == 0 || t.Bandwidth < minBandwidth {
			minBandwidth = t.Bandwidth
		}
		if maxBandwidth == 0 || t.Bandwidth > maxBandwidth {
			maxBandwidth = t.Bandwidth
		}
	}
	if minBandwidth == 0 {
		err = errors.New("cannot found valid audio tracks")
		return
	}
	s = `    <AdaptationSet` + "\n"
	s += fmt.Sprintf(`      group="%d"`, 1) + "\n"
	s += `      contentType="audio"` + "\n"
	s += `      lang="en"` + "\n"
	s += fmt.Sprintf(`      minBandwidth="%d"`, minBandwidth) + "\n"
	s += fmt.Sprintf(`      maxBandwidth="%d"`, maxBandwidth) + "\n"
	s += `      segmentAlignment="true"` + "\n"
	s += fmt.Sprintf(`      audioSamplingRate="%d"`, tracks[0].Config.Timescale) + "\n"
	s += `      mimeType="audio/mp4"` + "\n"
	s += `      codecs="mp4a.40.2">` + "\n"
	s += `      <AudioChannelConfiguration` + "\n"
	s += `        schemeIdUri="urn:mpeg:dash:23003:3:audio_channel_configuration:2011"` + "\n"
	s += fmt.Sprintf(`        value="%d">`, tracks[0].Config.Audio.NumberOfChannels) + "\n"
	s += `      </AudioChannelConfiguration>` + "\n"
	s += `      <SegmentTemplate` + "\n"
	s += fmt.Sprintf(`        timescale="%d"`, tracks[0].Config.Timescale) + "\n"
	s += fmt.Sprintf(`        initialization="%s-$RepresentationID$.dash"`, videoId) + "\n"
	s += fmt.Sprintf(`        media="%s-$RepresentationID$-$Number$.m4s"`, videoId) + "\n"
	s += `        startNumber="1"` + "\n"
	s += fmt.Sprintf(`        duration="%d">`, sDuration*tracks[0].Config.Timescale) + "\n"
	s += `      </SegmentTemplate>` + "\n"
	for _, t := range tracks {
		s += `      <Representation` + "\n"
		s += fmt.Sprintf(`        id="%s=%d"`, t.Name, t.Bandwidth) + "\n"
		s += fmt.Sprintf(`        bandwidth="%d">`, t.Bandwidth) + "\n"
		s += `      </Representation>` + "\n"
	}
	s += `    </AdaptationSet>` + "\n"

	return
}

func createVideoAdaptationSet(tracks []mp4.TrackEntry, videoId string, sDuration uint32) (s string, err error) {
	var minBandwidth uint64
	var maxBandwidth uint64
	var minWidth uint16
	var maxWidth uint16
	var minHeight uint16
	var maxHeight uint16

	minBandwidth = 0
	maxBandwidth = 0
	minWidth = 0
	maxWidth = 0
	minHeight = 0
	maxHeight = 0
	for _, t := range tracks {
		if minBandwidth == 0 || t.Bandwidth < minBandwidth {
			minBandwidth = t.Bandwidth
		}
		if maxBandwidth == 0 || t.Bandwidth > maxBandwidth {
			maxBandwidth = t.Bandwidth
		}
		if minWidth == 0 || t.Config.Video.Width < minWidth {
			minWidth = t.Config.Video.Width
		}
		if maxWidth == 0 || t.Config.Video.Width > maxWidth {
			maxWidth = t.Config.Video.Width
		}
		if minHeight == 0 || t.Config.Video.Height < minWidth {
			minHeight = t.Config.Video.Height
		}
		if maxHeight == 0 || t.Config.Video.Height > maxWidth {
			maxHeight = t.Config.Video.Height
		}
	}
	if minBandwidth == 0 {
		err = errors.New("cannot found valid video tracks")
		return
	}
	s = `    <AdaptationSet` + "\n"
	s += fmt.Sprintf(`      group="%d"`, 2) + "\n"
	s += `      contentType="video"` + "\n"
	s += `      lang="en"` + "\n"
	s += fmt.Sprintf(`      minBandwidth="%d"`, minBandwidth) + "\n"
	s += fmt.Sprintf(`      maxBandwidth="%d"`, maxBandwidth) + "\n"
	s += fmt.Sprintf(`      minWidth="%d"`, minWidth) + "\n"
	s += fmt.Sprintf(`      maxWidth="%d"`, maxWidth) + "\n"
	s += fmt.Sprintf(`      minHeight="%d"`, minHeight) + "\n"
	s += fmt.Sprintf(`      maxHeight="%d"`, maxHeight) + "\n"
	s += `      segmentAlignment="true"` + "\n"
	s += `      mimeType="video/mp4"` + "\n"
	s += `      startWithSAP="1">` + "\n"
	s += `      <SegmentTemplate` + "\n"
	s += fmt.Sprintf(`        timescale="%d"`, tracks[0].Config.Timescale) + "\n"
	s += fmt.Sprintf(`        initialization="%s-$RepresentationID$.dash"`, videoId) + "\n"
	s += fmt.Sprintf(`        media="%s-$RepresentationID$-$Number$.m4s"`, videoId) + "\n"
	s += `        startNumber="1"` + "\n"
	s += fmt.Sprintf(`        duration="%d">`, sDuration*tracks[0].Config.Timescale) + "\n"
	s += `      </SegmentTemplate>` + "\n"

	for _, t := range tracks {
		s += `      <Representation` + "\n"
		s += fmt.Sprintf(`        id="%s=%d"`, t.Name, t.Bandwidth) + "\n"
		s += fmt.Sprintf(`        bandwidth="%d"`, t.Bandwidth) + "\n"
		s += fmt.Sprintf(`        width="%d"`, t.Config.Video.Width) + "\n"
		s += fmt.Sprintf(`        height="%d"`, t.Config.Video.Height) + "\n"
		s += fmt.Sprintf(`        codecs="avc1.%.2X%.2X%.2X"`, t.Config.Video.CodecInfo[0], t.Config.Video.CodecInfo[1], t.Config.Video.CodecInfo[2]) + "\n"
		s += `        scanType="progressive">` + "\n"
		s += `      </Representation>` + "\n"
	}
	s += `    </AdaptationSet>` + "\n"

	return
}

func createDashManifest(jConf mp4.JsonConfig, videoId string) (dashManifest string) {
	dashManifest = ""
	dashManifest += `<?xml version="1.0" encoding="utf-8"?>` + "\n"
	dashManifest += `<!-- Created with Afrostream Media Server -->` + "\n"
	dashManifest += `<MPD` + "\n"
	dashManifest += `xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"` + "\n"
	dashManifest += `xmlns="urn:mpeg:dash:schema:mpd:2011"` + "\n"
	dashManifest += `xsi:schemaLocation="urn:mpeg:dash:schema:mpd:2011 http://standards.iso.org/ittf/PubliclyAvailableStandards/MPEG-DASH_schema_files/DASH-MPD.xsd"` + "\n"
	dashManifest += `type="static"` + "\n"
	var duration uint64
	if jConf.Tracks["video"] != nil {
		duration = uint64(jConf.Tracks["video"][0].Config.Duration) / uint64(jConf.Tracks["video"][0].Config.Timescale)
	} else {
		duration = uint64(jConf.Tracks["audio"][0].Config.Duration) / uint64(jConf.Tracks["audio"][0].Config.Timescale)
	}
	dashManifest += fmt.Sprintf(`mediaPresentationDuration="PT%dH%dM%d.%dS"`, duration/3600, (duration/60)%60, duration%60, (duration*1000)%1000) + "\n"
	dashManifest += fmt.Sprintf(`maxSegmentDuration="PT%dS"`, jConf.SegmentDuration) + "\n"
	dashManifest += fmt.Sprintf(`minBufferTime="PT%dS"`, jConf.SegmentDuration+1) + "\n"
	dashManifest += `profiles="urn:mpeg:dash:profile:isoff-live:2011">` + "\n"
	dashManifest += `  <Period>` + "\n"
	dashManifest += `    <BaseURL>dash/</BaseURL>` + "\n"

	a, err := createAudioAdaptationSet(jConf.Tracks["audio"], videoId, jConf.SegmentDuration)
	if err != nil {
		return
	}
	dashManifest += a
	a, err = createVideoAdaptationSet(jConf.Tracks["video"], videoId, jConf.SegmentDuration)
	if err != nil {
		return
	}
	dashManifest += a
	a, err = createExternalSubtitlesAdaptationSet(jConf.Tracks["subtitle"])
	if err != nil {
		return
	}
	dashManifest += a

	dashManifest += `  </Period>` + "\n"
	dashManifest += `</MPD>` + "\n"

	return
}

func httpServerLoadPage(path string) (content []byte, err error) {
	content, err = ioutil.ReadFile(path)

	return
}

func treatDashRequest(videoIdPath string, s []string, w http.ResponseWriter) {
	split1 := strings.Split(s[1], "-")
	split2 := strings.Split(split1[1], "=")
	split3 := strings.Split(split2[1], ".")
	split4 := strings.Split(split2[0], "_")

	trackName := split2[0]
	trackType := split4[0]
	var trackBandwidth uint64
	num, err := strconv.ParseUint(split3[0], 10, 64)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}
	trackBandwidth = num
	data, err := readFile("." + videoIdPath + ".json")
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}

	var jConfig mp4.JsonConfig
	err = json.Unmarshal(data, &jConfig)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}

	for _, t := range jConfig.Tracks[trackType] {
		if t.Name == trackName && t.Bandwidth == trackBandwidth {
			dashInit := mp4.CreateDashInitWithConf(*t.Config)
			b := mp4.MapToBytes(dashInit)
			w.Header().Set("Content-Length", strconv.Itoa(len(b)))
			_, err := w.Write(b)
			if err != nil {
				http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
				return
			}
		}
	}
}

func treatM4SRequest(videoIdPath string, s []string, w http.ResponseWriter) {
	split1 := strings.Split(s[1], "-")
	split2 := strings.Split(split1[1], "=")
	split3 := strings.Split(split1[2], ".")
	split4 := strings.Split(split2[0], "_")

	trackName := split2[0]
	trackType := split4[0]
	var trackBandwidth uint64
	num, err := strconv.ParseUint(split2[1], 10, 64)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}
	trackBandwidth = num
	num, err = strconv.ParseUint(split3[0], 10, 32)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}
	var segmentNumber uint32
	segmentNumber = uint32(num)

	data, err := readFile("." + videoIdPath + ".json")
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}
	var jConfig mp4.JsonConfig
	err = json.Unmarshal(data, &jConfig)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}

	for _, t := range jConfig.Tracks[trackType] {
		if t.Name == trackName && t.Bandwidth == trackBandwidth {
			t.File = path.Dir(videoIdPath) + "/" + t.File
			fragment := mp4.CreateDashFragmentWithConf(*t.Config, t.File, segmentNumber, jConfig.SegmentDuration)
			fb := mp4.MapToBytes(fragment)
			sizeToWrite := len(fb)
			w.Header().Set("Content-Length", strconv.Itoa(sizeToWrite))
			for sizeToWrite > 0 {
				num, err := w.Write(fb)
				if err != nil {
					http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
					return
				}
				sizeToWrite -= num
			}
			return
		}
	}
}

func treatMPDRequest(videoId string, videoIdPath string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/dash+xml")
	data, err := readFile("." + videoIdPath + ".json")
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}
	var jConfig mp4.JsonConfig
	err = json.Unmarshal(data, &jConfig)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}
	mpdContent := createDashManifest(jConfig, videoId)
	w.Write([]byte(mpdContent))
}

func treatM3U8Request(splitDirs[] string, videoIdPath string, w http.ResponseWriter) {

	w.Header().Set("Content-Type", "application/x-mpegURL")
	data, err := readFile("." + videoIdPath + ".json")
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}
	var jConfig mp4.JsonConfig
	err = json.Unmarshal(data, &jConfig)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}

	err = ts.TreatM3U8Request(splitDirs, jConfig, w)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}
}

func treatTSRequest(splitDirs[] string, videoId string, videoIdPath string, w http.ResponseWriter) {

	w.Header().Set("Content-Type", "video/MP2T")
	data, err := readFile("." + videoIdPath + ".json")
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}
	var jConfig mp4.JsonConfig
	err = json.Unmarshal(data, &jConfig)

	err = ts.TreatTSRequest(splitDirs, jConfig, videoIdPath + ".json", videoId, w)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		return
	}
}

func treatVTTRequest(splitDirs[] string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/octet-stream")
	originalPath := strings.Join(splitDirs[1:], "/")
	html, err := httpServerLoadPage(originalPath)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "file not found" }`, http.StatusNotFound)
		return
	} else {
		w.Header().Set("Content-Length", strconv.Itoa(len(html)))
		w.Write(html)
	}
}

func httpRootServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "DNT,X-CustomHeader,Keep-Alive,Range,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type")
	w.Header().Set("Connection", "close")
	log.Printf("[ REQUEST ] %+v", r.URL)
	pathStr := r.URL.Path[:]
	//splitDirs := strings.Split(pathStr, "/")
	var s []string
	s = strings.Split(pathStr, ".json")
	if s[0] == pathStr {
		s = strings.Split(pathStr, ".ism")
	}
	if s[0] != pathStr {
		videoIdPath := s[0]
		videoId := path.Base(videoIdPath)
		splitDirs := strings.Split(s[1], "/")

		if len(splitDirs) > 2 {
			if splitDirs[1] == "dash" {
				w.Header().Set("Content-Type", "video/mp4")
				switch path.Ext(pathStr) {
				case ".mpd":
					treatMPDRequest(videoId, videoIdPath, w)
				case ".dash":
					treatDashRequest(videoIdPath, s, w)
				case ".m4s":
					treatM4SRequest(videoIdPath, s, w)
				}
			} else if splitDirs[1] == "hls" {
				switch path.Ext(pathStr) {
				case ".m3u8":
					treatM3U8Request(splitDirs[2:], videoIdPath, w)
				case ".ts":
					treatTSRequest(splitDirs[2:], videoId, videoIdPath, w)
				case ".vtt":
					treatVTTRequest(splitDirs[2:], w)
				}
			} else {
				http.Error(w, `{ "status": "ERROR", "reason": "format is not supported" }`, http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, `{ "status": "ERROR", "reason": "format is not supported" }`, http.StatusInternalServerError)
			return
		}
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		html, err := httpServerLoadPage(pathStr)
		if err != nil {
			http.Error(w, `{ "status": "ERROR", "reason": "file not found" }`, http.StatusNotFound)
			return
		} else {
			w.Header().Set("Content-Length", strconv.Itoa(len(html)))
			w.Write(html)
		}
	}

	return
}

func main() {
	documentRoot := flag.String("d", "", "Document Root (default: none)")
	portNumber := flag.String("p", "80", "Port number used by AMS for listening connections")
	flag.Parse()

	if *documentRoot == "" {
		fmt.Printf("Please specify the document root for the web server with -d <document root>")
		return
	}

	mp4.Debug(false)
	ts.DebugPlaylist = true

	err := syscall.Chroot(*documentRoot)
	if err != nil {
		fmt.Printf("Please run Afrostream Media Server as root, cannot chroot the document root directory for security: %v\n", err)
		fmt.Println("Changing port to 8080")
		*portNumber = "8080"
		//return
	}

	listenPort := ":" + *portNumber
	log.Printf(" [*] Running Afrostream Media Server on %s, To exit press CTRL+C", listenPort)
	http.HandleFunc("/", httpRootServer)
	http.ListenAndServe(listenPort, nil)

	return
}
