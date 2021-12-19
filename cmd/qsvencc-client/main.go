package main

import (
	"encoding/binary"
	"encoding/json"
	"github.com/jessevdk/go-flags"
	qsvencc_tcp "github.com/zlm2012/qsvencc-tcp"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var opts struct {
	Input  string `short:"i" long:"input" description:"input file" required:"true"`
	Output string `short:"o" long:"output" description:"output file" required:"true"`
}

func main() {
	parser := flags.NewParser(&opts, flags.None)
	remainedArgs := make([]string, 0)
	parser.UnknownOptionHandler = func(option string, arg flags.SplitArgument, args []string) ([]string, error) {
		log.Println(option)
		if len(option) == 1 {
			remainedArgs = append(remainedArgs, "-"+option)
		} else {
			remainedArgs = append(remainedArgs, "--"+option)
		}
		if value, exists := arg.Value(); exists {
			remainedArgs = append(remainedArgs, value)
		} else if !strings.HasPrefix(args[0], "-") {
			remainedArgs = append(remainedArgs, args[0])
			args = args[1:]
		}
		return args, nil
	}
	_, err := parser.ParseArgs(os.Args)
	if err != nil {
		log.Fatalln("error on parse args; ", err)
	}
	var inputReader io.Reader = os.Stdin
	var outputWriter io.Writer = os.Stdout
	if opts.Input != "-" {
		inputReader, err = os.OpenFile(opts.Input, os.O_RDONLY, 0644)
		if err != nil {
			log.Fatalln("failed on read input file; ", err)
		}
	}
	if opts.Output != "-" {
		outputWriter, err = os.OpenFile(opts.Output, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalln("failed on write output file; ", err)
		}
	}

	cmdArgsJson, err := json.Marshal(remainedArgs)
	if err != nil {
		log.Fatalln("failed on marshaling args; ", err)
	}
	cmdArgsJson = append(cmdArgsJson, '\n')
	log.Println("marshaled args: ", string(cmdArgsJson))

	tcpAddr, err := net.ResolveTCPAddr("tcp", "192.168.122.1:11111")
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalln("failed on calling server; ", err)
	}
	defer conn.Close()
	content, err := qsvencc_tcp.TcpMsgVecEncode(qsvencc_tcp.QTWExec, cmdArgsJson)
	if err != nil {
		log.Fatalln("failed on encode exec request; ", err)
	}
	_, err = conn.Write(content)
	if err != nil {
		log.Fatalln("failed on write cmd args; ", err)
	}
	cmd, pipeId, err := qsvencc_tcp.TcpMsgDecode(conn)
	if cmd != qsvencc_tcp.QTWPipeIdentifier {
		log.Fatalln("unexpected response cmd, waiting for pipe identifier; ", err)
	}

	pipeConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalln("failed on calling server for pipe; ", err)
	}
	defer pipeConn.Close()
	content, err = qsvencc_tcp.TcpMsgVecEncode(qsvencc_tcp.QTWPipe, pipeId)
	if err != nil {
		log.Fatalln("failed on encode exec request; ", err)
	}
	_, err = pipeConn.Write(content)
	if err != nil {
		log.Fatalln("failed on write cmd args; ", err)
	}
	cmd, _, err = qsvencc_tcp.TcpMsgDecode(pipeConn)
	if cmd != qsvencc_tcp.QTWPipeReady {
		log.Fatalln("unexpected response cmd, waiting for pipe identifier; ", err)
	}
	err = pipeConn.SetNoDelay(false)
	if err != nil {
		log.Fatalln("failed on enable nagle")
	}
	go func() {
		io.Copy(pipeConn, inputReader)
		pipeConn.CloseWrite()
	}()
	go io.Copy(outputWriter, pipeConn)

	_, err = conn.Write([]byte{byte(qsvencc_tcp.QTWRun)})
	if err != nil {
		log.Fatalln("failed on run", err)
	}
	var retcode = 0
	for true {
		cmd, content, err := qsvencc_tcp.TcpMsgDecode(conn)
		if err != nil {
			log.Fatalln("failed on read conn; ", err)
		}
		if cmd == qsvencc_tcp.QTWStderr {
			os.Stderr.Write(content)
		} else if cmd == qsvencc_tcp.QTWCmdReturnCode {
			retcode = int(binary.BigEndian.Uint32(content))
			break
		} else {
			log.Fatalln("unexpected msg; ", cmd)
		}
	}
	os.Exit(retcode)
}
