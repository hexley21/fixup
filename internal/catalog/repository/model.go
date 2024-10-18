package repository

import "github.com/jackc/pgx/v5/pgtype"

type CategoryModel struct {
	ID     int32
	TypeID int32
	Name   string
}

type CategoryTypeModel struct {
	ID   int32
	Name string
}

type ProviderServiceModel struct {
	ProviderID int64
	ServiceID  int32
}

type ServiceModel struct {
	ID            int32
	SubcategoryID int32
	Name          string
	Description   pgtype.Text
}

type SubcategoryModel struct {
	ID         int32
	CategoryID int32
	Name       string
}
