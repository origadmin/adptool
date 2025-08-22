package sourcepkg2

// User represents a user in the system
type User struct {
	ID   int
	Name string
	Age  int
}

// UserService provides user-related operations
type UserService struct{}

// GetUserByID returns a user by ID
func (s *UserService) GetUserByID(id int) *User {
	return &User{
		ID:   id,
		Name: "John Doe",
		Age:  30,
	}
}

// GetUsers returns a list of users
func (s *UserService) GetUsers() []*User {
	return []*User{
		{ID: 1, Name: "John Doe", Age: 30},
		{ID: 2, Name: "Jane Smith", Age: 25},
	}
}