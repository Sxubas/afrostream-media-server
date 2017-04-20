package main

import (
	"ts"
	"fmt"
	"os"
)

func main() {
	//data := ts.NewData(4)
	//data.PushUInt(0x03, 2)

	pat := ts.NewPAT()
	pat.ToBytes().PrintHexFull()

	//mp4m := mp4.ParseFile("small.mp4", "en")
	//fragment := ts.CreateHLSFragment(mp4m.Boxes, 1, 2)
	//printFragments(fragment, 1)
	//writeBytes("sample.ts", fragment)
	//ts.PrintFragments(&fragments, 10)
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