package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/CPU-commits/Intranet_BFiles/controllers"
	"github.com/CPU-commits/Intranet_BFiles/middlewares"
	"github.com/CPU-commits/Intranet_BFiles/models"
	"github.com/CPU-commits/Intranet_BFiles/res"
	"github.com/CPU-commits/Intranet_BFiles/services"
	"github.com/CPU-commits/Intranet_BFiles/settings"
	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/natefinch/lumberjack.v2"

	// swaggerFiles "github.com/swaggo/files"     // swagger embed files
	// ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func ErrorHandler(c *gin.Context, info ratelimit.Info) {
	c.JSON(http.StatusTooManyRequests, &res.Response{
		Success: false,
		Message: "Too many requests. Try again in" + time.Until(info.ResetTime).String(),
	})
}

var settingsData = settings.GetSettings()

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found")
	}
}

func Init() {
	router := gin.New()
	// Proxies
	router.SetTrustedProxies([]string{"localhost"})
	// Zap logger
	// Create folder if not exists
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err := os.Mkdir("logs", os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	// Log file
	logEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	fileCore := zapcore.NewCore(logEncoder, zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
	}), zap.InfoLevel)
	// Log console
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.InfoLevel)
	// Combine cores for multi-output logging
	teeCore := zapcore.NewTee(fileCore, consoleCore)
	zapLogger := zap.New(teeCore)

	router.Use(ginzap.GinzapWithConfig(zapLogger, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  []string{"/api/annoucements/swagger"},
	}))
	router.Use(ginzap.RecoveryWithZap(zapLogger, true))

	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Server Internal Error: %s", err))
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, res.Response{
			Success: false,
			Message: "Server Internal Error",
		})
	}))

	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Server Internal Error: %s", err))
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, res.Response{
			Success: false,
			Message: "Server Internal Error",
		})
	}))
	// Docs
	/*docs.SwaggerInfo.BasePath = "/api/c/classroom"
	docs.SwaggerInfo.Version = "v1"
	docs.SwaggerInfo.Host = "localhost:8080"*/
	// CORS
	if settingsData.NODE_ENV == "prod" {
		httpOrigin := "http://" + settingsData.CLIENT_URL
		httpsOrigin := "https://" + settingsData.CLIENT_URL

		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{httpOrigin, httpsOrigin},
			AllowMethods:     []string{"GET", "OPTIONS", "PUT", "DELETE", "POST"},
			AllowCredentials: true,
			AllowWebSockets:  false,
			AllowHeaders:     []string{"*"},
			MaxAge:           12 * time.Hour,
		}))
	} else {
		router.Use(cors.New(cors.Config{
			AllowOrigins: []string{"*"},
		}))
	}
	// Secure
	sslUrl := "ssl." + settingsData.CLIENT_URL
	secureConfig := secure.Config{
		SSLHost:              sslUrl,
		STSSeconds:           315360000,
		STSIncludeSubdomains: true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
		IENoOpen:             true,
		ReferrerPolicy:       "strict-origin-when-cross-origin",
		SSLProxyHeaders: map[string]string{
			"X-Fowarded-Proto": "https",
		},
	}
	/*if settingsData.NODE_ENV == "prod" {
		secureConfig.AllowedHosts = []string{
			settingsData.CLIENT_URL,
			sslUrl,
		}
	}*/
	router.Use(secure.New(secureConfig))
	// Init nats subscribers
	services.InitFilesNats()
	// Rate limit
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: 7,
	})
	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: ErrorHandler,
		KeyFunc:      keyFunc,
	})
	router.Use(mw)
	// Routes
	files := router.Group(
		"/api/files",
		middlewares.JWTMiddleware(),
		middlewares.MaxSizePerFile(
			MAX_FILE_SIZE,
			MAX_FILE_SIZE_STR,
			MAX_FILES,
			"file",
		),
	)
	{
		// Init controllers
		filesController := new(controllers.FilesController)
		// Define routes
		files.GET(
			"/get_files",
			middlewares.RolesMiddleware([]string{
				models.DIRECTOR,
				models.DIRECTIVE,
				models.TEACHER,
			}),
			filesController.GetFiles,
		)
		files.GET(
			"/get_file/:idFile",
			filesController.GetFile,
		)
		files.POST(
			"/upload_file",
			middlewares.RolesMiddleware([]string{
				models.DIRECTOR,
				models.DIRECTIVE,
				models.TEACHER,
			}),
			filesController.UploadFile,
		)
		files.PUT(
			"/change_permissions/:idFile",
			middlewares.RolesMiddleware([]string{
				models.DIRECTOR,
				models.DIRECTIVE,
				models.TEACHER,
			}),
			filesController.ChangePermissions,
		)
		files.DELETE(
			"/delete_file/:idFile",
			middlewares.RolesMiddleware([]string{
				models.DIRECTOR,
				models.DIRECTIVE,
				models.TEACHER,
			}),
			filesController.DeleteFile,
		)
	}
	// Route docs
	// router.GET("/api/news/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// Route healthz
	router.GET("/api/files/healthz", func(ctx *gin.Context) {
		ctx.JSON(200, &res.Response{
			Success: true,
		})
	})
	// No route
	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, res.Response{
			Success: false,
			Message: "Not found",
		})
	})
	// Init server
	if err := router.Run(); err != nil {
		log.Fatalf("Error init server")
	}
}
