package db

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bouggo/log"
	"github.com/gogitdb/gitdb/v2/internal/crypto"
	"github.com/valyala/fastjson"
)

//Record represents a model stored in gitdb
type Record struct {
	id    string
	data  string
	index map[string]interface{}
	key   string

	p         fastjson.Parser
	decrypted bool
}

//newRecord constructs a Record
func newRecord(id, data string) *Record {
	return &Record{id: id, data: data, index: map[string]interface{}{}}
}

//ID returns record id
func (r *Record) ID() string {
	return r.id
}

//Data returns record data unmodified
func (r *Record) Data() string {
	return r.data
}

//Hydrate populates given interfacce with underlying record data
func (r *Record) Hydrate(model interface{}) error {
	r.decrypt(r.key)
	version := r.Version()
	switch version {
	case "v1":
		if err := json.Unmarshal([]byte(r.data), model); err != nil {
			return err
		}
		return nil
	case "v2": //TODO Optimize Unmarshall-Marshall technique
		v, err := r.p.Parse(r.data)
		if err != nil {
			return err
		}

		obj := v.GetObject("Data")
		buf := make([]byte, obj.Len())
		buf = obj.MarshalTo(buf)
		buf = bytes.Trim(buf, "\x00")

		// fmt.Printf("%s\n", oh)
		if err := json.Unmarshal(buf, model); err != nil {
			return err
		}

		obj = v.GetObject("Indexes")
		buf = make([]byte, obj.Len())
		buf = obj.MarshalTo(buf)
		buf = bytes.Trim(buf, "\x00")
		return json.Unmarshal(buf, &r.index)
	default:
		return fmt.Errorf("Unable to hydrate version : %s", version)
	}
}

func (r *Record) decrypt(key string) {
	if len(key) > 0 && !r.decrypted {
		log.Test("decrypting with: " + key)
		dec := crypto.Decrypt(key, r.data)
		if len(dec) > 0 {
			r.data = dec
		}
		r.decrypted = true
	}
}

//Indexes returns v2 indexes for GitDB
func (r *Record) Indexes() map[string]interface{} {
	var m map[string]interface{}
	r.Hydrate(&m)
	return r.index
}

//JSON returns data decrypted and indented
func (r *Record) JSON() string {
	var buf bytes.Buffer
	r.decrypt(r.key)
	if err := json.Indent(&buf, []byte(r.data), "", "\t"); err != nil {
		log.Error(err.Error())
	}

	return buf.String()
}

//Version returns the version of the record
func (r *Record) Version() string {
	v, err := r.p.Parse(r.data)
	if err != nil {
		return "v1"
	}

	versionBytes := v.GetStringBytes("Version")
	version := string(versionBytes)
	if len(version) == 0 {
		version = "v1"
	}

	return version
}

//ConvertModel converts a Model to a record
func ConvertModel(id string, m interface{}) *Record {
	b, _ := json.Marshal(m)
	return newRecord(id, string(b))
}

//collection represents a sortable slice of Records
type collection []*Record

func (c collection) Len() int {
	return len(c)
}
func (c collection) Less(i, j int) bool {
	return c[i].id < c[j].id
}
func (c collection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
