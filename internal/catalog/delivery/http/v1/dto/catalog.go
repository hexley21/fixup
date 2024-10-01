package dto

type CreateCategoryTypeDTO struct {
	Name string `json:"name" validate:"alpha,min=2,max=30,required"`
} // @name CreateCategoryTypeInput

type PatchCategoryTypeDTO struct {
	Name string `json:"name" validate:"alpha,min=2,max=30,required"`
} // @name UpdateCategoryTypeInput

type CategoryTypeDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
} // @name CategoryType
