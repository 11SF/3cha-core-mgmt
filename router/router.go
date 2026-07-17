package router

import (
	"net/http"
	"time"

	"portal/backend/app/agent"
	"portal/backend/app/cipher"
	"portal/backend/app/holiday"
	holidayaccess "portal/backend/app/holiday/access"
	"portal/backend/app/member"
	memberaccess "portal/backend/app/member/access"
	"portal/backend/app/queue"
	queueaccess "portal/backend/app/queue/access"
	"portal/backend/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(cfg config.Config, db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	if config.IsLocalEnv() {
		r.Use(gin.Logger())
	}

	corsConfig := cors.Config{
		AllowMethods:  []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}
	if cfg.CORS.AllowOrigin == "*" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = []string{cfg.CORS.AllowOrigin}
	}
	r.Use(cors.New(corsConfig))

	r.GET("/liveness", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.GET("/readiness", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.Status(http.StatusServiceUnavailable)
			return
		}
		c.Status(http.StatusOK)
	})

	{
		api := r.Group("/api/v1")
		registerQueueRoutes(cfg, api, db)
		registerMemberRoutes(api, db)
		registerHolidayRoutes(api, db)
		registerCipherRoutes(api)
	}
	registerAgentRoutes(r)

	return r
}

func registerQueueRoutes(cfg config.Config, api *gin.RouterGroup, db *gorm.DB) {
	queueHandler := queue.NewHandler(queue.HandlerConfig{
		Cfg:            cfg,
		MemberStorage:  queueaccess.NewMemberStorage(db),
		QueueStorage:   queueaccess.NewQueueStorage(db),
		HolidayStorage: queueaccess.NewHolidayStorage(db),
	})

	queues := api.Group("/queue")
	{
		queues.GET("/today", queueHandler.GetToday)
		queues.GET("", queueHandler.List)
		queues.POST("/reset", queueHandler.Reset)
		queues.PATCH("/:id/done", queueHandler.MarkDone)
		queues.PATCH("/:id/skip", queueHandler.Skip)
	}
}

func registerMemberRoutes(api *gin.RouterGroup, db *gorm.DB) {
	memberStorage := memberaccess.NewMemberStorage(db)
	memberHandler := member.NewHandler(member.HandlerConfig{
		MemberStorage: memberStorage,
	})

	members := api.Group("/members")
	{
		members.GET("", memberHandler.List)
		members.POST("", memberHandler.Create)
		members.PATCH("/reorder", memberHandler.Reorder)
		members.DELETE("/:id", memberHandler.Delete)
	}
}

func registerHolidayRoutes(api *gin.RouterGroup, db *gorm.DB) {
	holidayStorage := holidayaccess.NewHolidayStorage(db)
	holidayHandler := holiday.NewHandler(holiday.HandlerConfig{
		HolidayStorage: holidayStorage,
	})

	holidays := api.Group("/holidays")
	{
		holidays.GET("", holidayHandler.List)
		holidays.POST("", holidayHandler.Create)
		holidays.DELETE("/:id", holidayHandler.Delete)
	}
}

func registerCipherRoutes(api *gin.RouterGroup) {
	cipherHandler := cipher.NewHandler(cipher.HandlerConfig{})

	ciphers := api.Group("/cipher")
	{
		ciphers.POST("/encrypt", cipherHandler.Encrypt)
		ciphers.POST("/decrypt", cipherHandler.Decrypt)
	}
}

func registerAgentRoutes(r *gin.Engine) {
	agentHandler := agent.NewHandler(agent.HandlerConfig{})

	agentGroup := r.Group("/api/agent")
	{
		agentGroup.GET("/skill-md", agentHandler.SkillMD)
	}
}
