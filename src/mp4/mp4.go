// Copyright (c) 2015
//			Sebastien Petit & Afrostream - www.afrostream.tv - spebsd@gmail.com.
//			All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
// 1. Redistributions of source code must retain the above copyright
//		notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright
//		notice, this list of conditions and the following disclaimer in the
//		documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
//		may be used to endorse or promote products derived from this software
//		without specific prior written permission.
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

package mp4

import (
	"encoding/binary"
	"log"
	"os"
	"reflect"
	"strings"
)

var debugMode bool
var funcBoxes map[string]interface{}

type JsonConfig struct {
	SegmentDuration uint32
	Tracks          map[string][]TrackEntry
}

type TrackEntry struct {
	Name      string
	Bandwidth uint64
	File      string
	Lang      string
	Config    *DashConfig `json:",omitempty"`
}

type DashAudioEntry struct {
	// Sound fields if Type == "audio"
	NumberOfChannels uint16 // MP4A MP4 Box Info (eg: 2)
	SampleSize       uint16 // MP4A MP4 Box Info (eg: 16)
	CompressionId    uint16 // MP4A MP4 Box Info (eg: 0)
	SampleRate       uint32 // MP4A MP4 Box Info (eg: 3145728000)
}

type DashVideoEntry struct {
	// Video fields if Type == "video"
	Width                uint16  // AVC1 MP4 Box info (eg: 426)
	Height               uint16  // AVC1 MP4 Box info (eg: 240)
	HorizontalResolution uint32  // AVC1 MP4 Box info (eg: 4718592)
	VerticalResolution   uint32  // AVC1 MP4 Box info (eg: 4718592)
	EntryDataSize        uint32  // AVC1 MP4 Box info (eg: 0)
	FramesPerSample      uint16  // AVC1 MP4 Box info (eg: 1)
	BitDepth             uint16  // AVC1 MP4 Box info (eg: 24)
	ColorTableIndex      int16   // AVC1 MP4 Box info (eg: -1)
	CodecInfo            [3]byte // AVCC MP4 Box AVC Profile/Comptaiblity/Level Information (eg: []byte{ 0x42, 0xC0, 0x1E }
	NalUnitSize          byte    // AVCC MP4 Box info NALUnitLength field. upper 6-bits are reserved as 111111b aka | 0xf6 (eg: 0xFF)
	SPSEntryCount        uint8   // AVCC MP4 Box info (eg: 1)
	SPSSize              uint16  // AVCC MP4 Box info (eg: 23)
	SPSData              []byte  // AVCC MP4 Box info (eg: [103 66 192 30 219 2 128 191 229 192 68 0 0 15 164 0 7 83 0 60 88 187 128])
	PPSEntryCount        uint8   // AVCC MP4 Box info (eg: 1)
	PPSSize              uint16  // AVCC MP4 Box info (eg: 4)
	PPSData              []byte  // AVCC MP4 Box info (eg: 104 202 140 178)
	StssBoxOffset        int64
	StssBoxSize          uint32
	CttsBoxOffset        int64
	CttsBoxSize          uint32
}

type DashConfig struct {
	StszBoxOffset int64
	StszBoxSize   uint32
	MdatBoxOffset int64
	MdatBoxSize   uint32  // MDAT MP4 Box Size
	Type          string  // "audio" || "video
	Rate          int32   // Typically 0x00010000 (1.0)
	Volume        int16   // Typically 0x0100 (Full Volume)
	Duration      uint64  // MDHD MP4 Box info
	Timescale     uint32  // MDHD MP4 Box info (eg: for audio: 48000, for video: 60000)
	Language      [3]byte // ISO-639-2/T 3 letters code (eg: []byte{ 'e', 'n', 'g' }
	HandlerType   uint32  // HDLR MP4 Box info (eg: 1986618469)
	SampleDelta   uint32  // STTS MP4 Box SampleDelta via Entries[0] (eg: 1024)
	MediaTime     int64   // ELST MP4 Box MediaTime

	Audio *DashAudioEntry `json:",omitempty"`
	Video *DashVideoEntry `json:",omitempty"`
}

type Mp4 struct {
	Filename string
	Language string
	IsVideo  bool
	IsAudio  bool
	Boxes    map[string][]interface{}
}

type ParentBox struct {
	Name [4]byte
	Size uint32
}

type FtypBox struct {
	Size             uint32
	MajorBrand       [4]byte
	MinorVersion     uint32
	CompatibleBrands [][4]byte
}

type StypBox struct {
	Size             uint32
	MajorBrand       [4]byte
	MinorVersion     uint32
	CompatibleBrands [][4]byte
}

type FreeBox struct {
	Size uint32
	Data []byte
}

type MvhdBox struct {
	Size             uint32
	Version          byte
	Reserved         [3]byte
	CreationTime     uint64
	ModificationTime uint64
	Timescale        uint32
	Duration         uint64
	Rate             int32
	Volume           int16
	Reserved2        uint16
	Reserved3        uint64
	Matrix           [9]int32
	PreDefined       [6]uint32
	NextTrackID      uint32
}

type TkhdBox struct {
	Size             uint32
	Version          byte
	Flags            [3]byte
	CreationTime     uint64
	ModificationTime uint64
	TrackID          uint32
	Reserved         uint32
	Duration         uint64
	Reserved2        uint64
	Layer            int16
	AlternateGroup   int16
	Volume           int16
	Reserved3        uint16
	Matrix           [9]int32
	Width            uint32
	Height           uint32
}

type ElstBox struct {
	Size              uint32
	Version           byte
	Reserved          [3]byte
	EntryCount        uint32
	SegmentDuration   uint64
	MediaTime         int64
	MediaRateInteger  int16
	MediaRateFraction int16
}

type MdhdBox struct {
	Size             uint32
	Version          byte
	Reserved         [3]byte
	CreationTime     uint64
	ModificationTime uint64
	Timescale        uint32
	Duration         uint64
	Language         uint16
	PreDefined       uint16
}

type HdlrBox struct {
	Size        uint32
	Version     byte
	Reserved    [3]byte
	PreDefined  uint32
	HandlerType uint32
	Reserved2   [3]uint32
	Name        []byte
}

type VmhdBox struct {
	Size         uint32
	Version      byte
	Reserved     [3]byte
	GraphicsMode uint16
	OpColor      [3]uint16
}

type SmhdBox struct {
	Size      uint32
	Version   byte
	Reserved  [3]byte
	Balance   int16
	Reserved2 uint16
}

type HmhdBox struct {
	Version    byte
	Reserved   [3]byte
	MaxPDUSize uint16
	AvgPDUSize uint16
	MaxBitrate uint32
	AvgBitrate uint32
	Reserved2  uint32
}

type DrefBox struct {
	Size       uint32
	Version    byte
	Reserved   [3]byte
	EntryCount uint32
	UrlBox     []DrefUrlBox
	UrnBox     []DrefUrnBox
}

type DrefUrlBox struct {
	Size     uint32
	Location string
	Version  byte
	Flags    [3]byte
}

type DrefUrnBox struct {
	Size     uint32
	Name     string
	Location string
	Version  byte
	Flags    [3]byte
}

type SttsBox struct {
	Size       uint32
	Version    byte
	Reserved   [3]byte
	EntryCount uint32
	Entries    []SttsBoxEntry
}

type SttsBoxEntry struct {
	SampleCount uint32
	SampleDelta uint32
}

type CttsBox struct {
	Size       uint32
	Offset     int64
	Version    byte
	Reserved   [3]byte
	EntryCount uint32
	Entries    []CttsBoxEntry
}

type CttsBoxEntry struct {
	SampleCount  uint32
	SampleOffset uint32
}

type StssBox struct {
	Size         uint32
	Offset       int64
	Version      byte
	Reserved     [3]byte
	EntryCount   uint32
	SampleNumber []uint32
}

type MehdBox struct {
	Size             uint32
	Version          byte
	Reserved         [3]byte
	FragmentDuration uint64
}

type TrexBox struct {
	Size                          uint32
	Version                       byte
	Reserved                      [3]byte
	TrackID                       uint32
	DefaultSampleDescriptionIndex uint32
	DefaultSampleDuration         uint32
	DefaultSampleSize             uint32
	DefaultSampleFlags            uint32
}

type StsdBox struct {
	Size       uint32
	Version    byte
	Reserved   [3]byte
	EntryCount uint32
}

type Mp4aBox struct {
	Size               uint32
	Reserved           [6]byte
	DataReferenceIndex uint16
	Version            uint16
	RevisionLevel      uint16
	Vendor             uint32
	NumberOfChannels   uint16
	SampleSize         uint16
	CompressionId      uint16
	Reserved2          uint16
	SampleRate         uint32
}

type Avc1Box struct {
	Size                 uint32
	Reserved             [6]byte
	dataReferenceIndex   uint16
	Version              uint16
	RevisionLevel        uint16
	Vendor               uint32
	TemporalQuality      uint32
	SpacialQuality       uint32
	Width                uint16
	Height               uint16
	HorizontalResolution uint32
	VerticalResolution   uint32
	EntryDataSize        uint32
	FramesPerSample      uint16
	CompressorName       [32]byte
	BitDepth             uint16
	ColorTableIndex      int16
}

type AvcCBox struct {
	Size                 uint32
	ConfigurationVersion uint8 /* 1 */
	AVCProfileIndication uint8 /* profile idc in SPS */
	ProfileCompatibility uint8
	AVCLevelIndication   uint8  /* level idc in SPS */
	NalUnitSize          uint8  /* in bytes of the NALUnitLength field. upper 6-bits are reserved as 111111b aka | 0xf6 */
	SPSEntryCount        uint8  /* Number of Sequence Parameter Set Entries */
	SPSSize              uint16 /* Sequence Parameter Set Size upper 3-bits are reserved as 111b aka | 0xe0 */
	SPSData              []byte /* Sequence Parameter Set Datas */
	PPSEntryCount        uint8  /* Number of Picture Parameter Set Entries */
	PPSSize              uint16 /* Picture Parameter Set Size */
	PPSData              []byte /* Picture Parameter Set Datas */
}

/* MPEG-4 Bit Rate Box
 * This box signals the bit rate information of the AVC video stream. */
type BtrtBox struct {
	Size               uint32
	DecodingBufferSize uint32 /* the size of the decoding buffer for the elementary stream in bytes */
	MaxBitrate         uint32 /* the maximum rate in bits/second over any window of one second */
	AvgBitrate         uint32 /* the average rate in bits/second over the entire presentation */
}

type EsdsBox struct {
	Size    uint32
	Version uint32
	Data    []byte /* Unkown for the moment ??? */
}

type StscBox struct {
	Size       uint32
	Version    byte
	Flags      [3]byte
	EntryCount uint32
	Entries    []StscEntry
}

type StscEntry struct {
	FirstChunk             uint32
	SamplesPerChunk        uint32
	SampleDescriptionIndex uint32
}

type StszBox struct {
	Size        uint32
	Offset      int64
	Version     byte
	Reserved    [3]byte
	SampleSize  uint32
	SampleCount uint32
	EntrySize   []uint32
}

type SdtpBox struct {
	Size        uint32
	Version     byte
	SampleCount uint32
	Entries     []uint8
}

type StcoBox struct {
	Size        uint32
	Version     byte
	Reserved    [3]byte
	EntryCount  uint32
	ChunkOffset []uint32
}

/* MOOF SubBoxes */
type MfhdBox struct {
	Size           uint32
	Version        byte
	Reserved       [3]byte
	SequenceNumber uint32
}

type TfhdBox struct {
	Size    uint32
	Version byte
	Flags   [3]byte
	TrackID uint32
	// All of the following are optional fields
	BaseDataOffset         uint64
	SampleDescriptionIndex uint32
	DefaultSampleDuration  uint32
	DefaultSampleSize      uint32
	DefaultSampleFlags     uint32
}

type TrunBox struct {
	Size        uint32
	Version     byte
	Flags       [3]byte
	SampleCount uint32
	// All of the following are optional fields
	DataOffset       int32
	FirstSampleFlags uint32
	Samples          []TrunBoxSample
}

type TrunBoxSample struct {
	Duration              uint32
	Size                  uint32
	Flags                 uint32
	CompositionTimeOffset int64
}

type TfdtBox struct {
	Size                uint32
	Version             byte // Must be 1
	Reserved            [3]byte
	BaseMediaDecodeTime uint64
}

type FrmaBox struct {
	Size       uint32
	DataFormat [4]byte
}

type SchmBox struct {
	Size          uint32
	Version       byte
	Flags         [3]byte
	SchemeType    [4]byte
	SchemeVersion uint32
	SchemeUri     string
}

type MdatBox struct {
	Size     uint32
	Filename string
	Offset   int64
}

// ***
// *** Private functions
// ***

// Dump a box structure if debugMode is true
func dumpBox(boxPath string, box interface{}) {
	if debugMode {
		log.Printf("[ %s Box data ] %+v", boxPath, box)
	}
}

// Add a decoded box to the mp4 hashtable structure
// It's more easy to read box after
func addBox(mp4 map[string][]interface{}, boxPath string, box interface{}) {
	if mp4[boxPath] == nil {
		mp4[boxPath] = make([]interface{}, 1)
		mp4[boxPath][0] = box
	} else {
		mp4[boxPath] = append(mp4[boxPath], box)
	}
}

func replaceBox(mp4 map[string][]interface{}, boxPath string, box interface{}) {
	if mp4[boxPath] != nil {
		delete(mp4, boxPath)
	}
	mp4[boxPath] = make([]interface{}, 1)
	mp4[boxPath][0] = box
}

func (parent ParentBox) Bytes() (data []byte) {
	data = make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], parent.Size+8)
	copy(data[4:8], parent.Name[:])

	return
}

// Decode FTYP Box
func readFtypBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var ftyp FtypBox
	ftyp.Size = size
	copy(ftyp.MajorBrand[:], data[0:4])
	ftyp.MinorVersion = binary.BigEndian.Uint32(data[4:8])
	var entryCount uint32
	entryCount = (size - 8) / 4
	ftyp.CompatibleBrands = make([][4]byte, entryCount)
	var i uint32
	for i = 0; i < entryCount; i++ {
		copy(ftyp.CompatibleBrands[i][:], data[8+(i*4):12+(i*4)])
	}
	addBox(mp4, boxPath, ftyp)
	dumpBox(boxPath, ftyp)
}

func (ftyp FtypBox) Bytes() (data []byte) {
	boxSize := ftyp.Size + 8
	data = make([]byte, boxSize)
	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'f', 't', 'y', 'p'})
	copy(data[8:12], ftyp.MajorBrand[:])
	binary.BigEndian.PutUint32(data[12:16], ftyp.MinorVersion)
	i := 0
	for _, v := range ftyp.CompatibleBrands {
		copy(data[16+(i*4):20+(i*4)], v[:])
		i++
	}

	return
}

func readStypBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var styp StypBox
	styp.Size = size
	copy(styp.MajorBrand[:], data[0:4])
	styp.MinorVersion = binary.BigEndian.Uint32(data[4:8])
	var entryCount uint32
	entryCount = (size - 8) / 4
	styp.CompatibleBrands = make([][4]byte, entryCount)
	var i uint32
	for i = 0; i < entryCount; i++ {
		copy(styp.CompatibleBrands[i][:], data[8+(i*4):12+(i*4)])
	}
	addBox(mp4, boxPath, styp)
	dumpBox(boxPath, styp)
}

func (styp StypBox) Bytes() (data []byte) {
	boxSize := styp.Size + 8
	data = make([]byte, boxSize)
	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'s', 't', 'y', 'p'})
	copy(data[8:12], styp.MajorBrand[:])
	binary.BigEndian.PutUint32(data[12:16], styp.MinorVersion)
	i := 0
	for _, v := range styp.CompatibleBrands {
		copy(data[16+(i*4):20+(i*4)], v[:])
		i++
	}

	return
}

// Decode FREE Box
func readFreeBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var free FreeBox
	free.Size = size
	free.Data = make([]byte, size)
	_, err := f.Read(free.Data)
	if err != nil {
		panic(err)
	}
	addBox(mp4, boxPath, free)
	dumpBox(boxPath, free)
}

func (free FreeBox) Bytes() (data []byte) {
	boxSize := free.Size + 8
	data = make([]byte, boxSize)
	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'f', 'r', 'e', 'e'})
	copy(data[8:], free.Data[:])

	return
}

func readTkhdBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var tkhd TkhdBox
	var offset uint32
	offset = 4
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	tkhd.Size = size
	tkhd.Version = data[0]
	if tkhd.Version != 0 && tkhd.Version != 1 {
		if debugMode {
			log.Printf("ERROR: Unknown %s box version", boxPath)
		}
		return
	}
	copy(tkhd.Flags[:], data[1:4])
	if tkhd.Version == 0 {
		tkhd.CreationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
		tkhd.ModificationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
	} else {
		tkhd.CreationTime = binary.BigEndian.Uint64(data[offset : offset+8])
		offset += 8
		tkhd.ModificationTime = binary.BigEndian.Uint64(data[offset : offset+8])
		offset += 8
	}
	tkhd.TrackID = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	tkhd.Reserved = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	if tkhd.Version == 0 {
		tkhd.Duration = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
	} else {
		tkhd.Duration = binary.BigEndian.Uint64(data[offset : offset+8])
		offset += 8
	}
	tkhd.Reserved2 = binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8
	tkhd.Layer = int16(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	tkhd.AlternateGroup = int16(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	tkhd.Volume = int16(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	tkhd.Reserved3 = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	for i := 0; i < 9; i++ {
		tkhd.Matrix[i] = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
	}
	tkhd.Width = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	tkhd.Height = binary.BigEndian.Uint32(data[offset : offset+4])
	addBox(mp4, boxPath, tkhd)
	dumpBox(boxPath, tkhd)
}

func (tkhd TkhdBox) Bytes() (data []byte) {
	var offset uint32
	offset = 12
	boxSize := tkhd.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'t', 'k', 'h', 'd'})
	data[8] = tkhd.Version
	copy(data[9:12], tkhd.Flags[:])
	if tkhd.Version == 0 {
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(tkhd.CreationTime))
		offset += 4
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(tkhd.ModificationTime))
		offset += 4
	} else {
		binary.BigEndian.PutUint64(data[offset:offset+8], tkhd.CreationTime)
		offset += 8
		binary.BigEndian.PutUint64(data[offset:offset+8], tkhd.ModificationTime)
		offset += 8
	}
	binary.BigEndian.PutUint32(data[offset:offset+4], tkhd.TrackID)
	offset += 4
	binary.BigEndian.PutUint32(data[offset:offset+4], tkhd.Reserved)
	offset += 4
	if tkhd.Version == 0 {
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(tkhd.Duration))
		offset += 4
	} else {
		binary.BigEndian.PutUint64(data[offset:offset+8], tkhd.Duration)
		offset += 8
	}
	binary.BigEndian.PutUint64(data[offset:offset+8], tkhd.Reserved2)
	offset += 8
	binary.BigEndian.PutUint16(data[offset:offset+2], uint16(tkhd.Layer))
	offset += 2
	binary.BigEndian.PutUint16(data[offset:offset+2], uint16(tkhd.AlternateGroup))
	offset += 2
	binary.BigEndian.PutUint16(data[offset:offset+2], uint16(tkhd.Volume))
	offset += 2
	binary.BigEndian.PutUint16(data[offset:offset+2], tkhd.Reserved3)
	offset += 2
	for i := 0; i < 9; i++ {
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(tkhd.Matrix[i]))
		offset += 4
	}
	binary.BigEndian.PutUint32(data[offset:offset+4], tkhd.Width)
	offset += 4
	binary.BigEndian.PutUint32(data[offset:offset+4], tkhd.Height)

	return
}

func readElstBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var offset uint32
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	var elst ElstBox
	elst.Version = data[0]
	if elst.Version != 0 && elst.Version != 1 {
		if debugMode {
			log.Printf("ERROR: Unknown %s box version", boxPath)
		}
		return
	}
	copy(elst.Reserved[:], data[1:4])
	elst.EntryCount = binary.BigEndian.Uint32(data[4:8])
	offset = 8
	var i uint32
	for i = 0; i < elst.EntryCount; i++ {
		if elst.Version == 0 {
			elst.SegmentDuration = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
			offset += 4
			elst.MediaTime = int64(binary.BigEndian.Uint32(data[offset : offset+4]))
			offset += 4
		} else {
			elst.SegmentDuration = binary.BigEndian.Uint64(data[offset : offset+8])
			offset += 8
			elst.MediaTime = int64(binary.BigEndian.Uint64(data[offset : offset+8]))
			offset += 8
		}
		elst.MediaRateInteger = int16(binary.BigEndian.Uint16(data[offset : offset+2]))
		offset += 2
		elst.MediaRateFraction = int16(binary.BigEndian.Uint16(data[offset : offset+2]))
		offset += 2
	}
	addBox(mp4, boxPath, elst)
	dumpBox(boxPath, elst)
}

func readMdhdBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var mdhd MdhdBox
	var offset uint32
	offset = 4
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	mdhd.Size = size
	mdhd.Version = data[0]
	if mdhd.Version != 0 && mdhd.Version != 1 {
		if debugMode {
			log.Printf("ERROR: Unknown %s box version", boxPath)
		}
		return
	}
	copy(mdhd.Reserved[:], data[1:4])
	if mdhd.Version == 0 {
		mdhd.CreationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
		mdhd.ModificationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
	} else {
		mdhd.CreationTime = binary.BigEndian.Uint64(data[offset : offset+8])
		offset += 8
		mdhd.ModificationTime = binary.BigEndian.Uint64(data[offset : offset+8])
		offset += 8
	}
	mdhd.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	if mdhd.Version == 0 {
		mdhd.Duration = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
	} else {
		mdhd.Duration = binary.BigEndian.Uint64(data[offset : offset+8])
		offset += 8
	}
	mdhd.Language = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	mdhd.PreDefined = binary.BigEndian.Uint16(data[offset : offset+2])
	addBox(mp4, boxPath, mdhd)
	dumpBox(boxPath, mdhd)
}

func (mdhd MdhdBox) Bytes() (data []byte) {
	var offset uint32
	offset = 12
	boxSize := mdhd.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'m', 'd', 'h', 'd'})
	data[8] = mdhd.Version
	copy(data[9:12], mdhd.Reserved[:])
	if mdhd.Version == 0 {
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(mdhd.CreationTime))
		offset += 4
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(mdhd.ModificationTime))
		offset += 4
	} else {
		binary.BigEndian.PutUint64(data[offset:offset+8], mdhd.CreationTime)
		offset += 8
		binary.BigEndian.PutUint64(data[offset:offset+8], mdhd.ModificationTime)
		offset += 8
	}
	binary.BigEndian.PutUint32(data[offset:offset+4], mdhd.Timescale)
	offset += 4
	if mdhd.Version == 0 {
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(mdhd.Duration))
		offset += 4
	} else {
		binary.BigEndian.PutUint64(data[offset:offset+8], mdhd.Duration)
		offset += 8
	}
	binary.BigEndian.PutUint16(data[offset:offset+2], mdhd.Language)
	offset += 2
	binary.BigEndian.PutUint16(data[offset:offset+2], mdhd.PreDefined)

	return
}

func readHdlrBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	var hdlr HdlrBox
	hdlr.Size = size
	hdlr.Version = data[0]
	copy(hdlr.Reserved[:], data[1:4])
	hdlr.PreDefined = binary.BigEndian.Uint32(data[4:8])
	hdlr.HandlerType = binary.BigEndian.Uint32(data[8:12])
	for i := 0; i < 3; i++ {
		hdlr.Reserved2[i] = binary.BigEndian.Uint32(data[12+(i*4) : 16+(i*4)])
	}
	// String in utf8
	hdlr.Name = data[24:]
	addBox(mp4, boxPath, hdlr)
	dumpBox(boxPath, hdlr)
}

func (hdlr HdlrBox) Bytes() (data []byte) {
	boxSize := hdlr.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'h', 'd', 'l', 'r'})
	data[8] = hdlr.Version
	copy(data[9:12], hdlr.Reserved[:])
	binary.BigEndian.PutUint32(data[12:16], hdlr.PreDefined)
	binary.BigEndian.PutUint32(data[16:20], hdlr.HandlerType)
	for i := 0; i < 3; i++ {
		binary.BigEndian.PutUint32(data[20+(i*4):24+(i*4)], hdlr.Reserved2[i])
	}
	copy(data[32:], hdlr.Name[:])

	return
}

func readVmhdBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var vmhd VmhdBox
	vmhd.Size = size
	vmhd.Version = data[0]
	copy(vmhd.Reserved[:], data[1:4])
	vmhd.GraphicsMode = binary.BigEndian.Uint16(data[4:6])
	for i := 0; i < 3; i++ {
		vmhd.OpColor[i] = binary.BigEndian.Uint16(data[6+(i*2) : 8+(i*2)])
	}
	addBox(mp4, boxPath, vmhd)
	dumpBox(boxPath, vmhd)
}

func (vmhd VmhdBox) Bytes() (data []byte) {
	boxSize := vmhd.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'v', 'm', 'h', 'd'})
	data[8] = vmhd.Version
	copy(data[9:12], vmhd.Reserved[:])
	binary.BigEndian.PutUint16(data[12:14], vmhd.GraphicsMode)
	for i := 0; i < 3; i++ {
		binary.BigEndian.PutUint16(data[14+(i*2):16+(i*2)], vmhd.OpColor[i])
	}

	return
}

func readSmhdBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var smhd SmhdBox
	smhd.Size = size
	smhd.Version = data[0]
	copy(smhd.Reserved[:], data[1:4])
	smhd.Balance = int16(binary.BigEndian.Uint16(data[4:6]))
	smhd.Reserved2 = binary.BigEndian.Uint16(data[6:8])
	addBox(mp4, boxPath, smhd)
	dumpBox(boxPath, smhd)
}

func (smhd SmhdBox) Bytes() (data []byte) {
	boxSize := smhd.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'s', 'm', 'h', 'd'})
	data[8] = smhd.Version
	copy(data[9:12], smhd.Reserved[:])
	binary.BigEndian.PutUint16(data[12:14], uint16(smhd.Balance))
	binary.BigEndian.PutUint16(data[14:16], smhd.Reserved2)

	return
}

func readHmhdBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var hmhd HmhdBox
	hmhd.Version = data[0]
	copy(hmhd.Reserved[:], data[1:4])
	hmhd.MaxPDUSize = binary.BigEndian.Uint16(data[4:6])
	hmhd.AvgPDUSize = binary.BigEndian.Uint16(data[6:8])
	hmhd.MaxBitrate = binary.BigEndian.Uint32(data[8:12])
	hmhd.AvgBitrate = binary.BigEndian.Uint32(data[12:16])
	hmhd.Reserved2 = binary.BigEndian.Uint32(data[16:20])
	addBox(mp4, boxPath, hmhd)
	dumpBox(boxPath, hmhd)
}

func readDrefBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var offset uint32
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var dref DrefBox
	dref.Size = size
	dref.Version = data[0]
	copy(dref.Reserved[:], data[1:4])
	dref.EntryCount = binary.BigEndian.Uint32(data[4:8])
	offset = 8
	var i uint32
	for i = 0; i < dref.EntryCount; i++ {
		size := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		switch string(data[offset : offset+4]) {
		case "url ":
			offset += 4
			var url DrefUrlBox
			url.Size = size
			j := offset
			for data[j] != 0 {
				j++
			}
			url.Location = string(data[offset:j])
			offset = j
			url.Version = data[offset]
			offset++
			copy(url.Flags[0:3], data[offset:offset+3])
			offset += 3
			dref.UrlBox = append(dref.UrlBox, url)
		case "urn ":
			offset += 4
			var urn DrefUrnBox
			urn.Size = size
			var j uint32
			j = offset
			for data[j] != 0 {
				j++
			}
			urn.Name = string(data[offset:j])
			offset = j
			for data[j] != 0 {
				j++
			}
			urn.Location = string(data[offset:j])
			offset = j
			urn.Version = data[offset]
			offset++
			copy(urn.Flags[0:3], data[offset:offset+3])
			offset += 3
			dref.UrnBox = append(dref.UrnBox, urn)
		}
	}
	addBox(mp4, boxPath, dref)
	dumpBox(boxPath, dref)
}

func (dref DrefBox) Bytes() (data []byte) {
	var offset uint32
	boxSize := dref.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'d', 'r', 'e', 'f'})
	data[8] = dref.Version
	copy(data[9:12], dref.Reserved[:])
	binary.BigEndian.PutUint32(data[12:16], dref.EntryCount)
	offset = 16
	for _, v := range dref.UrlBox {
		binary.BigEndian.PutUint32(data[offset:offset+4], v.Size)
		offset += 4
		copy(data[offset:offset+4], []byte{'u', 'r', 'l', ' '}[:])
		offset += 4
		offset += uint32(copy(data[offset:offset+uint32(len(v.Location))], []byte(v.Location)))
		data[offset] = v.Version
		offset++
		offset += uint32(copy(data[offset:offset+3], v.Flags[0:3]))
	}
	for _, v := range dref.UrnBox {
		binary.BigEndian.PutUint32(data[offset:offset+4], v.Size)
		offset += 4
		copy(data[offset:offset+4], []byte{'u', 'r', 'n', ' '}[:])
		offset += uint32(copy(data[offset:uint32(len(v.Name))], []byte(v.Name)))
		offset += uint32(copy(data[offset:uint32(len(v.Location))], []byte(v.Location)))
		data[offset] = v.Version
		offset++
		offset += uint32(copy(data[offset:offset+3], v.Flags[0:3]))
	}

	return
}

func readMvhdBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var mvhd MvhdBox
	var offset uint32

	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	mvhd.Size = size
	mvhd.Version = data[0]
	if mvhd.Version != 0 && mvhd.Version != 1 {
		if debugMode {
			log.Printf("ERROR: Unknown %s box version", boxPath)
		}
		return
	}
	copy(mvhd.Reserved[:], data[1:4])
	offset = 4
	if mvhd.Version == 0 {
		mvhd.CreationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
		mvhd.ModificationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
	} else {
		mvhd.CreationTime = binary.BigEndian.Uint64(data[offset : offset+8])
		offset += 8
		mvhd.ModificationTime = binary.BigEndian.Uint64(data[offset : offset+8])
		offset += 8
	}
	mvhd.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	if mvhd.Version == 0 {
		mvhd.Duration = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
	} else {
		mvhd.Duration = binary.BigEndian.Uint64(data[offset : offset+8])
		offset += 8
	}
	mvhd.Rate = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4
	mvhd.Volume = int16(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	mvhd.Reserved2 = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	mvhd.Reserved3 = binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8
	for i := 0; i < 9; i++ {
		mvhd.Matrix[i] = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4
	}
	for i := 0; i < 6; i++ {
		mvhd.PreDefined[i] = binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
	}
	mvhd.NextTrackID = binary.BigEndian.Uint32(data[offset : offset+4])
	addBox(mp4, boxPath, mvhd)
	dumpBox(boxPath, mvhd)
}

func (mvhd MvhdBox) Bytes() (data []byte) {
	var offset uint32

	boxSize := mvhd.Size + 8
	data = make([]byte, boxSize)
	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'m', 'v', 'h', 'd'})
	data[8] = mvhd.Version
	copy(data[9:12], mvhd.Reserved[:])
	offset = 12
	if mvhd.Version == 0 {
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(mvhd.CreationTime))
		offset += 4
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(mvhd.ModificationTime))
		offset += 4
	} else {
		binary.BigEndian.PutUint64(data[offset:offset+8], mvhd.CreationTime)
		offset += 8
		binary.BigEndian.PutUint64(data[offset:offset+8], mvhd.ModificationTime)
		offset += 8
	}
	binary.BigEndian.PutUint32(data[offset:offset+4], mvhd.Timescale)
	offset += 4
	if mvhd.Version == 0 {
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(mvhd.Duration))
		offset += 4
	} else {
		binary.BigEndian.PutUint64(data[offset:offset+8], mvhd.Duration)
		offset += 8
	}
	binary.BigEndian.PutUint32(data[offset:offset+4], uint32(mvhd.Rate))
	offset += 4
	binary.BigEndian.PutUint16(data[offset:offset+2], uint16(mvhd.Volume))
	offset += 2
	binary.BigEndian.PutUint16(data[offset:offset+2], mvhd.Reserved2)
	offset += 2
	binary.BigEndian.PutUint64(data[offset:offset+8], mvhd.Reserved3)
	offset += 8
	for i := 0; i < 9; i++ {
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(mvhd.Matrix[i]))
		offset += 4
	}
	for i := 0; i < 6; i++ {
		binary.BigEndian.PutUint32(data[offset:offset+4], mvhd.PreDefined[i])
		offset += 4
	}
	binary.BigEndian.PutUint32(data[offset:offset+4], mvhd.NextTrackID)

	return
}

func readStsdBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, 8)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	var stsd StsdBox
	stsd.Size = size
	stsd.Version = data[0]
	copy(stsd.Reserved[:], data[1:4])
	stsd.EntryCount = binary.BigEndian.Uint32(data[4:8])

	addBox(mp4, boxPath, stsd)
	dumpBox(boxPath, stsd)

	readBoxes(f, size-8, level+1, boxPath, mp4)

	return
}

func (stsd StsdBox) Bytes() (data []byte) {
	boxSize := stsd.Size + 8
	data = make([]byte, 16)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'s', 't', 's', 'd'})
	data[8] = stsd.Version
	copy(data[9:12], stsd.Reserved[:])
	binary.BigEndian.PutUint32(data[12:16], stsd.EntryCount)

	return
}

func readMp4aBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, 28)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	var mp4a Mp4aBox
	mp4a.Size = size
	copy(mp4a.Reserved[0:6], data[0:6])
	mp4a.DataReferenceIndex = binary.BigEndian.Uint16(data[6:8])
	mp4a.Version = binary.BigEndian.Uint16(data[8:10])
	mp4a.RevisionLevel = binary.BigEndian.Uint16(data[10:12])
	mp4a.Vendor = binary.BigEndian.Uint32(data[12:16])
	mp4a.NumberOfChannels = binary.BigEndian.Uint16(data[16:18])
	mp4a.SampleSize = binary.BigEndian.Uint16(data[18:20])
	mp4a.CompressionId = binary.BigEndian.Uint16(data[20:22])
	mp4a.Reserved2 = binary.BigEndian.Uint16(data[22:24])
	mp4a.SampleRate = binary.BigEndian.Uint32(data[24:28])

	addBox(mp4, boxPath, mp4a)
	dumpBox(boxPath, mp4a)

	readBoxes(f, size-28, level+1, boxPath, mp4)

	return
}

func (mp4a Mp4aBox) Bytes() (data []byte) {
	boxSize := mp4a.Size + 8
	data = make([]byte, 36)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'m', 'p', '4', 'a'})
	copy(data[8:14], mp4a.Reserved[0:6])
	binary.BigEndian.PutUint16(data[14:16], mp4a.DataReferenceIndex)
	binary.BigEndian.PutUint16(data[16:18], mp4a.Version)
	binary.BigEndian.PutUint16(data[18:20], mp4a.RevisionLevel)
	binary.BigEndian.PutUint32(data[20:24], mp4a.Vendor)
	binary.BigEndian.PutUint16(data[24:26], mp4a.NumberOfChannels)
	binary.BigEndian.PutUint16(data[26:28], mp4a.SampleSize)
	binary.BigEndian.PutUint16(data[28:30], mp4a.CompressionId)
	binary.BigEndian.PutUint16(data[30:32], mp4a.Reserved2)
	binary.BigEndian.PutUint32(data[32:36], mp4a.SampleRate)

	return
}

func readEsdsBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	var esds EsdsBox
	esds.Size = size
	esds.Version = binary.BigEndian.Uint32(data[0:4])
	esds.Data = data[4:]

	addBox(mp4, boxPath, esds)
	dumpBox(boxPath, esds)

	return
}

func (esds EsdsBox) Bytes() (data []byte) {
	boxSize := esds.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'e', 's', 'd', 's'})
	binary.BigEndian.PutUint32(data[8:12], esds.Version)
	copy(data[12:], esds.Data[:])

	return
}

func readAvc1Box(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, 78)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	var avc1 Avc1Box
	avc1.Size = size
	copy(avc1.Reserved[0:6], data[0:6])
	avc1.dataReferenceIndex = binary.BigEndian.Uint16(data[6:8])
	avc1.Version = binary.BigEndian.Uint16(data[8:10])
	avc1.RevisionLevel = binary.BigEndian.Uint16(data[10:12])
	avc1.Vendor = binary.BigEndian.Uint32(data[12:16])
	avc1.TemporalQuality = binary.BigEndian.Uint32(data[16:20])
	avc1.SpacialQuality = binary.BigEndian.Uint32(data[20:24])
	avc1.Width = binary.BigEndian.Uint16(data[24:26])
	avc1.Height = binary.BigEndian.Uint16(data[26:28])
	avc1.HorizontalResolution = binary.BigEndian.Uint32(data[28:32])
	avc1.VerticalResolution = binary.BigEndian.Uint32(data[32:36])
	avc1.EntryDataSize = binary.BigEndian.Uint32(data[36:40])
	avc1.FramesPerSample = binary.BigEndian.Uint16(data[40:42])
	copy(avc1.CompressorName[0:32], data[42:74])
	avc1.BitDepth = binary.BigEndian.Uint16(data[74:76])
	avc1.ColorTableIndex = int16(binary.BigEndian.Uint16(data[76:78]))

	addBox(mp4, boxPath, avc1)
	dumpBox(boxPath, avc1)

	readBoxes(f, size-78, level+1, boxPath, mp4)

	return
}

func (avc1 Avc1Box) Bytes() (data []byte) {
	boxSize := avc1.Size + 8
	data = make([]byte, 86)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'a', 'v', 'c', '1'})
	copy(data[8:14], avc1.Reserved[0:6])
	binary.BigEndian.PutUint16(data[14:16], avc1.dataReferenceIndex)
	binary.BigEndian.PutUint16(data[16:18], avc1.Version)
	binary.BigEndian.PutUint16(data[18:20], avc1.RevisionLevel)
	binary.BigEndian.PutUint32(data[20:24], avc1.Vendor)
	binary.BigEndian.PutUint32(data[24:28], avc1.TemporalQuality)
	binary.BigEndian.PutUint32(data[28:32], avc1.SpacialQuality)
	binary.BigEndian.PutUint16(data[32:34], avc1.Width)
	binary.BigEndian.PutUint16(data[34:36], avc1.Height)
	binary.BigEndian.PutUint32(data[36:40], avc1.HorizontalResolution)
	binary.BigEndian.PutUint32(data[40:44], avc1.VerticalResolution)
	binary.BigEndian.PutUint32(data[44:48], avc1.EntryDataSize)
	binary.BigEndian.PutUint16(data[48:50], avc1.FramesPerSample)
	copy(data[50:82], avc1.CompressorName[0:32])
	binary.BigEndian.PutUint16(data[82:84], avc1.BitDepth)
	binary.BigEndian.PutUint16(data[84:86], uint16(avc1.ColorTableIndex))

	return
}

func readAvcCBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var offset uint32
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	var avcC AvcCBox
	avcC.Size = size
	avcC.ConfigurationVersion = data[0]
	avcC.AVCProfileIndication = data[1]
	avcC.ProfileCompatibility = data[2]
	avcC.AVCLevelIndication = data[3]
	avcC.NalUnitSize = data[4]
	avcC.SPSEntryCount = data[5] - 0xe0
	avcC.SPSSize = binary.BigEndian.Uint16(data[6:8])
	offset = 8
	avcC.SPSData = data[offset : offset+(uint32(avcC.SPSSize)*uint32(avcC.SPSEntryCount))]
	offset += uint32(avcC.SPSSize) * uint32(avcC.SPSEntryCount)
	avcC.PPSEntryCount = data[offset]
	offset++
	avcC.PPSSize = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	avcC.PPSData = data[offset : offset+(uint32(avcC.PPSSize)*uint32(avcC.PPSEntryCount))]

	addBox(mp4, boxPath, avcC)
	dumpBox(boxPath, avcC)

	return
}

func (avcC AvcCBox) Bytes() (data []byte) {
	var offset uint32
	boxSize := avcC.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'a', 'v', 'c', 'C'})
	data[8] = avcC.ConfigurationVersion
	data[9] = avcC.AVCProfileIndication
	data[10] = avcC.ProfileCompatibility
	data[11] = avcC.AVCLevelIndication
	data[12] = avcC.NalUnitSize
	data[13] = avcC.SPSEntryCount + 0xe0
	binary.BigEndian.PutUint16(data[14:16], avcC.SPSSize)
	offset = 16
	copy(data[offset:offset+(uint32(avcC.SPSSize)*uint32(avcC.SPSEntryCount))], avcC.SPSData[0:uint32(avcC.SPSEntryCount)*uint32(avcC.SPSSize)])
	offset += uint32(avcC.SPSSize) * uint32(avcC.SPSEntryCount)
	data[offset] = avcC.PPSEntryCount
	offset++
	binary.BigEndian.PutUint16(data[offset:offset+2], avcC.PPSSize)
	offset += 2
	copy(data[offset:offset+(uint32(avcC.PPSSize)*uint32(avcC.PPSEntryCount))], avcC.PPSData[0:uint32(avcC.PPSSize)*uint32(avcC.PPSEntryCount)])

	return
}

func readBtrtBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	var btrt BtrtBox
	btrt.Size = size
	btrt.DecodingBufferSize = binary.BigEndian.Uint32(data[0:4])
	btrt.MaxBitrate = binary.BigEndian.Uint32(data[4:8])
	btrt.AvgBitrate = binary.BigEndian.Uint32(data[8:12])

	addBox(mp4, boxPath, btrt)
	dumpBox(boxPath, btrt)

	return
}

func (btrt BtrtBox) Bytes() (data []byte) {
	boxSize := btrt.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'b', 't', 'r', 't'})
	binary.BigEndian.PutUint32(data[8:12], btrt.DecodingBufferSize)
	binary.BigEndian.PutUint32(data[12:16], btrt.MaxBitrate)
	binary.BigEndian.PutUint32(data[16:20], btrt.AvgBitrate)

	return
}

func readStscBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	var stsc StscBox
	stsc.Size = size
	stsc.Version = data[0]
	copy(stsc.Flags[:], data[1:4])
	stsc.EntryCount = binary.BigEndian.Uint32(data[4:8])
	stsc.Entries = make([]StscEntry, stsc.EntryCount)
	var i uint32
	for i = 0; i < stsc.EntryCount; i++ {
		stsc.Entries[i].FirstChunk = binary.BigEndian.Uint32(data[8+(i*12) : 12+(i*12)])
		stsc.Entries[i].SamplesPerChunk = binary.BigEndian.Uint32(data[12+(i*12) : 16+(i*12)])
		stsc.Entries[i].SampleDescriptionIndex = binary.BigEndian.Uint32(data[16+(i*12) : 20+(i*12)])
	}
	addBox(mp4, boxPath, stsc)
	dumpBox(boxPath, stsc)
}

func (stsc StscBox) Bytes() (data []byte) {
	var offset uint32
	boxSize := stsc.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'s', 't', 's', 'c'})
	data[8] = stsc.Version
	copy(data[9:12], stsc.Flags[:])
	binary.BigEndian.PutUint32(data[12:16], stsc.EntryCount)
	offset = 16
	for _, v := range stsc.Entries {
		binary.BigEndian.PutUint32(data[offset:offset+4], v.FirstChunk)
		offset += 4
		binary.BigEndian.PutUint32(data[offset:offset+4], v.SamplesPerChunk)
		offset += 4
		binary.BigEndian.PutUint32(data[offset:offset+4], v.SampleDescriptionIndex)
		offset += 4
	}

	return
}

func readStszBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var stsz StszBox
	stsz.Offset, _ = f.Seek(0, os.SEEK_CUR)
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	stsz.Size = size
	stsz.Version = data[0]
	copy(stsz.Reserved[:], data[1:4])
	stsz.SampleSize = binary.BigEndian.Uint32(data[4:8])
	stsz.SampleCount = binary.BigEndian.Uint32(data[8:12])
	if stsz.SampleSize == 0 {
		sampleCount := (size - 12) >> 2
		if sampleCount < stsz.SampleCount {
			stsz.SampleCount = sampleCount
		}
		stsz.EntrySize = make([]uint32, stsz.SampleCount)
		var i uint32
		for i = 0; i < stsz.SampleCount; i++ {
			stsz.EntrySize[i] = binary.BigEndian.Uint32(data[12+(i*4) : 16+(i*4)])
		}
	}
	addBox(mp4, boxPath, stsz)
	dumpBox(boxPath, stsz)
}

func (stsz StszBox) Bytes() (data []byte) {
	var offset uint32
	boxSize := stsz.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'s', 't', 's', 'z'})
	data[8] = stsz.Version
	copy(data[9:12], stsz.Reserved[:])
	binary.BigEndian.PutUint32(data[12:16], stsz.SampleSize)
	binary.BigEndian.PutUint32(data[16:20], stsz.SampleCount)
	offset = 20
	if stsz.SampleSize == 0 {
		sampleCount := (stsz.Size - 12) >> 2
		if sampleCount < stsz.SampleCount {
			stsz.SampleCount = sampleCount
		}
		var i uint32
		for i = 0; i < stsz.SampleCount; i++ {
			binary.BigEndian.PutUint32(data[offset:offset+4], stsz.EntrySize[i])
			offset += 4
		}
	}

	return
}

func readSdtpBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var sdtp SdtpBox
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	sdtp.Size = size
	sdtp.Version = data[0]
	if mp4["moov.trak.mdia.minf.stbl.stsz"] != nil {
		stsz := mp4["moov.trak.mdia.minf.stbl.stsz"][0].(StszBox)
		sdtp.SampleCount = stsz.SampleCount
		sdtp.Entries = make([]uint8, sdtp.SampleCount)
		var i uint32
		for i = 0; i < sdtp.SampleCount; i++ {
			sdtp.Entries[i] = uint8(data[1+i])
		}
	} else {
		sdtp.SampleCount = 0
	}

	return
}

func (sdtp SdtpBox) Bytes() (data []byte) {
	boxSize := sdtp.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'s', 'd', 't', 'p'})
	data[8] = sdtp.Version
	var i uint32
	for i = 0; i < sdtp.SampleCount; i++ {
		data[9+i] = byte(sdtp.Entries[i])
	}

	return
}

func readStcoBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var stco StcoBox
	stco.Size = size
	stco.Version = data[0]
	copy(stco.Reserved[:], data[1:4])
	stco.EntryCount = binary.BigEndian.Uint32(data[4:8])
	if stco.EntryCount > 0 {
		stco.ChunkOffset = make([]uint32, stco.EntryCount)
		var i uint32
		for i = 0; i < stco.EntryCount; i++ {
			stco.ChunkOffset[i] = binary.BigEndian.Uint32(data[8+(i*4) : 12+(i*4)])
		}
	}
	addBox(mp4, boxPath, stco)
	dumpBox(boxPath, stco)
}

func (stco StcoBox) Bytes() (data []byte) {
	var offset uint32
	boxSize := stco.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'s', 't', 'c', 'o'})
	data[8] = stco.Version
	copy(data[9:12], stco.Reserved[:])
	binary.BigEndian.PutUint32(data[12:16], stco.EntryCount)
	offset = 16
	if stco.EntryCount > 0 {
		for _, v := range stco.ChunkOffset {
			binary.BigEndian.PutUint32(data[offset:offset+4], v)
			offset += 4
		}
	}

	return
}

func readSttsBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	var stts SttsBox
	stts.Size = size
	stts.Version = data[0]
	copy(stts.Reserved[:], data[1:4])
	stts.EntryCount = binary.BigEndian.Uint32(data[4:8])
	stts.Entries = make([]SttsBoxEntry, stts.EntryCount)
	var i uint32
	for i = 0; i < stts.EntryCount; i++ {
		stts.Entries[i].SampleCount = binary.BigEndian.Uint32(data[8+(i*8) : 12+(i*8)])
		stts.Entries[i].SampleDelta = binary.BigEndian.Uint32(data[12+(i*8) : 16+(i*8)])
	}
	addBox(mp4, boxPath, stts)
	dumpBox(boxPath, stts)
}

func (stts SttsBox) Bytes() (data []byte) {
	var offset uint32
	boxSize := stts.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'s', 't', 't', 's'})
	data[8] = stts.Version
	copy(data[9:12], stts.Reserved[:])
	binary.BigEndian.PutUint32(data[12:16], stts.EntryCount)
	offset = 16
	for _, v := range stts.Entries {
		binary.BigEndian.PutUint32(data[offset:offset+4], v.SampleCount)
		offset += 4
		binary.BigEndian.PutUint32(data[offset:offset+4], v.SampleDelta)
		offset += 4
	}

	return
}

func readCttsBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var ctts CttsBox
	ctts.Offset, _ = f.Seek(0, os.SEEK_CUR)
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	ctts.Size = size
	ctts.Version = data[0]
	copy(ctts.Reserved[:], data[1:4])
	ctts.EntryCount = binary.BigEndian.Uint32(data[4:8])
	ctts.Entries = make([]CttsBoxEntry, ctts.EntryCount)
	var i uint32
	for i = 0; i < ctts.EntryCount; i++ {
		ctts.Entries[i].SampleCount = binary.BigEndian.Uint32(data[8+(i*8) : 12+(i*8)])
		ctts.Entries[i].SampleOffset = binary.BigEndian.Uint32(data[12+(i*8) : 16+(i*8)])
	}
	addBox(mp4, boxPath, ctts)
	dumpBox(boxPath, ctts)
}

func (ctts CttsBox) Bytes() (data []byte) {
	var offset uint32
	boxSize := ctts.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'c', 't', 't', 's'})
	data[8] = ctts.Version
	copy(data[9:12], ctts.Reserved[:])
	binary.BigEndian.PutUint32(data[12:16], ctts.EntryCount)
	offset = 16
	for _, v := range ctts.Entries {
		binary.BigEndian.PutUint32(data[offset:offset+4], v.SampleCount)
		offset += 4
		binary.BigEndian.PutUint32(data[offset:offset+4], v.SampleOffset)
		offset += 4
	}

	return
}

func readStssBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var stss StssBox
	stss.Offset, _ = f.Seek(0, os.SEEK_CUR)
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	stss.Size = size
	stss.Version = data[0]
	copy(stss.Reserved[:], data[1:4])
	stss.EntryCount = binary.BigEndian.Uint32(data[4:8])
	stss.SampleNumber = make([]uint32, stss.EntryCount)
	entryCount := (size - 8) >> 2
	if entryCount < stss.EntryCount {
		stss.EntryCount = entryCount
	}
	var i uint32
	for i = 0; i < stss.EntryCount; i++ {
		stss.SampleNumber[i] = binary.BigEndian.Uint32(data[8+(i*4) : 12+(i*4)])
	}
	addBox(mp4, boxPath, stss)
	dumpBox(boxPath, stss)
}

func (stss StssBox) Bytes() (data []byte) {
	boxSize := stss.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'s', 't', 's', 's'})
	data[8] = stss.Version
	copy(data[9:12], stss.Reserved[:])
	binary.BigEndian.PutUint32(data[12:16], stss.EntryCount)
	entryCount := (stss.Size - 8) >> 2
	if entryCount < stss.EntryCount {
		stss.EntryCount = stss.Size - 12
	}
	var i uint32
	for i = 0; i < stss.EntryCount; i++ {
		binary.BigEndian.PutUint32(data[16+(i*4):20+(i*4)], stss.SampleNumber[i])
	}

	return
}

func readMehdBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var offset uint32
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var mehd MehdBox
	mehd.Size = size
	mehd.Version = data[0]
	if mehd.Version != 0 && mehd.Version != 1 {
		if debugMode {
			log.Printf("ERROR: Unknown %s box version", boxPath)
		}
		return
	}
	copy(mehd.Reserved[:], data[1:4])
	offset = 4
	if mehd.Version == 0 {
		mehd.FragmentDuration = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
	} else {
		mehd.FragmentDuration = binary.BigEndian.Uint64(data[offset : offset+8])
	}
	addBox(mp4, boxPath, mehd)
	dumpBox(boxPath, mehd)
}

func (mehd MehdBox) Bytes() (data []byte) {
	var offset uint32
	boxSize := mehd.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'m', 'e', 'h', 'd'})
	data[8] = mehd.Version
	copy(data[9:12], mehd.Reserved[:])
	offset = 12
	if mehd.Version == 0 {
		binary.BigEndian.PutUint32(data[offset:offset+4], uint32(mehd.FragmentDuration))
	} else {
		binary.BigEndian.PutUint64(data[offset:offset+8], mehd.FragmentDuration)
	}

	return
}

func readTrexBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var trex TrexBox
	trex.Size = size
	trex.Version = data[0]
	copy(trex.Reserved[:], data[1:4])
	trex.TrackID = binary.BigEndian.Uint32(data[4:8])
	trex.DefaultSampleDescriptionIndex = binary.BigEndian.Uint32(data[8:12])
	trex.DefaultSampleDuration = binary.BigEndian.Uint32(data[12:16])
	trex.DefaultSampleSize = binary.BigEndian.Uint32(data[16:20])
	trex.DefaultSampleFlags = binary.BigEndian.Uint32(data[20:24])
	addBox(mp4, boxPath, trex)
	dumpBox(boxPath, trex)
}

func (trex TrexBox) Bytes() (data []byte) {
	boxSize := trex.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'t', 'r', 'e', 'x'})
	data[8] = trex.Version
	copy(data[9:12], trex.Reserved[:])
	binary.BigEndian.PutUint32(data[12:16], trex.TrackID)
	binary.BigEndian.PutUint32(data[16:20], trex.DefaultSampleDescriptionIndex)
	binary.BigEndian.PutUint32(data[20:24], trex.DefaultSampleDuration)
	binary.BigEndian.PutUint32(data[24:28], trex.DefaultSampleSize)
	binary.BigEndian.PutUint32(data[28:32], trex.DefaultSampleFlags)

	return
}

func readMfhdBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var mfhd MfhdBox
	mfhd.Size = size
	mfhd.Version = data[0]
	copy(mfhd.Reserved[:], data[1:4])
	mfhd.SequenceNumber = binary.BigEndian.Uint32(data[4:8])
	addBox(mp4, boxPath, mfhd)
	dumpBox(boxPath, mfhd)
}

func (mfhd MfhdBox) Bytes() (data []byte) {
	boxSize := mfhd.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'m', 'f', 'h', 'd'})
	data[8] = mfhd.Version
	copy(data[9:12], mfhd.Reserved[:])
	binary.BigEndian.PutUint32(data[12:16], mfhd.SequenceNumber)

	return
}

func readTfhdBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var tfhd TfhdBox
	tfhd.Size = size
	tfhd.Version = data[0]
	copy(tfhd.Flags[:], data[1:4])
	tfhd.TrackID = binary.BigEndian.Uint32(data[4:8])
	dataOffset := 8
	if tfhd.Flags[2]&0x01 != 0 {
		tfhd.BaseDataOffset = binary.BigEndian.Uint64(data[dataOffset : dataOffset+8])
		dataOffset += 8
	}
	if tfhd.Flags[2]&0x02 != 0 {
		tfhd.SampleDescriptionIndex = binary.BigEndian.Uint32(data[dataOffset : dataOffset+4])
		dataOffset += 4
	}
	if tfhd.Flags[2]&0x08 != 0 {
		tfhd.DefaultSampleDuration = binary.BigEndian.Uint32(data[dataOffset : dataOffset+4])
		dataOffset += 4
	}
	if tfhd.Flags[2]&0x10 != 0 {
		tfhd.DefaultSampleSize = binary.BigEndian.Uint32(data[dataOffset : dataOffset+4])
		dataOffset += 4
	}
	if tfhd.Flags[2]&0x20 != 0 {
		tfhd.DefaultSampleFlags = binary.BigEndian.Uint32(data[dataOffset : dataOffset+4])
		dataOffset += 4
	}
	addBox(mp4, boxPath, tfhd)
	dumpBox(boxPath, tfhd)
}

func (tfhd TfhdBox) Bytes() (data []byte) {
	boxSize := tfhd.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'t', 'f', 'h', 'd'})
	data[8] = tfhd.Version
	copy(data[9:12], tfhd.Flags[:])
	binary.BigEndian.PutUint32(data[12:16], tfhd.TrackID)
	dataOffset := 16
	if tfhd.Flags[2]&0x01 != 0 {
		binary.BigEndian.PutUint64(data[dataOffset:dataOffset+8], tfhd.BaseDataOffset)
		dataOffset += 8
	}
	if tfhd.Flags[2]&0x02 != 0 {
		binary.BigEndian.PutUint32(data[dataOffset:dataOffset+4], tfhd.SampleDescriptionIndex)
		dataOffset += 4
	}
	if tfhd.Flags[2]&0x08 != 0 {
		binary.BigEndian.PutUint32(data[dataOffset:dataOffset+4], tfhd.DefaultSampleDuration)
		dataOffset += 4
	}
	if tfhd.Flags[2]&0x10 != 0 {
		binary.BigEndian.PutUint32(data[dataOffset:dataOffset+4], tfhd.DefaultSampleSize)
		dataOffset += 4
	}
	if tfhd.Flags[2]&0x20 != 0 {
		binary.BigEndian.PutUint32(data[dataOffset:dataOffset+4], tfhd.DefaultSampleFlags)
		dataOffset += 4
	}

	return
}

func readTrunBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var trun TrunBox
	trun.Version = data[0]
	copy(trun.Flags[:], data[1:4])
	trun.SampleCount = binary.BigEndian.Uint32(data[4:8])
	var dataOffset uint32
	dataOffset = 8
	if trun.Flags[2]&0x01 != 0 {
		trun.DataOffset = int32(binary.BigEndian.Uint32(data[dataOffset : dataOffset+4]))
		dataOffset += 4
	}
	if trun.Flags[2]&0x04 != 0 {
		trun.FirstSampleFlags = binary.BigEndian.Uint32(data[dataOffset : dataOffset+4])
		dataOffset += 4
	}
	if trun.Flags[1] != 0 {
		trun.Samples = make([]TrunBoxSample, trun.SampleCount)
		var i uint32
		for i = 0; i < trun.SampleCount; i++ {
			if trun.Flags[1]&0x01 != 0 {
				trun.Samples[i].Duration = binary.BigEndian.Uint32(data[dataOffset : dataOffset+4])
				dataOffset += 4
			}
			if trun.Flags[1]&0x02 != 0 {
				trun.Samples[i].Size = binary.BigEndian.Uint32(data[dataOffset : dataOffset+4])
				dataOffset += 4
			}
			if trun.Flags[1]&0x04 != 0 {
				trun.Samples[i].Flags = binary.BigEndian.Uint32(data[dataOffset : dataOffset+4])
				dataOffset += 4
			}
			if trun.Flags[1]&0x08 != 0 {
				if trun.Version == 0 {
					trun.Samples[i].CompositionTimeOffset = int64(binary.BigEndian.Uint32(data[dataOffset : dataOffset+4]))
				} else {
					trun.Samples[i].CompositionTimeOffset = int64(int32(binary.BigEndian.Uint32(data[dataOffset : dataOffset+4])))
				}
				dataOffset += 4
			}
		}
	}
	addBox(mp4, boxPath, trun)
	dumpBox(boxPath, trun)
}

func (trun TrunBox) Bytes() (data []byte) {
	boxSize := trun.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'t', 'r', 'u', 'n'})
	data[8] = trun.Version
	copy(data[9:12], trun.Flags[:])
	binary.BigEndian.PutUint32(data[12:16], trun.SampleCount)
	dataOffset := 16
	if trun.Flags[2]&0x01 != 0 {
		binary.BigEndian.PutUint32(data[dataOffset:dataOffset+4], uint32(trun.DataOffset))
		dataOffset += 4
	}
	if trun.Flags[2]&0x04 != 0 {
		binary.BigEndian.PutUint32(data[dataOffset:dataOffset+4], trun.FirstSampleFlags)
		dataOffset += 4
	}
	if trun.Flags[1] != 0 {
		var i uint32
		for i = 0; i < trun.SampleCount; i++ {
			if trun.Flags[1]&0x01 != 0 {
				binary.BigEndian.PutUint32(data[dataOffset:dataOffset+4], trun.Samples[i].Duration)
				dataOffset += 4
			}
			if trun.Flags[1]&0x02 != 0 {
				binary.BigEndian.PutUint32(data[dataOffset:dataOffset+4], trun.Samples[i].Size)
				dataOffset += 4
			}
			if trun.Flags[1]&0x04 != 0 {
				binary.BigEndian.PutUint32(data[dataOffset:dataOffset+4], trun.Samples[i].Flags)
				dataOffset += 4
			}
			if trun.Flags[1]&0x08 != 0 {
				binary.BigEndian.PutUint32(data[dataOffset:dataOffset+4], uint32(trun.Samples[i].CompositionTimeOffset))
				dataOffset += 4
			}
		}
	}

	return
}

func readTfdtBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var offset uint32
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var tfdt TfdtBox
	tfdt.Size = size
	tfdt.Version = data[0]
	if tfdt.Version != 0 && tfdt.Version != 1 {
		if debugMode {
			log.Printf("ERROR: Unknown %s box version", boxPath)
		}
		return
	}
	copy(tfdt.Reserved[:], data[1:4])
	offset = 4
	if tfdt.Version == 0 {
		tfdt.BaseMediaDecodeTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
	} else {
		tfdt.BaseMediaDecodeTime = binary.BigEndian.Uint64(data[offset : offset+8])
	}
	addBox(mp4, boxPath, tfdt)
	dumpBox(boxPath, tfdt)
}

func (tfdt TfdtBox) Bytes() (data []byte) {
	boxSize := tfdt.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'t', 'f', 'd', 't'})
	data[8] = tfdt.Version
	copy(data[9:12], tfdt.Reserved[:])
	dataOffset := 12
	if tfdt.Version == 0 {
		binary.BigEndian.PutUint32(data[dataOffset:dataOffset+4], uint32(tfdt.BaseMediaDecodeTime))
		dataOffset += 4
	} else {
		binary.BigEndian.PutUint64(data[dataOffset:dataOffset+8], tfdt.BaseMediaDecodeTime)
		dataOffset += 8
	}

	return
}

func readFrmaBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var frma FrmaBox
	frma.Size = size
	copy(frma.DataFormat[:], data[0:4])
	addBox(mp4, boxPath, frma)
	dumpBox(boxPath, frma)
}

func (frma FrmaBox) Bytes() (data []byte) {
	boxSize := frma.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'f', 'r', 'm', 'a'})
	copy(data[8:12], frma.DataFormat[:])

	return
}

func readSchmBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	data := make([]byte, size)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	var schm SchmBox
	schm.Size = size
	schm.Version = data[0]
	copy(schm.Flags[:], data[1:4])
	copy(schm.SchemeType[:], data[4:8])
	schm.SchemeVersion = binary.BigEndian.Uint32(data[8:12])
	if schm.Flags[0] == 0x00 && schm.Flags[1] == 0x00 && schm.Flags[2] == 0x01 {
		schm.SchemeUri = string(data[12:])
	}
	addBox(mp4, boxPath, schm)
	dumpBox(boxPath, schm)
}

func (schm SchmBox) Bytes() (data []byte) {
	boxSize := schm.Size + 8
	data = make([]byte, boxSize)

	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'s', 'c', 'h', 'm'})
	data[8] = schm.Version
	copy(data[9:12], schm.Flags[0:3])
	copy(data[12:16], schm.SchemeType[0:4])
	binary.BigEndian.PutUint32(data[16:20], schm.SchemeVersion)
	if schm.Flags[0] == 0x00 && schm.Flags[1] == 0x00 && schm.Flags[2] == 0x01 {
		copy(data[20:], []byte(schm.SchemeUri)[:])
	}

	return
}

func readMdatBox(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var mdat MdatBox

	mdat.Size = size
	mdat.Filename = f.Name()
	offset, err := f.Seek(0, os.SEEK_CUR)
	if err != nil {
		panic(err)
	}
	mdat.Offset = offset

	_, err = f.Seek(int64(size), os.SEEK_CUR)
	if err != nil {
		panic(err)
	}

	addBox(mp4, boxPath, mdat)
	dumpBox(boxPath, mdat)
}

func (mdat MdatBox) Bytes() (data []byte) {
	boxSize := mdat.Size + 8
	data = make([]byte, boxSize)
	binary.BigEndian.PutUint32(data[0:4], boxSize)
	copy(data[4:8], []byte{'m', 'd', 'a', 't'})
	f, err := os.Open(mdat.Filename)
	if err != nil {
		panic(err)
	}
	_, err = f.ReadAt(data[8:], mdat.Offset)
	if err != nil {
		panic(err)
	}

	return
}

func (mdat MdatBox) ToBytes() (data []byte) {
	boxSize := mdat.Size
	data = make([]byte, boxSize)
	f, err := os.Open(mdat.Filename)
	if err != nil {
		panic(err)
	}
	_, err = f.ReadAt(data, mdat.Offset)
	if err != nil {
		panic(err)
	}

	return
}

// Read 8 bytes Box (4 bytes size and 4 bytes box name)
func readBox(f *os.File, level int) (boxSize uint32, boxName string) {
	data := make([]byte, 8)
	_, err := f.Read(data)
	if err != nil {
		panic(err)
	}

	boxSize = binary.BigEndian.Uint32(data[0:4])
	boxName = string(data[4:])

	if debugMode {
		log.Printf("( off %.8d ) [%s%s]", boxSize, strings.Repeat("*", level*2), boxName)
	}

	return
}

func readBoxes(f *os.File, size uint32, level int, boxPath string, mp4 map[string][]interface{}) {
	var offset uint32
	offset = 0

	for offset < size {
		boxSize, boxName := readBox(f, level)
		var boxFullPath string
		if boxPath == "" {
			boxFullPath = boxName
		} else {
			boxFullPath = boxPath + "." + boxName
		}

		if funcBoxes[boxFullPath] != nil {
			fb := reflect.ValueOf(funcBoxes[boxFullPath])
			rb := reflect.ValueOf(readBoxes)
			if fb.Pointer() == rb.Pointer() {
				var box ParentBox
				copy(box.Name[:], []byte(boxName)[0:4])
				box.Size = boxSize
				addBox(mp4, boxFullPath, box)
			}
			var callFunc func(*os.File, uint32, int, string, map[string][]interface{})
			callFunc = funcBoxes[boxFullPath].(func(*os.File, uint32, int, string, map[string][]interface{}))
			callFunc(f, boxSize-8, level+1, boxFullPath, mp4)
		} else {
			// Skip box because we don't know how to decode it
			if debugMode {
				log.Printf("ERROR: Unknown %s box", boxPath)
			}
			f.Seek(int64(boxSize-8), 1)
		}

		offset += boxSize
	}
}

func writeBox(f *os.File, boxName [4]byte, box interface{}) {
	var size uint32
	size = uint32(binary.Size(box))
	log.Printf("size of box is %u", size)
	err := binary.Write(f, binary.BigEndian, size)
	if err != nil {
		log.Printf("cannot write box size: %v", err)
		return
	}
	_, err = f.Write(boxName[:])
	if err != nil {
		log.Printf("Cannot write box name: %v", err)
		return
	}
	err = binary.Write(f, binary.BigEndian, box)
	if err != nil {
		log.Printf("Cannot write structure to file: %v", err)
		return
	}
}

// ***
// *** Public functions
// ***

func Debug(mode bool) {
	debugMode = mode
}

// Parse the mp4 file header and return all decoded box data in a map[string][]interface{}
func ParseFile(filename string, language string) (mp4 Mp4) {
	mp4.Boxes = make(map[string][]interface{})
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	finfo, err := f.Stat()
	if err != nil {
		panic(err)
	}
	readBoxes(f, uint32(finfo.Size()), 0, "", mp4.Boxes)
	if debugMode {
		log.Printf("[ MP4 STRUCTURE ] %+v", mp4.Boxes)
	}

	mp4.Filename = filename
	mp4.Language = language
	if mp4.Boxes["moov.trak.mdia.minf.stbl.stsd.mp4a"] != nil {
		mp4.IsAudio = true
	} else {
		mp4.IsAudio = false
	}

	if mp4.Boxes["moov.trak.mdia.minf.stbl.stsd.avc1"] != nil {
		mp4.IsVideo = true
	} else {
		mp4.IsVideo = false
	}

	return
}

func boxToBytes(box interface{}, boxFullPath string) []byte {
	boxNames := strings.Split(boxFullPath, ".")
	boxName := boxNames[len(boxNames)-1]
	switch boxName {
	case "ftyp":
		ftyp := box.(FtypBox)
		return ftyp.Bytes()
	case "free":
		free := box.(FreeBox)
		return free.Bytes()
	case "moov":
		moov := box.(ParentBox)
		return moov.Bytes()
	case "mvhd":
		mvhd := box.(MvhdBox)
		return mvhd.Bytes()
	case "trak":
		trak := box.(ParentBox)
		return trak.Bytes()
	case "tkhd":
		tkhd := box.(TkhdBox)
		return tkhd.Bytes()
	case "mdia":
		mdia := box.(ParentBox)
		return mdia.Bytes()
	case "mdhd":
		mdhd := box.(MdhdBox)
		return mdhd.Bytes()
	case "hdlr":
		hdlr := box.(HdlrBox)
		return hdlr.Bytes()
	case "minf":
		minf := box.(ParentBox)
		return minf.Bytes()
	case "smhd":
		smhd := box.(SmhdBox)
		return smhd.Bytes()
	case "vmhd":
		vmhd := box.(VmhdBox)
		return vmhd.Bytes()
	case "dinf":
		dinf := box.(ParentBox)
		return dinf.Bytes()
	case "dref":
		dref := box.(DrefBox)
		return dref.Bytes()
	case "stbl":
		stbl := box.(ParentBox)
		return stbl.Bytes()
	case "stsd":
		stsd := box.(StsdBox)
		return stsd.Bytes()
	case "mp4a":
		mp4a := box.(Mp4aBox)
		return mp4a.Bytes()
	case "esds":
		esds := box.(EsdsBox)
		return esds.Bytes()
	case "avc1":
		avc1 := box.(Avc1Box)
		return avc1.Bytes()
	case "avcC":
		avcC := box.(AvcCBox)
		return avcC.Bytes()
	case "btrt":
		btrt := box.(BtrtBox)
		return btrt.Bytes()
	case "stts":
		stts := box.(SttsBox)
		return stts.Bytes()
	case "ctts":
		ctts := box.(CttsBox)
		return ctts.Bytes()
	case "stsc":
		stsc := box.(StscBox)
		return stsc.Bytes()
	case "stsz":
		stsz := box.(StszBox)
		return stsz.Bytes()
	case "sdtp":
		sdtp := box.(SdtpBox)
		return sdtp.Bytes()
	case "stco":
		stco := box.(StcoBox)
		return stco.Bytes()
	case "stss":
		stss := box.(StssBox)
		return stss.Bytes()
	case "mvex":
		mvex := box.(ParentBox)
		return mvex.Bytes()
	case "mehd":
		mehd := box.(MehdBox)
		return mehd.Bytes()
	case "trex":
		trex := box.(TrexBox)
		return trex.Bytes()
	case "styp":
		styp := box.(StypBox)
		return styp.Bytes()
	case "moof":
		moof := box.(ParentBox)
		return moof.Bytes()
	case "mfhd":
		mfhd := box.(MfhdBox)
		return mfhd.Bytes()
	case "traf":
		traf := box.(ParentBox)
		return traf.Bytes()
	case "tfhd":
		tfhd := box.(TfhdBox)
		return tfhd.Bytes()
	case "tfdt":
		tfdt := box.(TfdtBox)
		return tfdt.Bytes()
	case "trun":
		trun := box.(TrunBox)
		return trun.Bytes()
	case "frma":
		frma := box.(FrmaBox)
		return frma.Bytes()
	case "mdat":
		mdat := box.(MdatBox)
		return mdat.Bytes()
	}

	return nil
}

func MapToBytes(mp4 map[string][]interface{}) (data []byte) {
	boxPathOrder := []string{
		"ftyp",
		"styp",
		"free",
		"moof",
		"moof.mfhd",
		"moof.traf",
		"moof.traf.tfhd",
		"moof.traf.tfdt",
		"moof.traf.trun",
		"moov",
		"moov.mvhd",
		"moov.trak",
		"moov.trak.tkhd",
		"moov.trak.mdia",
		"moov.trak.mdia.mdhd",
		"moov.trak.mdia.hdlr",
		"moov.trak.mdia.minf",
		"moov.trak.mdia.minf.smhd",
		"moov.trak.mdia.minf.vmhd",
		"moov.trak.mdia.minf.dinf",
		"moov.trak.mdia.minf.dinf.dref",
		"moov.trak.mdia.minf.stbl",
		"moov.trak.mdia.minf.stbl.stsd",
		"moov.trak.mdia.minf.stbl.stsd.mp4a",
		"moov.trak.mdia.minf.stbl.stsd.mp4a.esds",
		"moov.trak.mdia.minf.stbl.stsd.avc1",
		"moov.trak.mdia.minf.stbl.stsd.avc1.avcC",
		"moov.trak.mdia.minf.stbl.stsd.avc1.btrt",
		"moov.trak.mdia.minf.stbl.stts",
		"moov.trak.mdia.minf.stbl.ctts",
		"moov.trak.mdia.minf.stbl.stsc",
		"moov.trak.mdia.minf.stbl.stsz",
		"moov.trak.mdia.minf.stbl.sdtp",
		"moov.trak.mdia.minf.stbl.stco",
		"moov.mvex",
		"moov.mvex.mehd",
		"moov.mvex.trex",
		"mdat",
	}

	for _, v := range boxPathOrder {
		if mp4[v] == nil {
			continue
		}
		b := boxToBytes(mp4[v][0], v)
		if b == nil {
			break
		}
		data = append(data, b...)
	}

	return
}

// Create a DASH format mp4 Init header
func CreateDashInit(mp4 map[string][]interface{}) (mp4Init map[string][]interface{}) {
	var isVideo bool
	// DASH MP4 Init structure

	// Create All needed mp4 boxes
	// First copy all original boxes needed on DASH Init file
	mp4Init = make(map[string][]interface{})

	// Create FTYP
	var ftyp FtypBox
	ftyp.MajorBrand = [4]byte{'i', 's', 'o', '6'}
	ftyp.MinorVersion = 0
	ftyp.CompatibleBrands = make([][4]byte, 2)
	ftyp.CompatibleBrands[0] = [4]byte{'i', 's', 'o', '6'}
	ftyp.CompatibleBrands[1] = [4]byte{'d', 'a', 's', 'h'}
	ftyp.Size = 16
	replaceBox(mp4Init, "ftyp", ftyp)

	// Then modify some of them
	var free FreeBox
	free.Data = []byte("AMS by spebsd@gmail.com")
	free.Size = uint32(len(free.Data))
	replaceBox(mp4Init, "free", free)

	// Get and modify MVHD Box
	var moov ParentBox
	copy(moov.Name[:], []byte{'m', 'o', 'o', 'v'})
	if mp4["moov.mvhd"] != nil {
		mvhd := mp4["moov.mvhd"][0].(MvhdBox)
		mvhd.Duration = 0
		mvhd.Timescale = 1
		replaceBox(mp4Init, "moov.mvhd", mvhd)
		moov.Size = uint32(binary.Size(mvhd))
	}
	// Get and modify TRAK Box
	var trak ParentBox
	copy(trak.Name[:], []byte{'t', 'r', 'a', 'k'})
	replaceBox(mp4Init, "moov.trak", trak)
	var tkhd TkhdBox
	if mp4["moov.trak.tkhd"] != nil {
		tkhd = mp4["moov.trak.tkhd"][0].(TkhdBox)
		tkhd.ModificationTime = 0
		tkhd.Duration = 0
		tkhd.Flags[0] = 0x00
		tkhd.Flags[1] = 0x00
		tkhd.Flags[2] = 0x07
		tkhd.AlternateGroup = 0
		replaceBox(mp4Init, "moov.trak.tkhd", tkhd)
	}
	// Get and modify MDHD Box
	var mdia ParentBox
	copy(mdia.Name[:], []byte{'m', 'd', 'i', 'a'})
	var mdhd MdhdBox
	if mp4["moov.trak.mdia.mdhd"] != nil {
		mdhd = mp4["moov.trak.mdia.mdhd"][0].(MdhdBox)
		mdhd.Duration = 0
		replaceBox(mp4Init, "moov.trak.mdia.mdhd", mdhd)
	}

	hdlr := mp4["moov.trak.mdia.hdlr"][0].(HdlrBox)
	minf := mp4["moov.trak.mdia.minf"][0].(ParentBox)
	var smhd SmhdBox
	var vmhd VmhdBox
	if mp4["moov.trak.mdia.minf.smhd"] != nil {
		smhd = mp4["moov.trak.mdia.minf.smhd"][0].(SmhdBox)
		isVideo = false
	}
	if mp4["moov.trak.mdia.minf.vmhd"] != nil {
		vmhd = mp4["moov.trak.mdia.minf.vmhd"][0].(VmhdBox)
		isVideo = true
	}
	dinf := mp4["moov.trak.mdia.minf.dinf"][0].(ParentBox)
	dref := mp4["moov.trak.mdia.minf.dinf.dref"][0].(DrefBox)

	var stbl ParentBox
	copy(stbl.Name[:], []byte{'s', 't', 'b', 'l'})
	stsd := mp4["moov.trak.mdia.minf.stbl.stsd"][0].(StsdBox)
	// Audio AAC MP4a
	if mp4["moov.trak.mdia.minf.stbl.stsd.mp4a"] != nil {
		mp4a := mp4["moov.trak.mdia.minf.stbl.stsd.mp4a"][0].(Mp4aBox)
		esds := mp4["moov.trak.mdia.minf.stbl.stsd.mp4a.esds"][0].(EsdsBox)
		esdsData := []byte{0x03, 0x19, 0x00, 0x01, 0x00, 0x04, 0x11, 0x40, 0x15, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0xF3, 0xC2, 0x05, 0x02, 0x11, 0x90, 0x06, 0x01, 0x02}
		copy(esds.Data[:], esdsData[:])
		esdsOldSize := esds.Size
		esds.Size = 31
		stsd.Size += esds.Size - esdsOldSize
		mp4a.Size += esds.Size - esdsOldSize
		esds.Version = 0
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd.mp4a.esds", esds)
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd.mp4a", mp4a)
	}
	// Video AVC1 MP4
	var avc1 Avc1Box
	var avcC AvcCBox
	var btrt BtrtBox
	if mp4["moov.trak.mdia.minf.stbl.stsd.avc1"] != nil {
		avc1 = mp4["moov.trak.mdia.minf.stbl.stsd.avc1"][0].(Avc1Box)
		reserved := [6]byte{0, 0, 0, 0, 0, 0}
		copy(avc1.Reserved[0:6], reserved[0:6])
		avc1.dataReferenceIndex = 1
		avc1.Version = 0
		avc1.RevisionLevel = 0
		avc1.Vendor = 0
		avc1.TemporalQuality = 0
		avc1.SpacialQuality = 0
		compressorName := "AVC Coding"
		avc1.CompressorName[0] = byte(len(compressorName))
		copy(avc1.CompressorName[1:], []byte(compressorName)[:])
		avcC = mp4["moov.trak.mdia.minf.stbl.stsd.avc1.avcC"][0].(AvcCBox)
		btrt.Size = 12
		btrt.DecodingBufferSize = 0
		btrt.MaxBitrate = 0
		mdat := mp4["mdat"][0].(MdatBox)
		mdhd := mp4["moov.trak.mdia.mdhd"][0].(MdhdBox)
		btrt.AvgBitrate = uint32(float64(mdat.Size) / (float64(mdhd.Duration) / float64(mdhd.Timescale)) * 8)
		avc1.Size = 78 + avcC.Size + 8 + btrt.Size + 8
		stsd.Size = 8 + avc1.Size + 8
		hdlr.Name = []byte("AMS Video Handler\x00")
		hdlr.Size = 24 + uint32(len(hdlr.Name))
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd.avc1", avc1)
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd.avc1.avcC", avcC)
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd.avc1.btrt", btrt)
	}
	replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd", stsd)
	var stts SttsBox
	stts.Size = 8
	replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stts", stts)
	var stsc StscBox
	stsc.Size = 8
	replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsc", stsc)
	var stsz StszBox
	stsz.Size = 12
	replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsz", stsz)
	var stco StcoBox
	stco.Size = 8
	replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stco", stco)

	dinf.Size = dref.Size + 8
	stbl.Size = stsd.Size + 8 + stts.Size + 8 + stsc.Size + 8 + stsz.Size + 8 + stco.Size + 8
	if isVideo == false {
		minf.Size = smhd.Size + 8 + dinf.Size + 8 + stbl.Size + 8
	} else {
		minf.Size = vmhd.Size + 8 + dinf.Size + 8 + stbl.Size + 8
	}
	mdia.Size += mdhd.Size + 8 + hdlr.Size + 8 + minf.Size + 8
	trak.Size += tkhd.Size + 8 + mdia.Size + 8

	replaceBox(mp4Init, "moov.trak.mdia.minf.dinf.dref", dref)
	replaceBox(mp4Init, "moov.trak.mdia.minf.dinf", dinf)
	if isVideo == false {
		replaceBox(mp4Init, "moov.trak.mdia.minf.smhd", smhd)
	} else {
		replaceBox(mp4Init, "moov.trak.mdia.minf.vmhd", vmhd)
	}
	replaceBox(mp4Init, "moov.trak.mdia.minf.stbl", stbl)
	replaceBox(mp4Init, "moov.trak.mdia.minf", minf)
	replaceBox(mp4Init, "moov.trak.mdia.hdlr", hdlr)
	replaceBox(mp4Init, "moov.trak.mdia", mdia)
	replaceBox(mp4Init, "moov.trak", trak)

	var mvex ParentBox
	copy(mvex.Name[:], []byte{'m', 'v', 'e', 'x'})
	var elst ElstBox
	if mp4["moov.trak.edts.elst"] != nil {
		elst = mp4["moov.trak.edts.elst"][0].(ElstBox)
		/*var mehd MehdBox
		mehd.Size = 8
		mehd.Version = 0
		mehd.Reserved = [3]byte{ 0, 0, 0 }
		mehd.FragmentDuration = uint64(elst.SegmentDuration)*/
	}

	var trex TrexBox
	trex.Size = 24
	trex.Version = 0
	trex.Reserved = [3]byte{0, 0, 0}
	trex.TrackID = 1
	trex.DefaultSampleDescriptionIndex = 1
	trex.DefaultSampleDuration = uint32(elst.MediaTime)
	trex.DefaultSampleSize = 0
	trex.DefaultSampleFlags = 0
	replaceBox(mp4Init, "moov.mvex.trex", trex)
	mvex.Size += trex.Size + 8
	moov.Size += trak.Size + mvex.Size + 8

	replaceBox(mp4Init, "moov.mvex", mvex)
	replaceBox(mp4Init, "moov", moov)

	return
}

func CreateDashFragment(mp4 map[string][]interface{}, fragmentNumber uint32, fragmentDuration uint32) (fmp4 map[string][]interface{}) {
	fmp4 = make(map[string][]interface{})
	var isVideo bool

	// Check if the mp4 is audio or video
	if mp4["moov.trak.mdia.minf.smhd"] != nil {
		isVideo = false
	}
	if mp4["moov.trak.mdia.minf.vmhd"] != nil {
		isVideo = true
	}

	// STYP
	var styp StypBox
	styp.MajorBrand = [4]byte{'i', 's', 'o', '6'}
	styp.MinorVersion = 0
	styp.CompatibleBrands = make([][4]byte, 2)
	styp.CompatibleBrands[0] = [4]byte{'i', 's', 'o', '6'}
	styp.CompatibleBrands[1] = [4]byte{'m', 's', 'd', 'h'}
	styp.Size = 16
	replaceBox(fmp4, "styp", styp)

	// FREE
	var free FreeBox
	free.Data = []byte("AMS by spebsd@gmail.com")
	free.Size = uint32(len(free.Data))
	replaceBox(fmp4, "free", free)

	// MOOF Parent Box
	var moof ParentBox
	moof.Name = [4]byte{'m', 'o', 'o', 'f'}

	// MFHD
	var mfhd MfhdBox
	mfhd.Version = 0
	mfhd.Reserved = [3]byte{0, 0, 0}
	mfhd.SequenceNumber = fragmentNumber
	mfhd.Size = 8
	replaceBox(fmp4, "moof.mfhd", mfhd)

	// TRAF ParentBox
	var traf ParentBox
	traf.Name = [4]byte{'t', 'r', 'a', 'f'}

	// TFHD
	var tfhd TfhdBox
	tfhd.Version = 0
	if isVideo == false {
		tfhd.Flags[0] = 0x02                 // ISO/IEC 14496-12:2015 0x02 default-base-is-moof
		tfhd.Flags[1] = 0x00                 // Nothing
		tfhd.Flags[2] = 0x28                 // 0x28 -> default-sample-flags-present + default-sample-duration-present
		tfhd.DefaultSampleFlags = 0x02800040 // To check ISO/IEC ????
	} else {
		tfhd.Flags[0] = 0x02
		tfhd.Flags[1] = 0x00
		tfhd.Flags[2] = 0x08
		tfhd.DefaultSampleFlags = 0
	}
	tfhd.TrackID = 1
	//elst := mp4["moov.trak.edts.elst"][0].(ElstBox)
	//tfhd.DefaultSampleDuration = uint32(elst.MediaTime)
	stts := mp4["moov.trak.mdia.minf.stbl.stts"][0].(SttsBox)
	tfhd.DefaultSampleDuration = stts.Entries[0].SampleDelta
	if isVideo == false {
		tfhd.Size = 16
	} else {
		tfhd.Size = 12
	}
	replaceBox(fmp4, "moof.traf.tfhd", tfhd)

	// TRUN
	mdhd := mp4["moov.trak.mdia.mdhd"][0].(MdhdBox)
	var trun TrunBox
	trun.Version = 0
	if isVideo == false {
		trun.Flags[0] = 0x00 // Nothing
		trun.Flags[1] = 0x02 // ISO/IEC 14496-12:2015 0x02 sample-size-present
		trun.Flags[2] = 0x01 // ISO/IEC 14496-12:2015 0x01 data-offset-present
	} else {
		trun.Flags[0] = 0x00
		trun.Flags[1] = 0x06
		trun.Flags[2] = 0x01
	}
	sampleStart := (((int64(fragmentNumber) - 1) * int64(fragmentDuration)) * int64(mdhd.Timescale)) / int64(tfhd.DefaultSampleDuration)
	sampleEnd := (((int64(fragmentNumber) * int64(fragmentDuration)) * int64(mdhd.Timescale)) / int64(tfhd.DefaultSampleDuration)) - 1

	var stss StssBox
	if isVideo == true {
		// Must match an I-Frame
		stss = mp4["moov.trak.mdia.minf.stbl.stss"][0].(StssBox)
		var i uint32
		sampleStartSet := false
		for i = 0; int64(stss.SampleNumber[i]) < sampleEnd; i++ {
			if sampleStartSet == false && int64(stss.SampleNumber[i]) > sampleStart {
				sampleStart = int64(stss.SampleNumber[i] - 1)
				sampleStartSet = true
			}
		}
		sampleEnd = int64(stss.SampleNumber[i] - 2)
	}

	trun.SampleCount = uint32(sampleEnd - sampleStart + 1)
	trun.Size = 12
	trun.Samples = make([]TrunBoxSample, trun.SampleCount)
	stsz := mp4["moov.trak.mdia.minf.stbl.stsz"][0].(StszBox)
	mdat := mp4["mdat"][0].(MdatBox)
	var i int64
	for i = 0; i < sampleStart; i++ {
		mdat.Offset += int64(stsz.EntrySize[i])
	}
	mdat.Size = 0
	// Write video
	for i = sampleStart; i <= sampleEnd; i++ {
		trun.Samples[i-sampleStart].Size = stsz.EntrySize[i]
		trun.Size += 4
		if isVideo == true {
			trun.Samples[i-sampleStart].Flags = 21037248
			trun.Size += 4
		}
		mdat.Size += stsz.EntrySize[i]
	}
	if isVideo == true {
		var i uint32
		for i = 0; int64(stss.SampleNumber[i]) < sampleEnd; i++ {
			if int64(stss.SampleNumber[i]-1) >= sampleStart {
				trun.Samples[int64(stss.SampleNumber[i])-1-sampleStart].Flags = 37748800
			}
		}
	}

	// TFDT
	var tfdt TfdtBox
	tfdt.Version = 1
	tfdt.Reserved = [3]byte{0, 0, 0}
	tfdt.BaseMediaDecodeTime = uint64(sampleStart) * uint64(tfhd.DefaultSampleDuration)
	tfdt.Size = 12
	replaceBox(fmp4, "moof.traf.tfdt", tfdt)

	traf.Size = tfhd.Size + 8 + tfdt.Size + 8 + trun.Size + 8
	moof.Size = mfhd.Size + 8 + traf.Size + 8
	trun.DataOffset = int32(moof.Size + 8 + 8)
	replaceBox(fmp4, "moof.traf.trun", trun)
	replaceBox(fmp4, "moof.traf", traf)
	replaceBox(fmp4, "moof", moof)
	replaceBox(fmp4, "mdat", mdat)

	return
}

// Create DASH init header with a config struct
func CreateDashInitWithConf(dConf DashConfig) (mp4Init map[string][]interface{}) {
	mp4Init = make(map[string][]interface{})

	// Create FTYP Box
	var ftyp FtypBox
	ftyp.MajorBrand = [4]byte{'i', 's', 'o', '6'}
	ftyp.MinorVersion = 0
	ftyp.CompatibleBrands = make([][4]byte, 2)
	ftyp.CompatibleBrands[0] = [4]byte{'i', 's', 'o', '6'}
	ftyp.CompatibleBrands[1] = [4]byte{'d', 'a', 's', 'h'}
	ftyp.Size = 16
	replaceBox(mp4Init, "ftyp", ftyp)

	// Create FREE Box
	var free FreeBox
	free.Data = []byte("AMS by spebsd@gmail.com")
	free.Size = uint32(len(free.Data))
	replaceBox(mp4Init, "free", free)

	// Create MVHD Box
	var mvhd MvhdBox
	mvhd.Version = 0
	copy(mvhd.Reserved[0:3], []byte{0, 0, 0})
	mvhd.CreationTime = 0
	mvhd.ModificationTime = 0
	mvhd.Timescale = 1
	mvhd.Duration = 0
	mvhd.Rate = dConf.Rate
	mvhd.Volume = dConf.Volume
	mvhd.Reserved2 = 0
	mvhd.Reserved3 = 0
	mvhd.Matrix = [9]int32{0x00010000, 0, 0, 0, 0x00010000, 0, 0, 0, 0x40000000}
	mvhd.PreDefined = [6]uint32{0, 0, 0, 0, 0, 0}
	mvhd.NextTrackID = 2
	mvhd.Size = 100
	replaceBox(mp4Init, "moov.mvhd", mvhd)

	// Create TRAK/TKHD Box
	var trak ParentBox
	copy(trak.Name[:], []byte{'t', 'r', 'a', 'k'})
	replaceBox(mp4Init, "moov.trak", trak)
	var tkhd TkhdBox
	tkhd.Version = 0
	tkhd.Flags = [3]byte{0x00, 0x00, 0x07} // 0x000001 Track_enabled | 0x000002 Track_in_movie | 0x000004 Track_in_preview
	tkhd.CreationTime = 0
	tkhd.ModificationTime = 0
	tkhd.TrackID = 1
	tkhd.Reserved = 0
	tkhd.Duration = 0
	tkhd.Reserved2 = 0
	tkhd.Layer = 0
	tkhd.AlternateGroup = 0
	if dConf.Type == "video" {
		tkhd.Volume = 0
	} else {
		tkhd.Volume = 0x0100
	}
	tkhd.Reserved3 = 0
	tkhd.Matrix = [9]int32{0x00010000, 0, 0, 0, 0x00010000, 0, 0, 0, 0x40000000}
	if dConf.Type == "video" {
		tkhd.Width = uint32(dConf.Video.Width) * uint32(tkhd.Matrix[0])
		tkhd.Height = uint32(dConf.Video.Height) * uint32(tkhd.Matrix[4])
	} else {
		tkhd.Width = 0
		tkhd.Height = 0
	}
	tkhd.Size = 84
	replaceBox(mp4Init, "moov.trak.tkhd", tkhd)

	// Create MDHD Box
	var mdia ParentBox
	copy(mdia.Name[:], []byte{'m', 'd', 'i', 'a'})
	var mdhd MdhdBox
	mdhd.Version = 0
	mdhd.Reserved = [3]byte{0, 0, 0}
	mdhd.CreationTime = 0
	mdhd.ModificationTime = 0
	mdhd.Timescale = dConf.Timescale
	mdhd.Duration = 0
	mdhd.Language = 0x8000 | ((uint16((dConf.Language[0]-0x60)&0x1F) << 10) | (uint16((dConf.Language[1]-0x60)&0x1F) << 5) | (uint16(dConf.Language[2]-0x60) & 0x1F))
	mdhd.PreDefined = 0
	mdhd.Size = 24
	replaceBox(mp4Init, "moov.trak.mdia.mdhd", mdhd)

	// Create HDLR Box
	var hdlr HdlrBox
	hdlr.Version = 0
	hdlr.Reserved = [3]byte{0, 0, 0}
	hdlr.PreDefined = 0
	hdlr.HandlerType = dConf.HandlerType
	hdlr.Reserved2 = [3]uint32{0, 0, 0}
	if dConf.Type == "video" {
		hdlr.Name = []byte("AMS Video Handler\x00")
	} else {
		hdlr.Name = []byte("AMS Audio Handler\x00")
	}
	hdlr.Size = 24 + uint32(len(hdlr.Name))
	replaceBox(mp4Init, "moov.trak.mdia.hdlr", hdlr)

	// Create MINF Box
	var minf ParentBox
	minf.Name = [4]byte{'m', 'i', 'n', 'f'}

	// Create DREF Box
	var dref DrefBox
	dref.Version = 0
	dref.Reserved = [3]byte{0, 0, 0}
	dref.EntryCount = 1
	dref.UrlBox = make([]DrefUrlBox, 1)
	dref.UrlBox[0].Location = ""
	dref.UrlBox[0].Version = 0
	dref.UrlBox[0].Flags = [3]byte{0, 0, 1}
	dref.UrlBox[0].Size = 12
	dref.Size = 20
	replaceBox(mp4Init, "moov.trak.mdia.minf.dinf.dref", dref)

	// Create DINF Box
	var dinf ParentBox
	dinf.Name = [4]byte{'d', 'i', 'n', 'f'}
	dinf.Size = dref.Size + 8
	replaceBox(mp4Init, "moov.trak.mdia.minf.dinf", dinf)

	// Create STTS Box
	var stts SttsBox
	stts.Size = 8
	replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stts", stts)

	// Create STSC Box
	var stsc StscBox
	stsc.Size = 8
	replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsc", stsc)

	// Create STSZ Box
	var stsz StszBox
	stsz.Size = 12
	replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsz", stsz)

	// Create STCO Box
	var stco StcoBox
	stco.Size = 8
	replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stco", stco)

	// Create STSD Box
	var stsd StsdBox
	stsd.Version = 0
	stsd.Reserved = [3]byte{0, 0, 0}
	stsd.EntryCount = 1

	// Create MP4A Box if Type == "audio"
	if dConf.Type == "audio" {
		var smhd SmhdBox
		smhd.Version = 0
		smhd.Reserved = [3]byte{0, 0, 0}
		smhd.Balance = 0
		smhd.Reserved2 = 0
		smhd.Size = 8
		replaceBox(mp4Init, "moov.trak.mdia.minf.smhd", smhd)

		var esds EsdsBox
		esds.Data = []byte{0x03, 0x19, 0x00, 0x01, 0x00, 0x04, 0x11, 0x40, 0x15, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0xF3, 0xC2, 0x05, 0x02, 0x11, 0x90, 0x06, 0x01, 0x02}
		esds.Size = 31
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd.mp4a.esds", esds)

		var mp4a Mp4aBox
		mp4a.Reserved = [6]byte{0, 0, 0, 0, 0, 0}
		mp4a.DataReferenceIndex = 1
		mp4a.Version = 0
		mp4a.RevisionLevel = 0
		mp4a.Vendor = 0
		mp4a.NumberOfChannels = dConf.Audio.NumberOfChannels
		mp4a.SampleSize = dConf.Audio.SampleSize
		mp4a.CompressionId = dConf.Audio.CompressionId
		mp4a.Reserved2 = 0
		mp4a.SampleRate = dConf.Audio.SampleRate
		mp4a.Size = 28 + esds.Size + 8
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd.mp4a", mp4a)

		stsd.Size = 8 + mp4a.Size + 8

		// Create STBL Box
		var stbl ParentBox
		stbl.Name = [4]byte{'s', 't', 'b', 'l'}
		stbl.Size = stsd.Size + 8 + stts.Size + 8 + stsc.Size + 8 + stsz.Size + 8 + stco.Size + 8
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl", stbl)

		minf.Size = smhd.Size + 8 + dinf.Size + 8 + stbl.Size + 8
	}

	// Create AVC1 Box if Type == "video"
	if dConf.Type == "video" {
		var vmhd VmhdBox
		vmhd.Version = 0
		vmhd.Reserved = [3]byte{0, 0, 1}
		vmhd.GraphicsMode = 0
		vmhd.OpColor = [3]uint16{0, 0, 0}
		vmhd.Size = 12
		replaceBox(mp4Init, "moov.trak.mdia.minf.vmhd", vmhd)

		var avcC AvcCBox
		avcC.ConfigurationVersion = 1
		avcC.AVCProfileIndication = dConf.Video.CodecInfo[0]
		avcC.ProfileCompatibility = dConf.Video.CodecInfo[1]
		avcC.AVCLevelIndication = dConf.Video.CodecInfo[2]
		avcC.NalUnitSize = dConf.Video.NalUnitSize
		avcC.SPSEntryCount = dConf.Video.SPSEntryCount
		avcC.SPSSize = dConf.Video.SPSSize
		avcC.SPSData = dConf.Video.SPSData
		avcC.PPSEntryCount = dConf.Video.PPSEntryCount
		avcC.PPSSize = dConf.Video.PPSSize
		avcC.PPSData = dConf.Video.PPSData
		avcC.Size = 38
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd.avc1.avcC", avcC)

		var btrt BtrtBox
		btrt.DecodingBufferSize = 0
		btrt.MaxBitrate = 0
		btrt.AvgBitrate = uint32(float64(dConf.MdatBoxSize) / (float64(dConf.Duration) / float64(dConf.Timescale)) * 8)
		btrt.Size = 12
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd.avc1.btrt", btrt)

		var avc1 Avc1Box
		avc1.Reserved = [6]byte{0, 0, 0, 0, 0, 0}
		avc1.dataReferenceIndex = 1
		avc1.Version = 0
		avc1.RevisionLevel = 0
		avc1.Vendor = 0
		avc1.TemporalQuality = 0
		avc1.SpacialQuality = 0
		avc1.Width = dConf.Video.Width
		avc1.Height = dConf.Video.Height
		avc1.HorizontalResolution = dConf.Video.HorizontalResolution
		avc1.VerticalResolution = dConf.Video.VerticalResolution
		avc1.EntryDataSize = 0
		avc1.FramesPerSample = 1
		compressorName := "AVC Coding"
		avc1.CompressorName[0] = byte(len(compressorName))
		copy(avc1.CompressorName[1:], []byte(compressorName)[:])
		avc1.BitDepth = dConf.Video.BitDepth
		avc1.ColorTableIndex = dConf.Video.ColorTableIndex
		avc1.Size = 78 + avcC.Size + 8 + btrt.Size + 8
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd.avc1", avc1)

		stsd.Size = 8 + avc1.Size + 8

		// Create STBL Box
		var stbl ParentBox
		stbl.Name = [4]byte{'s', 't', 'b', 'l'}
		stbl.Size = stsd.Size + 8 + stts.Size + 8 + stsc.Size + 8 + stsz.Size + 8 + stco.Size + 8
		replaceBox(mp4Init, "moov.trak.mdia.minf.stbl", stbl)

		minf.Size = vmhd.Size + 8 + dinf.Size + 8 + stbl.Size + 8
	}

	mdia.Size += mdhd.Size + 8 + hdlr.Size + 8 + minf.Size + 8
	trak.Size += tkhd.Size + 8 + mdia.Size + 8

	replaceBox(mp4Init, "moov.trak.mdia.minf.stbl.stsd", stsd)
	replaceBox(mp4Init, "moov.trak.mdia.minf", minf)
	replaceBox(mp4Init, "moov.trak.mdia", mdia)
	replaceBox(mp4Init, "moov.trak", trak)

	// Create TREX Box
	var trex TrexBox
	trex.Size = 24
	trex.Version = 0
	trex.Reserved = [3]byte{0, 0, 0}
	trex.TrackID = 1
	trex.DefaultSampleDescriptionIndex = 1
	trex.DefaultSampleDuration = dConf.SampleDelta
	trex.DefaultSampleSize = 0
	trex.DefaultSampleFlags = 0
	replaceBox(mp4Init, "moov.mvex.trex", trex)

	// Create MVEX Box
	var mvex ParentBox
	copy(mvex.Name[:], []byte{'m', 'v', 'e', 'x'})
	mvex.Size += trex.Size + 8
	replaceBox(mp4Init, "moov.mvex", mvex)

	// And finaly MOOV Box
	var moov ParentBox
	copy(moov.Name[:], []byte{'m', 'o', 'o', 'v'})
	moov.Size = mvhd.Size + 8 + trak.Size + 8 + mvex.Size + 8
	replaceBox(mp4Init, "moov", moov)

	return
}

func CreateDashFragmentWithConf(dConf DashConfig, filename string, fragmentNumber uint32, fragmentDuration uint32) (fmp4 map[string][]interface{}) {
	lastSegment := false
	compositionTimeOffset := false

	f, err := os.Open(filename)
	if err != nil {
		return
	}
	fmp4 = make(map[string][]interface{})

	// FREE
	var free FreeBox
	free.Data = []byte("AMS by spebsd@gmail.com")
	free.Size = uint32(len(free.Data))
	replaceBox(fmp4, "free", free)

	// MOOF Parent Box
	var moof ParentBox
	moof.Name = [4]byte{'m', 'o', 'o', 'f'}

	// MFHD
	var mfhd MfhdBox
	mfhd.Version = 0
	mfhd.Reserved = [3]byte{0, 0, 0}
	mfhd.SequenceNumber = fragmentNumber
	mfhd.Size = 8
	replaceBox(fmp4, "moof.mfhd", mfhd)

	// TRAF ParentBox
	var traf ParentBox
	traf.Name = [4]byte{'t', 'r', 'a', 'f'}

	// TFHD
	var tfhd TfhdBox
	tfhd.Version = 0
	if dConf.Type == "audio" {
		tfhd.Flags[0] = 0x02 // ISO/IEC 14496-12:2015 0x02 default-base-is-moof
		tfhd.Flags[1] = 0x00 // Nothing
		tfhd.Flags[2] = 0x28 // 0x28 -> default-sample-flags-present + default-sample-duration-present
		tfhd.DefaultSampleFlags = 0x02800040
		tfhd.Size = 16
	} else {
		if dConf.Type == "video" {
			tfhd.Flags[0] = 0x02
			tfhd.Flags[1] = 0x00
			tfhd.Flags[2] = 0x08
			tfhd.DefaultSampleFlags = 0
			tfhd.Size = 12
		} else {
			fmp4 = nil
			return
		}
	}
	tfhd.TrackID = 1
	tfhd.DefaultSampleDuration = dConf.SampleDelta
	replaceBox(fmp4, "moof.traf.tfhd", tfhd)

	mp4 := make(map[string][]interface{})
	var ctts CttsBox
	if dConf.Type == "video" && dConf.Video.CttsBoxOffset != 0 {
		f.Seek(dConf.Video.CttsBoxOffset, 0)
		readCttsBox(f, dConf.Video.CttsBoxSize, 0, "moov.trak.mdia.minf.stbl.ctts", mp4)
		ctts = mp4["moov.trak.mdia.minf.stbl.ctts"][0].(CttsBox)
		compositionTimeOffset = true
	}

	// TRUN
	var trun TrunBox
	trun.Version = 0
	if dConf.Type == "audio" {
		trun.Flags[0] = 0x00 // Nothing
		trun.Flags[1] = 0x02 // ISO/IEC 14496-12:2015 0x02 sample-size-present
		trun.Flags[2] = 0x01 // ISO/IEC 14496-12:2015 0x01 data-offset-present
	} else {
		if dConf.Type == "video" {
			if compositionTimeOffset == true {
				trun.Flags[0] = 0x00
				trun.Flags[1] = 0x08 | 0x04 | 0x02 // sample-composition-time-offsets-present & sample-flags-present & sample-size-present
				trun.Flags[2] = 0x01
				trun.Version = 1
			} else {
				trun.Flags[0] = 0x00
				trun.Flags[1] = 0x04 | 0x02 // sample-flags-present & sample-size-present
				trun.Flags[2] = 0x01
			}
		} else {
			fmp4 = nil
			return
		}
	}

	sampleStart := uint32((((float64(fragmentNumber) - 1) * float64(fragmentDuration)) * float64(dConf.Timescale)) / float64(dConf.SampleDelta))
	sampleEnd := uint32(((float64(fragmentNumber) * float64(fragmentDuration)) * float64(dConf.Timescale)) / float64(dConf.SampleDelta))

	// Search Positions in STSS Box
	var stss StssBox
	var iFramesToSet []uint32
	if dConf.Type == "video" {
		f.Seek(dConf.Video.StssBoxOffset, 0)
		readStssBox(f, dConf.Video.StssBoxSize, 0, "moov.trak.mdia.minf.stbl.stss", mp4)
		// Must match an I-Frame
		stss = mp4["moov.trak.mdia.minf.stbl.stss"][0].(StssBox)
		var i uint32
		sampleStartSet := false
		for i = 0; (i < stss.EntryCount) && ((stss.SampleNumber[i] - 1) < sampleEnd); i++ {
			if (stss.SampleNumber[i] - 1) >= sampleStart {
				if sampleStartSet == false {
					sampleStart = stss.SampleNumber[i] - 1
					sampleStartSet = true
				}
				iFramesToSet = append(iFramesToSet, stss.SampleNumber[i]-1-sampleStart)
			}
		}
		if i < stss.EntryCount {
			sampleEnd = stss.SampleNumber[i] - 1
		} else {
			lastSegment = true
		}
	}
	sampleEnd--

	// Read STSZ Box
	f.Seek(dConf.StszBoxOffset, 0)
	var stszSize uint32
	stszSize = 12 + ((sampleEnd + 1) * 4)
	if stszSize > dConf.StszBoxSize {
		stszSize = dConf.StszBoxSize
	}
	readStszBox(f, stszSize, 0, "moov.trak.mdia.minf.stbl.stsz", mp4)
	stsz := mp4["moov.trak.mdia.minf.stbl.stsz"][0].(StszBox)
	if sampleEnd > (stsz.SampleCount - 1) {
		sampleEnd = stsz.SampleCount - 1
	}
	trun.SampleCount = uint32(sampleEnd - sampleStart + 1)
	trun.Size = 12
	trun.Samples = make([]TrunBoxSample, trun.SampleCount)
	var cttsOffset uint32
	var cttsSampleCount uint32
	cttsOffset = 0
	cttsSampleCount = 0
	var mdat MdatBox
	mdat.Offset = dConf.MdatBoxOffset
	mdat.Size = 0
	mdat.Filename = filename
	var i uint32

	// Get to the first element, count the number of ctts sample offset
	for i = 0; i < sampleStart; i++ {
		if stsz.SampleSize == 0 {
			mdat.Offset += int64(stsz.EntrySize[i])
		} else {
			mdat.Offset += int64(stsz.SampleSize)
		}

		// Getting the number of ctts offset
		if compositionTimeOffset == true {
			if cttsSampleCount > 0 {
				cttsSampleCount--
				if cttsSampleCount == 0 {
					cttsOffset++
				}
			} else {
				cttsSampleCount = ctts.Entries[cttsOffset].SampleCount - 1
				if cttsSampleCount == 0 {
					cttsOffset++
				}
			}
		}
	}
	var size uint32
	var lastCompositionTimeOffset int64
	lastCompositionTimeOffset = 0
	// Copy samples
	for i = sampleStart; i <= sampleEnd; i++ {
		if stsz.SampleSize == 0 {
			size = stsz.EntrySize[i]
		} else {
			size = stsz.SampleSize
		}
		trun.Samples[i-sampleStart].Size = size
		trun.Size += 4
		if dConf.Type == "video" {
			trun.Samples[i-sampleStart].Flags = 21037248
			trun.Size += 4
			if compositionTimeOffset == true {
				if lastCompositionTimeOffset != 0 {
					trun.Samples[i-sampleStart].Flags = 25231552
				}

				// Retrieve the last composition time offset
				trun.Samples[i-sampleStart].CompositionTimeOffset = int64(ctts.Entries[cttsOffset].SampleOffset) - dConf.MediaTime
				// If (bFrame scheme) reordering?
				if trun.Samples[i-sampleStart].CompositionTimeOffset > 0 {
					lastCompositionTimeOffset = trun.Samples[i-sampleStart].CompositionTimeOffset
				} else {
					// Set the lastCompositionOffset to the right time
					lastCompositionTimeOffset += trun.Samples[i-sampleStart].CompositionTimeOffset
				}
				trun.Size += 4
				if cttsSampleCount > 0 {
					cttsSampleCount--
					if cttsSampleCount == 0 {
						cttsOffset++
					}
				} else {
					cttsSampleCount = ctts.Entries[cttsOffset].SampleCount - 1
					if cttsSampleCount == 0 {
						cttsOffset++
					}
				}
			}
		}

		// Update the size of the element
		mdat.Size += size
	}
	if dConf.Type == "video" {
		for _, iframe := range iFramesToSet {
			trun.Samples[iframe].Flags = 37748800
		}
	}

	// TFDT
	var tfdt TfdtBox
	tfdt.Version = 1
	tfdt.Reserved = [3]byte{0, 0, 0}
	tfdt.BaseMediaDecodeTime = uint64(sampleStart) * uint64(tfhd.DefaultSampleDuration)
	tfdt.Size = 12
	replaceBox(fmp4, "moof.traf.tfdt", tfdt)

	// for loop to set each trun.Samples[X] from moov.trak.mdia.minf.stbl.stsz
	// trun.Samples[X].Duration = XXX trun.Samples[X].Size = XXX trun.Samples[X].Flags = XXX trun.Samples[X].CompositionTimeOffset = XXX
	traf.Size = tfhd.Size + 8 + tfdt.Size + 8 + trun.Size + 8
	moof.Size = mfhd.Size + 8 + traf.Size + 8
	trun.DataOffset = int32(moof.Size + 8 + 8)
	replaceBox(fmp4, "moof.traf.trun", trun)
	replaceBox(fmp4, "moof.traf", traf)
	replaceBox(fmp4, "moof", moof)
	replaceBox(fmp4, "mdat", mdat)

	// STYP
	var styp StypBox
	styp.MajorBrand = [4]byte{'i', 's', 'o', '6'}
	styp.MinorVersion = 0
	if lastSegment == true {
		styp.CompatibleBrands = make([][4]byte, 3)
		styp.CompatibleBrands[2] = [4]byte{'l', 'm', 's', 'g'}
		styp.Size = 20
	} else {
		styp.CompatibleBrands = make([][4]byte, 2)
		styp.Size = 16
	}
	styp.CompatibleBrands[0] = [4]byte{'i', 's', 'o', '6'}
	styp.CompatibleBrands[1] = [4]byte{'m', 's', 'd', 'h'}
	replaceBox(fmp4, "styp", styp)

	return
}

// ***
// *** Package initialization
// ***

func init() {
	debugMode = false
	funcBoxes = map[string]interface{}{
		"ftyp":                                         readFtypBox,
		"styp":                                         readStypBox,
		"free":                                         readFreeBox,
		"moov":                                         readBoxes,
		"moov.mvhd":                                    readMvhdBox,
		"moov.trak":                                    readBoxes,
		"moov.trak.tkhd":                               readTkhdBox,
		"moov.trak.edts":                               readBoxes,
		"moov.trak.edts.elst":                          readElstBox,
		"moov.trak.mdia":                               readBoxes,
		"moov.trak.mdia.mdhd":                          readMdhdBox,
		"moov.trak.mdia.hdlr":                          readHdlrBox,
		"moov.trak.mdia.minf":                          readBoxes,
		"moov.trak.mdia.minf.vmhd":                     readVmhdBox,
		"moov.trak.mdia.minf.smhd":                     readSmhdBox,
		"moov.trak.mdia.minf.hmhd":                     readVmhdBox,
		"moov.trak.mdia.minf.dinf":                     readBoxes,
		"moov.trak.mdia.minf.dinf.dref":                readDrefBox,
		"moov.trak.mdia.minf.stbl":                     readBoxes,
		"moov.trak.mdia.minf.stbl.stts":                readSttsBox,
		"moov.trak.mdia.minf.stbl.ctts":                readCttsBox,
		"moov.trak.mdia.minf.stbl.stsd":                readStsdBox,
		"moov.trak.mdia.minf.stbl.stsd.mp4a":           readMp4aBox,
		"moov.trak.mdia.minf.stbl.stsd.mp4a.esds":      readEsdsBox,
		"moov.trak.mdia.minf.stbl.stsd.avc1":           readAvc1Box,
		"moov.trak.mdia.minf.stbl.stsd.avc1.avcC":      readAvcCBox,
		"moov.trak.mdia.minf.stbl.stsd.avc1.btrt":      readBtrtBox,
		"moov.trak.mdia.minf.stbl.stsd.encv":           readAvc1Box,
		"moov.trak.mdia.minf.stbl.stsd.encv.avcC":      readAvcCBox,
		"moov.trak.mdia.minf.stbl.stsd.encv.btrt":      readBtrtBox,
		"moov.trak.mdia.minf.stbl.stsd.encv.sinf":      readBoxes,
		"moov.trak.mdia.minf.stbl.stsd.encv.sinf.frma": readFrmaBox,
		"moov.trak.mdia.minf.stbl.stsd.encv.sinf.schm": readSchmBox,
		"moov.trak.mdia.minf.stbl.stsd.encv.sinf.schi": readBoxes,
		"moov.trak.mdia.minf.stbl.stsc":                readStscBox,
		"moov.trak.mdia.minf.stbl.stsz":                readStszBox,
		"moov.trak.mdia.minf.stbl.sdtp":                readSdtpBox,
		"moov.trak.mdia.minf.stbl.stco":                readStcoBox,
		"moov.trak.mdia.minf.stbl.stss":                readStssBox,
		"moov.mvex":                                    readBoxes,
		"moov.mvex.mehd":                               readMehdBox,
		"moov.mvex.trex":                               readTrexBox,
		"moov.udta":                                    readBoxes,
		"moof":                                         readBoxes,
		"moof.mfhd":                                    readMfhdBox,
		"moof.traf":                                    readBoxes,
		"moof.traf.tfhd":                               readTfhdBox,
		"moof.traf.trun":                               readTrunBox,
		"moof.traf.tfdt":                               readTfdtBox,
		"mfra":                                         readBoxes,
		"skip":                                         readBoxes,
		"skip.udta":                                    readBoxes,
		"skip.udta.cprt":                               readBoxes,
		"meta":                                         readBoxes,
		"meta.dinf":                                    readBoxes,
		"meta.ipro":                                    readBoxes,
		"meta.ipro.sinf":                               readBoxes,
		"meta.ipro.sinf.frma":                          readFrmaBox,
		"meta.flin":                                    readBoxes,
		"meta.flin.paen":                               readBoxes,
		"meco":                                         readBoxes,
		"mdat":                                         readMdatBox,
	}
}
