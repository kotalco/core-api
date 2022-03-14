package tfa

import (
	"bytes"
	"image/png"

	"github.com/pquerna/otp/totp"
)

func CreateQRCode(accountName string) (bytes.Buffer, string, error) {
	key, err := totp.Generate(
		totp.GenerateOpts{
			Issuer:      "kotal.co",
			AccountName: accountName,
		})

	if err != nil {
		return bytes.Buffer{}, "", err
	}

	img, err := key.Image(200, 200)
	if err != nil {
		return bytes.Buffer{}, "", err
	}

	var qrCode bytes.Buffer
	png.Encode(&qrCode, img)
	return qrCode, key.Secret(), nil
}

func CheckOtp(userTOTPSecret string, otp string) bool {
	valid := totp.Validate(otp, userTOTPSecret)
	if valid {
		return true
	}

	return false
}
