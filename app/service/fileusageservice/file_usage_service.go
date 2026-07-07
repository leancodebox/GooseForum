package fileusageservice

import (
	"log/slog"
	"net/url"
	"path"
	"strings"

	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/fileUsage"
)

type Usage struct {
	FileName  string
	UsageType string
}

func ReplaceTopic(topicID uint64, userID uint64, content string) {
	replace(fileUsage.TargetTopic, topicID, []string{fileUsage.UsageInlineImage}, userID, namesToUsages(markdown2html.ExtractImageURLs(content), fileUsage.UsageInlineImage))
}

func ReplacePost(postID uint64, userID uint64, content string) {
	replace(fileUsage.TargetPost, postID, []string{fileUsage.UsageInlineImage}, userID, namesToUsages(markdown2html.ExtractImageURLs(content), fileUsage.UsageInlineImage))
}

func ReplaceAvatar(userId uint64, fileNames []string) {
	replace(fileUsage.TargetUser, userId, []string{fileUsage.UsageAvatar}, userId, namesToUsages(fileNames, fileUsage.UsageAvatar))
}

func AddAdminUpload(userId uint64, fileName string) {
	name := fileNameFromURL(fileName)
	if name == "" {
		return
	}
	if err := fileUsage.Create(&fileUsage.Entity{
		FileName:   name,
		TargetType: fileUsage.TargetAdminUpload,
		TargetId:   userId,
		UsageType:  fileUsage.UsageAdminUpload,
		UserId:     userId,
	}); err != nil {
		slog.Error("create admin upload usage failed", "userId", userId, "fileName", name, "err", err)
	}
}

func namesToUsages(values []string, usageType string) []Usage {
	usages := make([]Usage, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		name := fileNameFromURL(value)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		usages = append(usages, Usage{FileName: name, UsageType: usageType})
	}
	return usages
}

func replace(targetType string, targetId uint64, usageTypes []string, userId uint64, usages []Usage) {
	rows := make([]fileUsage.Entity, 0, len(usages))
	for _, usage := range usages {
		if usage.FileName == "" || usage.UsageType == "" {
			continue
		}
		rows = append(rows, fileUsage.Entity{
			FileName:   usage.FileName,
			TargetType: targetType,
			TargetId:   targetId,
			UsageType:  usage.UsageType,
			UserId:     userId,
		})
	}
	if err := fileUsage.ReplaceTargetUsages(targetType, targetId, usageTypes, rows); err != nil {
		slog.Error("replace file usages failed", "targetType", targetType, "targetId", targetId, "err", err)
	}
}

func fileNameFromURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if !parsed.IsAbs() && !strings.HasPrefix(parsed.Path, "/") {
		name := path.Clean(parsed.Path)
		if name == "." || strings.HasPrefix(name, "..") {
			return ""
		}
		return name
	}
	if parsed.IsAbs() && parsed.Host != "" && !strings.HasPrefix(parsed.Path, "/file/img/") {
		return ""
	}
	if !strings.HasPrefix(parsed.Path, "/file/img/") {
		return ""
	}
	name := strings.TrimPrefix(parsed.Path, "/file/img/")
	name = path.Clean("/" + name)
	return strings.TrimPrefix(name, "/")
}
