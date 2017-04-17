package mp4

import (
	"fmt"
	"bytes"
	"encoding/binary"
)

type Data struct {
	data []byte
	offset int
}

// Main packets
type Packet struct {
	Header Header
	AdaptationField AdaptationField
	Payload []byte
}

type PES struct {
	Packet
}

type PMT struct {
	Packet
	PointField byte // 1Byte
	Section ProgramMapSection
}

type PAT struct {
	Packet
	PointField byte // 1Byte
	Sections ProgramAssociationSection
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
func (pes PES) Bytes() (data []byte, offset int) {
	data = make([]byte, 188);
	// PID depends on stream


	return
}

func (packet Packet) Bytes() (data []byte, offset int) {
	offset = 0;
	data = make([]byte, 188);

	// Get header
	headerBytes := packet.Header.Bytes()
	offset += len(headerBytes)

	// Get adaptation field
	if packet.hasAdaptationField() {
		adaptationFieldBytes := packet.AdaptationField.Bytes()
		lenAdaptation := len(adaptationFieldBytes)
		copy(data[offset:offset+lenAdaptation], adaptationFieldBytes)
		offset += lenAdaptation
	}

	return
}

func (pat PAT) Bytes() (data []byte, offset int) {
	data, offset = pat.Packet.Bytes()

	// Set PointField
	data[offset] = pat.PointField
	offset++

	// Program association section


	return
}

func (pmt PMT) Bytes() (data []byte, offset int) {
	data = make([]byte, 188);



	return
}

// General parts
func (header Header) aBytes() (data Data) {
	data = *NewData(4)
	data.write(0x47)

	data.push(header.TransportErrorIndicator, 1)
	data.push(header.PayloadUnitStartIndicator, 1)
	data.push(header.TransportPriority, 1)
	data.push(header.PID, 13)
	data.push(header.TransportScramblingControl, 2)
	data.push(header.AdaptationFieldControl, 2)
	data.push(header.ContinuityCounter, 4)
	return
}

func (header Header) Bytes() (data []byte) {
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
	// Compute Adaptation length adding the first byte length
	adaptationLength := field.AdaptationFieldLength + 1

	data = make([]byte, adaptationLength);
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
	for offset < adaptationLength {
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
	data = make([]byte, 4)

	data[0] = section.TableID

	data[1]  = section.SectionSyntaxIndicator << 7
	data[1] |= byte(section.SectionLength >> 5)

	data[2]  = byte(section.SectionLength) << 5
	//data[2] |= byte(section.TransportStreamID >> 6) <<


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

// Tools

// Check if the current packet has adaptation field flag activated
func (packet Packet) hasAdaptationField() (bool bool){
	return packet.Header.AdaptationFieldControl & 2 != 0
}


// Write byte on data
//		object 			object to write
//		objectSize		part to write of this object
// Push the object on data
func (data *Data) push(object interface{}, objectSize int) {
	// Transform object in bytes
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, object)
	objectBytes := buf.Bytes()

	// While there are resulting bits to push
	bitIndex := 0;
	for bitIndex < objectSize {
		// Number of bits in actual Byte
		residualBits := data.getResidualBits()

		// Number of pushed bits
		pushedBits := min(residualBits, objectSize - bitIndex)

		// Get the corresponding byte to write
		writedByte := getByte(objectBytes,
			bitIndex,
			bitIndex + pushedBits) << byte(residualBits - pushedBits)

		// Write byte on data
		data.writeOR(writedByte)

		// Update data offset
		data.offset += pushedBits

		// Update bit index
		bitIndex += pushedBits
	}

	return
}
// Apply or operation to current byte and given byte
func (data *Data) writeOR(byte byte) {
	*data.getCurrentByte() |= byte
}

// Write byte on data
func (data *Data) write(byte byte) {
	*data.getCurrentByte() = byte
	data.offset += 8
}

// Create byte from start and end indices
//
// 		startIndex		start index read in bits
//		endIndex		end index to read in bits,
// 						startIndex - endIndex < 8
//
// 		shift			number of shift to add to the resulting byte
//
func getByte(data []byte, startIndex int, endIndex int) byte {
	var startIndexByte int = startIndex / 8
	var endIndexByte int = endIndex / 8
	var startIndexInByte = uint8(startIndex % 8)
	var endIndexInByte = uint8(endIndex % 8)

	// If start and end are on the same byte
	if startIndexByte == endIndexByte {
		return selectByte(data[startIndexByte], startIndexInByte, endIndexInByte)
	} else {

		// Compute shift from start
		shift := 8 - startIndexInByte

		// Get first part
		startByte := selectByte(data[startIndexByte], startIndexInByte, 7)

		// Get second part in the next byte
		endByte := selectByte(data[endIndexByte], 0, endIndexInByte - 1)

		// Create byte by coupling start and endByte
		return startByte | (endByte >> shift)
	}
}

// Select part of a byte
func selectByte(src byte, start, end uint8) byte {
	tmp := src << (8 - end)
	return tmp >> (8 - end + start)
}

// Get the current byte
func (data *Data) getCurrentByte() *byte {
	return &data.data[data.getByteIndex()]
}

// Set the current byte
func (data *Data) setCurrentByte(byte byte) {
	data.data[data.getByteIndex()] = byte
}

// Get the index of the current starting byte in bits
func (data *Data) getStartBitsIndex() int {
	return data.getByteIndex() * 8
}

// Get the index of the current starting byte in byte
func (data *Data) getByteIndex() int {
	return data.offset / 8
}

// Get number of residual bits in the current byte
func (data *Data) getResidualBits() int {
	return (8 - data.offset % 8)
}

// Create data with offset support
func NewData(length int) *Data {
	data := new(Data)
	data.data = make([]byte, length)
	return data
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}


func Test() {
	var header Header;
	header.TransportErrorIndicator = 1
	header.PayloadUnitStartIndicator = 0;
	header.PID = 513;
	header.AdaptationFieldControl = 3;
	header.ContinuityCounter = 3

	data := header.Bytes();
	fmt.Printf("%08b\n", data);
	fmt.Printf("% 8X\n", data);

	data2 := header.aBytes();
	fmt.Printf("%08b\n", data2.data);
	fmt.Printf("% 8X\n", data2.data);

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


