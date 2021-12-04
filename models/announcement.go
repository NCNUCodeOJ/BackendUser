package models

import "gorm.io/gorm"

// Announcement database model
type Announcement struct {
	gorm.Model
	Title   string `gorm:"type:varchar(255);not null"`
	Content string `gorm:"type:text;not null"`
	UserID  uint
	User    User `gorm:"foreignkey:UserID"`
}

// Create creates a new announcement
func (announcement *Announcement) Create() error {
	return DB.Create(&announcement).Error
}

// GetAllAnnouncements returns an announcement by its ID
func GetAllAnnouncements() (announcements []Announcement, err error) {
	err = DB.Preload("User").Find(&announcements).Error
	if err != nil {
		return
	}
	return
}

// DeleteAnnouncement deletes an announcement by its ID
func DeleteAnnouncement(id uint) error {
	return DB.Delete(&Announcement{}, id).Error
}
