package resp

import (
	"strings"
)

// first element of the array -> command, the rest -> arguments.
type Command struct {
	Com  string
	Args []string
}

// parseRESPCommand parses an array of bulk strings into a Command struct
// It returns the number of bytes consumed and a pointer to the Command struct.
// Assumes a request is an array of bulk strings (i.e. starts with '*')
func ParseRESPCommand(req []byte) (int, *Command) {
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

	var cmd Command = Command{
		Args: make([]string, 0, numElem-1),
	}

	for j := 0; j < numElem; j++ {
		var bulkString string = ""
		if req[i] == '$' {
			i++

			var size int = 0
			for ; i < len(req) && req[i] != '\r'; i++ {
				size = size*10 + int(req[i]-'0')
			}
			i += 2 // skip \r\n
			if i+size+2 > len(req) {
				return 0, nil
			}

			bulkString = string(req[i : i+size])
			i = i + size + 2 // skip \r\n
		}

		if j == 0 {
			cmd.Com = strings.ToUpper(bulkString)
		} else {
			cmd.Args = append(cmd.Args, bulkString)
		}

		if i >= len(req) {
			break
		}
	}

	return i, &cmd
}
