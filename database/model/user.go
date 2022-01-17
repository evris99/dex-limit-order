package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	TelegramID int64
	Orders     []Order
}
