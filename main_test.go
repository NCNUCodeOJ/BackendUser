package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/NCNUCodeOJ/BackendUser/models"
	"github.com/NCNUCodeOJ/BackendUser/router"
	"github.com/NCNUCodeOJ/BackendUser/views"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/assert.v1"
)

var d struct {
	Token string `json:"token"`
}
var userID string
var userName = "vincent"
var userPath = "/api/v1/user"
var password = "123456"
var announcementsLength = 0
var announcementsID int

func contentType() (string, string) {
	return "Content-Type", "application/json"
}

func init() {
	gin.SetMode(gin.TestMode)
	models.Setup()
	views.Setup()
}

func TestUserRegister(t *testing.T) {
	var data = []byte(`{
		"username": "` + userName + `",
		"password": "123456",
		"realname": "郭子緯",
		"email": "s107213004@ncnu.edu.tw",
		"student_id": "s107213004",
		"avatar": "https://avatars0.githubusercontent.com/u/1234?v=4"
	}`)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("POST", userPath, bytes.NewBuffer(data))
	req.Header.Set(contentType())
	r.ServeHTTP(w, req)

	s := struct {
		UserID string `json:"user_id"`
	}{}

	body, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(body, &s)
	userID = s.UserID

	id, err := strconv.Atoi(userID)
	assert.Equal(t, err, nil)

	user, err := models.UserDetailByID(uint(id))
	assert.Equal(t, err, nil)

	user.Admin = true
	err = models.UpdateUser(&user)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogin(t *testing.T) {
	var data = []byte(`{
		"username": "vincent",
		"password": "` + password + `"
	}`)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("POST", "/api/v1/token", bytes.NewBuffer(data))
	req.Header.Set(contentType())
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	body, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(body, &d)
}

func TestPing(t *testing.T) {
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("GET", "/ping", nil)
	req.Header.Set(contentType())
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	body, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(body, &d)
}

func TestRefresh(t *testing.T) {
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("GET", "/api/v1/token", bytes.NewBuffer(make([]byte, 1000)))
	req.Header.Set("Authorization", "Bearer "+d.Token)
	r.ServeHTTP(w, req)
	body, _ := ioutil.ReadAll(w.Body)
	old := d.Token
	json.Unmarshal(body, &d)
	assert.Equal(t, old, d.Token)
}

func TestUserInfo(t *testing.T) {
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("GET", userPath, nil)
	req.Header.Set("Authorization", "Bearer "+d.Token)
	r.ServeHTTP(w, req)
	body, _ := ioutil.ReadAll(w.Body)
	s := struct {
		Name string `json:"username"`
	}{}
	json.Unmarshal(body, &s)
	assert.Equal(t, "vincent", s.Name)
}

func TestUserChangeInfo(t *testing.T) {
	var data = []byte(`{
		"student_id": "vincentinttsh"
	}`)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("PATCH", userPath, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+d.Token)
	req.Header.Set(contentType())
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	req, _ = http.NewRequest("GET", userPath, bytes.NewBuffer(make([]byte, 1000)))
	req.Header.Set("Authorization", "Bearer "+d.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	body, _ := ioutil.ReadAll(w.Body)
	s := struct {
		StudentID string `json:"student_id"`
	}{}
	json.Unmarshal(body, &s)
	assert.Equal(t, "vincentinttsh", s.StudentID)
}

func TestChangeUserPermissions(t *testing.T) {
	var data = []byte(`{
		"user_id": ` + userID + `,
		"teacher": true
	}`)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("PATCH", "/api/v1/user/permission", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+d.Token)
	req.Header.Set(contentType())
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	req, _ = http.NewRequest("GET", userPath, bytes.NewBuffer(make([]byte, 1000)))
	req.Header.Set("Authorization", "Bearer "+d.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	body, _ := ioutil.ReadAll(w.Body)
	s := struct {
		Teacher bool `json:"teacher"`
		Admin   bool `json:"admin"`
	}{}
	json.Unmarshal(body, &s)
	assert.Equal(t, true, s.Teacher)
	assert.Equal(t, true, s.Admin)
}

func TestUserForgePassword(t *testing.T) {
	var data = []byte(`{
		"username": "vincent",
		"captcha_token": "10000000-aaaa-bbbb-cccc-000000000001"
	}`)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("POST", "/api/v1/forget_password", bytes.NewBuffer(data))
	req.Header.Set(contentType())
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserResetPassword(t *testing.T) {
	password = "12345678"

	var data = []byte(`{
		"username": "vincent",
		"verify_code": "test_code",
		"password": "` + password + `"
	}`)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("POST", "/api/v1/reset_password", bytes.NewBuffer(data))
	req.Header.Set(contentType())
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	TestLogin(t)
}

func TestUserName(t *testing.T) {
	oldUserID := userID
	userName = "vincentinttsh"
	TestUserRegister(t)
	var data = []byte(`{
		"user_id": ["` + userID + `","` + oldUserID + `"]
	}`)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("POST", "/api/v1/username", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+d.Token)
	req.Header.Set(contentType())
	r.ServeHTTP(w, req)
	body, _ := ioutil.ReadAll(w.Body)
	s := struct {
		UserList []struct {
			UserID   uint   `json:"user_id"`
			UserName string `json:"username"`
		} `json:"user_list"`
	}{}
	json.Unmarshal(body, &s)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 2, len(s.UserList))
	assert.Equal(t, "vincent", s.UserList[0].UserName)
	assert.Equal(t, "vincentinttsh", s.UserList[1].UserName)
}

func TestCreateAnnouncement(t *testing.T) {
	var data = []byte(`{
		"title": "test_title",
		"content": "test_content"
	}`)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("POST", "/api/v1/announcements", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+d.Token)
	req.Header.Set(contentType())
	r.ServeHTTP(w, req)
	body, _ := ioutil.ReadAll(w.Body)
	s := struct {
		AnnouncementID int `json:"announcement_id"`
	}{}
	json.Unmarshal(body, &s)
	announcementsID = s.AnnouncementID
	assert.Equal(t, http.StatusCreated, w.Code)
	announcementsLength++
}

func TestGetAllAnnouncement(t *testing.T) {
	TestCreateAnnouncement(t)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("GET", "/api/v1/announcements", nil)
	req.Header.Set(contentType())
	r.ServeHTTP(w, req)
	body, _ := ioutil.ReadAll(w.Body)
	s := struct {
		Announcements []struct {
			Publisher string `json:"publisher"`
		} `json:"announcements"`
	}{}
	json.Unmarshal(body, &s)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, announcementsLength, len(s.Announcements))
	assert.Equal(t, "vincent", s.Announcements[0].Publisher)
}

func TestDeleteAnnouncement(t *testing.T) {
	TestCreateAnnouncement(t)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("DELETE", "/api/v1/announcements/"+strconv.Itoa(announcementsID), nil)
	req.Header.Set("Authorization", "Bearer "+d.Token)
	req.Header.Set(contentType())
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	announcementsLength--
	TestGetAllAnnouncement(t)
}
func TestCleanup(t *testing.T) {
	e := os.Remove("test.db")
	if e != nil {
		t.Fail()
	}
}
