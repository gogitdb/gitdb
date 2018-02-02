package db

import "time"

type StringExpression func() string
type IndexFunction func() map[string]interface{}
type ModelConstructor func() ModelInterface

//ID interface for all schema structs
type ID struct {
	name     StringExpression
	blockId  StringExpression
	recordId StringExpression
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

func (a *ID) Init(def Schema) {
	a.name = def.Name
	a.blockId = def.Block
	a.recordId = def.Record
}

//Schema interface for all schemas structs
type Schema interface {
	Name() string
	Block() string
	Record() string
	Indexes() map[string]interface{}
}

//BaseSchema struct for all base schemas
type BaseSchema struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *BaseSchema) SetId(id string) {
	m.ID = id
}

func (m *BaseSchema) GetCreatedDate() time.Time {
	return m.CreatedAt
}

func (m *BaseSchema) GetUpdatedDate() time.Time {
	return m.UpdatedAt
}

func (m *BaseSchema) stampCreatedDate() {
	m.CreatedAt = time.Now()
}

func (m *BaseSchema) stampUpdatedDate() {
	m.UpdatedAt = time.Now()
}