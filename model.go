package gitdb

import (
	"time"
)

//Model interface describes methods GitDB supports
type Model interface {
	GetSchema() *Schema
	//Validate validates a Model
	Validate() error
	//ShouldEncrypt informs GitDb if a Model support encryption
	ShouldEncrypt() bool
	//BeforeInsert is called by gitdb before insert
	BeforeInsert() error
}

type LockableModel interface {
	//GetLockFileNames informs GitDb of files a Models using for locking
	GetLockFileNames() []string
}

//TimeStampedModel provides time stamp fields
type TimeStampedModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

//BeforeInsert implements Model.BeforeInsert
func (m *TimeStampedModel) BeforeInsert() error {
	stampTime := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = stampTime
	}
	m.UpdatedAt = stampTime

	return nil
}

type model struct {
	Version string
	Data    Model
}

func wrap(m Model) *model {
	return &model{
		Version: RecVersion,
		Data:    m,
	}
}

func (m *model) GetSchema() *Schema {
	return m.Data.GetSchema()
}

func (m *model) Validate() error {
	return m.Data.Validate()
}

func (m *model) ShouldEncrypt() bool {
	return m.Data.ShouldEncrypt()
}

func (m *model) BeforeInsert() error {
	err := m.Data.BeforeInsert()
	return err
}

func (g *gitdb) RegisterModel(dataset string, m Model) bool {
	if g.registry == nil {
		g.registry = make(map[string]Model)
	}
	g.registry[dataset] = m
	return true
}

func (g *gitdb) isRegistered(dataset string) bool {
	if _, ok := g.registry[dataset]; ok {
		return true
	}

	if g.config.Factory != nil && g.config.Factory(dataset) != nil {
		return true
	}

	return false
}
