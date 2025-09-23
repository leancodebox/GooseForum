package urlconfig

import "path"

func GetDefaultAvatar() string {
	return `/static/pic/default-avatar.webp`
}

func FilePath(filename string) string {
	return path.Join("/file/img", filename)
}
