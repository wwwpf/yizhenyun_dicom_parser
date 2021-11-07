package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"unsafe"
)

var (
	inputFile string
)

func init() {
	flag.StringVar(&inputFile, "f", "", "输入文件")
}

type Parser struct {
	*bytes.Reader
}

func (p Parser) ParseData(l int) []byte {
	b := make([]byte, l)
	n, err := p.Read(b)
	if n != l {
		panic("read data length error")
	}
	if err != nil {
		panic(err)
	}
	return b
}

func (p Parser) ParseNumber(typename string) interface{} {
	switch typename {
	case "uint8":
		b := p.ParseData(int(unsafe.Sizeof(uint8(0))))
		return b[0]
	case "uint16":
		b := p.ParseData(int(unsafe.Sizeof(uint16(0))))
		return binary.BigEndian.Uint16(b)
	case "uint32":
		b := p.ParseData(int(unsafe.Sizeof(int32(0))))
		return binary.BigEndian.Uint32(b)
	case "int32":
		b := p.ParseData(int(unsafe.Sizeof(int32(0))))
		return int32(binary.BigEndian.Uint32(b))
	case "uint64":
		b := p.ParseData(int(unsafe.Sizeof(uint64(0))))
		return binary.BigEndian.Uint64(b)
	default:
		return 0
	}
}

func (p Parser) ParseUint8() uint8 {
	v := p.ParseNumber("uint8")
	return v.(uint8)
}

func (p Parser) ParseBit() bool {
	return p.ParseUint8() != 0
}

func (p Parser) ParseUint16() uint16 {
	v := p.ParseNumber("uint16")
	return v.(uint16)
}

func (p Parser) ParseInt32() int32 {
	v := p.ParseNumber("int32")
	return v.(int32)
}

func (p Parser) ParseUint32() uint32 {
	v := p.ParseNumber("uint32")
	return v.(uint32)
}

func (p Parser) ParseUint64() uint64 {
	v := p.ParseNumber("uint64")
	return v.(uint64)
}

func (p Parser) ParseBitset() []bool {
	n := p.ParseInt32()
	nn := int(n)
	bitset := make([]bool, nn)
	for i := 0; i < nn; i++ {
		bitset[i] = p.ParseBit()
	}
	return bitset
}

func (p Parser) ParseBytes() []byte {
	n := p.ParseInt32()
	if n == -1 {
		return make([]byte, 0)
	}
	return p.ParseData(int(n))
}

func (p Parser) ParseString() string {
	n := p.ParseInt32()
	nn := int(n)
	data := p.ParseData(nn)
	s := make([]byte, 0, n)
	for i := 0; i < nn; i += 2 {
		if data[i] == 0 {
			s = append(s, data[i+1])
		} else {
			s = append(s, data[i], data[i+1])
		}
	}
	return string(s)
}

type PkgItem struct {
	parser Parser

	studyID           uint64
	seriesID          uint64
	imageInstanceUID  string
	imageNumber       int32
	extendString      string
	packageInfo       string
	thumbnailBuffer   []byte
	studyInstanceUID  string
	seriesInstanceUID string
	enables           []bool
	doubles           []bool
	stage             uint32
	result            bool
	startPos          uint64
	size              uint64
	content           []byte
}

func NewPkgItem(p Parser) PkgItem {
	return PkgItem{parser: p}
}

func (p *PkgItem) Parse() {
	p.studyID = p.parser.ParseUint64()
	p.seriesID = p.parser.ParseUint64()
	p.imageInstanceUID = p.parser.ParseString()
	p.imageNumber = p.parser.ParseInt32()
	p.extendString = p.parser.ParseString()
	p.packageInfo = p.parser.ParseString()
	p.thumbnailBuffer = p.parser.ParseBytes()
	p.studyInstanceUID = p.parser.ParseString()
	p.seriesInstanceUID = p.parser.ParseString()
	p.enables = p.parser.ParseBitset()
	p.doubles = p.parser.ParseBitset()
	p.stage = p.parser.ParseUint32()
	p.result = p.parser.ParseBit()
	p.startPos = p.parser.ParseUint64()
	p.size = p.parser.ParseUint64()
	p.content = p.parser.ParseBytes()

	outputFile := p.imageInstanceUID + ".dcm"
	fmt.Printf("export file: %s\n", outputFile)
	if err := os.WriteFile(outputFile, p.content, os.ModePerm); err != nil {
		panic(err)
	}
}

type PkgFile struct {
	parser Parser

	file   string
	header string

	items []PkgItem
}

func (p *PkgFile) Open(f string) {
	p.file = f
	var err error
	var data []byte
	if data, err = os.ReadFile(p.file); err != nil {
		panic(err)
	}
	p.parser = Parser{bytes.NewReader(data)}
}

func (p *PkgFile) Parse() {
	p.header = p.parser.ParseString()

	fileCount := int(p.parser.ParseUint16())
	p.items = make([]PkgItem, 0, fileCount)
	for i := 0; i < fileCount; i++ {
		item := NewPkgItem(p.parser)
		item.Parse()
	}
}

func main() {
	flag.Parse()

	var p PkgFile
	p.Open(inputFile)
	p.Parse()
}
