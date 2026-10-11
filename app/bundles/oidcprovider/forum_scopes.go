package oidcprovider

import "slices"

const (
	ScopeForumRead    = "forum:read"
	ScopeTopicsCreate = "topics:create"
	ScopePostsCreate  = "posts:create"
	ScopeTopicsEdit   = "topics:edit"
	ScopePostsEdit    = "posts:edit"
	ScopeImagesUpload = "images:upload"
)

// Forum scopes bind an opaque access token to the GooseForum resource API.
func HasForumScope(scopes []string) bool {
	return slices.Contains(scopes, ScopeForumRead) || slices.Contains(scopes, ScopeTopicsCreate) || slices.Contains(scopes, ScopePostsCreate) || slices.Contains(scopes, ScopeImagesUpload) || slices.Contains(scopes, ScopeTopicsEdit) || slices.Contains(scopes, ScopePostsEdit)
}

func ValidForumScopes(scopes []string) bool {
	return !HasForumScope(scopes) || slices.Contains(scopes, ScopeForumRead)
}
