package db

type DbUser struct {
	Name  string
	Email string
}

func (u *DbUser) AuthorName() string {
	return u.Name + " <" + u.Email + ">"
}

func (u *DbUser) String() string {
	return u.AuthorName()
}

func NewUser(name string, email string) *DbUser {
	return &DbUser{Name: name, Email: email}
}
