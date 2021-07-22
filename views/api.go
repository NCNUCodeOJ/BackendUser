package views

import (
	"errors"
	"net/http"

	"github.com/NCNUCodeOJ/BackendUser/models"
	"github.com/NCNUCodeOJ/BackendUser/pkg"
	"github.com/gin-gonic/gin"
	"github.com/vincentinttsh/replace"
	"github.com/vincentinttsh/zero"
	"gorm.io/gorm"
)

// UserRegister 註冊
func UserRegister(c *gin.Context) {
	var user models.User
	var data struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		RealName string `json:"realname"`
		UserName string `json:"username"`
	}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "未按照格式填寫或未使用json",
		})
		return
	}
	if zero.IsZero(data) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "未填寫完成",
		})
		return
	}
	u, err := models.UserDetailByUserName(data.UserName)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "系統錯誤",
		})
		return
	}
	if u.ID != 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "此 username 已被使用",
		})
		return
	}
	pwd, err := pkg.Encrypt(data.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "系統錯誤",
		})
		return
	}
	user.RealName = data.RealName
	user.UserName = data.UserName
	user.Email = data.Email
	user.Password = pwd
	if err := models.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "註冊失敗",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "註冊成功",
	})
}

// UserInfo 使用者自己的資訊
func UserInfo(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	user, err := models.UserDetailByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "無此使用者",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"username": user.UserName,
		"realname": user.RealName,
		"email":    user.Email,
		"admin":    user.Admin,
	})
}

// Pong test server is operating
func Pong(c *gin.Context) {
	if models.Ping() != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "server error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// Login login
func Login(c *gin.Context) (interface{}, error) {
	var d struct {
		Name     *string `json:"username"`
		Password *string `json:"password"`
	}
	if err := c.BindJSON(&d); err != nil {
		return nil, errors.New("未按照格式填寫或未使用json")
	}
	if zero.IsZero(d) {
		return nil, errors.New("未填寫完成")
	}
	u, err := models.UserDetailByUserName(*d.Name)
	if err != nil {
		return nil, errors.New("帳號或密碼錯誤")
	}
	if pkg.Compare(u.Password, *d.Password) == nil {
		return &u, nil
	}
	return nil, errors.New("帳號或密碼錯誤")
}

// UserChangeInfo 使用者更改自己的資訊
func UserChangeInfo(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	user, err := models.UserDetailByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "無此使用者",
		})
		return
	}
	var data struct {
		RealName *string `json:"realname"`
		UserName *string `json:"username"`
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "未按照格式填寫",
		})
		return
	}
	if !zero.IsZero(data.Password) {
		pwd, err := pkg.Encrypt(*data.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "系統錯誤",
			})
			return
		}
		data.Password = &pwd
	}
	replace.Replace(&user, &data)
	if err := models.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "更改失敗",
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "更改成功",
	})
}
