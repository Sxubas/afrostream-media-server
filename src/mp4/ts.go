package mp4


import "fmt"

// Main packets
type PES struct {
	Header Header;
	AdaptationField AdaptationField;
	Payload []byte;
}

type PMT struct {
	Header Header;
	AdaptationField AdaptationField;
	PointField byte; // 1Byte
	Sections []ProgramMapSection;
}

type PAT struct {
	Header Header;
	AdaptationField AdaptationField;
	PointField byte; // 1Byte
	Sections []ProgramAssociationSection;
}


// General parts
type Header struct {
	TransportErrorIndicator byte;
	PayloadUnitStartIndicator byte;
	TransportPriority byte;
	PID uint16; // 13b
	TransportScramblingControl byte; // 2b
	AdaptationFieldControl byte; // 2b
	ContinuityCounter byte; // 4b
}

type AdaptationField struct {
	AdaptationFieldLength byte; // 1Byte
	DiscontinuityIndicator byte;
	RandomAccessIndicator byte;
	ElementaryStreamPriorityIndicator byte;
	PCR_Flag byte;
	OPCR_Flag byte;
	SplicingPointFlag byte;
	TransportPrivateDataFlag byte;
	AdaptationFieldExtensionFlag byte;
	PCR PCR;
}

type PCR struct {
	ProgramClockReferenceBase uint64; // 33b;
	Reserved byte; // 6b;
	ProgramClockReferenceExtension uint16; // 9b
}

// Specifics parts
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

type ProgramAssociationSection struct {
	TableID byte; // 1Byte
	SectionSyntaxIndicator byte;
	SectionLength uint16; // 12b;
	TransportStreamID byte; // 2b
	VersionNumber byte; // 5b
	CurrentNextIndicator byte;
	SectionNumber byte; // 1Byte
	LastSectionNumber byte; // 1Byte
	Sections []ProgramAssociationSubSection;
	crc32 byte; // 4b
}

type ProgramAssociationSubSection struct {
	ProgramNumber byte; // 2b
	ProgramMapID uint16; // 13b
}

type ProgramMapSection struct {
	TableId byte; // 1Byte
	SectionSyntaxIndicator byte;
	SectionLength uint16; // 12b
	ProgramNumber uint16; // 2Bytes
	VersionNumber byte; // 5b
	CurrentNextIndicator byte;
	SectionNumber byte; // 1Byte
	LastSectionNumber byte; // 1Byte
	PCR_PID uint16; // 13b
	ProgramInfoLength uint16; // 12b
	Sections []ProgramMapSubSection;
	crc32 uint32; // 4Bytes
}

type ProgramMapSubSection struct {
	StreamType byte; // 1Byte
	ElementaryPID uint16; // 13b
	ESInfoLength [12]byte; // 12b
}

// Functions
// Main packets
func (pes PES) Bytes() (data []byte) {
	data = make([]byte, 188);

	return
}

func (pat PAT) Bytes() (data []byte) {
	return
}

func (pmt PMT) Bytes() (data []byte) {
	return
}

// General parts
func (header Header) Bytes() (data[] byte) {
	data = make([]byte, 4)
	data[0] = 0x47

	data[1]  = header.TransportErrorIndicator << 7
	data[1] |= header.PayloadUnitStartIndicator << 6
	data[1] |= header.TransportPriority << 5
	data[1] |= byte(header.PID >> 8);

	data[2] = byte(header.PID)

	data[3]  = header.TransportScramblingControl << 6
	data[3] |= header.AdaptationFieldControl << 4
	data[3] |= header.ContinuityCounter
	return
}

func (pcr PCR) Bytes() (data []byte) {
	data = make([]byte, 4)
	data[0] = byte(pcr.ProgramClockReferenceBase >> 17)
	data[1] = byte(pcr.ProgramClockReferenceBase >> 1)

	data[2]  = byte(pcr.ProgramClockReferenceBase) << 7
	data[2] |= pcr.Reserved << 6
	data[2] |= byte(pcr.ProgramClockReferenceExtension >> 8)

	data[3]  = byte(pcr.ProgramClockReferenceExtension)
	return
}

func (field AdaptationField) Bytes() (data[] byte) {
	data = make([]byte, field.AdaptationFieldLength);
	data[0] = field.AdaptationFieldLength

	data[1]  = field.DiscontinuityIndicator << 7
	data[1] |= field.RandomAccessIndicator << 6
	data[1] |= field.ElementaryStreamPriorityIndicator << 5
	data[1] |= field.PCR_Flag << 4
	data[1] |= field.OPCR_Flag << 3
	data[1] |= field.SplicingPointFlag << 2
	data[1] |= field.TransportPrivateDataFlag << 1
	data[1] |= field.AdaptationFieldLength

	var offset byte = 2
	if field.PCR_Flag == 1 {
		copy(data[offset:offset+4], field.PCR.Bytes())
		offset += 4
	}

	// Stuffing bytes
	for offset < field.AdaptationFieldLength {
		data[offset] = 0xff
		offset++
	}

	return
}

// Specifics parts
func (section PESSection) Bytes() (data []byte) {
	return
}

func (section ProgramAssociationSection) Bytes() (data []byte) {
	return
}

func (section ProgramAssociationSubSection) Bytes() (data []byte) {
	return
}

func (section ProgramMapSection) Bytes() (data []byte) {
	return
}

func (section ProgramMapSubSection) Bytes() (data []byte) {
	return
}


func Test() {
	var header Header;
	header.TransportErrorIndicator = 1
	header.PayloadUnitStartIndicator = 1;
	header.PID = 17;
	header.AdaptationFieldControl = 1;

	data := header.Bytes();
	fmt.Printf("%08b\n", data);
	fmt.Printf("% 8X\n", data);
}

func CreateHLSInit() {

}

func CreateHLSFragment(mp4 map[string][]interface{},
	fragmentNumber uint32,
	fragmentDuration uint32,
	PMTSpacing uint32) (ts []interface{}){

	ts = make([]interface{}, 0);

	// Retrieve main information

	// Create default PAT

	// Create Default PMT

	// Fill with Stream


	// Create PAT 0
	// Create PMT 4096
	// Continue with mdat
		// Video stream 256
		// Audio stream 255
	return
}

func CreateHLSInitWithConf() {

}

func CreateHLSFragmentWithConf() {

}


