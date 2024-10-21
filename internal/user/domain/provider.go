package domain

type (
	Provider struct {
		UserID       int64
		PersonalInfo ProviderPersonalInfo
	}
	ProviderPersonalInfo struct {
		PersonalIDNumber  []byte
		PersonalIDPreview string
	}
)
