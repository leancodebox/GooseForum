package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/agent"
)

func agentRoute(engine *gin.Engine) { agent.New().Register(engine) }
