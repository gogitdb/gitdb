package gitdb

import (
	"encoding/json"
	"fmt"
)

//record represents a model stored in gitdb
type record struct {
	id    string
	data  string
	index map[string]interface{}
	key   string
}

func newRecord(id, data string) *record {
	return &record{id: id, data: data}
}

//Hydrate assumes record has already been decrypted because
//it is usually called after a Fetch or Search which means
//block.grecords would have been called, which inturn
//decrypts all records it returns
func (r *record) Hydrate(model Model) error {
	return r.hydrate(model)
}

func (r *record) hydrateUsingKey(model Model, key string) error {
	if len(key) > 0 {
		r.decrypt(key)
	}

	return r.hydrate(model)
}

func (r *record) decrypt(key string) {
	logTest("decrypting with: " + key)
	dec := decrypt(key, r.data)
	if len(dec) > 0 {
		r.data = dec
	}

}

func (r *record) version() string {
	//TODO can we optimize this?
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(r.data), &v); err != nil {
		return "v1"
	}

	ver, ok := v["Version"]
	if !ok {
		return "v1"
	}

	return ver.(string)
}

func (r *record) hydrate(model interface{}) error {
	return r.hydrateByVersion(model, r.version())
}

func (r *record) hydrateByVersion(model interface{}, version string) error {
	switch version {
	case "v1":
		if err := json.Unmarshal([]byte(r.data), model); err != nil {
			return err
		}
		return nil
	case "v2": //TODO Optimize Unmarshall-Marshall technique
		var rawV2 struct {
			Indexes map[string]interface{}
			Data    map[string]interface{}
		}

		if err := json.Unmarshal([]byte(r.data), &rawV2); err != nil {
			return err
		}

		b, err := json.Marshal(rawV2.Data)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, model); err != nil {
			return err
		}

		r.index = rawV2.Indexes
	default:
		return fmt.Errorf("Unable to hydrate version : %s", version)
	}

	return nil
}

//indexes takes a factory for read-only backward compatibility with earlier versions of GitDB
func (r *record) indexes(dataset string, factory func(name string) Model) map[string]interface{} {

	if r.version() == "v1" && factory != nil {
		//todo cache this model?
		model := factory(dataset)
		r.hydrateByVersion(model, "v1")
	} else {
		var m map[string]interface{}
		r.hydrate(&m)
	}

	return r.index
}

type collection []*record

func (c collection) Len() int {
	return len(c)
}
func (c collection) Less(i, j int) bool {
	return c[i].id < c[j].id
}
func (c collection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
