package wallet

import (
	"bytes"
	"errors"
	"fmt"
	"syscall"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/term"
)

var ErrIncorrectAccounts error = errors.New("must have exactly one account")

type Wallet struct {
	KeyStore *keystore.KeyStore
	Account  accounts.Account
}

// Returns the public address of the wallet
func (w *Wallet) GetAddress() common.Address {
	return w.Account.Address
}

// Creates a new wallet with a password from stdin and returns it
func GenerateFromTerm(path string) (*Wallet, error) {
	fmt.Printf("Enter a password for your account: ")
	password, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return nil, fmt.Errorf("could not read password: %w", err)
	}
	fmt.Printf("\n")

	return Generate(path, string(bytes.TrimSpace(password)))
}

// Creates a new wallet
func Generate(path, password string) (*Wallet, error) {
	wallet := &Wallet{
		KeyStore: keystore.NewKeyStore(path, keystore.StandardScryptN, keystore.StandardScryptP),
	}

	var err error
	wallet.Account, err = wallet.KeyStore.NewAccount(password)
	return wallet, err
}

// Loads an existing wallet and unlocks it with a password from stdin
func Load(path string) (*Wallet, error) {
	fmt.Printf("Enter the password for the account: ")
	password, err := term.ReadPassword(syscall.Stdin)
	fmt.Printf("\n")
	if err != nil {
		return nil, fmt.Errorf("could not read password: %w", err)
	}

	return LoadWithPassword(path, string(bytes.TrimSpace(password)))
}

// Loads an existing wallet and unlocks it with the given password
func LoadWithPassword(path, password string) (*Wallet, error) {
	wallet := &Wallet{
		KeyStore: keystore.NewKeyStore(path, keystore.StandardScryptN, keystore.StandardScryptP),
	}

	if len(wallet.KeyStore.Accounts()) != 1 {
		//TODO: Make user choose account
		return nil, ErrIncorrectAccounts
	}

	wallet.Account = wallet.KeyStore.Accounts()[0]
	err := wallet.KeyStore.Unlock(wallet.Account, password)
	return wallet, err
}
