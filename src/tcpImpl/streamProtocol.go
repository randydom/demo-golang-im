package tcpImpl

import "encoding/binary"

const (
    kStreamProtocol4HeaderLength = 4
    kStreamProtocol2HeaderLength = 2
)

func getStreamMaxLength(headerBytes uint32) uint64 {
    return 1<<(8*headerBytes) - 1
}

// StreamProtocol4
// Binary format : | 4 byte (total stream length) | data ... (total stream length - 4) |
//  implement default stream protocol
//  stream protocol interface for 4 bytes header
type StreamProtocol4 struct {
}

func NewStreamProtocol4() *StreamProtocol4 {
    return &StreamProtocol4{}
}

func (s *StreamProtocol4) Init() {

}

func (s *StreamProtocol4) GetHeaderLength() uint32 {
    return kStreamProtocol4HeaderLength
}

func (s *StreamProtocol4) UnserializeHeader(buf []byte) uint32 {
    if len(buf) < kStreamProtocol4HeaderLength {
        return 0
    }
    return binary.BigEndian.Uint32(buf)
}

func (s *StreamProtocol4) SerializeHeader(body []byte) []byte {
    if uint64(len(body)+kStreamProtocol4HeaderLength) > uint64(getStreamMaxLength(kStreamProtocol4HeaderLength)) {
        //	stream is too long
        return nil
    }

    var ln uint32 = uint32(len(body) + kStreamProtocol4HeaderLength)
    var buffer [kStreamProtocol4HeaderLength]byte
    binary.BigEndian.PutUint32(buffer[0:], ln)
    return buffer[0:]
}