package db

import (
	"time"
)

type ModelInterface interface {
	Id() string
	SetId(string)
	String() string
	Validate() bool
	IsLockable() bool
	GetLockFileNames() []string
	SetSchema(*Schema)
	GetSchema() *Schema
	GetCreatedDate() time.Time
	GetUpdatedDate() time.Time
	StampCreatedDate()
	StampUpdatedDate()
	ShouldEncrypt() bool
}

type Model struct {
	//extends..
	schema *Schema
}

func (m *Model) Id() string {
	return m.GetSchema().Id()
}

func (m *Model) String() string {
	return m.GetSchema().Id()
}

func (m *Model) Validate() bool {
	return true
}

func (m *Model) IsLockable() bool {
	return false
}

func (m *Model) GetLockFileNames() []string {
	return []string{}
}

func (m *Model) SetSchema(s *Schema) {
	m.schema = s
}

func (m *Model) GetSchema() *Schema {
	if m.schema == nil {
		m.schema = &Schema{}
		//m.schema.SetDef(m)
	}

	return m.schema
}

func (m *Model) ShouldEncrypt() bool {
	return false
}
