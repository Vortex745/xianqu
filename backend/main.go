package main

import (
	"embed"
	"fmt"
	"gotest/config"
	"gotest/internal/controllers"
	"gotest/internal/middleware"
	"gotest/internal/services" // 新增引入
	"gotest/pkg/ws"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS Middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
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

//go:embed dist/*
var content embed.FS

func main() {
	// 1. Anti-crash
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Critical Error:", err)
		}
	}()

	// 2. Setup paths
	workDir, _ := os.Getwd()
	fmt.Println(">>> Current working directory:", workDir)

	// 3. Init DB
	config.InitDB()

	// 4. Create uploads dir
	uploadDir := filepath.Join(workDir, "uploads")
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	// 5. Init WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// ★★★ 新增：启动订单超时定时任务 ★★★
	services.StartOrderSweeper()

	// 6. Init Gin
	r := gin.Default()
	// ★★★ 核心修复：提升文件上传大小限制到 100MB ★★★
	r.MaxMultipartMemory = 100 << 20
	r.Use(CORSMiddleware())
	r.StaticFS("/uploads", gin.Dir(uploadDir, true))

	// 7. Frontend Hosting
	frontendFS, err := fs.Sub(content, "dist")
	if err == nil {
		indexHtmlBytes, _ := fs.ReadFile(frontendFS, "index.html")
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// 如果是 /api 开头但没匹配到路由，返回 JSON 错误
			if strings.HasPrefix(path, "/api") {
				c.JSON(404, gin.H{"error": "API not found: " + path}) // 返回具体路径方便调试
				return
			}
			// 否则尝试返回前端静态文件
			file, err := frontendFS.Open(strings.TrimPrefix(path, "/"))
			if err == nil {
				defer file.Close()
				stat, _ := file.Stat()
				if !stat.IsDir() {
					c.FileFromFS(path, http.FS(frontendFS))
					return
				}
			}
			// SPA 回退到 index.html
			c.Data(200, "text/html; charset=utf-8", indexHtmlBytes)
		})
	}

	// 9. Initialize Controllers
	chatController := &controllers.ChatController{Hub: hub}
	userController := new(controllers.UserController)
	productController := new(controllers.ProductController)
	fileController := new(controllers.FileController)
	orderController := new(controllers.OrderController)
	cartController := new(controllers.CartController)
	adminController := new(controllers.AdminController)
	authController := new(controllers.AuthController)       // ★ 邮箱验证码登录
	aiModelController := new(controllers.AIModelController) // ★ AI模型管理

	// 10. API Routes
	api := r.Group("/api")
	{
		api.POST("/register", userController.Register)
		api.POST("/login", userController.Login)

		// ★ 邮箱验证码登录
		api.POST("/auth/send-code", authController.SendCode)
		api.POST("/auth/verify-login", authController.VerifyLogin)

		// WebSocket Endpoint
		api.GET("/ws", chatController.Connect)

		api.GET("/products", productController.List)
		api.GET("/products/:id", productController.GetDetail)
		api.GET("/categories", productController.Categories)

		// Protected Routes (需要登录)
		userGroup := api.Group("/", middleware.Auth())
		{
			userGroup.GET("/user/data", userController.GetMyData)
			userGroup.PUT("/user/profile", userController.UpdateProfile)
			userGroup.PUT("/user/password", userController.ChangePassword)
			userGroup.GET("/users/:id", userController.GetUserInfo) // 聊天用

			// ★★★ 核心修复：对齐前端 ProductDetail.vue 的接口 ★★★

			// 1. 收藏相关 (Favorites)
			// 前端: /api/favorites/check
			userGroup.GET("/favorites/check", userController.CheckFavorite)
			// 前端: /api/favorites/add (复用 Toggle 逻辑)
			userGroup.POST("/favorites/add", userController.ToggleFavorite)
			// 前端: /api/favorites/remove (复用 Toggle 逻辑)
			userGroup.POST("/favorites/remove", userController.ToggleFavorite)

			// 2. 购物车相关 (Cart)
			// 前端: /api/cart/add
			userGroup.POST("/cart/add", cartController.Add)
			// 前端: /api/cart (获取列表)
			userGroup.GET("/cart", cartController.List)
			userGroup.DELETE("/cart/:id", cartController.Delete)

			// 3. 聊天相关
			chatGroup := userGroup.Group("/chat")
			{
				chatGroup.GET("/contacts", chatController.GetContacts)
				chatGroup.GET("/messages", chatController.GetHistory)
			}

			// 4. 其他业务
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

		// ★ 前台公共接口：获取可用AI模型列表
		api.GET("/ai-models/active", aiModelController.ActiveList)
		// ★ 内部接口：AI Service 获取解密模型信息
		api.GET("/ai-models/secret/:id", aiModelController.GetModelSecret)
		// ★ 用量上报接口（AI Service 调用）
		api.POST("/ai-models/usage", aiModelController.ReportUsage)

		// Admin Routes
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

				// ★★★ AI模型管理路由 ★★★
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

	fmt.Println(">>> Server started successfully: http://localhost:8081")
	r.Run(":8081")
}
