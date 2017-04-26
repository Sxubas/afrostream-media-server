package ts

// Variables data used to create our fragment
type FragmentData struct {
	pes []PES

	PCR_Emitter IEmitter
	DTS_Emitter IEmitter
	CTS_Emitter IEmitter

	pat PAT
	pmt PMT

	PAT_Emitter IEmitter
	PMT_Emitter IEmitter
}

type Emitter interface {
	ResetEmitter()
	HasEmitLeft()
	Emit()
}

type IEmitter struct {
	Emitter
	NEmit int
	Min_emit int
}


func (emitter IEmitter) ResetEmitter() {
	emitter.NEmit = emitter.Min_emit
}

func (emitter IEmitter) Emit() {
	emitter.NEmit -= 1
}

func (emitter IEmitter) HasEmitLeft() (bool) {
	return emitter.NEmit > 0
}


func NewDefaultFragment() (fragmentData FragmentData) {
	return
}
