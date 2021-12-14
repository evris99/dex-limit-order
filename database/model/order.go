package model

import (
	"strings"

	"github.com/evris99/dex-limit-order/order"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	Path         string
	Type         string
	SellAmount   string
	BuyAmount    string
	Slippage     float64
	GasLimit     uint64
	GasPriceMult float64
	UserID       uint //foreign key
}

// Converts a database order to an order struct
func (model Order) ToOrder() (*order.Order, error) {
	o := &order.Order{
		ID:           model.ID,
		Path:         strings.Split(model.Path, ","),
		SellAmount:   model.SellAmount,
		BuyAmount:    model.BuyAmount,
		Slippage:     model.Slippage,
		GasPriceMult: model.GasPriceMult,
		GasLimit:     model.GasLimit,
	}

	switch model.Type {
	case "limit":
		o.Type = order.Limit
	case "stop_limit":
		o.Type = order.StopLimit
	default:
		return nil, order.ErrInvalidType
	}

	return o, nil
}

// Converts an order struct to a database order
func FromOrder(o *order.Order) (*Order, error) {
	m := &Order{
		Path:         strings.Join(o.Path, ","),
		SellAmount:   o.SellAmount,
		BuyAmount:    o.BuyAmount,
		Slippage:     o.Slippage,
		GasLimit:     o.GasLimit,
		GasPriceMult: o.GasPriceMult,
		UserID:       o.UserID,
	}

	switch o.Type {
	case order.Limit:
		m.Type = "limit"
	case order.StopLimit:
		m.Type = "stop_limit"
	default:
		return nil, order.ErrInvalidType
	}

	return m, nil
}
