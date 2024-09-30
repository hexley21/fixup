package entity

type Category struct {
	ID     int32  `json:"id"`
	TypeID int32  `json:"type_id"`
	Name   string `json:"name"`
}
