package main

import (
	"encoding/json"
	"github.com/google/uuid"
	qsvencc_tcp "github.com/zlm2012/qsvencc-tcp"
	"io"
	"log"
	"net"
	"os/exec"
)

type procCtxt struct {
	cmd      *exec.Cmd
	ready    bool
	finished chan interface{}
}

var procIdMap = map[string]*procCtxt{}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "192.168.122.1:11111")
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		// クライアントからのコネクション情報を受け取る
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
		}

		// ハンドラーに接続情報を渡す
		go handler(conn)
	}
}

func execHandler(content []byte, conn *net.TCPConn) {
	var cmdArgs []string
	err := json.Unmarshal(content, &cmdArgs)
	if err != nil {
		log.Println("failed on parse cmdline: ", err)
		return
	}
	log.Println("cmd line: ", cmdArgs)
	cmdArgs = append([]string{"-i", "-", "-o", "-"}, cmdArgs...)
	qsvCmd := exec.Command("qsvencc", cmdArgs...)
	pipeId := uuid.New().String()
	procIdMap[pipeId] = &procCtxt{qsvCmd, false, make(chan interface{})}
	writeContent, err := qsvencc_tcp.TcpMsgVecEncode(qsvencc_tcp.QTWPipeIdentifier, []byte(pipeId))
	if err != nil {
		log.Println("failed on encoding pipe id: ", err)
		delete(procIdMap, pipeId)
		return
	}
	_, err = conn.Write(writeContent)
	if err != nil {
		log.Println("failed on writing on tcp conn: ", err)
		delete(procIdMap, pipeId)
		return
	}
	defer func() {
		if !procIdMap[pipeId].ready {
			delete(procIdMap, pipeId)
		} else {
			procIdMap[pipeId].finished <- true
		}
	}()
	// waiting for run request after pipe setup
	cmd, content, err := qsvencc_tcp.TcpMsgDecode(conn)
	if err != nil {
		log.Println("failed on waiting run command: ", err)
		return
	}
	if cmd != qsvencc_tcp.QTWRun {
		log.Println("unexpected command", cmd, "exit")
		return
	}
	if !procIdMap[pipeId].ready {
		log.Println("pipe not initialized, exit")
		return
	}
	pr, pw := io.Pipe()
	qsvCmd.Stderr = pw
	stderrBuf := make([]byte, 1024, 1024)
	log.Println("ready to run", pipeId, qsvCmd)
	err = qsvCmd.Start()
	if err != nil {
		log.Println("failed on run command: ", err)
	}
	go func() {
		for {
			n, err := pr.Read(stderrBuf[:])
			if err != nil && err != io.EOF {
				log.Println("failed on reading from stderr: ", err)
				qsvCmd.Process.Kill()
				break
			}
			if n > 0 {
				content, _ = qsvencc_tcp.TcpMsgVecEncode(qsvencc_tcp.QTWStderr, stderrBuf[:n])
				_, err2 := conn.Write(content)
				if err2 != nil {
					qsvCmd.Process.Kill()
					log.Println("failed on writing to tcp conn: ", err)
				}
			}
			if err == io.EOF {
				log.Println("stderr eof")
				break
			}
		}
	}()
	log.Println("wait for proc exit")
	qsvCmd.Wait()
	log.Println("exited")
	conn.Write(qsvencc_tcp.TcpMsgReturnCodeEncode(qsvCmd.ProcessState.ExitCode()))
	log.Println("finished on running", pipeId)
}

func pipeHandler(content []byte, conn *net.TCPConn) {
	procId := string(content)
	ctxt, ok := procIdMap[procId]
	if !ok {
		log.Println("proc id not found, exit")
		return
	}
	ctxt.cmd.Stdin = conn
	ctxt.cmd.Stdout = conn
	ctxt.ready = true
	_, err := conn.Write([]byte{byte(qsvencc_tcp.QTWPipeReady)})
	if err != nil {
		log.Println("failed on send ready: ", err)
		ctxt.ready = false
		return
	}
	log.Println("pipe ready for", procId, ctxt)
	defer func() {
		delete(procIdMap, procId)
	}()
	_ = <-ctxt.finished
}

func handler(conn *net.TCPConn) {
	defer conn.Close()
	cmd, content, err := qsvencc_tcp.TcpMsgDecode(conn)
	if err != nil {
		log.Println("failed on decode msg", err)
		return
	}
	switch cmd {
	case qsvencc_tcp.QTWExec:
		execHandler(content, conn)
	case qsvencc_tcp.QTWPipe:
		pipeHandler(content, conn)
	default:
		log.Println("unsupported pattern of request, exit")
		return
	}
}
