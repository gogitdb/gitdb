package db

type User struct {
	Name  string
	Email string
}

func (u *User) AuthorName() string {
	return u.Name + " <" + u.Email + ">"
}

func (u *User) String() string{
	return u.AuthorName()
}

func NewUser(name string, email string) *User {
	return &User{Name: name, Email: email}
}


