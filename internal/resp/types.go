package resp

type RESPType interface {
	isRESPType()
}

type SimpleString string
type ErrorString string
type Integer int
type BulkString string
type Array []RESPType
type Null struct{}

func (SimpleString) isRESPType() {}
func (ErrorString) isRESPType() {}
func (Integer) isRESPType() {}
func (BulkString) isRESPType() {}
func (Array) isRESPType() {}
func (Null) isRESPType() {}