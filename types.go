package qsvencc_tcp

type TcpCmd byte

const (
	QTWReserved TcpCmd = iota
	QTWExec
	QTWPipe
	QTWRun
)

const (
	QTWPipeIdentifier TcpCmd = iota + 100
	QTWCmdReturnCode
	QTWPipeReady
	QTWStderr
)
