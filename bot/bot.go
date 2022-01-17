package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/evris99/dex-limit-order/database"
	"github.com/evris99/dex-limit-order/manager"
	"github.com/evris99/dex-limit-order/order"

	"gopkg.in/tucnak/telebot.v2"
)

// Creates and returns a new telebot instance
func New(db *database.DB, apiKey string) (*telebot.Bot, error) {
	// Use long polling
	poller := &telebot.LongPoller{Timeout: 15 * time.Second}

	// Limit usage to specific users
	userFilter := telebot.NewMiddlewarePoller(poller, filter(db))

	//Initialize bot
	bot, err := telebot.NewBot(telebot.Settings{
		Poller: userFilter,
		Token:  apiKey,
	})

	return bot, err
}

// Starts the bot listening proccess
func Listen(bot *telebot.Bot, db *database.DB, manager *manager.Manager) {

	// Button for listing orders
	mainMenu := &telebot.ReplyMarkup{}
	listBtn := mainMenu.Data("List orders", "list")
	mainMenu.Inline(mainMenu.Row(listBtn))

	// Button for removing order
	orderMenu := &telebot.ReplyMarkup{}
	removeBtn := mainMenu.Data("Remove order", "remove")

	bot.Handle("/start", func(m *telebot.Message) {
		bot.Send(m.Sender, "Welcome to PancakeSwap limit order bot 🙂", mainMenu)
	})

	bot.Handle(telebot.OnDocument, onFile(db, bot, manager))
	bot.Handle(&removeBtn, onRemoveBtn(bot, manager))

	bot.Handle(&listBtn, func(c *telebot.Callback) {
		defer bot.Respond(c, &telebot.CallbackResponse{})
		orders, err := db.GetOrdersByUser(c.Sender.ID)
		if err != nil {
			bot.Send(c.Sender, "Could not get orders")
		}

		if len(orders) == 0 {
			bot.Send(c.Sender, "No active orders")
		}

		for _, order := range orders {
			// Add remove order button with correct order ID
			removeBtn.Data = strconv.Itoa(int(order.ID))
			orderMenu.Inline(orderMenu.Row(removeBtn))
			bot.Send(c.Sender, getOrderMsg(*order), orderMenu)
		}
	})

	bot.Start()
}

// Filters incoming messages from users not in database
func filter(db *database.DB) func(*telebot.Update) bool {
	return func(u *telebot.Update) bool {
		if u.Message == nil {
			return true
		}

		// Filter users that are not in config
		_, err := db.GetUserFromTelegramID(u.Message.Sender.ID)
		if err != nil {
			log.Println(err)
			return false
		}

		return true
	}
}

// Returns a callback for handling a document upload
func onFile(db *database.DB, bot *telebot.Bot, manager *manager.Manager) func(*telebot.Message) {
	return func(m *telebot.Message) {
		if m.Document == nil {
			return
		}

		if m.Document.MIME != "application/json" {
			return
		}

		file, err := bot.GetFile(m.Document.MediaFile())
		if err != nil {
			log.Println(err)
			bot.Send(m.Sender, "Could not get file")
			return
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		order := new(order.Order)
		if err := decoder.Decode(order); err != nil {
			log.Println(err)
			bot.Send(m.Sender, "Invalid file")
			return
		}
		user, err := db.GetUserFromTelegramID(m.Sender.ID)
		if err != nil {
			log.Println(err)
			bot.Send(m.Sender, "Invalid user")
			return
		}

		order.UserID = user.ID
		order, err = manager.AddOrder(order)
		if err != nil {
			log.Println(err)
			bot.Send(m.Sender, "Could not add order")
			return
		}

		bot.Send(m.Sender, "Order added:\n%s", getOrderMsg(*order))
	}
}

// Returns a callback for handling a remove button click
func onRemoveBtn(bot *telebot.Bot, manager *manager.Manager) func(*telebot.Callback) {
	return func(c *telebot.Callback) {
		defer bot.Respond(c, &telebot.CallbackResponse{})
		id, err := strconv.Atoi(c.Data)
		if err != nil {
			log.Println(err)
			bot.Send(c.Sender, "something went wrong")
			return
		}

		manager.RemoveOrder(uint(id))
		bot.Send(c.Sender, "Order removed successfully")
	}
}

// Returns the corresponding telegram message for the order
func getOrderMsg(o order.Order) string {
	path := fmt.Sprintf("Path: %s", o.Path[0])
	for i := 1; i < len(o.Path); i++ {
		path += fmt.Sprintf(" -> %s", o.Path[i])
	}
	path += "\n"

	var orderType string
	switch o.Type {
	case order.Limit:
		orderType = "Type: limit"
	case order.StopLimit:
		orderType = "Type: stop_limit"
	default:
		panic("Invalid type")
	}

	amounts := fmt.Sprintf("Buy amount: %v\nSell amount: %v\n", o.BuyAmount, o.SellAmount)
	gasInfo := fmt.Sprintf("Slippage: %v\nGas price multiplier: %v%%\nGas Limit: %v\n", o.Slippage, o.GasPriceMult, o.GasLimit)
	return fmt.Sprintf("ID: %v\n%v%v%s%s", o.ID, orderType, path, amounts, gasInfo)
}
