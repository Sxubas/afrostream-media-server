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

package util

import (
	"errors"
	"path"
	"strings"

	"github.com/Sxubas/afrostream-media-server/src/mp4"
)

func SplitFilename(filename string) (string, string) {
	ext := path.Ext(filename)
	return filename[:len(filename)-len(ext)], ext
}

func ParseBasename(filename string) (string, string, string, string, error) {
	splitted := strings.Split(filename, "_")

	l := len(splitted)
	if l < 4 {
		return "", "", "", "", errors.New("Invalid filename format")
	}

	return strings.Join(splitted[:len(splitted)-3], "_"), splitted[l-3], splitted[l-2], splitted[l-1], nil
}

func NumberOfSegments(track mp4.TrackEntry, jConfig mp4.JsonConfig) uint32 {
	numberOfSegments := uint32(track.Config.Duration) / uint32(jConfig.SegmentDuration*track.Config.Timescale)
	if int(track.Config.Duration)%int(jConfig.SegmentDuration*track.Config.Timescale) != 0 {
		numberOfSegments++
	}
	return numberOfSegments
}
