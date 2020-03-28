package gitdb

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/distatus/battery"
)

//RandStr returns a random string of length n
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

	m.SetID(m.GetSchema().RecordID())

	metadata := &metaData{}
	metadata.Indexes = m.GetSchema().Indexes()
	metadata.Encrypted = m.ShouldEncrypt()

	m.SetMetaData(metadata)

}

//ParseID parses a GitDB record id and returns it's metadata
func ParseID(id string) (dataDir string, block string, record string, err error) {
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

func hasSufficientBatteryPower(threshold float64) bool {
	batt, err := battery.Get(0)
	if err != nil {
		return false
	}

	percentageCharge := batt.Current / batt.Full * 100

	log(fmt.Sprintf("Battery Level: %6.2f%%", percentageCharge))

	//return true if battery life is above threshold
	return percentageCharge >= threshold
}
