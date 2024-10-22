package domain

type (
	Provider struct {
		UserID       int64
		PersonalInfo ProviderPersonalInfo
	} // Provider domain entity
	ProviderPersonalInfo struct {
		PersonalIDNumber  []byte
		PersonalIDPreview string
	} // Provider personal info value object
)
