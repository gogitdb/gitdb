package gitdb

import (
	"os"
	"path/filepath"
	"time"
	"io/ioutil"
	"encoding/json"
)

type Mail struct {
	subject string
	body    string
	date    time.Time
}

func newMail(subject string, body string) *Mail {
	return &Mail{subject:subject, body:body, date:time.Now()}
}

func (m *Mail) GetSubject() string {
	return m.subject
}

func (m *Mail) GetBody() string {
	return m.body
}

func (m *Mail) GetDate() time.Time {
	return m.date
}

func (m *Mail) send() error {

	if _, err := os.Stat(mailDir()); err != nil {
		os.MkdirAll(mailDir(), 0744)
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(mailDir(), time.Now().Format("20060102150405")+".json"), bytes, 0744)
	if err != nil {
		log("Could not send notification - " + err.Error())
	}

	return err
}

func GetMails() []*Mail {

	var mails []*Mail
	files, err := ioutil.ReadDir(mailDir())
	if err != nil {
		logError(err.Error())
		return mails
	}

	var fileName string
	for _, file := range files {
		fileName = filepath.Join(mailDir(), file.Name())
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

			mails = append(mails, mail)
		}
	}

	return mails
}