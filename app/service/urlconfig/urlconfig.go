package urlconfig

import "path/filepath"

func GetDefaultAvatar() string {
	return `/api/assets/default-avatar.png`
}

func FilePath(filename string) string {
	return filepath.Join("/file/img", filename)
}
