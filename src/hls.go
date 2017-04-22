package main

import (
	"ts"
	"fmt"
	"os"
	"mp4"
)

func debug() {
	debug := ts.NewStartStreamDebug(63000)
	debug.Payload.Data = []byte{0x00, 0x00, 0x01, 0xe0, 0x00, 0x00, 0x80, 0xc0}
	debug.ToBytes().PrintHexFull()
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
	mp4m := mp4.ParseFile("small.mp4", "en")
	fragment := ts.CreateHLSFragmentDebug(mp4m.Boxes, 1, 10)
	//printFragments(fragment, 10)
	writeBytes("sample.ts", fragment)
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