package sourcepkg

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

// GetProducts returns a list of products
func (s *ProductService) GetProducts() []*Product {
	return []*Product{
		{ID: 1, Name: "Product 1", Price: 19.99},
		{ID: 2, Name: "Product 2", Price: 29.99},
	}
}