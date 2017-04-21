package main

import (
	"ts"
	"fmt"
	"os"
	"mp4"
)

func main() {
	startStream := ts.NewPes()
	startStream.PID = 256
	startStream.PCR_Flag = 1
	startStream.PayloadUnitStartIndicator = 1
	startStream.RandomAccessIndicator = 1
	startStream.PCR.BaseMediaDecodeTime = 6300 //baseMediaDecodeTime
	startStream.AdaptationFieldControl = 0x03      // Adaptation field only, no payload
	startStream.AdaptationFieldLength = 7
	bytes := startStream.ToBytes()
	bytes.PrintHexFull()

	mp4m := mp4.ParseFile("small.mp4", "en")
	fragment := ts.CreateHLSFragment(mp4m.Boxes, 1, 10)
	printFragments(fragment, 10)
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