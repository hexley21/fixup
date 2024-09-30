package entity

import "github.com/jackc/pgx/v5/pgtype"

type Service struct {
	ID            int32      `json:"id"`
	SubcategoryID int32      `json:"subcategory_id"`
	Name          string     `json:"name"`
	Description   pgtype.Text `json:"description"`
}
