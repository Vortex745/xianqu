package app

import (
	"gotest/config"
	"gotest/core/controllers"
	"gotest/core/middleware"
	"gotest/core/services"
	"gotest/pkg/ws"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var engine *gin.Engine

func SetupEngine(frontendFS fs.FS) *gin.Engine {
	if engine != nil {
		return engine
	}

	// 1. Setup paths
	workDir, _ := os.Getwd()

	// 2. Init DB
	config.InitDB()

	// 3. Create uploads dir
	uploadDir := filepath.Join(workDir, "uploads")
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	// 4. Init WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// 启动订单超时定时任务
	services.StartOrderSweeper()

	// 5. Init Gin
	r := gin.Default()
	r.MaxMultipartMemory = 100 << 20
	r.Use(CORSMiddleware())
	r.StaticFS("/uploads", gin.Dir(uploadDir, true))

	// 6. Frontend Hosting (Optional)
	if frontendFS != nil {
		indexHtmlBytes, _ := fs.ReadFile(frontendFS, "index.html")
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api") {
				c.JSON(404, gin.H{"error": "API not found: " + path})
				return
			}
			file, err := frontendFS.Open(strings.TrimPrefix(path, "/"))
			if err == nil {
				defer file.Close()
				stat, _ := file.Stat()
				if !stat.IsDir() {
					c.FileFromFS(path, http.FS(frontendFS))
					return
				}
			}
			c.Data(200, "text/html; charset=utf-8", indexHtmlBytes)
		})
	}

	// 7. Initialize Controllers
	chatController := &controllers.ChatController{Hub: hub}
	userController := new(controllers.UserController)
	productController := new(controllers.ProductController)
	fileController := new(controllers.FileController)
	orderController := new(controllers.OrderController)
	cartController := new(controllers.CartController)
	adminController := new(controllers.AdminController)
	authController := new(controllers.AuthController)
	aiModelController := new(controllers.AIModelController)

	// 8. API Routes
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "message": "Backend is alive!"})
		})
		api.POST("/register", userController.Register)
		api.POST("/login", userController.Login)
		api.POST("/auth/send-code", authController.SendCode)
		api.POST("/auth/verify-login", authController.VerifyLogin)
		api.GET("/ws", chatController.Connect)
		api.GET("/products", productController.List)
		api.GET("/products/:id", productController.GetDetail)
		api.GET("/categories", productController.Categories)

		userGroup := api.Group("/", middleware.Auth())
		{
			userGroup.GET("/user/data", userController.GetMyData)
			userGroup.PUT("/user/profile", userController.UpdateProfile)
			userGroup.PUT("/user/password", userController.ChangePassword)
			userGroup.GET("/users/:id", userController.GetUserInfo)
			userGroup.GET("/favorites/check", userController.CheckFavorite)
			userGroup.POST("/favorites/add", userController.ToggleFavorite)
			userGroup.POST("/favorites/remove", userController.ToggleFavorite)
			userGroup.POST("/cart/add", cartController.Add)
			userGroup.GET("/cart", cartController.List)
			userGroup.DELETE("/cart/:id", cartController.Delete)
			chatGroup := userGroup.Group("/chat")
			{
				chatGroup.GET("/contacts", chatController.GetContacts)
				chatGroup.GET("/messages", chatController.GetHistory)
			}
			userGroup.POST("/upload", fileController.Upload)
			userGroup.POST("/products", productController.Create)
			userGroup.PUT("/products/:id", productController.Update)
			userGroup.POST("/orders", orderController.Create)
			userGroup.POST("/orders/batch", orderController.BatchCreate)
			userGroup.GET("/orders", orderController.List)
			userGroup.POST("/orders/:id/pay", orderController.Pay)
			userGroup.POST("/orders/:id/confirm_pay", orderController.ConfirmPay)
			userGroup.PUT("/orders/:id/confirm", orderController.ConfirmReceive)
			userGroup.PUT("/orders/:id/refund", orderController.Refund)
		}

		api.GET("/ai-models/active", aiModelController.ActiveList)
		api.GET("/ai-models/secret/:id", aiModelController.GetModelSecret)
		api.POST("/ai-models/usage", aiModelController.ReportUsage)

		adminGroup := api.Group("/admin")
		{
			adminGroup.POST("/login", adminController.Login)
			adminGroup.GET("/init", adminController.Init)
			authGroup := adminGroup.Group("/", middleware.AdminAuth())
			{
				authGroup.GET("/info", adminController.GetInfo)
				authGroup.GET("/stats", adminController.GetStats)
				authGroup.GET("/users", adminController.GetUsers)
				authGroup.PUT("/users/:id/status", adminController.UpdateUserStatus)
				authGroup.GET("/products", adminController.GetProducts)
				authGroup.PUT("/products/:id/audit", adminController.AuditProduct)
				authGroup.GET("/orders", adminController.GetOrders)
				aiGroup := authGroup.Group("/ai-models")
				{
					aiGroup.GET("", aiModelController.List)
					aiGroup.POST("", aiModelController.Create)
					aiGroup.GET("/:id", aiModelController.GetOne)
					aiGroup.PUT("/:id", aiModelController.Update)
					aiGroup.DELETE("/:id", aiModelController.Delete)
					aiGroup.PUT("/:id/status", aiModelController.ToggleStatus)
					aiGroup.GET("/dashboard", aiModelController.Dashboard)
					aiGroup.POST("/detect", aiModelController.DetectModels)
				}
			}
		}
	}
	engine = r
	return engine
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
