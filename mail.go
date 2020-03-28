package gitdb

import (
	"time"
)

type mail struct {
	Subject string
	Body    string
	Date    time.Time
}

func newMail(subject string, body string) *mail {
	return &mail{Subject: subject, Body: body, Date: time.Now()}
}

func (g *gitdb) GetMails() []*mail {
	mails := g.mails
	g.mails = []*mail{}
	return mails
}

func (g *gitdb) sendMail(m *mail) {
	g.mails = append(g.mails, m)

}
