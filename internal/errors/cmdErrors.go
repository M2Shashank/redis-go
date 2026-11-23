package Errors

import (
	"errors"
)

var ERR_INVALID = errors.New("ERR Invalid or Unsupported Command")

func NewInvalidSyntaxError(command string) error {
	return errors.New("ERR wrong number of arguments for '" + command + "' command")
}

func NewInvalidArgumentsError(command string) error {
	return errors.New("ERR Invalid or Unsupported Arguments for '" + command + "' command")
}

// Authentication Related Errors
var ERR_NOAUTH = errors.New("NOAUTH Authentication required.")
var ERR_WRONGPASS = errors.New("WRONGPASS invalid username-password pair or user is disabled.")

// Client Transaction Related Errors
var ERR_WATCH_INSIDE_MULTI = errors.New("ERR WATCH inside MULTI is not allowed")
var ERR_MULTI_NESTED = errors.New("ERR MULTI calls can not be nested")
var ERR_EXEC_WITHOUT_MULTI = errors.New("ERR EXEC without MULTI")
var ERR_DISCARD_WITHOUT_MULTI = errors.New("ERR DISCARD without MULTI")

func ERRPermissionDenied(command string) error {
	return errors.New("ERR Can't execute '" + command + "': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context")
}
