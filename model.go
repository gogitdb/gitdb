package db

import (
	"time"
)

type DataFormat string

const (
	JSON DataFormat = "json"
	BSON DataFormat = "bson"
	CSV  DataFormat = "csv"
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
	GetDataFormat() DataFormat
	GetValidationErrors() []error
}

type BaseModel struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	errors []error
}

func (m *BaseModel) Id() string {
	return m.ID
}

func (m *BaseModel) SetId(id string) {
	m.ID = id
}

func (m *BaseModel) GetCreatedDate() time.Time {
	return m.CreatedAt
}

func (m *BaseModel) GetUpdatedDate() time.Time {
	return m.UpdatedAt
}

func (m *BaseModel) stampCreatedDate() {
	m.CreatedAt = time.Now()
}

func (m *BaseModel) stampUpdatedDate() {
	m.UpdatedAt = time.Now()
}

func (m *BaseModel) String() string {
	return m.ID
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

func (m *BaseModel) ShouldEncrypt() bool {
	return false
}

func (m *BaseModel) GetDataFormat() DataFormat {
	return JSON
}

func (m *BaseModel) GetValidationErrors() []error {
	return m.errors
}

type collection []Model

func (c collection) Len() int {
	return len(c)
}
func (c collection) Less(i, j int) bool {
	return c[i].GetCreatedDate().Before(c[j].GetCreatedDate())
}
func (c collection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
