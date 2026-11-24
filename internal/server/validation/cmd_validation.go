package validation

import (
	"strings"

	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func ValidateCmdArguments(cmd *resp.Command) error {

	switch strings.ToUpper(cmd.Com) {
	case "PSYNC", "WAIT", "PUBLISH", "AUTH":
		if len(cmd.Args) != 2 {
			return Error.ERR_InvalidSyntax(cmd.Com)
		}
	case "INFO", "PING", "UNWATCH", "MULTI", "EXEC", "DISCARD":
		if len(cmd.Args) > 1 {
			return Error.ERR_InvalidSyntax(cmd.Com)
		}
	case "SUBSCRIBE", "UNSUBSCRIBE", "WATCH":
		if len(cmd.Args) < 1 {
			return Error.ERR_InvalidSyntax(cmd.Com)
		}
	case "REPLCONF":
		if len(cmd.Args) < 2 {
			return Error.ERR_InvalidSyntax("replconf")
		}
		str := strings.ToUpper(cmd.Args[0])

		if str != "ACK" && str != "LISTENING-PORT" && str != "CAPA" && str != "GETACK" {
			return Error.ERR_InvalidSyntax("replconf")
		}
	case "CONFIG":
		if len(cmd.Args) != 2 {
			return Error.ERR_InvalidSyntax("config")
		}
		var typeOf string = cmd.Args[0]
		if typeOf != "GET" {
			return Error.ERR_InvalidArguments("config")
		}
	case "ACL":
		if len(cmd.Args) < 1 {
			return Error.ERR_InvalidSyntax("acl")
		}
		if cmd.Args[0] == "GETUSER" && len(cmd.Args) != 2 {
			return Error.ERR_InvalidSyntax("acl getuser")
		}
		if cmd.Args[0] == "SETUSER" && len(cmd.Args) < 3 {
			return Error.ERR_InvalidSyntax("acl setuser")
		}
		if cmd.Args[0] == "WHOAMI" && len(cmd.Args) != 1 {
			return Error.ERR_InvalidSyntax("acl whoami")
		}
	default:
		return storeCmdArgumentsValidation(cmd)
	}

	return nil
}
