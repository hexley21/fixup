package domain

type (
	Category struct {
		ID   int32
		Info CategoryInfo
	}

	CategoryInfo struct {
		TypeID int32
		Name   string
	}
)
