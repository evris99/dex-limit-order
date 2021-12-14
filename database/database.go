package database

import (
	"fmt"

	"github.com/evris99/dex-limit-order/order"

	"github.com/evris99/dex-limit-order/database/model"

	"gorm.io/gorm"
)

func GetUserFromTelegramID(DB *gorm.DB, telegramID int64) (*model.User, error) {
	user := new(model.User)
	res := DB.Where("telegram_id = ?", telegramID).First(user)

	return user, res.Error
}

// Queries the database and returns an slice of all the orders
func GetOrders(DB *gorm.DB) ([]*order.Order, error) {
	var models []*model.Order
	res := DB.Find(&models)
	if res.Error != nil {
		return nil, fmt.Errorf("could not get orders from db: %w", res.Error)
	}

	orders := make([]*order.Order, len(models))
	for i := range models {
		var err error
		orders[i], err = models[i].ToOrder()
		if err != nil {
			return nil, err
		}
	}

	return orders, nil
}

// Queries the database and returns a slice of orders matching the user
func GetOrdersByUser(DB *gorm.DB, telegramID int64) ([]*order.Order, error) {
	var user model.User
	res := DB.Where("telegram_id = ?", telegramID).First(&user)

	if res.Error != nil {
		return nil, fmt.Errorf("could not get orders from db: %w", res.Error)
	}

	orders := make([]*order.Order, len(user.Orders))
	for i := range user.Orders {
		var err error
		orders[i], err = user.Orders[i].ToOrder()
		if err != nil {
			return nil, err
		}
	}

	return orders, nil
}

// Adds the given order to the database and returns its ID
func AddOrder(DB *gorm.DB, order *order.Order) (uint, error) {
	model, err := model.FromOrder(order)
	if err != nil {
		return 0, fmt.Errorf("could not convert order to model: %w", err)
	}
	res := DB.Create(model)
	if res.Error != nil {
		return 0, fmt.Errorf("could not create order: %w", err)
	}

	return model.ID, nil
}

func AddUser(DB *gorm.DB, telegramID int64, username string) error {
	user := &model.User{
		TelegramID:       telegramID,
		TelegramUserName: username,
	}

	return DB.Create(user).Error
}
