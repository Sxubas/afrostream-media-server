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
	Reset()
	HasEmitLeft()
	Emit()
}

type IEmitter struct {
	Emitter
	NEmit int
	Min_emit int
}


func (emitter IEmitter) Reset() {
	emitter.NEmit = emitter.Min_emit
}

func (emitter IEmitter) Emit() (bool) {
	if emitter.HasEmitLeft() {
		emitter.NEmit -= 1
		return false
	}
	return true
}

func (emitter IEmitter) HasEmitLeft() (bool) {
	return emitter.NEmit > 0
}


func NewDefaultFragment() (fragmentData FragmentData) {
	return
}
