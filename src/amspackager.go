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
	"errors"
	"flag"
	"fmt"
	"mp4"
	"os"
	"path"
)

type fileSlice []string
type languageSlice []string

type inputFile struct {
	Filename string
	Language string
}

// Global vars for Flags
var inputFilenames fileSlice
var languageCodes languageSlice

func (s *fileSlice) String() string {
	return fmt.Sprintf("%+v", *s)
}

func (s *fileSlice) Set(value string) error {
	*s = append(*s, value)

	return nil
}

func (s *languageSlice) String() string {
	return fmt.Sprintf("%+v", *s)
}

func (s *languageSlice) Set(value string) error {
	if inputFilenames == nil {
		return errors.New("no input filenames specified before -l option")
	}
	if len(value) != 3 {
		return errors.New("ISO-639-2 language code is 3 character size (eg: eng)")
	}
	for len(*s) < len(inputFilenames)-1 {
		*s = append(*s, "")
	}
	*s = append(*s, value)

	return nil
}

func parseMp4Files(files []inputFile) (mp4Files map[string][]mp4.Mp4) {
	mp4Files = make(map[string][]mp4.Mp4)
	for _, in := range files {
		fmt.Printf("-- Parsing file='%s' language='%s'\n", in.Filename, in.Language)
		mp4File := mp4.ParseFile(in.Filename, in.Language)
		if mp4File.IsVideo == true {
			mp4Files["video"] = append(mp4Files["video"], mp4File)
		}
		if mp4File.IsAudio == true {
			mp4Files["audio"] = append(mp4Files["audio"], mp4File)
		}
	}

	return
}


func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Afrostream Media Server version 0.1     Sebastien Petit <spebsd@gmail.com>\n")
		fmt.Printf("Usage: amspackager -o [json output filename] < -d [segment duration] > { -i [mp4 or vtt input file] < -l [language] > ... }\n")
		fmt.Printf("  < ... > options are optional\n")
		fmt.Printf("  [mp4 or vtt input file]     must be audio mp4a / video avc1 / vtt subtitles files\n")
		fmt.Printf("                              only one stream per mp4 file is supported\n")
		fmt.Printf("  -d [segment duration]       duration of each segments in seconds\n")
		fmt.Printf("                              default value: 10\n")
		fmt.Printf("  -l [language]               ISO-639-2 language code for the input file preceeding this argument\n")
		fmt.Printf("\n")
		fmt.Printf("Example: amspackager -o video.json -d 8 -i video-384k.mp4 -i video-1500k.mp4 -i video-2950k.mp4 -i audio-128k.mp4 -i sub_fr.vtt -l fra -i sub_en.vtt -l eng\n")

		return
	}

	jsonFilename := flag.String("o", "video.json", "JSON output filename (default: video.json)")
	segmentDuration := flag.Uint("d", 10, "segment duration (default: 10)")
	flag.Var(&inputFilenames, "i", "MP4 or VTT input filename")
	flag.Var(&languageCodes, "l", "ISO-639-2 language code")
	flag.Parse()

	var mp4FileSlice []inputFile
	var vttFileSlice []inputFile
	for i, inputFilename := range inputFilenames {
		switch path.Ext(inputFilename) {
		case ".mp4":
			var in inputFile
			in.Filename = inputFilename
			if i < len(languageCodes) && languageCodes[i] != "" {
				in.Language = languageCodes[i]
			} else {
				in.Language = "eng"
			}
			mp4FileSlice = append(mp4FileSlice, in)
		case ".vtt":
			var in inputFile
			in.Filename = inputFilename
			if i < len(languageCodes) && languageCodes[i] != "" {
				in.Language = languageCodes[i]
			} else {
				in.Language = "eng"
			}
			vttFileSlice = append(vttFileSlice, in)
		default:
			fmt.Printf("Sorry, but the file %s is unkwown and can't be packaged. Please use .mp4 or .vtt extensions for your files\n", inputFilename)
		}
	}

	fmt.Printf("AMSPackager -- spebsd@gmail.com / Afrostream\n\n")

	mp4Files := parseMp4Files(mp4FileSlice)

	var jConf mp4.JsonConfig
	//jConf.Tracks = make([][]mp4.TrackEntry, 2)
	jConf.Tracks = make(map[string][]mp4.TrackEntry)
	jConf.SegmentDuration = uint32(*segmentDuration)

	for _, mp4File := range mp4Files["video"] {
		mdat := mp4File.Boxes["mdat"][0].(mp4.MdatBox)
		mdhd := mp4File.Boxes["moov.trak.mdia.mdhd"][0].(mp4.MdhdBox)
		hdlr := mp4File.Boxes["moov.trak.mdia.hdlr"][0].(mp4.HdlrBox)
		stts := mp4File.Boxes["moov.trak.mdia.minf.stbl.stts"][0].(mp4.SttsBox)
		var ctts mp4.CttsBox
		cttsBoxPresent := false
		if mp4File.Boxes["moov.trak.mdia.minf.stbl.ctts"] != nil {
			ctts = mp4File.Boxes["moov.trak.mdia.minf.stbl.ctts"][0].(mp4.CttsBox)
			cttsBoxPresent = true
		}
		stsz := mp4File.Boxes["moov.trak.mdia.minf.stbl.stsz"][0].(mp4.StszBox)
		stss := mp4File.Boxes["moov.trak.mdia.minf.stbl.stss"][0].(mp4.StssBox)
		avc1 := mp4File.Boxes["moov.trak.mdia.minf.stbl.stsd.avc1"][0].(mp4.Avc1Box)
		avcC := mp4File.Boxes["moov.trak.mdia.minf.stbl.stsd.avc1.avcC"][0].(mp4.AvcCBox)
		elst := mp4File.Boxes["moov.trak.edts.elst"][0].(mp4.ElstBox)
		var t mp4.TrackEntry
		t.Bandwidth = uint64(float64(mdat.Size) / (float64(mdhd.Duration) / float64(mdhd.Timescale)) * 8)
		t.Name = "video_" + mp4File.Language
		t.File = mp4File.Filename
		t.Lang = mp4File.Language
		t.Config = new(mp4.Conf)
		t.Config.StszBoxOffset = stsz.Offset
		t.Config.StszBoxSize = stsz.Size
		t.Config.MdatBoxOffset = mdat.Offset
		t.Config.MdatBoxSize = mdat.Size
		t.Config.Type = "video"
		t.Config.Rate = 0x00010000
		t.Config.Volume = 0x0100
		t.Config.Duration = mdhd.Duration
		t.Config.Timescale = mdhd.Timescale
		t.Config.Language[0] = byte((0x7c00&mdhd.Language)>>10) + 0x60
		t.Config.Language[1] = byte((0x03e0&mdhd.Language)>>5) + 0x60
		t.Config.Language[2] = byte(0x1f&mdhd.Language) + 0x60
		t.Config.HandlerType = hdlr.HandlerType
		t.Config.SampleDelta = stts.Entries[0].SampleDelta
		t.Config.MediaTime = elst.MediaTime
		t.Config.Video = new(mp4.DashVideoEntry)
		t.Config.Video.Width = avc1.Width
		t.Config.Video.Height = avc1.Height
		t.Config.Video.HorizontalResolution = avc1.HorizontalResolution
		t.Config.Video.VerticalResolution = avc1.VerticalResolution
		t.Config.Video.EntryDataSize = avc1.EntryDataSize
		t.Config.Video.FramesPerSample = avc1.FramesPerSample
		t.Config.Video.BitDepth = avc1.BitDepth
		t.Config.Video.ColorTableIndex = avc1.ColorTableIndex
		t.Config.Video.CodecInfo = [3]byte{avcC.AVCProfileIndication, avcC.ProfileCompatibility, avcC.AVCLevelIndication}
		t.Config.Video.NalUnitSize = avcC.NalUnitSize & 0x03
		t.Config.Video.SPSEntryCount = avcC.SPSEntryCount
		t.Config.Video.SPSSize = avcC.SPSSize
		t.Config.Video.SPSData = avcC.SPSData
		t.Config.Video.PPSEntryCount = avcC.PPSEntryCount
		t.Config.Video.PPSSize = avcC.PPSSize
		t.Config.Video.PPSData = avcC.PPSData
		t.Config.Video.StssBoxOffset = stss.Offset
		t.Config.Video.StssBoxSize = stss.Size
		if cttsBoxPresent == true {
			t.Config.Video.CttsBoxOffset = ctts.Offset
			t.Config.Video.CttsBoxSize = ctts.Size
		}
		jConf.Tracks["video"] = append(jConf.Tracks["video"], t)
	}

	for _, mp4File := range mp4Files["audio"] {
		mdat := mp4File.Boxes["mdat"][0].(mp4.MdatBox)
		mdhd := mp4File.Boxes["moov.trak.mdia.mdhd"][0].(mp4.MdhdBox)
		hdlr := mp4File.Boxes["moov.trak.mdia.hdlr"][0].(mp4.HdlrBox)
		stts := mp4File.Boxes["moov.trak.mdia.minf.stbl.stts"][0].(mp4.SttsBox)
		stsz := mp4File.Boxes["moov.trak.mdia.minf.stbl.stsz"][0].(mp4.StszBox)
		mp4a := mp4File.Boxes["moov.trak.mdia.minf.stbl.stsd.mp4a"][0].(mp4.Mp4aBox)
		elst := mp4File.Boxes["moov.trak.edts.elst"][0].(mp4.ElstBox)
		var t mp4.TrackEntry
		t.Bandwidth = uint64(float64(mdat.Size) / (float64(mdhd.Duration) / float64(mdhd.Timescale)) * 8)
		t.Name = "audio_" + mp4File.Language
		t.File = mp4File.Filename
		t.Lang = mp4File.Language
		t.Config = new(mp4.Conf)
		t.Config.StszBoxOffset = stsz.Offset
		t.Config.StszBoxSize = stsz.Size
		t.Config.MdatBoxOffset = mdat.Offset
		t.Config.MdatBoxSize = mdat.Size
		t.Config.Type = "audio"
		t.Config.Rate = 0x00010000
		t.Config.Volume = 0x0100
		t.Config.Duration = mdhd.Duration
		t.Config.Timescale = mdhd.Timescale
		t.Config.Language[0] = byte((0x7c00&mdhd.Language)>>10) + 0x60
		t.Config.Language[1] = byte((0x03e0&mdhd.Language)>>5) + 0x60
		t.Config.Language[2] = byte(0x1f&mdhd.Language) + 0x60
		t.Config.HandlerType = hdlr.HandlerType
		t.Config.SampleDelta = stts.Entries[0].SampleDelta
		t.Config.MediaTime = elst.MediaTime
		t.Config.Audio = new(mp4.DashAudioEntry)
		t.Config.Audio.NumberOfChannels = mp4a.NumberOfChannels
		t.Config.Audio.SampleSize = mp4a.SampleSize
		t.Config.Audio.CompressionId = mp4a.CompressionId
		t.Config.Audio.SampleRate = mp4a.SampleRate
		jConf.Tracks["audio"] = append(jConf.Tracks["audio"], t)
	}

	for _, vttFile := range vttFileSlice {
		var t mp4.TrackEntry
		t.Bandwidth = 256
		t.Name = "caption_" + vttFile.Language
		t.File = vttFile.Filename
		t.Lang = vttFile.Language
		jConf.Tracks["subtitle"] = append(jConf.Tracks["subtitle"], t)
	}

	jsonStr, err := json.Marshal(jConf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n-- Creating package file '%s'\n", *jsonFilename)
	f, err := os.Create(*jsonFilename)
	if err != nil {
		fmt.Printf("Cannot open filename '%s': %v", jsonFilename, err)
		return
	}
	defer f.Close()

	f.WriteString(string(jsonStr))

	fmt.Printf("\nAll files has been packaged successfully\n")

	return
}
