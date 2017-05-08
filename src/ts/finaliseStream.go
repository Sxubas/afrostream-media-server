package ts

func FinaliseFragment(data *FragmentData) (bytes []byte) {

	bytes = make([]byte, 0)

	pmt := data.pmt.ToBytes().Data
	pat := data.pat.ToBytes().Data
	patEmitter := data.PAT_Emitter
	pmtEmitter := data.PMT_Emitter

	patEmitter.Min_emit = 200
	pmtEmitter.Min_emit = 40

	bytes = append(bytes, pat...)
	bytes = append(bytes, pmt...)

	for i, packet := range data.pes {

		if patEmitter.Emit() {
			//fmt.Println("patEmit")
			//bytes = append(bytes, pat...)
			patEmitter.Reset()
		}

		if pmtEmitter.Emit() {
			//fmt.Println("pmtEmit")
			//bytes = append(bytes, pmt...)
			pmtEmitter.Reset()
		}

		//fmt.Println("PES")
		//fmt.Println(i)
		packet.ContinuityCounter = byte(i % 16)
		bytes = append(bytes, packet.ToBytes().Data...)
	}

	return
}

