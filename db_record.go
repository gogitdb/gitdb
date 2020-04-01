package gitdb

import (
	"encoding/json"
	"fmt"
)

//record represents a model stored in gitdb
type record struct {
	//used by UI
	Content string

	id    string
	data  string
	index map[string]interface{}
	key   string
}

func newRecord(id, data string) *record {
	return &record{id: id, data: data}
}

func (r *record) Hydrate(model Model) error {
	//check if decryption is required
	if model.ShouldEncrypt() {
		r = r.decrypt(r.key)
	}

	return r.hydrate(model)
}

func (r *record) gHydrate(model Model, key string) error {
	if model.ShouldEncrypt() && len(key) > 0 {
		r = r.decrypt(key)
	}

	return r.hydrate(model)
}

func (r *record) decrypt(key string) *record {
	r2 := *r
	logTest("decrypting with: " + key)
	dec := decrypt(key, r2.data)
	if len(dec) > 0 {
		r2.data = dec
	}
	return &r2
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

func (r *record) indexes(key string) map[string]interface{} {
	var m map[string]interface{}
	r.hydrate(&m)
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
