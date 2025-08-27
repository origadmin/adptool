package pkg3

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