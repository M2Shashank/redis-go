package resp

type RESPType interface {
	isRESPType()
	Type() string
}

type SimpleString string
type SimpleError string
type Integer int
type BulkString string
type Array []RESPType
type Null struct{}

func (SimpleString) isRESPType() {}
func (SimpleError) isRESPType() {}
func (Integer) isRESPType() {}
func (BulkString) isRESPType() {}
func (Array) isRESPType() {}
func (Null) isRESPType() {}

func (s SimpleString) Type() string { return "SimpleString" }
func (s SimpleError) Type() string  { return "SimpleError" }
func (i Integer) Type() string      { return "Integer" }
func (b BulkString) Type() string   { return "BulkString" }
func (a Array) Type() string        { return "Array" }
func (n Null) Type() string         { return "Null" }