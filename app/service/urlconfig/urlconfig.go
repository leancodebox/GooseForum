package urlconfig

import "path/filepath"

func GetDefaultAvatar() string {
	return `/static/pic/default-avatar.png`
}

func FilePath(filename string) string {
	return filepath.Join("/file/img", filename)
}
