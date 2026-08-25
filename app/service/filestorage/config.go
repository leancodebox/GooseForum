package filestorage

import (
	"fmt"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func ConfigureFromPreferences() error {
	driver := strings.ToLower(strings.TrimSpace(preferences.GetString("storage.driver", DatabaseDriver)))
	switch driver {
	case "", DatabaseDriver:
		return Configure(NewDatabaseStore())
	case S3Driver:
		store, err := NewS3Store(S3Config{
			Endpoint:     preferences.GetString("storage.s3.endpoint"),
			Bucket:       preferences.GetString("storage.s3.bucket"),
			AccessKey:    preferences.GetString("storage.s3.accessKey"),
			SecretKey:    preferences.GetString("storage.s3.secretKey"),
			SessionToken: preferences.GetString("storage.s3.sessionToken"),
			Region:       preferences.GetString("storage.s3.region"),
			Secure:       preferences.GetBool("storage.s3.secure", true),
			PathStyle:    preferences.GetBool("storage.s3.pathStyle", false),
		})
		if err != nil {
			return err
		}
		return Configure(store, NewDatabaseStore())
	default:
		return fmt.Errorf("unsupported file storage driver %q", driver)
	}
}
