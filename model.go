package db

import (
	"time"
)

type ModelInterface interface {
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
	Indexes() map[string]interface{}
	Init(schema Schema)
}

type Model struct {
	//extends..
	_Id *ID
}

func (m *Model) String() string {
	return m.GetID().String()
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

func (m *Model) GetID() *ID {
	if m._Id == nil {
		m._Id = &ID{}
	}

	return m._Id
}

func (m *Model) Init(schema Schema) {
	m.GetID().Init(schema)
}

func (m *Model) ShouldEncrypt() bool {
	return false
}
