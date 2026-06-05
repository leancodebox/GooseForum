// Package setting exposes process-level configuration helpers.
package setting

const storage = "./storage/"
const tmp = "./storage/tmp/"

// GetStorage returns the default storage directory.
func GetStorage() string {
	return storage
}

// GetTmp returns the default temporary storage directory.
func GetTmp() string {
	return tmp
}
