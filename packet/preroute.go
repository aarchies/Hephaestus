package packet

import (
	"encoding/binary"
	"errors"
	"sync"
	"unsafe"
)

type FiveTuple struct {
	SrcIP    uint32  // 4
	DstIP    uint32  // 4
	SrcPort  uint16  // 2
	DstPort  uint16  // 2
	Protocol uint8   // 1
	_        [3]byte // 3 // padding to 16 bytes
}

type FastParser struct {
	bufferPool *sync.Pool
}

func NewFastParser() *FastParser {
	return &FastParser{
		bufferPool: &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, 1518) // 标准MTU大小
				return &buf
			},
		},
	}
}

func (p *FastParser) Parse(packet []byte) (FiveTuple, error) {
	bufPtr := p.bufferPool.Get().(*[]byte)
	defer p.bufferPool.Put(bufPtr)

	// 重置切片长度（保留容量）
	*bufPtr = (*bufPtr)[:0]

	// 使用指针操作数据
	*bufPtr = append(*bufPtr, packet...)

	return p.parseUnsafe(*bufPtr)
}

// func ipToStr(ip uint32) string {
// 	b0 := byte(ip >> 24)
// 	b1 := byte(ip >> 16)
// 	b2 := byte(ip >> 8)
// 	b3 := byte(ip)
// 	buf := make([]byte, 0, 15) // 最大"255.255.255.255"长度
// 	buf = strconv.AppendUint(buf, uint64(b0), 10)
// 	buf = append(buf, '.')
// 	buf = strconv.AppendUint(buf, uint64(b1), 10)
// 	buf = append(buf, '.')
// 	buf = strconv.AppendUint(buf, uint64(b2), 10)
// 	buf = append(buf, '.')
// 	buf = strconv.AppendUint(buf, uint64(b3), 10)
// 	return string(buf)
// }

//go:nosplit
func (p *FastParser) parseUnsafe(packet []byte) (FiveTuple, error) {
	const (
		ethHeaderLen = 14
		ipHeaderMin  = 20
	)

	if len(packet) < ethHeaderLen+ipHeaderMin {
		return FiveTuple{}, errors.New("packet too short")
	}

	// 提取以太网头部
	ipHeader := *(*[20]byte)(unsafe.Pointer(&packet[ethHeaderLen]))

	// 提取IP头字段
	version := ipHeader[0] >> 4
	if version != 4 {
		return FiveTuple{}, errors.New("non-IPv4 packet")
	}

	ipTuple := FiveTuple{
		SrcIP: binary.BigEndian.Uint32(ipHeader[12:16]),
		DstIP: binary.BigEndian.Uint32(ipHeader[16:20]),
	}

	// 传输层解析
	proto := ipHeader[9]
	ihl := int(ipHeader[0]&0x0F) << 2
	if len(packet) < ethHeaderLen+ihl+4 {
		return FiveTuple{}, errors.New("invalid transport header")
	}

	transHeader := packet[ethHeaderLen+ihl:]
	switch proto {
	case 6, 17: // TCP/UDP
		ipTuple.SrcPort = binary.BigEndian.Uint16(transHeader)
		ipTuple.DstPort = binary.BigEndian.Uint16(transHeader[2:])
		ipTuple.Protocol = proto
	default:
		return FiveTuple{}, errors.New("unsupported protocol")
	}

	return ipTuple, nil
}
