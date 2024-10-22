package domain

type (
	Service struct {
		ID   int32
		Info ServiceInfo
	} // Service Domain Entity
	ServiceInfo struct {
		SubcategoryID int32
		Name          string
		Description   string
	} // Service info Value Object
)

func NewService(id int32, subcategoryID int32, name string, description string) Service {
	info := NewServiceInfo(subcategoryID, name, description)
	return Service{
		ID:   id,
		Info: info,
	}
}

func NewServiceInfo(subcategoryID int32, name string, description string) ServiceInfo {
	return ServiceInfo{
		SubcategoryID: subcategoryID,
		Name:          name,
		Description:   description,
	}
}
