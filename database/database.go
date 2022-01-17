package database

import (
	"fmt"

	"github.com/evris99/dex-limit-order/database/model"
	"github.com/evris99/dex-limit-order/order"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	SQL *gorm.DB
}

// Creates a new database connection
func New(dbPath string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) Migrate() error {
	if err := db.SQL.AutoMigrate(&model.Order{}); err != nil {
		return err
	}

	return db.SQL.AutoMigrate(&model.User{})
}

// Queries the database and returns an slice of all the orders
func (db *DB) GetOrders() ([]*order.Order, error) {
	var models []*model.Order
	res := db.SQL.Find(&models)
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
func (db *DB) GetOrdersByUser(telegramID int64) ([]*order.Order, error) {
	var user model.User
	res := db.SQL.Preload("Orders").Where("telegram_id = ?", telegramID).First(&user)

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

// Searches a user in the database based on the telegram ID
func (db *DB) GetUserFromTelegramID(telegramID int64) (*model.User, error) {
	user := new(model.User)
	res := db.SQL.Where("telegram_id = ?", telegramID).First(user)

	return user, res.Error
}

// Adds the given order to the database and returns its ID
func (db *DB) AddOrder(order *order.Order) (uint, error) {
	model, err := model.FromOrder(order)
	if err != nil {
		return 0, fmt.Errorf("could not convert order to model: %w", err)
	}
	res := db.SQL.Create(model)
	if res.Error != nil {
		return 0, fmt.Errorf("could not create order: %w", err)
	}

	return model.ID, nil
}

// Inserts a new user with the given fields to the database
func (db *DB) AddUser(telegramID int64, username string) error {
	user := &model.User{
		TelegramID:       telegramID,
		TelegramUserName: username,
	}

	return db.SQL.Create(user).Error
}

func (db *DB) DeleteOrder(orderID uint) error {
	return db.SQL.Delete(&model.Order{}, orderID).Error
}
