package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NCNUCodeOJ/BackendUser/models"
	"github.com/NCNUCodeOJ/BackendUser/router"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/assert.v1"
)

var d struct {
	Token string `json:"token"`
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
		"email": "s107213004@main"
	}`)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("POST", "/api/v1/user", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
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
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	body, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(body, &d)
}

func TestPing(t *testing.T) {
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("GET", "/ping", nil)
	req.Header.Set("Content-Type", "application/json")
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
	req, _ := http.NewRequest("GET", "/api/v1/user", bytes.NewBuffer(make([]byte, 1000)))
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
		"username": "vincentinttsh"
	}`)
	r := router.SetupRouter()
	w := httptest.NewRecorder() // 取得 ResponseRecorder 物件
	req, _ := http.NewRequest("PATCH", "/api/v1/user", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+d.Token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	req, _ = http.NewRequest("GET", "/api/v1/user", bytes.NewBuffer(make([]byte, 1000)))
	req.Header.Set("Authorization", "Bearer "+d.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	body, _ := ioutil.ReadAll(w.Body)
	s := struct {
		Name string `json:"username"`
	}{}
	json.Unmarshal(body, &s)
	assert.Equal(t, "vincentinttsh", s.Name)
}

func TestCleanup(t *testing.T) {
	e := os.Remove("test.db")
	if e != nil {
		t.Fail()
	}
}
