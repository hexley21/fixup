package rest

type response[T any] struct {
	Data T `json:"data"`
}

func NewApiResponse[T any](data T) *response[T] {
	return &response[T]{Data: data}
}
