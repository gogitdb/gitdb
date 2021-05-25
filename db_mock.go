package gitdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/bouggo/log"
	"github.com/gogitdb/gitdb/v2/internal/db"
)

type mockdb struct {
	config Config
	data   map[string]Model
	index  map[string]map[string]interface{}
	locks  map[string]bool
}

type mocktransaction struct {
	name       string
	operations []operation
	db         *mockdb
}

func (t *mocktransaction) Commit() error {
	for _, o := range t.operations {
		if err := o(); err != nil {
			log.Info("Reverting transaction: " + err.Error())
			return err
		}
	}
	return nil
}

func (t *mocktransaction) AddOperation(o operation) {
	t.operations = append(t.operations, o)
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

func (g *mockdb) Get(id string, result Model) error {

	if reflect.ValueOf(result).Kind() != reflect.Ptr || reflect.ValueOf(result).IsNil() {
		return errors.New("Second argument to Get must be a non-nil pointer")
	}

	model, exists := g.data[id]
	if exists {

		resultType := reflect.ValueOf(result).Type()
		modelType := reflect.ValueOf(model).Type()
		if resultType != modelType {
			return fmt.Errorf("Second argument to Get must be of type %v", modelType)
		}

		b, err := json.Marshal(model)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, result)
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

func (g *mockdb) Fetch(dataset string, blocks ...string) ([]*db.Record, error) {
	var result []*db.Record
	blockStream := "|" + strings.Join(blocks, "|") + "|"
	for id, model := range g.data {
		ds, b, _, err := ParseID(id)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if ds == dataset && strings.Contains(blockStream, "|"+b+"|") {
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

func (g *mockdb) Upload() *Upload {
	//todo
	return nil
}

func (g *mockdb) GetMails() []*mail {
	return []*mail{}
}

func (g *mockdb) StartTransaction(name string) Transaction {
	//todo return mock transaction
	return &mocktransaction{name: name, db: g}
}

func (g *mockdb) GetLastCommitTime() (time.Time, error) {
	return time.Now(), nil
}

func (g *mockdb) SetUser(user *User) error {
	g.config.User = user
	return nil
}

func (g *mockdb) Migrate(from Model, to Model) error {

	migrate := []Model{}
	records, err := g.Fetch(from.GetSchema().dataset)
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := record.Hydrate(to); err != nil {
			return err
		}

		migrate = append(migrate, to)
	}

	if err := g.InsertMany(migrate); err != nil {
		return err
	}

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

func (g *mockdb) Sync() error {
	return nil
}
