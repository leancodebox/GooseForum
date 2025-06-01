package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
)

func PostV2(c *gin.Context) {
	viewrender.Render(c, "list.gohtml", map[string]any{
		"title": "newgooseforum",
	})
}
