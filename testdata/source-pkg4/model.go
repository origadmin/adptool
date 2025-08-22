package source_pkg4

// Model represents a generic model structure
type Model struct {
	ID   int
	Name string
}

// Data represents data structure
type Data struct {
	Model
	Value string
}

// Service provides operations for models
type Service struct {
	Name string
}

// GetModel returns a model by ID
func (s *Service) GetModel(id int) *Model {
	return &Model{
		ID:   id,
		Name: "Test Model",
	}
}

// GetData returns data by ID
func (s *Service) GetData(id int) *Data {
	return &Data{
		Model: Model{
			ID:   id,
			Name: "Test Data",
		},
		Value: "Sample Value",
	}
}