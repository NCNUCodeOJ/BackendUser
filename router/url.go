package router

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/NCNUCodeOJ/BackendUser/models"
	"github.com/NCNUCodeOJ/BackendUser/views"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func getUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(jwt.ExtractClaims(c)["id"].(string))
		if err != nil {
			c.Abort()
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "系統錯誤",
				"error":   err.Error(),
			})
		} else {
			c.Set("userID", uint(id))
			c.Set("teacher", jwt.ExtractClaims(c)["teacher"].(bool))
			c.Set("admin", jwt.ExtractClaims(c)["admin"].(bool))
			c.Next()
		}
	}
}

// SetupRouter index
func SetupRouter() *gin.Engine {
	if os.Getenv("GIN_MODE") != "release" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Error loading .env file")
		}
	}
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:            "NCNUOJ",
		SigningAlgorithm: "HS512",
		Key:              []byte(os.Getenv("SECRET_KEY")),
		Timeout:          4 * time.Hour,
		MaxRefresh:       time.Hour,
		Authenticator:    views.Login,
		TimeFunc:         time.Now,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.User); ok {
				return jwt.MapClaims{
					"id":       strconv.FormatUint(uint64(v.ID), 10),
					"username": v.UserName,
					"admin":    v.Admin,
					"teacher":  v.Teacher,
				}
			}
			return jwt.MapClaims{}
		},
	})
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	baseURL := "api/v1"
	r := gin.Default()
	// CORS
	if os.Getenv("FrontendURL") != "" {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{os.Getenv("FrontendURL")},
			AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "DELETE"},
			AllowHeaders:     []string{"Origin, Authorization, Content-Type, Accept"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}
	r.GET("/ping", views.Pong)
	r.POST(baseURL+"/user", views.UserRegister)
	r.POST(baseURL+"/token", authMiddleware.LoginHandler)
	r.POST(baseURL+"/forget_password", views.UserForgetPassword)
	r.POST(baseURL+"/reset_password", views.UserResetPassword)
	auth := r.Group(baseURL + "/token")
	auth.Use(authMiddleware.MiddlewareFunc())
	auth.GET("", authMiddleware.RefreshHandler)
	user := r.Group(baseURL + "/user")
	user.Use(authMiddleware.MiddlewareFunc())
	user.Use(getUserInfo())
	{
		user.GET("", views.UserInfo)
		user.PATCH("", views.UserChangeInfo)
		user.PATCH("/permission", views.ChangeUserPermissions)
	}
	username := r.Group(baseURL + "/username")
	username.Use(authMiddleware.MiddlewareFunc())
	{
		username.POST("", views.GetUserName)
	}
	announcement := r.Group(baseURL + "/announcements")
	announcement.Use(authMiddleware.MiddlewareFunc())
	announcement.Use(getUserInfo())
	{
		announcement.GET("", views.GetAllAnnouncements)
		announcement.POST("", views.CreateAnnouncement)
		announcement.DELETE("/:id", views.DeleteAnnouncement)
	}
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Page not found",
		})
	})
	return r
}
