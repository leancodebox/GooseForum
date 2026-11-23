package pageConfig

import (
	"encoding/json"
	"errors"
)

// decodeSecuritySettings preserves defaults for fields absent in older settings.
func decodeSecuritySettings(raw string, defaults SecurityAndRegistration) (SecurityAndRegistration, error) {
	cfg := &defaults
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return SecurityAndRegistration{}, err
	}
	if cfg == nil {
		return SecurityAndRegistration{}, errors.New("empty security settings")
	}
	return *cfg, nil
}
