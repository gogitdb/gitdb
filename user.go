package gitdb

//DbUser represents the user currently connected to the database
//and will be used to identify who made changes to it
type DbUser struct {
	Name  string
	Email string
}

//AuthorName return commit author git style
func (u *DbUser) AuthorName() string {
	return u.Name + " <" + u.Email + ">"
}

//String is an alias for AuthorName
func (u *DbUser) String() string {
	return u.AuthorName()
}

//NewUser constructs a *DbUser
func NewUser(name string, email string) *DbUser {
	return &DbUser{Name: name, Email: email}
}

//SetUser sets the user connection
func (g *gitdb) SetUser(user *DbUser) error {
	g.config.User = user
	return nil
}
