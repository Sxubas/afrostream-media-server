package ts

type PES struct {
	Packet
	OptionalHeader *OptionalPESHeader
	Section        PESSection
	crc32          uint32 // 4Bytes
}

type PESSection struct {
	Bytes
	PacketStartCodePrefix uint32  // 24b
	StreamId              uint32  // 1Byte
	PES_PacketLength      uint16  // 2Bytes
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

	if pes.HasPayload() {
		if pes.Payload.Size() > data.GetSpaceLeftInByte() {
			panic("Not enough space to write payload")
		}
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

func (section PESSection) Size() int {
	return int(6 + section.PES_PacketLength)
}

func NewPes() (pes *PES) {
	pes = new(PES)
	return
}

func NewStream(PID uint16) (pes *PES) {
	pes = NewPes()
	pes.PID = PID
	return
}

func NewStartStream(PID uint16, streamId uint32) (pes *PES) {
	pes = NewPes()

	pes.PID = PID

	pes.AdaptationFieldControl = 0x03 // Adaptation field + payload

	pes.PayloadUnitStartIndicator = 1
	pes.Section.StreamId = streamId

	return
}

