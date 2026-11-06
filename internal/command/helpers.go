package command

import (
	//"errors"
	"fmt"
	"strconv"
	"strings"
)

func validateStreamID(ID string, command string) (int64, int64, error) {
	var mst, seq int64
	if ID == "-" {
		return -1, -1, nil
	}
	if ID == "+" {
		return -2, -2, nil
	}

	if strings.Contains(ID, "-") {
		var parts []string = strings.Split(ID, "-")
		if len(parts) != 2 {
			return -1, -1, fmt.Errorf("ERR Invalid Arguments for %s command", command)
		}
		val1, err1 := strconv.ParseInt(parts[0], 10, 64)
		val2, err2 := strconv.ParseInt(parts[1], 10, 64)
		if err1 != nil || err2 != nil {
			return -1, -1, fmt.Errorf("ERR Invalid Arguments for %s command", command)
		}
		mst = val1
		seq = val2
	} else {
		val, err := strconv.ParseInt(ID, 10, 64)
		if err != nil {
			return -1, -1, fmt.Errorf("ERR Invalid Arguments for %s command", command)
		}
		mst = val
		seq = -1
	}
	return mst, seq, nil
}