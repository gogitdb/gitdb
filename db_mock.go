package gitdb

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bouggo/log"
	"github.com/fobilow/gitdb/v2/internal/db"
)

type mockdb struct {
	config Config
	data   map[string]Model
	index  map[string]map[string]interface{}
	locks  map[string]bool
}

func newMockConnection() *mockdb {
	db := &mockdb{
		data:  make(map[string]Model),
		index: make(map[string]map[string]interface{}),
		locks: make(map[string]bool),
	}
	return db
}

func (g *mockdb) Close() error {
	g.data = nil
	return nil
}

func (g *mockdb) Insert(m Model) error {
	g.data[ID(m)] = m

	for name, value := range m.GetSchema().indexes {
		key := m.GetSchema().dataset + "." + name
		if _, ok := g.index[key]; !ok {
			g.index[key] = make(map[string]interface{})
		}
		g.index[key][ID(m)] = value
	}

	return nil
}

func (g *mockdb) InsertMany(m []Model) error {
	for _, model := range m {
		g.Insert(model)
	}
	return nil
}

func (g *mockdb) Get(id string, m Model) error {
	model, exists := g.data[id]
	if exists {
		m = model
		return nil
	}

	dataset, _, _, _ := ParseID(id)
	return fmt.Errorf("Record %s not found in %s", id, dataset)
}

func (g *mockdb) Exists(id string) error {
	_, exists := g.data[id]
	if !exists {
		dataset, _, _, _ := ParseID(id)
		return fmt.Errorf("Record %s not found in %s", id, dataset)
	}

	return nil
}

func (g *mockdb) Fetch(dataset string) ([]*db.Record, error) {
	result := []*db.Record{}
	for id, model := range g.data {
		ds, _, _, err := ParseID(id)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if ds == dataset {
			result = append(result, db.ConvertModel(ID(model), model))
		}
	}

	return result, nil
}

func (g *mockdb) Search(dataset string, searchParams []*SearchParam, searchMode SearchMode) ([]*db.Record, error) {
	result := []*db.Record{}
	for _, searchParam := range searchParams {
		key := dataset + "." + searchParam.Index

		queryValue := strings.ToLower(searchParam.Value)
		for recordID, value := range g.index[key] {
			addResult := false
			dbValue := strings.ToLower(value.(string))
			switch searchMode {
			case SearchEquals:
				addResult = dbValue == queryValue
			case SearchContains:
				addResult = strings.Contains(dbValue, queryValue)
			case SearchStartsWith:
				addResult = strings.HasPrefix(dbValue, queryValue)
			case SearchEndsWith:
				addResult = strings.HasSuffix(dbValue, queryValue)
			}

			if addResult {
				result = append(result, db.ConvertModel(recordID, g.data[recordID]))
			}
		}
	}

	return result, nil
}

func (g *mockdb) Delete(id string) error {
	delete(g.data, id)
	return nil
}

func (g *mockdb) DeleteOrFail(id string) error {
	_, exists := g.data[id]
	if !exists {
		return fmt.Errorf("record %s does not exist", id)
	}

	delete(g.data, id)
	return nil
}

func (g *mockdb) Lock(m Model) error {

	if !m.IsLockable() {
		return errors.New("Model is not lockable")
	}

	for _, l := range m.GetLockFileNames() {
		key := m.GetSchema().dataset + "." + l
		if _, ok := g.locks[l]; !ok {
			g.locks[key] = true
		} else {
			return errors.New("Lock file already exist: " + l)
		}
	}
	return nil
}

func (g *mockdb) Unlock(m Model) error {
	if !m.IsLockable() {
		return errors.New("Model is not lockable")
	}

	for _, l := range m.GetLockFileNames() {
		key := m.GetSchema().dataset + "." + l
		delete(g.locks, key)
	}
	return nil
}

func (g *mockdb) GetMails() []*mail {
	return []*mail{}
}

func (g *mockdb) StartTransaction(name string) *transaction {
	//todo return mock transaction
	return nil
}

func (g *mockdb) GetLastCommitTime() (time.Time, error) {
	return time.Now(), nil
}

func (g *mockdb) SetUser(user *User) error {
	g.config.User = user
	return nil
}

func (g *mockdb) Migrate(from Model, to Model) error {
	return nil
}

func (g *mockdb) Config() Config {
	return g.config
}

func (g *mockdb) configure(cfg Config) {
	if len(cfg.ConnectionName) == 0 {
		cfg.ConnectionName = defaultConnectionName
	}
	g.config = cfg
}
