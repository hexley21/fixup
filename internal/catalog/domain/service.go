package domain

type (
	Service struct {
		ID   int32
		Info ServiceInfo
	}

	ServiceInfo struct {
		SubcategoryID int32
		Name          string
		Description   string
	}
)
