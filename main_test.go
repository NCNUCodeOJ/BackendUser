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
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/assert.v1"
)

var d struct {
	Token string `json:"token"`
}
var userID string
var userPath = "/api/v1/user"

func contentType() (string, string) {
	return "Content-Type", "application/json"
}

func init() {
	gin.SetMode(gin.TestMode)
	models.Setup()
}

func TestUserRegister(t *testing.T) {
	var data = []byte(`{
		"username": "vincent",
		"password": "1234",
		"realname": "郭子緯",
		"email": "s107213004@main",
		"student_id": "s107213004"
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
		"password": "1234"
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
	req, _ := http.NewRequest("GET", userPath, bytes.NewBuffer(make([]byte, 1000)))
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

func TestCleanup(t *testing.T) {
	e := os.Remove("test.db")
	if e != nil {
		t.Fail()
	}
}
