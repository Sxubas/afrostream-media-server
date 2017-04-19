package ts

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
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
			panic("Object size inferior to requested size")
		}
	}

	data.Push(objectBytes, objectSize)
	return
}

func (data *Data) PushAll(bytes []byte) {
	data.Push(bytes, len(bytes))
}

func (data *Data) PushData(dataPushed Data) {
	data.Push(dataPushed.Data, dataPushed.Offset)
}

// Write bytes on Data
//		bytes 			bytes to write
//		bytesSize		part to write of these bytes
// Push the object on Data
func (data *Data) Push(bytes []byte, sizeToPush int) {
	// While there are resulting bits to push
	bitIndex := 0
	for bitIndex < sizeToPush {
		// Number of bits in actual Byte
		residualBits := data.GetResidualBits()

		// Number of pushed bits
		pushedBits := Min(residualBits, sizeToPush-bitIndex)

		// Get the corresponding byte to write
		writtenByte := GetByte(bytes,
			bitIndex,
			bitIndex + pushedBits - 1) << byte(residualBits-pushedBits)

		// Write byte on Data
		data.WriteOR(writtenByte)

		// Update Data offset
		data.Offset += pushedBits

		// Update bit index
		bitIndex += pushedBits
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
func (data *Data) FillRemaining(byte byte) {
	lenData := len(data.Data) * 8

	if data.Offset % 8 != 0 {
		panic("Not aligned data")
	}

	for data.Offset != lenData {
		data.Write(byte)
	}
}

// Create byte from start and end indices
//
// 		startIndex		start index read in bits
//		endIndex		end index to read in bits,
// 						startIndex - endIndex < 8
//
// 		shift			number of shift to add to the resulting byte
//
func GetByte(data []byte, startIndex int, endIndex int) byte {
	var startIndexByte int = startIndex / 8
	var endIndexByte int = endIndex / 8
	var startIndexInByte = uint8(startIndex % 8)
	var endIndexInByte = uint8(endIndex % 8)

	// If start and end are on the same byte
	if startIndexByte == endIndexByte {
		return SelectByte(data[startIndexByte], startIndexInByte, endIndexInByte)
	} else {

		// Compute shift from start
		shift := 8 - startIndexInByte

		// Get first part
		startByte := SelectByte(data[startIndexByte], startIndexInByte, 7)

		// Get second part in the next byte
		endByte := SelectByte(data[endIndexByte], 0, endIndexInByte-1)

		// Create byte by coupling start and endByte
		return startByte | (endByte >> shift)
	}
}

// Select part of a byte
func SelectByte(src byte, start, end uint8) byte {
	tmp := src << (8 - end)
	return tmp >> (8 - end + start)
}

// Get the current byte
func (data *Data) GetCurrentByte() *byte {

	if data.GetByteIndex() >= len(data.Data) {
		panic("Data: No more memory left")
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

func (data *Data) PrintBinary() {
	var i uint32 = 0
	size := uint32(data.Offset)
	for i < size {
		PrintLine("%08b    ", i, size, data.Data)
		i += 16
	}
}

func (data *Data) PrintHex() {
	var i uint32 = 0
	size := uint32(data.Offset)
	for i < size {
		PrintLine("% 8X   ", i, size, data.Data)
		i += 16
	}
}

func (data *Data) PrintHexFull() {
	var i uint32 = 0
	size := uint32(data.Offset)
	fmt.Printf("\nData-Bytes:\n")
	for i < size {
		fmt.Printf(" %04X:    ", uint16(i))
		PrintLine("% 8X   ", i, size, data.Data)
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

func (data *Data) GenerateCRC32() uint32 {
	return crc32.Checksum(data.Data, crc32.IEEETable)
}
