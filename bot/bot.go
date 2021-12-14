package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/evris99/dex-limit-order/order"

	"github.com/evris99/dex-limit-order/manager"

	"github.com/evris99/dex-limit-order/database"

	"gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

// Creates and returns a new telebot instance
func New(DB *gorm.DB, apiKey string) (*telebot.Bot, error) {
	// Use long polling
	poller := &telebot.LongPoller{Timeout: 15 * time.Second}

	// Limit usage to specific users
	userFilter := telebot.NewMiddlewarePoller(poller, filter(DB))

	//Initialize bot
	bot, err := telebot.NewBot(telebot.Settings{
		Poller: userFilter,
		Token:  apiKey,
	})

	return bot, err
}

// Starts the bot listening proccess
func Listen(bot *telebot.Bot, DB *gorm.DB, manager *manager.Manager) {
	mainMenu := &telebot.ReplyMarkup{}
	listBtn := mainMenu.Data("List orders", "list")
	mainMenu.Inline(mainMenu.Row(listBtn))

	orderMenu := &telebot.ReplyMarkup{}
	removeBtn := mainMenu.Data("Remove order", "remove")

	bot.Handle("/start", func(m *telebot.Message) {
		bot.Send(m.Sender, "Welcome to PancakeSwap limit order bot 🙂", mainMenu)
	})

	bot.Handle(telebot.OnDocument, onFile(DB, bot, manager))
	bot.Handle(&removeBtn, onRemoveBtn(bot, manager))

	bot.Handle(&listBtn, func(c *telebot.Callback) {
		defer bot.Respond(c, &telebot.CallbackResponse{})
		orders, err := database.GetOrdersByUser(DB, c.Sender.ID)
		if err != nil {
			bot.Send(c.Sender, "Could not get orders")
		}

		for _, order := range orders {
			removeBtn.Data = strconv.Itoa(int(order.ID))
			orderMenu.Inline(orderMenu.Row(removeBtn))
			bot.Send(c.Sender, getOrderMsg(*order), orderMenu)
		}
	})

	bot.Start()
}

// Filters incoming messages from users not in database
func filter(db *gorm.DB) func(*telebot.Update) bool {
	return func(u *telebot.Update) bool {
		if u.Message == nil {
			return true
		}

		// Filter users that are not in config
		// TODO: Use database along with redist
		_, err := database.GetUserFromTelegramID(db, u.Message.Sender.ID)
		if err != nil {
			log.Println(err)
			return false
		}

		return true
	}
}

// Returns a callback for handling a document upload
func onFile(DB *gorm.DB, bot *telebot.Bot, manager *manager.Manager) func(*telebot.Message) {
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
		user, err := database.GetUserFromTelegramID(DB, m.Sender.ID)
		if err != nil {
			log.Println(err)
			bot.Send(m.Sender, "Invalid user")
			return
		}

		order.UserID = user.ID
		if err := manager.AddOrder(order); err != nil {
			log.Println(err)
			bot.Send(m.Sender, "Could not add order")
			return
		}

		bot.Send(m.Sender, "Order added")
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
