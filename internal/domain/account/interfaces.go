package account

type Repository interface {
	CreateUser(user *User) error
	GetUser(userID int) (*User, error)
}
