package services

import (
	"github.com/wing8169/golang-htmx-chat-app/dto"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers() []*dto.UserDto {

	password, _ := bcrypt.GenerateFromPassword([]byte("jxiong"), 8)

	return []*dto.UserDto{
		{
			ID:       "1",
			Username: "jxiong",
			Password: string(password),
		},
	}
}
