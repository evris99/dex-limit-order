package main

import (
	"flag"
	"log"

	"github.com/evris99/dex-limit-order/bot"
	"github.com/evris99/dex-limit-order/config"
	"github.com/evris99/dex-limit-order/database"
	"github.com/evris99/dex-limit-order/manager"
	"github.com/evris99/dex-limit-order/wallet"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load the configuration from the given file
	confPath := flag.String("c", "./configuration.toml", "the path to the configuration flag")
	flag.Parse()
	config, err := config.Load(*confPath)
	if err != nil {
		log.Fatalln(err)
	}

	// Initializes a wallet instance
	wallet, err := wallet.LoadWithPassword(config.WalletPath, config.Password)
	if err != nil {
		log.Fatalln(err)
	}

	// Creates a database connection
	DB, err := database.New(config.DBPath)
	if err != nil {
		log.Fatalln(err)
	}

	// Migrates the database
	if err := DB.Migrate(); err != nil {
		log.Fatalln(err)
	}

	// Initialize telegram bot
	telegramBot, err := bot.New(DB, config.TelegramAPIKey)
	if err != nil {
		log.Fatalln(err)
	}

	// Create order manager
	manager, err := manager.New(telegramBot, wallet, DB, config.ChainID, config.EndpointURL, config.PancakeRouterHex)
	if err != nil {
		log.Fatalln(err)
	}

	// Start watching existing orders from the database
	if err := manager.StartDBOrders(); err != nil {
		log.Fatalln(err)
	}

	bot.Listen(telegramBot, DB, manager)
}
