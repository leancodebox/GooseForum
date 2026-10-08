package oidcprovider

import "slices"

const (
	ScopeForumRead    = "forum:read"
	ScopeTopicsCreate = "topics:create"
	ScopePostsCreate  = "posts:create"
)

// Forum scopes bind an opaque access token to the GooseForum resource API.
func HasForumScope(scopes []string) bool {
	return slices.Contains(scopes, ScopeForumRead) || slices.Contains(scopes, ScopeTopicsCreate) || slices.Contains(scopes, ScopePostsCreate)
}

func ValidForumScopes(scopes []string) bool {
	return !HasForumScope(scopes) || slices.Contains(scopes, ScopeForumRead)
}
