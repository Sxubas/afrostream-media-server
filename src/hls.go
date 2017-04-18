package main

import (
	"mp4"
	"ts"
	"fmt"
)



func main() {
	mp4m := mp4.ParseFile("small.mp4", "en")
	ts.CreateHLSFragment(mp4m.Boxes, 1, 2)
	pmt := ts.Header{}
	fmt.Printf("% 8X\n", pmt.ToBytes())
	/*for i := 0; i < len(fragment); i++ {
		fmt.Printf("% 8X\n", fragment[i].ToBytes())
	}*/


	//ts.PrintFragments(&fragments, 10)
}

