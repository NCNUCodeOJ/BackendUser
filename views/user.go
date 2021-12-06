package views

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/NCNUCodeOJ/BackendUser/models"
	"github.com/NCNUCodeOJ/BackendUser/pkg"
	"github.com/gin-gonic/gin"
	"github.com/meyskens/go-hcaptcha"
	"github.com/vincentinttsh/replace"
	"github.com/vincentinttsh/zero"
	"gorm.io/gorm"
)

var captchaClient *hcaptcha.HCaptcha

func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// UserRegister 註冊
func UserRegister(c *gin.Context) {
	var user models.User
	var err error
	var data struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		RealName  string `json:"realname"`
		StudentID string `json:"student_id"`
		UserName  string `json:"username"`
		Avatar    string `json:"avatar"`
	}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "json format error",
		})
		return
	}

	userAvatar := data.Avatar
	data.Avatar = "default"

	if zero.IsZero(data) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "data is not complete",
		})
		return
	}

	data.Avatar = userAvatar

	if data.Avatar != "" && !isValidURL(data.Avatar) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	isValidUserName := regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString
	if !isValidUserName(data.UserName) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "username can only contain letters and numbers",
		})
		return
	}

	_, err = models.UserDetailByUserName(data.UserName)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "server error",
		})
		return
	}
	if err == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "username is already used",
		})
		return
	}

	if len(data.Password) < 6 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "password is too short, at least 6 characters",
		})
		return
	}

	pwd, err := pkg.Encrypt(data.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "system error",
		})
		return
	}

	user.RealName = data.RealName
	user.UserName = data.UserName
	user.StudentID = data.StudentID
	user.Email = data.Email
	user.Password = pwd
	user.Avatar = data.Avatar

	if err := models.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "register failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "register success",
		"user_id": strconv.FormatUint(uint64(user.ID), 10),
	})
}

// UserInfo 使用者自己的資訊
func UserInfo(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	user, err := models.UserDetailByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "no such user",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":    strconv.FormatUint(uint64(user.ID), 10),
		"username":   user.UserName,
		"realname":   user.RealName,
		"email":      user.Email,
		"student_id": user.StudentID,
		"admin":      user.Admin,
		"teacher":    user.Teacher,
		"avatar":     user.Avatar,
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
		return nil, errors.New("json format error")
	}
	if zero.IsZero(d) {
		return nil, errors.New("data is not complete")
	}

	u, err := models.UserDetailByUserName(*d.Name)
	if err != nil {
		return nil, errors.New("username or password is wrong")
	}

	if pkg.Compare(u.Password, *d.Password) == nil {
		return &u, nil
	}

	return nil, errors.New("username or password is wrong")
}

// UserChangeInfo 使用者更改自己的資訊
func UserChangeInfo(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	user, err := models.UserDetailByID(uint(userID))

	var pwd string

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "no such user",
		})
		return
	}
	var data struct {
		RealName  *string `json:"realname"`
		Email     *string `json:"email"`
		Password  *string `json:"password"`
		StudentID *string `json:"student_id"`
		Avatar    *string `json:"avatar"`
	}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "json format error",
		})
		return
	}

	if needLog {
		log.Println(
			"data: realname:",
			*data.RealName,
			"email:",
			*data.Email,
			"password:",
			*data.Password,
			"student_id:",
			*data.StudentID,
			"avatar:",
			*data.Avatar,
		)
	}

	if len(*data.Password) < 6 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "password is too short, at least 6 characters",
		})
		return
	}

	if !zero.IsZero(data.Password) {
		pwd, err = pkg.Encrypt(*data.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "system error",
			})
			return
		}
		data.Password = &pwd
	}

	replace.Replace(&user, &data)

	if needLog {
		log.Printf("%+v\n", user)
	}

	if err := models.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "update failed",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "update success",
	})
}

// ChangeUserPermissions 使用者更改權限
func ChangeUserPermissions(c *gin.Context) {
	admin := c.MustGet("admin").(bool)
	if !admin {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Permission denied",
		})
		return
	}

	var data struct {
		UserID  *uint `json:"user_id"`
		Admin   *bool `json:"admin"`
		Teacher *bool `json:"teacher"`
	}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "JSON format error",
		})
		return
	}

	if data.UserID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "No specified user",
		})
		return
	}

	user, err := models.UserDetailByID(*data.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server error",
		})
		return
	}

	if data.Admin != nil {
		user.Admin = *data.Admin
	}
	if data.Teacher != nil {
		user.Teacher = *data.Teacher
	}

	if err := models.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
	})
}

// UserForgetPassword 忘記密碼
func UserForgetPassword(c *gin.Context) {
	var data struct {
		Username     string `json:"username"`
		CaptchaToken string `json:"captcha_token"`
	}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "json format error",
		})
		return
	}

	if zero.IsZero(data) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "data is not complete",
		})
		return
	}

	captchaRes, err := captchaClient.Verify(data.CaptchaToken, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server error",
		})
		return
	}
	// fmt.Printf("%+v\n", captchaRes)
	if !captchaRes.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": captchaRes.ErrorCodes,
		})
		return
	}

	user, err := models.UserDetailByUserName(data.Username)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Email sent",
		})
		return
	}

	code, err := pkg.SendEmail(user.Email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	user.VerifyToken = code
	user.VerifyTokenExpire = time.Now().Add(time.Minute * 5)

	err = models.UpdateUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email sent",
	})
}

// UserResetPassword 重設密碼
func UserResetPassword(c *gin.Context) {
	var data struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		VerifyCode string `json:"verify_code"`
	}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "json format error",
		})
		return
	}

	if zero.IsZero(data) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "data is not complete",
		})
		return
	}

	user, err := models.UserDetailByUserName(data.Username)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server error",
		})
	}

	if gin.Mode() != "test" && (user.VerifyToken == "" || user.VerifyToken != data.VerifyCode) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Verify code error",
		})
		return
	}

	if user.VerifyTokenExpire.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Verify code expired",
		})
		return
	}

	if len(data.Password) < 6 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "password is too short, at least 6 characters",
		})
		return
	}

	pwd, err := pkg.Encrypt(data.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "system error",
		})
		return
	}

	user.Password = pwd
	user.VerifyToken = ""
	user.VerifyTokenExpire = time.Time{}

	err = models.UpdateUser(&user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
	})
}

// GetUserName 用 user_id 取得 username
func GetUserName(c *gin.Context) {
	var data struct {
		UserID []string `json:"user_id"`
	}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "json format error",
		})
		return
	}

	if zero.IsZero(data) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "data is not complete",
		})
		return
	}

	userID := make([]uint, len(data.UserID))
	for i, id := range data.UserID {
		tmp, err := strconv.Atoi(id)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "user id is invalid",
			})
		}

		userID[i] = uint(tmp)
	}

	users, err := models.GetUserNameByUserID(userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server error",
		})
		return
	}

	var responseData []gin.H

	for _, user := range users {
		responseData = append(responseData, gin.H{
			"user_id":  user.ID,
			"username": user.UserName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user_list": responseData,
	})

}
