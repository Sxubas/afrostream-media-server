package ts

type Frame struct {
	FrameType int
	Pos int
	Frame uint32
	Time uint32
	pts uint32
	dts uint32
}

type PTS struct {
	PTS uint32
	pos int
	dts uint32
	spos int
}
