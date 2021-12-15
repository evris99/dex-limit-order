package manager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/evris99/dex-limit-order/bsc_client"
	"github.com/evris99/dex-limit-order/database"
	"github.com/evris99/dex-limit-order/database/model"
	"github.com/evris99/dex-limit-order/order"
	"github.com/evris99/dex-limit-order/price"
	"github.com/evris99/dex-limit-order/wallet"

	"gopkg.in/tucnak/telebot.v2"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

type Manager struct {
	Wallet        *wallet.Wallet
	Client        *bsc_client.Client
	ChainID       int64
	RouterHex     string
	OrderChannels map[uint]chan bool
	channelsMutex sync.Mutex
	DB            *database.DB
	Bot           *telebot.Bot
}

// Returns a new order manager
func New(bot *telebot.Bot, wallet *wallet.Wallet, db *database.DB, chainID int64, rpcURL, routerHex string) (*Manager, error) {
	client, err := bsc_client.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to endpoint: %w", err)
	}

	return &Manager{
		Wallet:        wallet,
		Client:        client,
		ChainID:       chainID,
		RouterHex:     routerHex,
		OrderChannels: make(map[uint]chan bool),
		DB:            db,
		Bot:           bot,
	}, nil
}

// Add the existing database orders to the manager
func (m *Manager) StartDBOrders() error {
	orders, err := m.DB.GetOrders()
	if err != nil {
		return err
	}

	for _, order := range orders {
		if err := order.Init(m.Client, m.RouterHex, m.ChainID); err != nil {
			return fmt.Errorf("could not initialize order: %w", err)
		}

		go m.checkPrices(order)
	}

	return nil
}

// Adds a new order to the manager
func (m *Manager) AddOrder(order *order.Order) error {
	if err := order.Init(m.Client, m.RouterHex, m.ChainID); err != nil {
		return fmt.Errorf("could not initialize order: %w", err)
	}

	// Approve Order
	tx, err := order.ApproveMax(m.Client, m.Wallet)
	if err != nil {
		return fmt.Errorf("could not approve order: %w", err)
	}

	// Wait for it to be approved
	if tx == nil {
		log.Printf("Token %s was already approved\n", order.GetSellToken().Address.Hex())
	} else {
		_, err := m.Client.GetPendingTxReceipt(context.Background(), tx)
		if err != nil {
			return fmt.Errorf("could not wait for token approval: %w", err)
		}
	}

	order.ID, err = m.DB.AddOrder(order)
	if err != nil {
		return fmt.Errorf("could not add order to db: %w", err)
	}

	go m.checkPrices(order)
	return nil
}

// Removes an order from the manager
func (m *Manager) RemoveOrder(id uint) error {
	m.channelsMutex.Lock()
	channel, ok := m.OrderChannels[id]
	m.channelsMutex.Unlock()

	if !ok {
		return ErrOrderNotFound
	}

	channel <- true
	m.channelsMutex.Lock()
	delete(m.OrderChannels, id)
	m.channelsMutex.Unlock()

	return m.DB.SQL.Delete(&model.Order{}, id).Error
}

// TODO: Clean up the code and add sending on success or failure
// Compares a stream o prices and swaps the tokens if the price is right
func (m *Manager) checkPrices(order *order.Order) {
	done, stopChan := make(chan bool), make(chan bool)

	// Saves the channel for stopping the order
	m.channelsMutex.Lock()
	m.OrderChannels[order.ID] = stopChan
	m.channelsMutex.Unlock()

	priceChan, errChan := order.GetPriceStream(m.Client, done)
	log.Printf("Started watching order #%d\n", order.ID)
	for {
		select {
		case currentPrice := <-priceChan:
			floatPrice := price.RemoveDecimals(currentPrice, order.GetBuyToken().Decimals)
			log.Printf("Current price for swap: %s\n", floatPrice.Text('f', 18))
			tx, err := order.CompAndSwap(m.Client, m.Wallet, currentPrice)
			if err != nil {
				log.Println(err)
				return
			}

			if tx != nil {
				done <- true
				log.Printf("Swap transaction %s <-> %s pending: %s\n", order.GetSellToken().Address.Hex(), order.GetBuyToken().Address.Hex(), tx.Hash().Hex())
				_, err := m.Client.GetPendingTxReceipt(context.Background(), tx)
				if err != nil {
					log.Println(err)
					return
				}

				log.Printf("Swap transaction %s <-> %s completed: %s\n", order.GetSellToken().Address.Hex(), order.GetBuyToken().Address.Hex(), tx.Hash().Hex())
				if err := m.RemoveOrder(order.ID); err != nil {
					log.Printf("Could not remove order %d\n", order.ID)
					return
				}
				return
			}
		case err := <-errChan:
			log.Println(err)
		case <-stopChan:
			done <- true

			m.channelsMutex.Lock()
			delete(m.OrderChannels, order.ID)
			m.channelsMutex.Unlock()

			return
		}
	}
}
