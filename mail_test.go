package gitdb

import (
	"testing"
	"time"
)

func TestMailGetSubject(t *testing.T) {
	subject := "Hello"
	body := "world"
	mail := newMail(nil, subject, body)
	if mail.GetSubject() != subject {
		t.Errorf("mail subject should be %s", subject)
	}
}

func TestMailGetBody(t *testing.T) {
	subject := "Hello"
	body := "world"
	mail := newMail(nil, subject, body)
	if mail.GetBody() != body {
		t.Errorf("mail body should be %s", body)
	}
}

func TestMailGetDate(t *testing.T) {
	mail := newMail(nil, "Hello", "world")
	y, m, d := mail.GetDate().Date()
	ty, tm, td := time.Now().Date()
	if y != ty || m != tm || d != td {
		t.Errorf("mail.GetDate() should be %d, %d %d", ty, tm, td)
	}
}

func TestMailGetMails(t *testing.T) {
	dbConn := getDbConn()
	defer dbConn.Close()
	mail := newMail(dbConn.db(), "Hello", "world")
	if err := mail.send(); err != nil {
		t.Errorf("mail.send() returned error - %s", err)
	}

	mails := dbConn.GetMails()
	if len(mails) <= 0 {
		t.Errorf("dbConn.GetMails() should return at least 1 mail")
	}
}
