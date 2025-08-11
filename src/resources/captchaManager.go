package resources

import (
	"CodeStream/src"

	recaptcha2 "github.com/xinguang/go-recaptcha"
)

func ValidateCaptcha(captchaResponse string) bool {
	recaptcha, err := recaptcha2.NewWithSecert(src.Config.GoogleCaptchaKey)
	if err != nil {
		panic(err)
	}
	err = recaptcha.Verify(captchaResponse)
	if err != nil {
		return false
	}
	return true
}
