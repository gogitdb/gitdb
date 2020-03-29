package gitdb

import (
	"time"
)

//Model interface describes methods GitDB supports
type Model interface {
	GetSchema() *Schema
	//Validate validates a Model
	Validate() error
	//IsLockable informs GitDb if a Model support locking
	IsLockable() bool
	//GetLockFileNames informs GitDb of files a Models using for locking
	GetLockFileNames() []string
	//ShouldEncrypt informs GitDb if a Model support encryption
	ShouldEncrypt() bool
	//SetBaseModel sets shared fields and is called by gitdb before insert
	SetBaseModel()
}

//TimeStampedModel provides time stamp fields
type TimeStampedModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

//SetBaseModel sets shared fields and is called by gitdb before insert
func (m *TimeStampedModel) SetBaseModel() {
	stampTime := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = stampTime
	}
	m.UpdatedAt = stampTime
}

type gRecord struct {
	Indexes map[string]interface{}
	Data    Model
}

func wrapModel(m Model) *gRecord {
	return &gRecord{
		Indexes: m.GetSchema().indexes(),
		Data:    m,
	}
}

func (m *gRecord) GetSchema() *Schema {
	return m.Data.GetSchema()
}

func (m *gRecord) Validate() error {
	return m.Data.Validate()
}
func (m *gRecord) IsLockable() bool {
	return m.Data.IsLockable()
}
func (m *gRecord) ShouldEncrypt() bool {
	return m.Data.ShouldEncrypt()
}
func (m *gRecord) GetLockFileNames() []string {
	return m.Data.GetLockFileNames()
}

func (m *gRecord) SetBaseModel() {
	m.Data.SetBaseModel()
}
