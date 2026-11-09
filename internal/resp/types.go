package resp

type RESPType interface {
	isRESPType()
}

type SimpleString string
type SimpleError string
type Integer int
type BulkString string
type Array []RESPType
type Null struct{}
type RDBFileFormat string

func (SimpleString) isRESPType() {}
func (SimpleError) isRESPType() {}
func (Integer) isRESPType() {}
func (BulkString) isRESPType() {}
func (Array) isRESPType() {}
func (Null) isRESPType() {}
func (RDBFileFormat) isRESPType() {}