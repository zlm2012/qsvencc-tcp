package qsvencc_tcp

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
)

func readByte(r io.Reader) (byte, error) {
	rb := []byte{0}
	i, err := r.Read(rb)
	if err != nil {
		return 0, err
	}
	if i > 1 {
		log.Fatal("unexpected read length: ", i)
	}
	return rb[0], nil
}

func TcpMsgDecode(r io.Reader) (TcpCmd, []byte, error) {
	rb, err := readByte(r)
	if err != nil {
		return QTWReserved, nil, err
	}
	cmd := TcpCmd(rb)
	switch cmd {
	case QTWExec:
		fallthrough
	case QTWPipe:
		fallthrough
	case QTWPipeIdentifier:
		fallthrough
	case QTWStderr:
		lenb := []byte{0, 0}
		_, err = io.ReadFull(r, lenb)
		if err != nil {
			return QTWReserved, nil, err
		}
		leni := binary.BigEndian.Uint16(lenb)
		content := make([]byte, leni, leni)
		_, err = io.ReadFull(r, content)
		if err != nil {
			return QTWReserved, nil, err
		}
		return cmd, content, nil
	case QTWRun:
		fallthrough
	case QTWPipeReady:
		return cmd, nil, nil
	case QTWCmdReturnCode:
		retb := make([]byte, 4, 4)
		_, err = io.ReadFull(r, retb)
		if err != nil {
			return QTWReserved, nil, err
		}
		return cmd, retb, nil
	default:
		return QTWReserved, nil, errors.New("not supported cmd")
	}
}
