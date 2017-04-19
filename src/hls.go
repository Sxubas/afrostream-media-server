package main

import (
	"ts"
	"mp4"
	"fmt"
)

func main() {
	mp4m := mp4.ParseFile("small.mp4", "en")
	fragment := ts.CreateHLSFragment(mp4m.Boxes, 1, 2)

	for i := 0; i < len(fragment); i++ {
		fmt.Printf("\nPacket-%d\n", i)
		fragment[i].ToBytes().PrintHexFull()
	}

	//ts.PrintFragments(&fragments, 10)
}

