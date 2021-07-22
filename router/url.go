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
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func getUserID() gin.HandlerFunc {
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
			c.Next()
		}
	}
}

// SetupRouter index
func SetupRouter() *gin.Engine {
	if os.Getenv("GIN_MOD") != "release" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Error loading .env file")
		}
	}
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:            "NCNUOJ",
		SigningAlgorithm: "HS512",
		Key:              []byte(os.Getenv("SECRET_KEY")),
		MaxRefresh:       time.Hour,
		Authenticator:    views.Login,
		TimeFunc:         time.Now,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.User); ok {
				return jwt.MapClaims{
					"id":       strconv.FormatUint(uint64(v.ID), 10),
					"username": v.UserName,
					"admin":    v.Admin,
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
	r.POST(baseURL+"/user", views.UserRegister)
	r.POST(baseURL+"/token", authMiddleware.LoginHandler)
	auth := r.Group(baseURL + "/token")
	auth.Use(authMiddleware.MiddlewareFunc())
	auth.GET("", authMiddleware.RefreshHandler)
	user := r.Group(baseURL + "/user")
	user.Use(authMiddleware.MiddlewareFunc())
	user.Use(getUserID())
	{
		user.GET("", views.UserInfo)
		user.PATCH("", views.UserChangeInfo)
	}
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Page not found",
		})
	})
	return r
}
