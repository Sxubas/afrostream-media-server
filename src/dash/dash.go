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

package dash

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "path"
    "strconv"
    "strings"

    "logger"
    "mp4"
    "util"
)

const (
    drm_system_id_widevine = "edef8ba979d64acea3c827dcd51d21ed"
)

func createExternalSubtitlesAdaptationSet(tracks []mp4.TrackEntry) (s string, err error) {
    s = ""
    for _, t := range tracks {
        s += fmt.Sprintf(`    <AdaptationSet mimeType="text/vtt" lang="%s">`, t.Lang) + "\n"
        s += fmt.Sprintf(`      <Representation id="%s" bandwidth="%d">`, t.Name, t.Bandwidth) + "\n"
        s += fmt.Sprintf(`        <BaseURL>%s</BaseURL>`, t.File) + "\n"
        s += `      </Representation>` + "\n"
        s += `    </AdaptationSet>` + "\n"
    }

    return
}

func createAudioAdaptationSet(tracks []mp4.TrackEntry, videoId string, segmentDuration uint32) (s string, err error) {
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
    s += fmt.Sprintf(`        initialization="%s_$RepresentationID$.dash"`, videoId) + "\n"
    s += fmt.Sprintf(`        media="%s_$RepresentationID$-$Number$.m4s"`, videoId) + "\n"
    s += `        startNumber="1"` + "\n"
    s += fmt.Sprintf(`        duration="%d">`, segmentDuration * tracks[0].Config.Timescale) + "\n"
    s += `      </SegmentTemplate>` + "\n"
    for _, t := range tracks {
        s += `      <Representation` + "\n"
        s += fmt.Sprintf(`        id="audio_%s_%d"`, t.Lang, t.Bandwidth) + "\n"
        s += fmt.Sprintf(`        bandwidth="%d">`, t.Bandwidth) + "\n"
        s += `      </Representation>` + "\n"
    }
    s += `    </AdaptationSet>` + "\n"

    return
}

func createVideoAdaptationSet(tracks []mp4.TrackEntry, videoId string, segmentDuration uint32) (s string, err error) {
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
    s += fmt.Sprintf(`        initialization="%s_$RepresentationID$.dash"`, videoId) + "\n"
    s += fmt.Sprintf(`        media="%s_$RepresentationID$-$Number$.m4s"`, videoId) + "\n"
    s += `        startNumber="1"` + "\n"
    s += fmt.Sprintf(`        duration="%d">`, segmentDuration * tracks[0].Config.Timescale) + "\n"
    s += `      </SegmentTemplate>` + "\n"

    for _, t := range tracks {
        s += `      <Representation` + "\n"
        s += fmt.Sprintf(`        id="video_%s_%d"`, t.Lang, t.Bandwidth) + "\n"
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

func createContentProtection(jConf mp4.JsonConfig, videoId string) (s string) {
    s = ""
    s += `<ContentProtection>`
    s += `</ContentProtection>`

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
    dashManifest += fmt.Sprintf(`mediaPresentationDuration="PT%dH%dM%d.%dS"`, duration / 3600, (duration / 60) % 60, duration % 60, (duration * 1000) % 1000) + "\n"
    dashManifest += fmt.Sprintf(`maxSegmentDuration="PT%dS"`, jConf.SegmentDuration) + "\n"
    dashManifest += fmt.Sprintf(`minBufferTime="PT%dS"`, jConf.SegmentDuration + 1) + "\n"
    dashManifest += `profiles="urn:mpeg:dash:profile:isoff-live:2011">` + "\n"
    dashManifest += `  <Period>` + "\n"
    dashManifest += `    <BaseURL>./</BaseURL>` + "\n"

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

func HandleDashManifestRequest(w http.ResponseWriter, r *http.Request, dir string, filename string) {
    basename := util.Basename(filename)

    data, err := ioutil.ReadFile(path.Join(dir, basename + ".json"))

    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
        return
    }

    var jConfig mp4.JsonConfig
    err = json.Unmarshal(data, &jConfig)
    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
        return
    }

    manifest := createDashManifest(jConfig, basename)

    w.Header().Set("Content-Type", "application/dash+xml")
    w.Header().Set("Content-Length", strconv.Itoa(len(manifest)))
    w.Write([]byte(manifest))
}

func HandleDashContentRequest(w http.ResponseWriter, r *http.Request, dir string, filename string) {
    w.Header().Set("Content-Type", "video/mp4")
    
    fileBase, fileExt := util.SplitFilename(filename)
    trackName, trackType, trackLang, trackId, err := util.ParseBasename(fileBase)
    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
        return
    }

    trackIds := strings.Split(trackId, "-")
    var trackBandwidth uint64
    trackBandwidth, err = strconv.ParseUint(trackIds[0], 10, 64)
    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
        return
    }

    data, err := ioutil.ReadFile(path.Join(dir, trackName + ".json"))
    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
        return
    }

    var jConfig mp4.JsonConfig
    err = json.Unmarshal(data, &jConfig)
    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
        return
    }

    for _, t := range jConfig.Tracks[trackType] {
        if t.Name == trackName && t.Lang == trackLang && t.Bandwidth == trackBandwidth {
            t.File = "/" + t.File
            var dashData map[string][]interface{}
            switch fileExt {
                case ".dash":
                    dashData = mp4.CreateDashInitWithConf(*t.Config) // InitData
                case ".m4s":
                    if len(trackIds) !=2 {
                        http.Error(w, `{ "status": "ERROR", "reason": "Invalid track Id" }`, http.StatusInternalServerError)
                        logger.Error("Invalid track Id")
                        return
                    }
                    var num uint64
                    num, err = strconv.ParseUint(trackIds[1], 10, 32)
                    if err != nil {
                        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
                        logger.Error("%s", err.Error())
                        return
                    }
                    var segmentNumber uint32
                    segmentNumber = uint32(num)
                    dashData = mp4.CreateDashFragmentWithConf(*t.Config, t.File, segmentNumber, jConfig.SegmentDuration) // Fragment
                default :
                    http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
                    logger.Error("%s", err.Error())
                    return
            }
            b := mp4.MapToBytes(dashData) // InitData or Fragment
            w.Header().Set("Content-Length", strconv.Itoa(len(b)))
            _, err := w.Write(b)
            if err != nil {
                http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
                logger.Error("%s", err.Error())
                return
            }
            return
        }
    }

    http.Error(w, `{ "status": "ERROR", "reason": "File not found : ` + path.Join(dir, filename) + `" }`, http.StatusInternalServerError)
    logger.Error("File not found")
}

