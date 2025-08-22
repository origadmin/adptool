package sourcepkg3

// User represents a user in the system
type User struct {
	ID   int
	Name string
	Age  int
}

// Product represents a product in the system
type Product struct {
	ID    int
	Name  string
	Price float64
}

// CommonService provides common operations
type CommonService struct{}

// GetUserByID returns a user by ID
func (s *CommonService) GetUserByID(id int) *User {
	return &User{
		ID:   id,
		Name: "John Doe",
		Age:  30,
	}
}

// GetProductByID returns a product by ID
func (s *CommonService) GetProductByID(id int) *Product {
	return &Product{
		ID:    id,
		Name:  "Sample Product",
		Price: 99.99,
	}
}