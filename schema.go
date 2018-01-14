package db

import "time"

type StringExpression func() string
type IndexFunction func() map[string]interface{}
type ModelConstructor func() ModelInterface

//Schema interface for all schema structs
type Schema struct {
	Name       StringExpression
	Block      StringExpression
	Record     StringExpression
	Indexes    IndexFunction
	Definition Definition
}

func (a *Schema) Id() string {
	return a.Name() + "/" + a.Block() + "/" + a.Record()
}

func (a *Schema) RecordId() string {
	return a.Name() + "/" + a.Block() + "/" + a.Record()
}

func (a *Schema) BlockId() string {
	return a.Name() + "/" + a.Block()
}

func (a *Schema) SetDef(def Definition) {
	a.Definition = def
	a.Name = def.Name
	a.Block = def.Block
	a.Record = def.Record
	a.Indexes = def.Indexes
}

//Definition interface for all def structs
type Definition interface {
	Name() string
	Block() string
	Record() string
	Indexes() map[string]interface{}
}

//BaseDef struct for all base schemas
type BaseDef struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *BaseDef) SetId(id string) {
	m.ID = id
}

func (m *BaseDef) GetCreatedDate() time.Time {
	return m.CreatedAt
}

func (m *BaseDef) GetUpdatedDate() time.Time {
	return m.UpdatedAt
}

func (m *BaseDef) StampCreatedDate() {
	m.CreatedAt = time.Now()
}

func (m *BaseDef) StampUpdatedDate() {
	m.UpdatedAt = time.Now()
}

func (m *BaseDef) TimeStamp() {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
}
