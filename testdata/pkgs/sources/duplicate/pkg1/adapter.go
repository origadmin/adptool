package pkg1

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

// Product represents a product in the system
type Product struct {
	ID    int
	Name  string
	Price float64
}

// ProductService provides product-related operations
type ProductService struct{}

// GetProductByID returns a product by ID
func (s *ProductService) GetProductByID(id int) *Product {
	return &Product{
		ID:    id,
		Name:  "Sample Product",
		Price: 99.99,
	}
}