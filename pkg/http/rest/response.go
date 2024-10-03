package rest

type ApiResponse[T any] struct {
	Data T `json:"data"`
}

func NewApiResponse[T any](data T) *ApiResponse[T] {
	return &ApiResponse[T]{Data: data}
}
