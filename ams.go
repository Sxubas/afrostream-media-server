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
    "flag"
    "io/ioutil"
    "net/http"
    "path"
    "strconv"
    "strings"
    "syscall"

    "dash"
    "logger"
)

func handleFileRequest(w http.ResponseWriter, r *http.Request, dir string, filename string) {
    html, err := ioutil.ReadFile(path.Join(dir, filename)) // dir + filename may be different from r.URL.Path
    if err != nil {
        http.Error(w, `{ "status": "ERROR", "reason": "` + err.Error() + `" }`, http.StatusInternalServerError)
        logger.Error("%s", err.Error())
        return
    }

    switch path.Ext(filename) {
        case ".html":
            w.Header().Set("Content-Type", "text/html")
        default:
            w.Header().Set("Content-Type", "application/octet-stream")
    }

    w.Header().Set("Content-Length", strconv.Itoa(len(html)))
    w.Write(html)
}

func handleContentRequest(w http.ResponseWriter, r *http.Request, dir string, filename string) {
    switch path.Ext(filename) {
        case ".mpd":
            dash.HandleDashManifestRequest(w, r, dir, filename)
            return
        case ".dash":
            dash.HandleDashContentRequest(w, r, dir, filename)
            return
        case ".m4s":
            dash.HandleDashContentRequest(w, r, dir, filename)
            return

        case ".m3u8":
            http.Error(w, `{ "status": "ERROR", "reason": "Not implemented yet" }`, http.StatusNotImplemented)
            logger.Error("Not implemented yet")
            return
        case ".hls":
            http.Error(w, `{ "status": "ERROR", "reason": "Not implemented yet" }`, http.StatusNotImplemented)
            logger.Error("Not implemented yet")
            return
        case ".ts":
            http.Error(w, `{ "status": "ERROR", "reason": "Not implemented yet" }`, http.StatusNotImplemented)
            logger.Error("Not implemented yet")
            return

        case ".vtt":
            handleFileRequest(w, r, dir, filename)
        default:
            http.Error(w, `{ "status": "ERROR", "reason": "Format is not supported" }`, http.StatusInternalServerError)
            logger.Error("Format is not supported")
            return
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

    paths := strings.Split(dir[1:], "/")

    // Switch between video (manifest, video, audio, subtitles) request and other request type

    switch paths[0] {
        case "video":
            handleContentRequest(w, r, dir[6:], filename) // Remove relative path /video/ -> /
        default:
            handleFileRequest(w, r, dir, filename)
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

