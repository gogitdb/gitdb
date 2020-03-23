package gitdb

import (
	"errors"
	"math/rand"
	"strings"
	"time"
)

func RandStr(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func stamp(m Model) {
	stampTime := time.Now()
	if m.GetCreatedDate().IsZero() {
		m.SetCreatedDate(stampTime)
	}
	m.SetUpdatedDate(stampTime)

	m.SetId(m.GetSchema().RecordId())
}

func ParseId(id string) (dataDir string, block string, record string, err error) {
	recordMeta := strings.Split(id, "/")
	if len(recordMeta) != 3 {
		err = errors.New("Invalid record id: " + id)
	} else {
		dataDir = recordMeta[0]
		block = recordMeta[1]
		record = recordMeta[2]
	}

	return dataDir, block, record, err
}
