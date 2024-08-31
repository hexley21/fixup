package rest

type apiResponse[T any] struct {
	Data T `json:"data"`
}

func NewApiResponse[T any](data T) *apiResponse[T] {
	return &apiResponse[T]{Data: data}
}
