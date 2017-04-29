package ts


func FinaliseFragment(data *FragmentData) (bytes []byte) {

	bytes = make([]byte, 0)

	pmt := data.pmt.ToBytes().Data
	pat := data.pat.ToBytes().Data
	patEmitter := data.PAT_Emitter
	pmtEmitter := data.PMT_Emitter

	patEmitter.Min_emit = 200
	pmtEmitter.Min_emit = 40

	for _, packet := range data.pes {

		if patEmitter.Emit() {
			bytes = append(bytes, pat...)
			patEmitter.Reset()
		}

		if pmtEmitter.Emit() {
			bytes = append(bytes, pmt...)
			pmtEmitter.Reset()
		}

		bytes = append(bytes, packet.ToBytes().Data...)
	}

	return
}

