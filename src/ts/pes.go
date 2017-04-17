package ts

type PES struct {
	Packet
}

type PESSection struct {
	PacketStartCodePrefix uint32; // 24b
	StreamId byte; // 1Byte
	PES_PacketLength uint16; // 2Bytes
}

type OptionalPESHeader struct {
	PES_ScramblingControl byte; // 2Byte
	PES_Priority byte;
	DataAlignmentIndicator byte;
	Copyright byte;
	OriginalOrCopy byte;
	PTS_DTS_Flags byte; // 2b
	ESCR_Flag byte;
	ES_RateFlag byte;
	DSM_TrickModeFlag byte;
	AdditionalCopyInfoGlag byte;
	PES_CRC_Flag byte;
	PES_ExtensionFlag byte;
	PES_HeaderDataLength byte; // 1Byte
	PTS_DTS [10]byte; // 10Bytes
}


// To bytes
func (pes PES) Bytes() (data Data) {
	data = *NewData(188)
	// PID depends on stream


	return
}

func (section PESSection) Bytes() (data Data) {
	return
}