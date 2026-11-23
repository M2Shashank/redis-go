package validation

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"

	Error "github.com/codecrafters-io/redis-starter-go/internal/Errors"
)

func storeCmdArgumentsValidation(cmd *resp.Command) error {
	switch strings.ToUpper(cmd.Com) {
	case "BITCOUNT":
		if len(cmd.Args) < 1 || len(cmd.Args) > 3 {
			return Error.NewInvalidSyntaxError("bitcount")
		}
	case "ECHO", "TYPE", "KEYS", "GET", "INCR", "LLEN", "ZCARD", "STRLEN":
		if len(cmd.Args) != 1 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "SET":
		if len(cmd.Args) == 4 {
			var expiryType string = strings.ToUpper(cmd.Args[2])
			if expiryType != "EX" && expiryType != "PX" {
				return Error.NewInvalidArgumentsError(cmd.Com)
			}
		} else if len(cmd.Args) != 2 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "RPUSH", "LPUSH", "ZREM", "GEOPOS":
		if len(cmd.Args) < 2 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "LRANGE", "XRANGE", "ZRANGE", "SETBIT", "GEODIST":
		if len(cmd.Args) != 3 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "LPOP":
		if len(cmd.Args) != 1 && len(cmd.Args) != 2 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "BLPOP", "ZRANK", "ZSCORE", "GETBIT":
		if len(cmd.Args) != 2 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "XADD":
		if len(cmd.Args) < 3 || len(cmd.Args)%2 != 0 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "XREAD", "ZADD":
		if len(cmd.Args) < 3 || len(cmd.Args)%2 != 1 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "BITOP":
		if len(cmd.Args) < 3 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "GEOADD":
		if len(cmd.Args) < 4 || len(cmd.Args)%3 != 1 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "GEOSEARCH":
		if len(cmd.Args) != 7 || cmd.Args[1] != "FROMLONLAT" || cmd.Args[4] != "BYRADIUS" || cmd.Args[6] != "m" {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	default:
		return Error.ERR_INVALID
	}
	return nil
}