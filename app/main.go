package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// first element of the array -> command, the rest -> arguments.
type ReqCmd struct {
	Cmd  string
	Args []string
}

// parseCmd parses an array of bulk strings into a ReqCmd struct
// It returns the number of bytes consumed and a pointer to the ReqCmd struct.
// Assumes a request is an array of bulk strings (i.e. starts with '*')
func parseCmd(req []byte) (int, *ReqCmd) {
	if len(req) == 0 || req[0] != '*' {
		return 0, nil
	}
	var i int = 1

	var numElem int = 0
	for ; i < len(req) && req[i] != '\r'; i++ {
		numElem = numElem*10 + int(req[i]-'0')
	}
	i += 2 // skip \r\n

	if numElem <= 0 {
		return 0, nil
	}

	var cmd ReqCmd = ReqCmd{
		Args: make([]string, 0, numElem-1),
	}

	for j := 0; j < numElem; j++ {
		var val string = ""
		if req[i] == '$' {
			i++

			var size int = 0
			for ; i < len(req) && req[i] != '\r'; i++ {
				size = size*10 + int(req[i]-'0')
			}
			i += 2 // skip \r\n

			val = string(req[i : i+size])
			if i+size+2 > len(req) {
				return 0, nil
			}
			i = i + size + 2 // skip \r\n
		}

		if j == 0 {
			cmd.Cmd = strings.ToLower(val)
		} else {
			cmd.Args = append(cmd.Args, val)
		}
	}

	return i, &cmd
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	fmt.Println("Server listening on port 6379")

	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	var buff []byte = make([]byte, 1024)
	var pending []byte

	for {
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			break
		}

		pending = append(pending, buff[:n]...)

		for {
			n, cmd := parseCmd(pending)
			if cmd == nil || n == 0 {
				break
			}

			// ------ Handle commands ------
			if cmd.Cmd == "ping" {
				conn.Write([]byte("+PONG\r\n"))
			}

			if cmd.Cmd == "echo" {
				if len(cmd.Args) != 1 {
					conn.Write([]byte("-ERR Invalid Command\r\n"))
				} else {
					var arg string = cmd.Args[0]
					var resp string = fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
					conn.Write([]byte(resp))
				}
			}
			// ---------------------------

			pending = pending[n:]
		}
	}
}
