package views

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/meyskens/go-hcaptcha"
)

// Setup setup api
func Setup() {

	if gin.Mode() == "test" {
		captchaClient = hcaptcha.New("0x0000000000000000000000000000000000000000")
		return
	}
	captchaClient = hcaptcha.New(os.Getenv("HCAPTCHA_SECRET"))
}
