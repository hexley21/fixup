package dto

type CreateCategoryDTO struct {
	Name   string `json:"name" validate:"alpha,min=2,max=30,required"`
	TypeID string `json:"type_id" validate:"number"`
} // @name CreateCategoryInput

type PatchCategoryDTO struct {
	Name   string `json:"name" validate:"alpha,min=2,max=30,required"`
	TypeID string `json:"type_id" validate:"number"`
} // @name CreateCategoryInput

type CategoryDTO struct {
	ID     string `json:"id"`
	TypeID string `json:"type_id"`
	Name   string `json:"name"`
} // @name Category
