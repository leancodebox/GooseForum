package agent

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/oidcproviderservice"
)

//go:embed SKILL.md
var skillDocument []byte

//go:embed API.md
var apiDocument []byte

func (h *Handler) apiReference(c *gin.Context) {
	serveMarkdown(c, apiDocument)
}

func (h *Handler) skill(c *gin.Context) {
	serveMarkdown(c, skillDocument)
}

func serveMarkdown(c *gin.Context, document []byte) {
	etag := fmt.Sprintf(`"%x"`, sha256.Sum256(document))
	c.Header("ETag", etag)
	c.Header("Cache-Control", "public, max-age=300")
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}
	c.Data(200, "text/markdown; charset=utf-8", document)
}

func (h *Handler) site(c *gin.Context) {
	config := hotdataserve.GetPostingSettingsConfigCache().TextControl
	status := oidcproviderservice.Status()
	success(c, 200, gin.H{
		"name":       hotdataserve.GetSiteSettingsConfigCache().SiteName,
		"apiVersion": "v1", "enabled": enabled(), "contentFormat": "markdown",
		"capabilities": []string{"read", "search", "createTopic", "createPost", "submissions"},
		"limits": gin.H{
			"lengthUnit": "utf8_bytes", "minTitleLength": config.MinTitleLength, "maxTitleLength": config.MaxTitleLength,
			"minContentLength": config.MinPostLength, "maxContentLength": config.MaxPostLength,
			"maxCategories": 3, "maxPageSize": 50, "maxRequestBytes": 1 << 20,
			"maxDailyTopicsPerUser": config.MaxDailyTopicsPerUser, "newUserPostCooldownMinutes": config.NewUserPostCooldownMinutes,
		},
		"auth": gin.H{
			"browserAuthorizationAvailable": status.Available, "issuer": status.Issuer,
			"discoveryUrl":          baseURL() + "/oauth2/.well-known/openid-configuration",
			"authorizationEndpoint": baseURL() + "/oauth2/authorize", "tokenEndpoint": baseURL() + "/oauth2/token",
			"revocationEndpoint":     baseURL() + "/oauth2/revoke",
			"registration":           "Ask a site administrator to register an OIDC client with forum:read and S256 PKCE",
			"manualTokensEnabled":    hotdataserve.GetAgentSettingsConfigCache().ManualTokens,
			"manualTokenSettingsUrl": baseURL() + "/settings?tab=agent-tokens",
			"scopes":                 []string{"openid", "forum:read", "topics:create", "posts:create", "offline_access"},
		},
	})
}

func schemaRef(name string) gin.H { return gin.H{"$ref": "#/components/schemas/" + name} }
func stringSchema() gin.H         { return gin.H{"type": "string"} }
func objectSchema(properties gin.H, required ...string) gin.H {
	return gin.H{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}
func arraySchema(items any) gin.H { return gin.H{"type": "array", "items": items} }

func (h *Handler) openapi(c *gin.Context) {
	c.Header("Cache-Control", "no-cache")
	data, err := json.Marshal(openAPIDocument(baseURL()))
	if err != nil {
		fail(c, 500, "internal_error", "Unable to build API contract")
		return
	}
	etag := fmt.Sprintf(`"%x"`, sha256.Sum256(data))
	c.Header("ETag", etag)
	if c.GetHeader("If-None-Match") == etag {
		c.Status(304)
		return
	}
	c.Data(200, "application/json; charset=utf-8", data)
}

func openAPIDocument(base string) gin.H {
	id := schemaRef("ID")
	requestID := gin.H{"type": "string", "format": "uuid"}
	schemas := gin.H{
		"ID":          gin.H{"type": "string", "pattern": "^[1-9][0-9]*$", "description": "Positive uint64 decimal string"},
		"CreateTopic": objectSchema(gin.H{"title": stringSchema(), "content": stringSchema(), "categoryIds": gin.H{"type": "array", "items": id, "minItems": 1, "maxItems": 3, "uniqueItems": true}, "clientRequestId": requestID}, "title", "content", "categoryIds", "clientRequestId"),
		"CreatePost":  objectSchema(gin.H{"content": stringSchema(), "replyToPostId": id, "clientRequestId": requestID}, "content", "clientRequestId"),
		"Submission":  objectSchema(gin.H{"submissionId": id, "clientRequestId": requestID, "topicId": id, "postId": id, "postNo": gin.H{"type": "integer"}, "url": stringSchema(), "visible": gin.H{"type": "boolean"}, "moderationStatus": gin.H{"type": "string", "enum": []string{"none", "pending", "approved", "rejected", "denied"}}, "reused": gin.H{"type": "boolean"}}, "submissionId", "clientRequestId", "topicId", "postId", "visible", "moderationStatus", "reused", "postNo", "url"),
		"Error":       objectSchema(gin.H{"error": objectSchema(gin.H{"code": stringSchema(), "message": stringSchema(), "details": gin.H{"type": "object", "additionalProperties": true}}, "code", "message", "details"), "requestId": requestID}, "error", "requestId"),
	}
	author := objectSchema(gin.H{"id": id}, "id")
	stamp := gin.H{"type": "string", "format": "date-time"}
	schemas["Post"] = objectSchema(gin.H{"id": id, "topicId": id, "author": author, "postNo": gin.H{"type": "integer"}, "replyToPostId": gin.H{"type": "string", "nullable": true}, "content": stringSchema(), "contentFormat": gin.H{"type": "string", "enum": []string{"markdown"}}, "sourceVersion": gin.H{"type": "integer"}, "createdAt": stamp, "updatedAt": stamp, "url": stringSchema()}, "id", "topicId", "author", "postNo", "content", "contentFormat", "sourceVersion", "replyToPostId", "createdAt", "updatedAt", "url")
	schemas["Topic"] = objectSchema(gin.H{"id": id, "title": stringSchema(), "author": author, "categoryIds": arraySchema(id), "createdAt": stamp, "updatedAt": stamp, "url": stringSchema(), "capabilities": objectSchema(gin.H{"reply": gin.H{"type": "boolean"}}, "reply"), "firstPost": schemaRef("Post")}, "id", "title", "author", "categoryIds", "createdAt", "updatedAt", "url", "capabilities")
	schemas["Category"] = objectSchema(gin.H{"id": id, "name": stringSchema(), "description": stringSchema(), "slug": stringSchema(), "capabilities": objectSchema(gin.H{"createTopic": gin.H{"type": "boolean"}, "reply": gin.H{"type": "boolean"}}, "createTopic", "reply")}, "id", "name", "description", "slug", "capabilities")
	schemas["Me"] = objectSchema(gin.H{"id": id, "username": stringSchema(), "clientId": stringSchema(), "scopes": arraySchema(stringSchema()), "canWrite": gin.H{"type": "boolean"}, "restriction": stringSchema()}, "id", "username", "clientId", "scopes", "canWrite", "restriction")
	schemas["SubmissionSummary"] = objectSchema(gin.H{"submissionId": id, "topicId": id, "clientRequestId": requestID}, "submissionId", "topicId", "clientRequestId")
	schemas["TopicDetail"] = gin.H{"allOf": []any{schemaRef("Topic"), gin.H{"type": "object", "required": []string{"firstPost"}}}}
	schemas["CursorPagination"] = objectSchema(gin.H{"hasMore": gin.H{"type": "boolean"}, "nextCursor": stringSchema()}, "hasMore", "nextCursor")
	schemas["SearchPagination"] = objectSchema(gin.H{"hasMore": gin.H{"type": "boolean"}, "page": gin.H{"type": "integer", "minimum": 1}}, "hasMore", "page")
	security := func(scope string) []gin.H {
		return []gin.H{{"oauth": []string{"openid", scope}}, {"manualToken": []string{}}}
	}
	param := func(name, in string, schema any, required bool) gin.H {
		return gin.H{"name": name, "in": in, "required": required, "schema": schema}
	}
	paths := gin.H{}
	type operation struct {
		path, method, name, scope, response, body string
		anonymous, list                           bool
		params                                    []gin.H
	}
	listParams := []gin.H{param("limit", "query", gin.H{"type": "integer", "minimum": 1, "maximum": 50, "default": 20}, false), param("cursor", "query", stringSchema(), false)}
	ops := []operation{
		{path: "/site", method: "get", name: "getSite", anonymous: true, response: "Site"},
		{path: "/me", method: "get", name: "getMe", scope: "forum:read", response: "Me"},
		{path: "/categories", method: "get", name: "listCategories", scope: "forum:read", anonymous: true, list: true, response: "Category"},
		{path: "/topics", method: "get", name: "listTopics", scope: "forum:read", anonymous: true, list: true, response: "Topic", params: append(append([]gin.H{}, listParams...), param("categoryId", "query", id, false), param("sort", "query", gin.H{"type": "string", "enum": []string{"newest"}}, false))},
		{path: "/topics/{topicId}", method: "get", name: "getTopic", scope: "forum:read", anonymous: true, response: "TopicDetail"},
		{path: "/topics/{topicId}/posts", method: "get", name: "listPosts", scope: "forum:read", anonymous: true, list: true, response: "Post", params: listParams},
		{path: "/posts/{postId}", method: "get", name: "getPost", scope: "forum:read", anonymous: true, response: "Post"},
		{path: "/search", method: "get", name: "searchTopics", scope: "forum:read", anonymous: true, list: true, response: "Topic", params: []gin.H{param("q", "query", gin.H{"type": "string", "maxLength": 500}, true), param("page", "query", gin.H{"type": "integer", "minimum": 1, "maximum": 1000, "default": 1}, false), listParams[0]}},
		{path: "/topics", method: "post", name: "createTopic", scope: "topics:create", response: "Submission", body: "CreateTopic"},
		{path: "/topics/{topicId}/posts", method: "post", name: "createPost", scope: "posts:create", response: "Submission", body: "CreatePost"},
		{path: "/me/submissions", method: "get", name: "listSubmissions", scope: "forum:read", list: true, response: "SubmissionSummary", params: append(append([]gin.H{}, listParams...), param("clientRequestId", "query", requestID, false))},
		{path: "/me/submissions/{submissionId}", method: "get", name: "getSubmission", scope: "forum:read", response: "Submission"},
	}
	limitFields := gin.H{"lengthUnit": gin.H{"type": "string", "enum": []string{"utf8_bytes"}}}
	for _, key := range []string{"minTitleLength", "maxTitleLength", "minContentLength", "maxContentLength", "maxCategories", "maxPageSize", "maxRequestBytes", "maxDailyTopicsPerUser", "newUserPostCooldownMinutes"} {
		limitFields[key] = gin.H{"type": "integer", "minimum": 0}
	}
	authFields := gin.H{"browserAuthorizationAvailable": gin.H{"type": "boolean"}, "manualTokensEnabled": gin.H{"type": "boolean"}, "scopes": arraySchema(stringSchema())}
	for _, key := range []string{"issuer", "discoveryUrl", "authorizationEndpoint", "tokenEndpoint", "revocationEndpoint", "registration", "manualTokenSettingsUrl"} {
		authFields[key] = stringSchema()
	}
	schemas["Site"] = objectSchema(gin.H{"name": stringSchema(), "apiVersion": stringSchema(), "enabled": gin.H{"type": "boolean"}, "contentFormat": stringSchema(), "capabilities": arraySchema(stringSchema()), "auth": objectSchema(authFields, "browserAuthorizationAvailable", "manualTokensEnabled", "issuer", "discoveryUrl", "authorizationEndpoint", "tokenEndpoint", "revocationEndpoint", "registration", "manualTokenSettingsUrl", "scopes"), "limits": objectSchema(limitFields, "lengthUnit", "minTitleLength", "maxTitleLength", "minContentLength", "maxContentLength", "maxCategories", "maxPageSize", "maxRequestBytes", "maxDailyTopicsPerUser", "newUserPostCooldownMinutes")}, "name", "apiVersion", "enabled", "contentFormat", "capabilities", "auth", "limits")
	errorDescriptions := map[string]string{"400": "Invalid request", "401": "Authentication required or invalid token", "403": "Insufficient scope or permission", "404": "Resource unavailable", "409": "clientRequestId conflict", "410": "Submission deleted", "413": "Request exceeds 1 MiB", "422": "Invalid content or categories", "429": "Rate limit, quota or cooldown", "500": "Internal error", "503": "Service unavailable"}
	sharedResponses := gin.H{}
	for status, description := range errorDescriptions {
		response := gin.H{"description": description, "content": gin.H{"application/json": gin.H{"schema": schemaRef("Error")}}}
		if status == "429" {
			response["headers"] = gin.H{"Retry-After": gin.H{"schema": stringSchema(), "description": "Seconds to wait, when known"}}
		}
		sharedResponses["Error"+status] = response
	}
	operationErrors := map[string][]string{
		"listCategories": {"500"}, "listTopics": {"400", "500"}, "getTopic": {"400", "404", "500"},
		"listPosts": {"400", "404", "500"}, "getPost": {"400", "404", "500"}, "searchTopics": {"400", "500"},
		"createTopic": {"400", "404", "409", "410", "413", "422", "500"}, "createPost": {"400", "404", "409", "410", "413", "422", "500"},
		"listSubmissions": {"400", "404", "410", "500"}, "getSubmission": {"400", "404", "410", "500"},
	}
	for _, op := range ops {
		data := any(schemaRef(op.response))
		if op.list {
			data = arraySchema(data)
		}
		props := gin.H{"data": data, "requestId": requestID}
		if len(op.params) > 0 && op.list {
			pagination := "CursorPagination"
			if op.name == "searchTopics" {
				pagination = "SearchPagination"
			}
			props["pagination"] = schemaRef(pagination)
		}
		required := []string{"data", "requestId"}
		if props["pagination"] != nil {
			required = append(required, "pagination")
		}
		var result any = objectSchema(props, required...)
		if op.name == "listSubmissions" {
			result = gin.H{"oneOf": []any{result, objectSchema(gin.H{"data": schemaRef("Submission"), "requestId": requestID}, "data", "requestId")}}
		}
		response := gin.H{"description": "Success", "content": gin.H{"application/json": gin.H{"schema": result}}}
		responses := gin.H{"200": response}
		if op.body != "" {
			responses["201"] = response
		}
		if op.name != "getSite" {
			for _, status := range append([]string{"401", "403", "429", "503"}, operationErrors[op.name]...) {
				responses[status] = gin.H{"$ref": "#/components/responses/Error" + status}
			}
		}
		sec := security(op.scope)
		if op.scope == "" {
			sec = nil
		}
		if op.anonymous {
			sec = append(sec, gin.H{})
		}
		params := append([]gin.H{}, op.params...)
		for _, name := range []string{"topicId", "postId", "submissionId"} {
			if containsPathParam(op.path, name) {
				params = append(params, param(name, "path", id, true))
			}
		}
		entry := gin.H{"operationId": op.name, "security": sec, "parameters": params, "responses": responses}
		if op.body != "" {
			entry["requestBody"] = gin.H{"required": true, "content": gin.H{"application/json": gin.H{"schema": schemaRef(op.body)}}}
		}
		if paths[op.path] == nil {
			paths[op.path] = gin.H{}
		}
		paths[op.path].(gin.H)[op.method] = entry
	}
	return gin.H{"openapi": "3.0.3", "info": gin.H{"title": "GooseForum Agent API", "version": "1.0.0", "description": "Quick reference: API.md. OAuth uses S256 PKCE. Operation scopes also apply to manual tokens; ID tokens and cookies are not API credentials."}, "externalDocs": gin.H{"url": base + "/api/agent/v1/API.md"}, "servers": []gin.H{{"url": base + "/api/agent/v1"}}, "paths": paths, "components": gin.H{"schemas": schemas, "responses": sharedResponses, "securitySchemes": gin.H{"oauth": gin.H{"type": "oauth2", "flows": gin.H{"authorizationCode": gin.H{"authorizationUrl": base + "/oauth2/authorize", "tokenUrl": base + "/oauth2/token", "scopes": gin.H{"openid": "Confirm identity", "forum:read": "Read accessible forum content", "topics:create": "Publish topics", "posts:create": "Reply to topics", "offline_access": "Refresh access after expiry"}}}}, "manualToken": gin.H{"type": "http", "scheme": "bearer"}}}}
}

func containsPathParam(path, name string) bool {
	for i := 0; i+len(name)+2 <= len(path); i++ {
		if path[i:i+len(name)+2] == "{"+name+"}" {
			return true
		}
	}
	return false
}
