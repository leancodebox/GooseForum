package pageConfig

import "encoding/json"

// DecodeSecuritySettings preserves defaults for fields absent in older settings.
func DecodeSecuritySettings(raw string, defaults SecurityAndRegistration) SecurityAndRegistration {
	if err := json.Unmarshal([]byte(raw), &defaults); err != nil {
		return defaults
	}
	return defaults
}

func GetSecuritySettings(defaults SecurityAndRegistration) SecurityAndRegistration {
	entity := GetByPageType(SecuritySettings)
	if entity.Id == 0 {
		return defaults
	}
	return DecodeSecuritySettings(entity.Config, defaults)
}
