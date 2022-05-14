package socks5

import (
	"bufio"
	"encoding/binary"
	"errors"
	"forward/internal/stream"
	"io"
	"net"
)

type StreamError struct {
	msg string
}

func (s StreamError) Error() string {
	return s.msg
}

var hand = []byte{0x05, 0x00}
var ack = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10}

//读 5 1 0 写回 5 0
func Handshake(reader *bufio.Reader, clientCon net.Conn) error {
	versionAndMethodNum, err := reader.Peek(2)
	if err != nil {
		return err
	}
	if versionAndMethodNum[0] == 0x05 {
		methodNum := int(versionAndMethodNum[1])
		reader.Discard(2)
		methodBytes := make([]byte, methodNum)
		_, err = io.ReadFull(reader, methodBytes)
		if err != nil {
			return err
		}
		stream.WriteAll(clientCon, hand)
		return nil
	} else {
		return StreamError{"版本号不为5"}
	}
}

func ParseRequest(reader *bufio.Reader, clientCon net.Conn) (string, int, error) {
	VerCmdRsvAtyp, err := reader.Peek(4)
	reader.Discard(4)
	if err != nil {
		return "", 0, err
	} else if VerCmdRsvAtyp[0] == 0x05 && VerCmdRsvAtyp[1] == 0x01 && VerCmdRsvAtyp[2] == 0x00 {
		if VerCmdRsvAtyp[3] == 3 {
			domainLen, err := reader.ReadByte()
			if err != nil {
				return "", 0, err
			}
			domainBytes := make([]byte, domainLen)
			_, err = io.ReadFull(reader, domainBytes)
			if err != nil {
				return "", 0, err
			}
			portBytes := make([]byte, 2)
			_, err = io.ReadFull(reader, portBytes)
			if err != nil {
				return "", 0, err
			}
			writeErr := stream.WriteAll(clientCon, ack)
			return string(domainBytes), int(binary.BigEndian.Uint16(portBytes)), writeErr
		} else if VerCmdRsvAtyp[3] == 1 {
			domainPortBytesLen := 6
			domainPortBytes := make([]byte, domainPortBytesLen)
			_, err = io.ReadFull(reader, domainPortBytes)
			if err != nil {
				return "", 0, err
			}
			writeErr := stream.WriteAll(clientCon, ack)
			return net.IPv4(domainPortBytes[0], domainPortBytes[1], domainPortBytes[2], domainPortBytes[3]).String(), int(binary.BigEndian.Uint16(domainPortBytes[4:])), writeErr
		} else {
			return "", 0, errors.New("不能处理ipv6")
		}

	} else {
		return "", 0, errors.New("不能处理非CONNECT请求")
	}
}
