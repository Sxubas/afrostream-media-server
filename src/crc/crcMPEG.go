package crc

//http://www.sunshine2k.de/articles/coding/crc/understanding_crc.html#ch6
type Table [256]uint32

var MPEGTable *Table = CreateCRCTable(0x04C11DB7)

func CreateCRCTable(polynomial uint32) (table *Table) {
	table = new(Table)

	for divident := uint32(0); divident < 256; divident++ {
		curByte := divident << 24

		for bit := uint32(0); bit < 8; bit++ {

			if (curByte & 0x80000000) != 0 {
				curByte <<= 1
				curByte ^= polynomial
			} else {
				curByte <<= 1
			}
		}
		table[divident] = curByte
	}
	return
}

func ComputeCRC32(bytes []byte, table *Table) (crc uint32) {
	crc = 0xFFFFFFFF

	for _,b := range bytes {
		pos := byte((crc^(uint32(b) << 24)) >> 24)
		crc = uint32((crc << 8) ^ uint32(table[pos]))
	}

	return
}

func ComputeCRC32Simple(bytes []byte, polynomial uint32) (crc uint32) {
	crc = 0xFFFFFFFF

	for _,b := range bytes {
		crc ^= uint32(b) << 24

		for i := 0; i < 8; i++ {
			if (crc & uint32(0x80000000)) != 0 {
				crc = uint32(crc << 1) ^ polynomial
			} else {
				crc <<= 1
			}
		}
	}
	return
}

func ComputeCRC32MPEG(bytes []byte) uint32 {
	return ComputeCRC32(bytes, MPEGTable)
}