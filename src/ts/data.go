package ts

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"crc"
)

type Data struct {
	Data   []byte
	Offset int
}

// Push int on Data
func (data *Data) PushUInt(object uint32, objectSize int) {
	data.PushObj(object, objectSize)
}

// Push int on Data
func (data *Data) PushInt(object int64, objectSize int) {
	data.PushObj(object, objectSize)
}

// Push object on Data
//		object 			object to write
//		objectSize		part to write of this object
// Push the object on Data
func (data *Data) PushObj(object interface{}, objectSize int) {
	// Transform object in bytes
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, object)
	objectBytes := buf.Bytes()

	if len(objectBytes) * 8 < objectSize {

		switch object.(type) {
		case int:
			fmt.Println("Warning: Int processed as object")
			data.PushInt(int64(object.(int)), objectSize)

		case uint:
			fmt.Println("Warning: UInt processed as object")
			data.PushUInt(uint32(object.(uint)), objectSize)

		default:
			panic("Object: Wrong Requested Size")
		}
	}

	data.Push(objectBytes, objectSize)
	return
}

func (data *Data) PushAll(bytes []byte) {
	data.Push(bytes, len(bytes))
}

func (data *Data) PushData(dataPushed Data) {
	data.Push(dataPushed.Data, len(dataPushed.Data) * 8)
}

// Write bytes on Data
//		bytes 			bytes to write
//		bytesSize		part to write of these bytes
// Push the object on Data
func (data *Data) Push(bytes []byte, sizeToPush int) {

	nBits := len(bytes) * 8

	// While there are resulting bits to push
	remainingSize := sizeToPush
	for remainingSize != 0 {
		// Number of bits in actual Byte
		residualBits := data.GetResidualBits()

		// Number of pushed bits
		pushedBits := Min(residualBits, remainingSize)

		// Compute the lastBitIndex and startIndex
		// Bit at index 0 is the last bit of the bytes
		lastBitIndex := nBits - remainingSize
		firstBitIndex := nBits - (remainingSize - pushedBits)

		// Get the corresponding byte to write
		writtenByte := GetByte(bytes,
			lastBitIndex,
			firstBitIndex)
		writtenByte = writtenByte << byte(residualBits - pushedBits)

		// Write byte on Data
		data.WriteOR(writtenByte)

		// Update Data offset
		data.Offset += pushedBits

		// Update bit index
		remainingSize -= pushedBits
	}

}

func (data *Data) PushBytes(bytes Bytes) {
	toBytes := bytes.ToBytes()
	data.PushData(toBytes)
}

// Apply or operation to current byte and given byte
func (data *Data) WriteOR(byte byte) {
	*data.GetCurrentByte() |= byte
}

// Write byte on Data
func (data *Data) Write(byte byte) {
	*data.GetCurrentByte() = byte
	data.Offset += 8
}

// Fill rest of Data with byte
func (data *Data) FillRemaining(pushed byte) {
	lenData := len(data.Data) * 8
	data.FillTo(pushed, lenData)
}

func (data *Data) FillTo(pushed byte, offsetBitAddress int) {
	// Jump to next byte
	restingBits := data.Offset % 8
	if restingBits != 0 {
		data.PushObj(byte(pushed), restingBits)
	}

	for data.Offset != offsetBitAddress {
		data.Write(pushed)
	}
}

// Create byte from start and end indices, from last bits to first
//
// 		lastBitIndex	last index read in bits
//		firstBitIndex	firstBit index to read in bits,
// 						endIndex - startIndex < 8
//						Copy from last bits to first
//
// 		shift			number of shift to add to the resulting byte
//
func GetByte(data []byte, lastBitIndex int, firstBitIndex int) byte {
	firstBitIndexIncluded := firstBitIndex - 1

	lastByteIndex := lastBitIndex / 8
	firstByteIndex := firstBitIndexIncluded / 8
	lastIndexInByte := uint8(lastBitIndex% 8)
	firstIndexInByte := uint8(firstBitIndexIncluded) % 8

	// If start and end are on the same byte
	if firstByteIndex == lastByteIndex {
		return SelectByte(data[firstByteIndex], lastIndexInByte, firstIndexInByte)
	} else {

		// Compute shift from start
		shift := 8 - lastIndexInByte

		// Get first part
		startByte := SelectByte(data[firstByteIndex], lastIndexInByte, 7)

		// Get second part in the next byte
		endByte := SelectByte(data[lastByteIndex], 0, firstIndexInByte)

		// Create byte by coupling start and endByte
		return startByte | (endByte >> shift)
	}
}

// Select part of a byte
func SelectByte(src byte, lastIndexInByte, firstIndexInByte uint8) byte {
	tmp := src << (lastIndexInByte)
	return tmp >> (lastIndexInByte + firstIndexInByte - 7)
}

// Get the current byte
func (data *Data) GetCurrentByte() *byte {

	if data.GetByteIndex() >= len(data.Data) {
		panic("Data: Out Of Memory")
	}

	return &data.Data[data.GetByteIndex()]
}

// Set the current byte
func (data *Data) SetCurrentByte(byte byte) {
	data.Data[data.GetByteIndex()] = byte
}

// Get the index of the current starting byte in bits
func (data *Data) GetStartBitsIndex() int {
	return data.GetByteIndex() * 8
}

// Get the index of the current starting byte in byte
func (data *Data) GetByteIndex() int {
	return data.Offset / 8
}

// Get number of residual bits in the current byte
func (data *Data) GetResidualBits() int {
	return 8 - data.Offset%8
}

// Create Data with offset support
func NewData(length int) *Data {
	data := new(Data)
	data.Data = make([]byte, length)
	return data
}

func (data Data) PrintBinary() {
	var i uint32 = 0
	size := uint32(len(data.Data))
	for i < size {
		PrintLine("%08b    ", i, size, data.Data)
		i += 16
	}
}

func (data Data) PrintHex() {
	var i uint32 = 0
	size := uint32(len(data.Data))
	for i < size {
		PrintLine("% 8x   ", i, size, data.Data)
		i += 16
	}
}

func (data Data) PrintHexFull() {
	var i uint32 = 0
	size := uint32(len(data.Data))
	for i < size {
		fmt.Printf(" %04x:    ", uint16(i))
		PrintLine("% 8x   ", i, size, data.Data)
		i += 16
	}
}

func PrintLine(format string, start uint32, size uint32, bytes []byte) {
	for w := 0; w < 2 && start < size; w++ {
		endMax := Min32(start + 8, size)
		fmt.Printf(format, bytes[start:endMax])
		start += 8
	}
	fmt.Printf("\n")
}

func (data *Data) GenerateCRC32ToOffset() uint32 {
	return crc.ComputeCRC32MPEG(data.Data[0:data.GetByteIndex()])
}

func (data *Data) GenerateCRC32() uint32 {
	return crc.ComputeCRC32MPEG(data.Data)
}
