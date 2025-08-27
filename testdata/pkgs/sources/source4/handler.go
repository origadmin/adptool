package source_pkg4

// Handler handles requests
type Handler struct {
	Service *Service
}

// HandleModel handles model requests
func (h *Handler) HandleModel(id int) *Model {
	return h.Service.GetModel(id)
}

// HandleData handles data requests
func (h *Handler) HandleData(id int) *Data {
	return h.Service.GetData(id)
}

// GenericHandler handles generic requests
type GenericHandler[T any] struct {
	Data T
}

// Handle handles generic data
func (h *GenericHandler[T]) Handle() T {
	return h.Data
}