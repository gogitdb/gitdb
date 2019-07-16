package gitdb

import (
	"math/rand"
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