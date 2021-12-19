package qsvencc_tcp

import (
	"encoding/binary"
	"errors"
	"math"
)

func TcpMsgVecEncode(cmd TcpCmd, content []byte) ([]byte, error) {
	if cmd != QTWExec && cmd != QTWPipe && cmd != QTWPipeIdentifier && cmd != QTWStderr {
		return nil, errors.New("cmd not supported")
	}
	if len(content) > math.MaxUint16 {
		return nil, errors.New("content is over max length supported")
	}
	result := make([]byte, 3)
	result[0] = byte(cmd)
	binary.BigEndian.PutUint16(result[1:], uint16(len(content)))
	return append(result, content...), nil
}

func TcpMsgReturnCodeEncode(returnCode int) []byte {
	result := make([]byte, 5)
	result[0] = byte(QTWCmdReturnCode)
	binary.BigEndian.PutUint32(result[1:], uint32(returnCode))
	return result
}
