package ts

type PES struct {
	Packet
	OptionalHeader *OptionalPESHeader
	Section        PESSection
	crc32          uint32 // 4Bytes
}

type PESSection struct {
	Bytes
	PacketStartCodePrefix uint32 // 24b
	StreamId              byte   // 1Byte
	PES_PacketLength      uint16 // 2Bytes
}

type OptionalPESHeader struct {
	Bytes
	PES_ScramblingControl  byte // 2Byte
	PES_Priority           byte
	DataAlignmentIndicator byte
	Copyright              byte
	OriginalOrCopy         byte
	PTS_DTS_Flags          byte // 2b
	ESCR_Flag              byte
	ES_RateFlag            byte
	DSM_TrickModeFlag      byte
	AdditionalCopyInfoGlag byte
	PES_CRC_Flag           byte
	PES_ExtensionFlag      byte
	PES_HeaderDataLength   byte     // 1Byte
	PTS_DTS                [10]byte // 10Bytes
}

// To bytes
func (pes PES) ToBytes() (data Data) {
	data = pes.Packet.ToBytes()

	// Optional Header
	if pes.OptionalHeader != nil {
		data.PushBytes(pes.OptionalHeader)
	}

	// PES Section
	//data.PushBytes(pes.Section)

	if pes.HasPayload() {
		// Push payload
 		data.PushBytes(pes.Payload)
	}

	// Fill remaining bytes with 0xff
	data.FillRemaining(0xff)

	return
}

func (section PESSection) ToBytes() (data Data) {
	data = *NewData(6)

	data.PushObj(section.PacketStartCodePrefix, 24)
	data.PushObj(section.StreamId, 8)
	data.PushObj(section.PES_PacketLength, 8)

	return
}

func NewPes() (pes *PES) {
	pes = new(PES)
	return
}
