package ts

import (
	"mp4"
	"hash/crc32"
)




func CreateHLSInit();

func CreateHLSFragment() {
	// Create SDT 17 ??
	// Create PAT 0
	// Create PMT 4096
	// Continue with mdat
		// Video stream 256
		// Audio stream 255
}

func CreateHLSInitWithConf();

func CreateHLSFragmentWithConf();

func FillHeader() {
Header:
	sync_byte(1B)：0x47
	transport_error_indicator (1b)：0
	payload_unit_start_indicator(1b)：
	1，  该字段置1表示携带的是PSI或PES第一个包
	transport_priority(1b)：0
	PID(13b)：0 0000 0000 0000，说明携带的是PAT表
	Transport_scrambling_control(2b)：00
	Adaptation_field_control(2b)：11，标识既有adaptation_field又有payload
}

func FillAdaptationField() {
adaptation_field_length（1B）：0xa6
discontinuity_indecator(1b)：0
random_access_indicator(1b)：0
elementary_stream_priority_indicator(1b)：0
PCR_flag(1b)：0
OPCR_flag(1b)：0
splicing_point_flag(1b)：0
transport_private_data_flag(1b)：0
adaptation_field_extension_flag（1b）：0
stuffing_bytes：填充0xff，结束地址：0x00c6
}

func CreatePAT() {
	table_id(1B)：0x00，PAT表固定值
	section_syntax_indicator(1b)：1
	‘0’
	‘11’
	section_length(12b)：0x 0 0d
	transport_stream_id(2B)：0x 4b 7c,为何是这个值？？？？？？
	‘11’
	version_number(5b)：00 101
	current_next_indicator(1b)：0
	section_number(1B)：0x 00
	last_section_number(1B)：0x 00
	进入N _loop循环
	program_number(2B)：0x 00 01
	‘111’
	program_map_PID(13b)：0 0000 0100 0010（0x42）

	CRC32(4B)：0xbec6 cee3

}



type Header struct {
	TransportErrorIndicator uint8;
	PayloadUnitStartIndicator uint8;
	TransportPriority uint8;
	PID [13]byte;
	TransportScramblingControl uint16;
	AdaptationFieldControl uint16;
}

type AdaptationField struct {
	AdaptationFieldLength uint8;
	DiscontinuityIndicator uint8;
	RandomAccessIndicator uint8;
	ElementaryStreamPriorityIndicator uint8;
	PCR_Flag uint8;
	OPCR_Flag uint8;
	SplicingPointFlag uint8;
	TransportPrivateDataFlag uint8;
	AdaptationFieldExtensionFlag uint8;
}

type AdaptationFieldStuffed struct {
	StuffingBytes uint16;
}

type AdaptationFieldExtended struct {
	AdaptationFieldExtensionFlag uint8;
	ProgramClockReferenceBase [33]byte;
	Reserved [6]byte;
	ProgramClockReferenceExtension [9]byte;
}

type ProgramAssociationSection struct {
	TableID uint8;
	SectionSyntaxIndicator uint8;
	SectionLength [12]byte;
	TransportStreamID uint16;
	VersionNumber [5]byte;
	CurrentNextIndicator uint8;
	SectionNumber uint8;
	LastSectionNumber uint8;
	subSections []ProgramAssociationSubSection;
	crc32 [4]byte;
}

type ProgramAssociationSubSection struct {
	ProgramNumber uint16;
	ProgramMapID [13]byte;
}

func CreatePMT {
	Program Map Section：
	table_id(1B)：0x02，PMT表固定值
	section_syntax_indicator(1b)：1
	‘0’
	‘11’
	section_length(12b)：0x 0 17
	program number(2B)：0x00 01
	‘11’
	version_number(5b)：11 100
	current_next_indicator(1b)：1
	section_number(1B)：0x 00
	last_section_number(1B)：0x 00
	‘111’
	PCR_PID(13b)：0 0000 0100 0100（0x44）
	‘1111’
	program_info_length(12b)：0000 0000 0000
	进入N _loop循环
	stream_type(1B)：0x 02 (13818-2 Video stream)
	elementarty_PID(13b)：0 0000 0100 0100（0x44）//为何为0x44，用户自定义？？？
	‘1111’
	ES_info_length(12b)：0000 0000 0000

	stream_type(1B)：0x03 (ISO/IEC 11172Audio，为什么不是04 13818-3video？？)
	‘111’
	elementary_PID(13b)：0 0000 0100 1001（0x45）
	‘1111’
	ES_info_length(12b)：0000 0000 0000

	CRC_32(4B)：0x73cf 72ce
}

type ProgramMapSection struct {
	TableId uint8;
	SectionSyntaxIndicator uint8;
	SectionLength [12]byte;
	ProgramNumber uint16;
	VersionNumber [5]byte;
	CurrentNextIndicator uint8;
	SectionNumber uint8;
	LastSectionNumber uint8;
	PCR_PID [13]byte;
	ProgramInfoLength [12]byte;
	subSections []ProgramMapSubSection;
	crc32 [4]byte;
}

type ProgramMapSubSection struct {
	StreamType uint8;
	ElementaryPID [13]byte;
	ESInfoLength [12]byte;
}

func createPES {
	packet_start_code_prefix(24b)：0x00 00 01
	stream_id(1B)：0xe0
	PES_packet_length(2B)：0x20cf，PS上为0x07ec？？？？？？？
	optional_PES_HEADER：
	‘10’
	PES_scrambling_control(2b)：00
	PES_priority(1b)：0
	data_alignment_indicator(1b)：0
	copyright(1b)：0
	original_or_copy(1b)：0
	PTS_DTS_flags(2b)：11
	ESCR_flag(1b)：0
	ES_rate_flag(1b)：0
	DSM_trick_mode_flag(1b)：0
	additional_copy_info_flag(1b)：0
	PES_CRC_flag(1b)：0
	PES_extension_flag(1b)：0
	PES_header_data_length（1B）：0x0a

	PTS_DTS(10B)：0x 35 91af 1293 1591 adf5 41,
		PS中为 0x 31 00 01 a9 f3 11 00 01 8c a1,为何不同？
	video_sequence：和PS中完全相同
}

type PESSection struct {
	PacketStartCodePrefix [24]byte;
	StreamId uint8;
	PES_PacketLength uint16;
}

type OptionalPESHeader struct {
	PES_ScramblingControl uint16;
	PES_Priority uint8;
	DataAlignmentIndicator uint8;
	Copyright uint8;
	OriginalOrCopy uint8;
	PTS_DTS_Flags uint16;
	ESCR_Flag uint8;
	ES_RateFlag uint8;
	DSM_TrickModeFlag uint8;
	AdditionalCopyInfoGlag uint8;
	PES_CRC_Flag uint8;
	PES_ExtensionFlag uint8;
	PES_HeaderDataLength uint8;
	PTS_DTS [10]byte;
}

