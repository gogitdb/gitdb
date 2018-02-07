package db

type StringFunc func() string
type IndexFunction func() map[string]interface{}
type ModelConstructor func() Model

//ID interface for all schema structs
type ID struct {
	name     StringFunc
	blockId  StringFunc
	recordId StringFunc
	indexes  IndexFunction
}

func NewID(name StringFunc, block StringFunc, record StringFunc, indexes IndexFunction) *ID {
	return &ID{name, block, record, indexes}
}

func (a *ID) Name() string {
	return a.name()
}

func (a *ID) Id() string {
	return a.RecordId()
}

func (a *ID) RecordId() string {
	return a.BlockId() + "/" + a.recordId()
}

func (a *ID) BlockId() string {
	return a.name() + "/" + a.blockId()
}

func (a *ID) String() string {
	return a.RecordId()
}

func (a *ID) Indexes() map[string]interface{} {
	return a.indexes()
}
