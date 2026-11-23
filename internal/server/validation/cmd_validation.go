package validation

import (
	"strings"

	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func ValidateCmdArguments(cmd *resp.Command) error {

	switch strings.ToUpper(cmd.Com) {
	case "PSYNC", "WAIT":
		if len(cmd.Args) != 2 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "REPLCONF":
		if len(cmd.Args) < 2 {
			return Error.NewInvalidSyntaxError("replconf")
		}
		str := strings.ToUpper(cmd.Args[0])

		if str != "ACK" && str != "LISTENING-PORT" && str != "CAPA" && str != "GETACK" {
			return Error.NewInvalidSyntaxError("replconf")
		}
	case "INFO", "PING":
		if len(cmd.Args) > 1 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "SUBSCRIBE", "UNSUBSCRIBE", "WATCH", "UNWATCH", "MULTI", "EXEC", "DISCARD":
		if len(cmd.Args) < 1 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "PUBLISH", "AUTH":
		if len(cmd.Args) != 2 {
			return Error.NewInvalidSyntaxError(cmd.Com)
		}
	case "CONFIG":
		if len(cmd.Args) != 2 {
			return Error.NewInvalidSyntaxError("config")
		}
		var typeOf string = cmd.Args[0]
		if typeOf != "GET" {
			return Error.NewInvalidArgumentsError("config")
		}
	case "ACL":
		if len(cmd.Args) < 1 {
			return Error.NewInvalidSyntaxError("acl")
		}
		if cmd.Args[0] == "GETUSER" && len(cmd.Args) != 2 {
			return Error.NewInvalidSyntaxError("acl getuser")
		}
		if cmd.Args[0] == "SETUSER" && len(cmd.Args) < 3 {
			return Error.NewInvalidSyntaxError("acl setuser")
		}
		if cmd.Args[0] == "WHOAMI" && len(cmd.Args) != 1 {
			return Error.NewInvalidSyntaxError("acl whoami")
		}
	default:
		return storeCmdArgumentsValidation(cmd)
	}

	return nil
}
