package attachmentservice

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/leancodebox/GooseForum/app/models/forum/attachments"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/filestorage"
	"github.com/leancodebox/GooseForum/app/service/trustlevelservice"
)

const ObjectPrefix = "attachment-"
const HardMaxSize = 50 * 1024 * 1024

// Attachments deliberately use the database store, independent of public image S3 settings.
func store() *filestorage.Service {
	s, err := filestorage.New(filestorage.NewDatabaseStore())
	if err != nil {
		panic(err)
	}
	return s
}

func fail(code string) error {
	status := 400
	if code == "permission.denied" {
		status = 403
	}
	return &trustlevelservice.PolicyError{Status: status, Code: code}
}

func ValidateConfig(c pageConfig.AttachmentSettings) error {
	if !slices.Contains([]int{0, 100, 200, 1000}, c.MinTrustLevel) || c.MaxFileSizeKb < 1 || c.MaxFileSizeKb > HardMaxSize/1024 || c.MaxFilesPerPost < 1 || c.MaxFilesPerPost > 20 {
		return fail("common.request.invalidParams")
	}
	seen := map[string]bool{}
	for _, ext := range c.Extensions {
		if !slices.Contains([]string{".pdf", ".txt", ".zip"}, ext) || seen[ext] {
			return fail("common.request.invalidParams")
		}
		seen[ext] = true
	}
	return nil
}

func Policy(userID uint64) (pageConfig.AttachmentSettings, error) {
	c := hotdataserve.GetPostingSettingsConfigCache().Attachments
	if err := ValidateConfig(c); err != nil {
		return c, err
	}
	p, err := trustlevelservice.ResolvePolicy(userID)
	if err != nil {
		return c, err
	}
	if !c.Enabled || int(p.Account.TrustLevel) < c.MinTrustLevel {
		return c, fail("permission.denied")
	}
	return c, nil
}

func ValidateFile(name string, data []byte, c pageConfig.AttachmentSettings) error {
	if !utf8.ValidString(name) || len(name) == 0 || len(name) > 255 || strings.ContainsAny(name, "/\\") || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return fail("attachment.invalidFile")
	}
	if len(data) == 0 || len(data) > c.MaxFileSizeKb*1024 {
		return fail("attachment.tooLarge")
	}
	ext := strings.ToLower(filepath.Ext(name))
	if !slices.Contains(c.Extensions, ext) {
		return fail("attachment.invalidFile")
	}
	valid := false
	switch ext {
	case ".pdf":
		valid = bytes.HasPrefix(data, []byte("%PDF-"))
	case ".zip":
		valid = bytes.HasPrefix(data, []byte("PK\x03\x04")) || bytes.HasPrefix(data, []byte("PK\x05\x06"))
	case ".txt":
		valid = utf8.Valid(data) && !bytes.ContainsRune(data, 0)
	}
	if !valid {
		return fail("attachment.invalidFile")
	}
	return nil
}

func Upload(ctx context.Context, userID uint64, name string, body io.Reader) (attachments.Info, error) {
	c, err := Policy(userID)
	if err != nil {
		return attachments.Info{}, err
	}
	p, err := trustlevelservice.ResolvePolicy(userID)
	if err != nil {
		return attachments.Info{}, err
	}
	if err = p.ValidateAttachmentUpload(1); err != nil {
		return attachments.Info{}, err
	}
	data, err := io.ReadAll(io.LimitReader(body, int64(c.MaxFileSizeKb)*1024+1))
	if err != nil {
		return attachments.Info{}, err
	}
	if err = ValidateFile(name, data, c); err != nil {
		return attachments.Info{}, err
	}
	objectName := ObjectPrefix + uuid.NewString()
	_, err = store().Put(ctx, filestorage.PutRequest{Name: objectName, ContentType: "application/octet-stream", Size: int64(len(data)), UserId: userID, Body: bytes.NewReader(data)})
	if err != nil {
		return attachments.Info{}, err
	}
	e := attachments.Entity{UserID: userID, ObjectName: objectName, Filename: name, Size: int64(len(data))}
	if err = attachments.Create(&e); err != nil {
		if cleanupErr := store().Delete(context.WithoutCancel(ctx), objectName); cleanupErr != nil {
			slog.Error("attachment upload cleanup failed", "name", objectName, "error", cleanupErr)
		}
		return attachments.Info{}, err
	}
	return info(e), nil
}

func info(e attachments.Entity) attachments.Info {
	return attachments.Info{ID: e.ID, Filename: e.Filename, Size: e.Size, URL: fmt.Sprintf("/file/attachments/%d", e.ID)}
}

// Nil preserves legacy clients' existing attachments; an empty list explicitly removes them.
func Prepare(userID, postID uint64, current []attachments.Info, ids *[]uint64) ([]attachments.Info, error) {
	if ids == nil {
		return current, nil
	}
	if len(*ids) > 20 {
		return nil, fail("attachment.limit")
	}
	rows, err := attachments.GetMany(*ids)
	if err != nil {
		return nil, err
	}
	byID := map[uint64]attachments.Entity{}
	for _, r := range rows {
		byID[r.ID] = r
	}
	seen := map[uint64]bool{}
	result := make([]attachments.Info, 0, len(*ids))
	added := false
	var policy *pageConfig.AttachmentSettings
	for _, id := range *ids {
		r, ok := byID[id]
		if !ok || seen[id] || r.UserID != userID || (r.PostID != 0 && r.PostID != postID) {
			return nil, fail("attachment.invalidReference")
		}
		seen[id] = true
		old := slices.ContainsFunc(current, func(f attachments.Info) bool { return f.ID == id })
		if !old {
			added = true
			if policy == nil {
				c, err := Policy(userID)
				if err != nil {
					return nil, err
				}
				policy = &c
			}
			if r.Size > int64(policy.MaxFileSizeKb)*1024 || !slices.Contains(policy.Extensions, strings.ToLower(filepath.Ext(r.Filename))) {
				return nil, fail("attachment.invalidFile")
			}
		}
		result = append(result, info(r))
	}
	if added && len(result) > policy.MaxFilesPerPost {
		return nil, fail("attachment.limit")
	}
	return result, nil
}

func CanRead(userID uint64, e attachments.Entity) bool {
	if e.PostID == 0 {
		return userID != 0 && userID == e.UserID
	}
	p := posts.Get(e.PostID)
	if p.Id == 0 || !slices.ContainsFunc(p.Attachments, func(f attachments.Info) bool { return f.ID == e.ID }) {
		return false
	}
	t := topics.GetSimple(p.TopicId)
	if t.Id == 0 {
		return false
	}
	actor, err := accesscontrol.Resolve(userID)
	if err != nil {
		return false
	}
	if actor.CanManageAnyCategory(t.CategoryIds) {
		return true
	}
	if !actor.CanReadCategory(t.MainCategoryId) {
		return false
	}
	if t.Status != 1 {
		return userID != 0 && t.UserId == userID && p.UserId == userID
	}
	// A manual topic block denies its author too, just like the topic page.
	if t.ProcessStatus != 0 {
		return false
	}
	return p.ProcessStatus == 0 || (userID != 0 && p.UserId == userID)
}

func Open(ctx context.Context, userID, id uint64) (attachments.Entity, *filestorage.Object, error) {
	e, err := attachments.Get(id)
	if err != nil || !CanRead(userID, e) {
		return e, nil, attachments.ErrInvalid
	}
	obj, err := store().Open(ctx, e.ObjectName)
	if err == nil && obj.Metadata.UserId != e.UserID {
		_ = obj.Body.Close()
		return e, nil, attachments.ErrInvalid
	}
	return e, obj, err
}

func Cancel(ctx context.Context, userID, id uint64) error {
	e, err := attachments.Get(id)
	if err != nil {
		return err
	}
	removed, err := attachments.DeleteUnbound(id, userID)
	if err != nil {
		return err
	}
	if !removed {
		return attachments.ErrInvalid
	}
	if err := store().Delete(ctx, e.ObjectName); err != nil {
		slog.Error("attachment object deletion failed", "name", e.ObjectName, "error", err)
		return err
	}
	return nil
}

func Cleanup(ctx context.Context) error {
	rows, err := attachments.UnboundBefore(time.Now().Add(-14 * 24 * time.Hour))
	if err != nil {
		return err
	}
	for _, e := range rows {
		if err := Cancel(ctx, e.UserID, e.ID); err != nil {
			slog.Warn("attachment cleanup failed", "id", e.ID, "error", err)
		}
	}
	return nil
}
