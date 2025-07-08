package urlconfig

import "path/filepath"

func GetDefaultAvatar() string {
	return `/static/pic/default-avatar.webp`
}

func FilePath(filename string) string {
	return filepath.Join("/file/img", filename)
}
