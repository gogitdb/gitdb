package db

import (
	"time"
)

type Model interface {
	SetId(string)
	String() string
	Validate() bool
	IsLockable() bool
	GetLockFileNames() []string
	GetID() *ID
	GetCreatedDate() time.Time
	GetUpdatedDate() time.Time
	stampCreatedDate()
	stampUpdatedDate()
	ShouldEncrypt() bool
	Init(schema Schema)
}

type ModelSchema interface {
	Model
	Schema
}

type BaseModel struct {
	//extends..
	_Id *ID
}

func (m *BaseModel) String() string {
	return m.GetID().String()
}

func (m *BaseModel) Validate() bool {
	return true
}

func (m *BaseModel) IsLockable() bool {
	return false
}

func (m *BaseModel) GetLockFileNames() []string {
	return []string{}
}

func (m *BaseModel) GetID() *ID {
	if m._Id == nil {
		m._Id = &ID{}
	}

	return m._Id
}

func (m *BaseModel) Init(schema Schema) {
	m.GetID().Init(schema)
}

func (m *BaseModel) ShouldEncrypt() bool {
	return false
}
