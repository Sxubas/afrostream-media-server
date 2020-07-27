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

package ams

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/Sxubas/afrostream-media-server/src/dash"
	"github.com/Sxubas/afrostream-media-server/src/hls"
	"github.com/Sxubas/afrostream-media-server/src/logger"
	"github.com/Sxubas/afrostream-media-server/src/mp4"
	"github.com/Sxubas/afrostream-media-server/src/ts"
	"github.com/Sxubas/afrostream-media-server/src/util"
)

const (
	contentTypeHtml = "text/html"
	contentTypeFile = "application/octet-stream"
)

func handleFileRequest(w http.ResponseWriter, path string, contentType string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		logger.Error("%s", err.Error())
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	_, err = w.Write(data)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		logger.Error("%s", err.Error())
	}
}

func handleManifestRequest(w http.ResponseWriter, dir string, basename string, extension string) {
	data, err := ioutil.ReadFile(path.Join(dir, basename+".json"))
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		logger.Error("%s", err.Error())
		return
	}

	var jConfig mp4.JsonConfig
	err = json.Unmarshal(data, &jConfig)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		logger.Error("%s", err.Error())
		return
	}

	var manifest string
	switch extension {
	case ".mpd":
		manifest = dash.CreateDashManifest(jConfig, basename)
		w.Header().Set("Content-Type", "application/dash+xml")
	case ".m3u8":
		manifest = hls.CreateMainDescriptor(jConfig, basename)
		w.Header().Set("Content-Type", "application/x-mpegURL")
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(manifest)))
	_, err = w.Write([]byte(manifest))
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		logger.Error("%s", err.Error())
	}
}

func handleMediaRequest(w http.ResponseWriter, dir string, basename string, extension string) {
	trackName, trackType, trackLang, trackId, err := util.ParseBasename(basename)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		logger.Error("%s", err.Error())
		return
	}

	trackIds := strings.Split(trackId, "-")
	var trackBandwidth uint64
	trackBandwidth, err = strconv.ParseUint(trackIds[0], 10, 64)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		logger.Error("%s", err.Error())
		return
	}

	data, err := ioutil.ReadFile(path.Join(dir, trackName+".json"))
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		logger.Error("%s", err.Error())
		return
	}

	var jConfig mp4.JsonConfig
	err = json.Unmarshal(data, &jConfig)
	if err != nil {
		http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
		logger.Error("%s", err.Error())
		return
	}

	for _, t := range jConfig.Tracks[trackType] {
		if t.Lang == trackLang && t.Bandwidth == trackBandwidth {
			t.File = "/" + t.File

			var b []byte

			// DASH : InitData or Fragment
			// HLS  : Playlist or Fragment
			// Subititles

			switch extension {
			case ".dash":
				content := mp4.CreateDashInitWithConf(*t.Config) // InitData
				b = mp4.MapToBytes(content)
				w.Header().Set("Content-Type", "video/mp4")
			case ".m4s":
				if len(trackIds) != 2 {
					http.Error(w, `{ "status": "ERROR", "reason": "Invalid track Id" }`, http.StatusInternalServerError)
					logger.Error("Invalid track Id")
					return
				}
				var num uint64
				num, err = strconv.ParseUint(trackIds[1], 10, 32)
				if err != nil {
					http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
					logger.Error("%s", err.Error())
					return
				}
				var segmentNumber uint32
				segmentNumber = uint32(num)
				content := mp4.CreateDashFragmentFromFileWithConf(*t.Config, t.File, segmentNumber, jConfig.SegmentDuration) // Fragment
				b = mp4.MapToBytes(content)
				w.Header().Set("Content-Type", "video/mp4")

			case ".hls":
				if trackType == "subtitle" {
					b = []byte(hls.CreateSubtitlesDescriptor(trackName, trackLang, trackBandwidth))
				} else {
					segmentNumber := util.NumberOfSegments(t, jConfig)
					b = []byte(hls.CreateMediaDescriptor(jConfig.SegmentDuration, segmentNumber, trackName, trackType, trackLang, trackBandwidth))
				}
				w.Header().Set("Content-Type", "application/x-mpegURL")
			case ".ts":
				var num uint64
				num, err = strconv.ParseUint(trackIds[1], 10, 32)
				if err != nil {
					http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
					logger.Error("%s", err.Error())
					return
				}
				var segmentNumber uint32
				segmentNumber = uint32(num)
				b = ts.CreateHLSFragmentWithConf(*t.Config, t.File, segmentNumber, jConfig.SegmentDuration)
				w.Header().Set("Content-Type", "video/MP2T")

			case ".vtt":
				handleFileRequest(w, t.File, contentTypeFile)
				return
			}

			w.Header().Set("Content-Length", strconv.Itoa(len(b)))
			_, err := w.Write(b)
			if err != nil {
				http.Error(w, `{ "status": "ERROR", "reason": "`+err.Error()+`" }`, http.StatusInternalServerError)
				logger.Error("%s", err.Error())
				return
			}
			return
		}
	}

	http.Error(w, `{ "status": "ERROR", "reason": "Media not found : `+path.Join(dir, basename+extension)+`" }`, http.StatusInternalServerError)
	logger.Error("Media not found : %s", path.Join(dir, basename+extension))
}

func handleContentRequest(w http.ResponseWriter, dir string, basename string, extension string) {
	switch extension {
	case ".mpd":
		handleManifestRequest(w, dir, basename, extension)
	case ".dash":
		handleMediaRequest(w, dir, basename, extension)
	case ".m4s":
		handleMediaRequest(w, dir, basename, extension)

	case ".m3u8":
		handleManifestRequest(w, dir, basename, extension)
	case ".hls":
		handleMediaRequest(w, dir, basename, extension)
	case ".ts":
		handleMediaRequest(w, dir, basename, extension)

	case ".vtt":
		handleMediaRequest(w, dir, basename, extension)

	default:
		http.Error(w, `{ "status": "ERROR", "reason": "Format is not supported" }`, http.StatusInternalServerError)
		logger.Error("Format is not supported")
	}
}

func handleHttpRequest(w http.ResponseWriter, r *http.Request) {
	logger.Info("Request -> %s", r.URL.Path)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "DNT,X-CustomHeader,Keep-Alive,Range,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type")
	w.Header().Set("Connection", "close")

	dir, filename := path.Split(path.Clean(r.URL.Path))
	basename, extension := util.SplitFilename(filename)

	paths := strings.Split(dir[1:], "/")

	// Switch between video (manifest, video, audio, subtitles) request and other request type

	switch paths[0] {
	case "video":
		handleContentRequest(w, dir[6:], basename, extension) // Remove relative path /video/ -> /
	default:
		switch extension {
		case ".html":
			handleFileRequest(w, r.URL.Path, contentTypeHtml)
		default:
			handleFileRequest(w, r.URL.Path, contentTypeFile)
		}
	}
}
