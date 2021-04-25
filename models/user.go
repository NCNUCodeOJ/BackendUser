package models

import "gorm.io/gorm"

//User Database - database
type User struct {
	gorm.Model
	SchoolID  int    `gorm:"NOT NULL;"`
	StudentID string `gorm:"type:varchar(15) NOT NULL;"`
	Email     string `gorm:"type:varchar(40) NOT NULL;"`
	UserName  string `gorm:"type:varchar(20) NOT NULL;"`
	Password  string `gorm:"type:varchar(100) NOT NULL;"`
	RealName  string `gorm:"type:varchar(30)"`
	Admin     bool   `gorm:"default:false"`
}

// CreateUser 新增 user
func CreateUser(user *User) (err error) {
	err = DB.Create(&user).Error
	return
}

// UpdateUser 更新 user
func UpdateUser(user *User) (err error) {
	err = DB.Save(&user).Error
	return
}

// UserDetailByID 透過 id 取得 user
func UserDetailByID(id uint) (user User, err error) {
	err = DB.Where("id = ?", id).First(&user).Error
	return
}

// UserDetailByUserName 透過 UserName 取得 username
func UserDetailByUserName(name string) (user User, err error) {
	err = DB.Where("user_name = ?", name).First(&user).Error
	return
}
