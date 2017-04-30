package ts

import ()

type Packet struct {
	Header
	AdaptationField
	Payload
}

// Structures
type Header struct {
	Bytes
	TransportErrorIndicator    byte
	PayloadUnitStartIndicator  byte
	TransportPriority          byte
	PID                        uint16 // 13b
	TransportScramblingControl byte   // 2b
	AdaptationFieldControl     byte   // 2b
	ContinuityCounter          byte   // 4b
}

type AdaptationField struct {
	Bytes
	AdaptationFieldLength             byte // 1Byte
	DiscontinuityIndicator            byte
	RandomAccessIndicator             byte
	ElementaryStreamPriorityIndicator byte
	PCR_Flag                          byte
	OPCR_Flag                         byte
	SplicingPointFlag                 byte
	TransportPrivateDataFlag          byte
	AdaptationFieldExtensionFlag      byte
	PCR                               PCR
}

type PCR struct {
	Bytes
	BaseMediaDecodeTime uint64 // 33b;
	Extension           uint16 // 9b
}

// Bytes
func (packet Packet) ToBytes() (data Data) {
	data = *NewData(188)

	// Get header
	data.PushBytes(packet.Header)

	// Get adaptation field
	if packet.HasAdaptationField() {
		data.PushBytes(packet.AdaptationField)
	}

	return
}

func (packet Packet) RestingSize() (size int) {

	// Get header
	size = packet.Header.Size()

	// Get adaptation field
	if packet.HasAdaptationField() {
		size += packet.Header.Size()
	}

	return
}

func (header Header) ToBytes() (data Data) {
	data = *NewData(header.Size())
	data.Write(0x47)

	data.PushObj(header.TransportErrorIndicator, 1)
	data.PushObj(header.PayloadUnitStartIndicator, 1)
	data.PushObj(header.TransportPriority, 1)
	data.PushObj(header.PID, 13)
	data.PushObj(header.TransportScramblingControl, 2)
	data.PushObj(header.AdaptationFieldControl, 2)
	data.PushObj(header.ContinuityCounter, 4)
	return
}


func (header Header) Size() (int) {
	return 4
}

func (packet Packet) Size() (int) {
	return 188
}

func (pcr PCR) ToBytes() (data Data) {
	data = *NewData(6)

	// PCR = Base * 300 + extension
	data.PushObj(pcr.BaseMediaDecodeTime, 33)
	data.PushUInt(0x3f, 6)
	data.PushObj(pcr.Extension, 9)

	return
}

func (field AdaptationField) ToBytes() (data Data) {
	// Compute Adaptation length adding the first byte length
	adaptationLength := field.GetAdaptationLength()
	data = *NewData(field.Size())

	data.PushObj(adaptationLength, 8)
	data.PushObj(field.DiscontinuityIndicator, 1)
	data.PushObj(field.RandomAccessIndicator, 1)
	data.PushObj(field.ElementaryStreamPriorityIndicator, 1)
	data.PushObj(field.PCR_Flag, 1)
	data.PushObj(field.OPCR_Flag, 1)
	data.PushObj(field.SplicingPointFlag, 1)
	data.PushObj(field.TransportPrivateDataFlag, 1)
	data.PushObj(field.AdaptationFieldExtensionFlag, 1)

	if field.PCR_Flag == 1 {
		data.PushBytes(field.PCR)
	}

	// Fill with stuffing bytes
	// Don't forget the byte counting adaptationLength
	result := int(adaptationLength + 1) * 8
	data.FillTo(0xff, result)
	return
}

func (field AdaptationField) Size() (int) {
	return int(field.GetAdaptationLength()+ 1)
}

func (field AdaptationField) GetAdaptationLength() (byte) {
	if field.AdaptationFieldLength != 0 {
		return field.AdaptationFieldLength
	}

	var adaptationLength byte = 1

	if field.PCR_Flag == 1 {
		adaptationLength += 6
	}

	return adaptationLength
}

// Check if the current packet has adaptation field flag activated
func (packet Packet) HasAdaptationField() bool {
	return packet.Header.AdaptationFieldControl&2 != 0
}

// Check if the current packet has payload field flag activated
func (packet Packet) HasPayload() bool {
	return packet.Header.AdaptationFieldControl&1 != 0
}

func (field *AdaptationField) setPCR(PCR PCR) {
	field.PCR = PCR
	field.PCR_Flag = 1
	field.AdaptationFieldLength += 6
}

func (packet *Packet) setAdaptationControl(adaptationFieldFlag bool, payloadFlag bool) {
	if adaptationFieldFlag {
		if payloadFlag {
			packet.AdaptationFieldControl = 0x03
		} else {
			packet.AdaptationFieldControl = 0x02
		}
	} else {
		if payloadFlag {
			packet.AdaptationFieldControl = 0x01
		} else {
			packet.AdaptationFieldControl = 0x00
		}
	}

}