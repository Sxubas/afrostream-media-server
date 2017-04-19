package main

import (
	"ts"
	"mp4"
	"fmt"
)

func main() {
	mp4m := mp4.ParseFile("small.mp4", "en")
	fragment := ts.CreateHLSFragment(mp4m.Boxes, 1, 2)
	printFragments(fragment, 4)

	//ts.PrintFragments(&fragments, 10)
}

func printFragments(fragments []ts.Bytes, max int) {
	for i := 0; i < ts.Min(len(fragments), max); i++ {
		fmt.Printf("\nPacket-%d\n", i)
		fragments[i].ToBytes().PrintHexFull()
	}
}