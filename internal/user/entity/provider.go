package entity

type Provider struct {
	PersonalIDNumber  []byte `json:"personal_id_number"`
	PersonalIDPreview string `json:"personal_id_preview"`
	UserID            int64  `json:"user_id"`
}
