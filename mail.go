package gitdb

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type mail struct {
	Subject string
	Body    string
	Date    time.Time
}

type Mail struct {
	privateMail *mail
	db          *gitdb
}

// Implement json.Unmarshaller
func (m *Mail) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &m.privateMail)
}

func (m *Mail) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.privateMail)
}

func newMail(db *gitdb, subject string, body string) *Mail {
	return &Mail{
		privateMail: &mail{Subject: subject, Body: body, Date: time.Now()},
		db:          db,
	}
}

func (m *Mail) GetSubject() string {
	return m.privateMail.Subject
}

func (m *Mail) GetBody() string {
	return m.privateMail.Body
}

func (m *Mail) GetDate() time.Time {
	return m.privateMail.Date
}

func (m *Mail) send() error {

	if _, err := os.Stat(m.db.mailDir()); err != nil {
		err := os.MkdirAll(m.db.mailDir(), 0744)
		if err != nil {
			return err
		}
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(m.db.mailDir(), time.Now().Format("20060102150405")+".json"), bytes, 0744)
	if err != nil {
		log("Could not send notification - " + err.Error())
	}

	return err
}

func (c *Connection) GetMails() []*mail {
	g := c.db()
	var mails []*mail
	files, err := ioutil.ReadDir(g.mailDir())
	if err != nil {
		logError(err.Error())
		return mails
	}

	var fileName string
	for _, file := range files {
		fileName = filepath.Join(g.mailDir(), file.Name())
		if filepath.Ext(fileName) == ".json" {
			data, err := ioutil.ReadFile(fileName)
			if err != nil {
				logError(err.Error())
				continue
			}

			var mail *Mail
			fmtErr := json.Unmarshal(data, &mail)
			if fmtErr != nil {
				logError(err.Error())
				continue
			}

			mails = append(mails, mail.privateMail)
		}
	}

	return mails
}
