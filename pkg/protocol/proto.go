package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/lucheng0127/vtun/pkg/cipher"
)

const (
	HDR_FLG_UNKNOW uint16 = 0x1b00
	HDR_FLG_REQ    uint16 = 0x1b00 | 0x01
	HDR_FLG_ACK    uint16 = 0x1b00 | (0x01 << 1)
	HDR_FLG_PSH    uint16 = 0x1b00 | (0x01 << 2)
	HDR_FLG_DAT    uint16 = 0x1b00 | (0x01 << 3)
	HDR_FLG_FIN    uint16 = 0x1b00 | (0x01 << 4)
	HDR_FLG_IPS    uint16 = 0x1b00 | (0x01 << 5)
	HDR_FLG_ROU    uint16 = 0x1b00 | (0x01 << 6)

	MAX_FRG_SIZE int = 4096
	HDR_LEN      int = 6
	MIN_PLEN     int = 12 // Min payload length = Min ethernet traffic length - header sum(eth+ip+udp+VT) = 64 - 18 - 20 - 8 - 6
)

type VTHeader struct {
	Flag uint16
	PLen uint16
	NLen uint16
}

type VTPacket struct {
	VTHeader
	Payload []byte
	Noise   []byte
}

// Payload should be the crypted data
func NewVTPkt(flag uint16, payload []byte, encipher cipher.Cipher) (*VTPacket, error) {
	var err error
	ePayload := payload

	if len(payload) != 0 {
		// Encrypt payload if not empty
		ePayload, err = encipher.Encrypt(payload)
		if err != nil {
			return nil, err
		}
	}

	// Build pkt
	pkt := new(VTPacket)
	pkt.Flag = flag
	pkt.PLen = uint16(len(ePayload))
	pkt.NLen = uint16(0)
	pkt.Payload = ePayload
	pkt.Noise = make([]byte, 0)

	// Add noise if payload not enough
	nLen := MIN_PLEN - len(payload)
	if nLen > 0 {
		pkt.NLen = uint16(nLen)
		pkt.Noise = make([]byte, nLen)
	}

	return pkt, nil
}

func (pkt *VTPacket) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	// BigEndian is easy to read
	if err := binary.Write(buf, binary.BigEndian, pkt.Flag); err != nil {
		return make([]byte, 0), err
	}

	if err := binary.Write(buf, binary.BigEndian, pkt.PLen); err != nil {
		return make([]byte, 0), err
	}

	if err := binary.Write(buf, binary.BigEndian, pkt.NLen); err != nil {
		return make([]byte, 0), err
	}

	stream := append(buf.Bytes(), pkt.Payload...)
	stream = append(stream, pkt.Noise...)
	return stream, nil
}

func Decode(stream []byte, encipher cipher.Cipher) (uint16, []byte, error) {
	sLen := len(stream)
	if sLen < (HDR_LEN + MIN_PLEN) {
		return HDR_FLG_UNKNOW, make([]byte, 0), errors.New("invalidate data fragment")
	}

	flag := binary.BigEndian.Uint16(stream[:2])
	pLen := binary.BigEndian.Uint16(stream[2:4])
	nLen := binary.BigEndian.Uint16(stream[4:6])

	if sLen != (HDR_LEN + int(pLen) + int(nLen)) {
		return HDR_FLG_UNKNOW, make([]byte, 0), errors.New("invalidate data fragment")
	}

	if pLen == 0 {
		// Return no payload
		return flag, make([]byte, 0), nil
	}

	ePayload := stream[6 : 6+int(pLen)]

	// Decrypt payload
	payload, err := encipher.Decrypt(ePayload)
	if err != nil {
		return HDR_FLG_UNKNOW, make([]byte, 0), fmt.Errorf("decrypt data %s", err)
	}
	return flag, payload, nil
}
