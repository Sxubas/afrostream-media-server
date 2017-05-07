package main

import (
	"ts"
	"fmt"
	"os"
	"mp4"
	"encoding/json"
)

func debug() {
}

func main() {
	isDebug := false
	if isDebug {
		debug()
	} else {
		launch()
	}
}


func launch() {
	data, _ := readFile1("samll_audio" + ".json")
	var jConfig mp4.JsonConfig
	json.Unmarshal(data, &jConfig)

	track := jConfig.Tracks["audio"][0]
	fragment := ts.CreateHLSFragmentWithConf(*track.Config, track.File, 1, jConfig.SegmentDuration)
	writeSample("sample.ts", fragment)

}

func printFragments(fragment []ts.Bytes, max int) {
	for i := 0; i < ts.Min(len(fragment), max); i++ {
		fmt.Printf("\nPacket-%d\n", i + 1)
		fragment[i].ToBytes().PrintHexFull()
	}
}

func writeBytes(filename string, fragment []ts.Bytes) {
	f, _ := os.Create(filename)

	defer f.Close()

	for i := 0; i < len(fragment); i++ {
		f.Write(fragment[i].ToBytes().Data)
	}
}


func writeSample(filename string, fragment []byte) {
	f, _ := os.Create(filename)

	defer f.Close()

	f.Write(fragment)
}

func readFile1(filename string) (data []byte, r error) {
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
