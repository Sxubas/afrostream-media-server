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
    "encoding/json"
    "flag"
    "io/ioutil"
    "net/http"
    "path"
    "strconv"
    "strings"
    "syscall"

    "dash"
    "logger"
    "mp4"
    "util"
)

func handleHtmlRequest(w http.ResponseWriter, r *http.Request) {
    html, err := ioutil.ReadFile(r.URL.Path)
    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
        return
    }

    w.Header().Set("Content-Type", "text/html")
    w.Header().Set("Content-Length", strconv.Itoa(len(html)))

    _, err = w.Write(html)
    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
    }
}

func handleFileRequest(w http.ResponseWriter, path string) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
        return
    }

    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("Content-Length", strconv.Itoa(len(data)))

    _, err = w.Write(data)
    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
    }
}

func handleManifestRequest(w http.ResponseWriter, r *http.Request, dir string, basename string, extension string) {
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

    var manifest string
    switch extension {
        case ".mpd":
            manifest = dash.CreateDashManifest(jConfig, basename)
    }

    w.Header().Set("Content-Type", "application/dash+xml")
    w.Header().Set("Content-Length", strconv.Itoa(len(manifest)))
    _, err = w.Write([]byte(manifest))
    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
    }
}

func handleMediaRequest(w http.ResponseWriter, r *http.Request, dir string, basename string, extension string) {
    //w.Header().Set("Content-Type", "video/mp4")
    
    trackName, trackType, trackLang, trackId, err := util.ParseBasename(basename)
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
        if t.Lang == trackLang && t.Bandwidth == trackBandwidth {
            t.File = "/" + t.File

            var content map[string][]interface{}

            // DASH : InitData or Fragment
            // HLS  : Playlist or Fragment
            // Subititles

            switch extension {
                case ".dash":
                    content = mp4.CreateDashInitWithConf(*t.Config) // InitData
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
                    content = mp4.CreateDashFragmentWithConf(*t.Config, t.File, segmentNumber, jConfig.SegmentDuration) // Fragment
                case ".vtt":
                    handleFileRequest(w, t.File)
                    return
            }

            b := mp4.MapToBytes(content)
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

    http.Error(w, `{ "status": "ERROR", "reason": "Media not found : ` + path.Join(dir, basename + extension) + `" }`, http.StatusInternalServerError)
    logger.Error("Media not found : %s", path.Join(dir, basename + extension))
}

func handleContentRequest(w http.ResponseWriter, r *http.Request, dir string, basename string, extension string) {
    switch extension {
        case ".mpd":
            handleManifestRequest(w, r, dir, basename, extension)
        case ".dash":
            handleMediaRequest(w, r, dir, basename, extension)
        case ".m4s":
            handleMediaRequest(w, r, dir, basename, extension)

        case ".m3u8":
            http.Error(w, `{ "status": "ERROR", "reason": "Not implemented yet" }`, http.StatusNotImplemented)
            logger.Error("Not implemented yet")
        case ".hls":
            http.Error(w, `{ "status": "ERROR", "reason": "Not implemented yet" }`, http.StatusNotImplemented)
            logger.Error("Not implemented yet")
        case ".ts":
            http.Error(w, `{ "status": "ERROR", "reason": "Not implemented yet" }`, http.StatusNotImplemented)
            logger.Error("Not implemented yet")

        case ".vtt":
            handleMediaRequest(w, r, dir, basename, extension)

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
            handleContentRequest(w, r, dir[6:], basename, extension) // Remove relative path /video/ -> /
        default:
            handleFileRequest(w, r.URL.Path)
    }
}

func help() {
    logger.Message("Afrostream Media Server version 0.1     Sebastien Petit <spebsd@gmail.com>")
    logger.Message("Usage: ams -d [directory] < -p [port] -log [filename] >")
    logger.Message("  < ... > are optional\n")
    flag.PrintDefaults()
    logger.Message("\nExample: amspackager -d public_html -p 80")
}

func main() {
    flag.Usage = help

    var flag_help bool
    flag.BoolVar(&flag_help, "help", false, "Display this help and exit")

    var logfile string
    flag.StringVar(&logfile, "log", "", "Log `filename`, stdout by default")

    var directory string
    flag.StringVar(&directory, "d", "", "Root `directory` of AMS web server")

    var port string
    flag.StringVar(&port, "p", "80", "Listening `port` of AMS web server")

    flag.Parse()

    if flag_help {
        help()
        return
    }

    logger.Init(logfile, logger.F_Debug)

    if directory == "" {
        logger.Message("Please specify the root directory for AMS web server\n")
        help()
        return
    }

    err := syscall.Chroot(directory)
    if err != nil {
        logger.Message("Please run Afrostream Media Server as root, cannot chroot the root directory for security: %v", err)
        return
    }

    logger.Message("[*] Running Afrostream Media Server on port %s, press CTRL+C to exit", port)

    http.HandleFunc("/", handleHttpRequest)
    logger.Error("%s", http.ListenAndServe(":" + port, nil))

    return
}

