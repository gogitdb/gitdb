package gitdb

//User represents the user currently connected to the database
//and will be used to identify who made changes to it
type User struct {
	Name  string
	Email string
}

//AuthorName return commit author git style
func (u *User) AuthorName() string {
	return u.Name + " <" + u.Email + ">"
}

//String is an alias for AuthorName
func (u *User) String() string {
	return u.AuthorName()
}

//NewUser constructs a *DbUser
func NewUser(name string, email string) *User {
	return &User{Name: name, Email: email}
}

//SetUser sets the user connection
func (g *gitdb) SetUser(user *User) error {
	g.config.User = user
	return nil
}
