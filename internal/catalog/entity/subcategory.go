package entity

type Subcategory struct {
	ID         int32  `json:"id"`
	CategoryID int32  `json:"category_id"`
	Name       string `json:"name"`
}
