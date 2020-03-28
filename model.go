package gitdb

import (
	"time"
)

//Model interface describes methods GitDB supports
type Model interface {
	ID() string
	SetID(string)
	SetMetaData(m *metaData)
	String() string
	Validate() bool
	IsLockable() bool
	GetLockFileNames() []string
	GetSchema() *Schema
	GetCreatedDate() time.Time
	GetUpdatedDate() time.Time
	SetCreatedDate(time.Time)
	SetUpdatedDate(time.Time)
	ShouldEncrypt() bool
	GetValidationErrors() []error
}

type metaData struct {
	Indexes   map[string]interface{}
	Encrypted bool
}

//BaseModel provides sensible defaults for a Model struct by composition
type BaseModel struct {
	RecordID  string `json:"ID"`
	Meta      *metaData
	CreatedAt time.Time
	UpdatedAt time.Time
	errors    []error
}

//ID returns record id
func (m *BaseModel) ID() string {
	return m.RecordID
}

//SetID sets record id
func (m *BaseModel) SetID(id string) {
	m.RecordID = id
}

//SetMetaData sets metadata
func (m *BaseModel) SetMetaData(md *metaData) {
	m.Meta = md
}

//GetCreatedDate returns created time of Model
func (m *BaseModel) GetCreatedDate() time.Time {
	return m.CreatedAt
}

//GetUpdatedDate returns updated time of Model
func (m *BaseModel) GetUpdatedDate() time.Time {
	return m.UpdatedAt
}

//SetCreatedDate sets created time of Model
func (m *BaseModel) SetCreatedDate(t time.Time) {
	m.CreatedAt = t
}

//SetUpdatedDate sets updated time of Model
func (m *BaseModel) SetUpdatedDate(t time.Time) {
	m.UpdatedAt = t
}

func (m *BaseModel) String() string {
	return m.RecordID
}

//Validate validates a Model
//this func is here provide a sensible default
func (m *BaseModel) Validate() bool {
	return true
}

//IsLockable informs GitDb if a Model support locking
//this func is here provide a sensible default
func (m *BaseModel) IsLockable() bool {
	return false
}

//GetLockFileNames informs GitDb of files a Models using for locking
//It only works with lockable Models
//this func is here provide a sensible default
func (m *BaseModel) GetLockFileNames() []string {
	return []string{}
}

//ShouldEncrypt informs GitDb if a Model support encryption
//this func is here provide a sensible default
func (m *BaseModel) ShouldEncrypt() bool {
	return false
}

//GetValidationErrors returns any errors from Validate
//this func is here provide a sensible default
func (m *BaseModel) GetValidationErrors() []error {
	return m.errors
}
