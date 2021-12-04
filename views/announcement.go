package views

import (
	"net/http"
	"strconv"

	"github.com/NCNUCodeOJ/BackendUser/models"
	"github.com/gin-gonic/gin"
	"github.com/vincentinttsh/zero"
)

// CreateAnnouncement is a function to create announcement
func CreateAnnouncement(c *gin.Context) {
	admin := c.MustGet("admin").(bool)
	if !admin {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Permission denied",
		})
		return
	}
	userID := c.MustGet("userID").(uint)

	user, err := models.UserDetailByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "no such user",
		})
		return
	}

	var announcement models.Announcement
	var data struct {
		Title   string `json:"title"`
		Content string `json:"content"`
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

	announcement.Title = data.Title
	announcement.Content = data.Content
	announcement.User = user

	if err := announcement.Create(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "create announcement error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "create announcement success",
		"announcement_id": announcement.ID,
	})
}

// GetAllAnnouncements is a function to get all announcement
func GetAllAnnouncements(c *gin.Context) {
	var announcements []models.Announcement
	var err error
	var announcementsData []gin.H

	if announcements, err = models.GetAllAnnouncements(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "get all announcements error",
		})
		return
	}

	for _, announcement := range announcements {
		announcementsData = append(announcementsData, gin.H{
			"announcement_id": announcement.ID,
			"title":           announcement.Title,
			"content":         announcement.Content,
			"created_at":      announcement.CreatedAt.Unix(),
			"publisher":       announcement.User.UserName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "get all announcements success",
		"announcements": announcementsData,
	})
}

// DeleteAnnouncement is a function to delete announcement
func DeleteAnnouncement(c *gin.Context) {
	var announcementID uint
	admin := c.MustGet("admin").(bool)
	if !admin {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Permission denied",
		})
		return
	}

	if ID, err := strconv.Atoi(c.Params.ByName("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "announcement id error",
		})
	} else {
		announcementID = uint(ID)
	}

	if err := models.DeleteAnnouncement(announcementID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "delete announcement error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "delete announcement success",
	})

}
