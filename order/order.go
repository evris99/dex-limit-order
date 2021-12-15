package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/evris99/dex-limit-order/blockclient"
	"github.com/evris99/dex-limit-order/contracts/bep20_token"
	"github.com/evris99/dex-limit-order/contracts/factory"
	"github.com/evris99/dex-limit-order/contracts/pair"
	"github.com/evris99/dex-limit-order/contracts/router"
	"github.com/evris99/dex-limit-order/price"
	"github.com/evris99/dex-limit-order/wallet"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/sync/errgroup"
)

const (
	// The interval to check the price in Milliseconds
	CheckIntervalMil time.Duration = 500
	// The deadline for completing the swap in Minutes
	SwapDeadlineMin time.Duration = 20
)

var (
	ErrConv        = errors.New("could not convert amount to big int")
	ErrInvalidType = errors.New("invalid type value")
)

type Type int

const (
	Limit Type = iota
	StopLimit
)

// Custom json unmarshal for Type
func (t *Type) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "limit":
		*t = Limit
	case "stop_limit":
		*t = StopLimit
	default:
		return ErrInvalidType
	}

	return nil
}

// Custom json unmarshal for Type
func (t Type) MarshalJSON() ([]byte, error) {
	var s string
	switch t {
	case Limit:
		s = "limit"
	case StopLimit:
		s = "stop_limit"
	default:
		return nil, ErrInvalidType
	}

	return json.Marshal(s)
}

type Token struct {
	Address  common.Address
	Decimals *big.Int
	Amount   *big.Int
	Instance *bep20_token.Bep20Token
}

type Router struct {
	Address  common.Address
	Instance *router.Router
}

type Order struct {
	ID           uint     `json:"id"`
	Type         Type     `json:"type"`
	Path         []string `json:"path"`
	SellAmount   string   `json:"sell_amount"`
	BuyAmount    string   `json:"buy_amount"`
	Slippage     float64  `json:"slippage"`
	GasPriceMult float64  `json:"gas_price_multiplier"`
	GasLimit     uint64   `json:"gas_limit"`
	UserID       uint
	tokens       []*Token
	router       *Router
	chainID      *big.Int
}

// Initializes the Order
func (o *Order) Init(client bind.ContractBackend, routerHex string, chainID int64) error {
	o.router = &Router{Address: common.HexToAddress(routerHex)}
	o.chainID = big.NewInt(chainID)

	var err error
	o.router.Instance, err = router.NewRouter(o.router.Address, client)
	if err != nil {
		return fmt.Errorf("could not initialize router contract: %w", err)
	}

	o.tokens = make([]*Token, len(o.Path))
	for i, address := range o.Path {
		o.tokens[i], err = getToken(address, client)
		if err != nil {
			return fmt.Errorf("could not get token %s: %w", address, err)
		}
	}

	// TODO: Make handle multiple paths
	if len(o.Path) != 2 {
		return fmt.Errorf("supports only 2 addesses")
	}

	o.tokens[0].Amount, err = getAmount(o.tokens[0], o.SellAmount)
	if err != nil {
		return err
	}

	o.tokens[len(o.tokens)-1].Amount, err = getAmount(o.tokens[len(o.tokens)-1], o.BuyAmount)
	return err
}

// Returns the addresses of the path
func (o *Order) GetAddreses() []common.Address {
	addrs := make([]common.Address, len(o.Path))
	for i := range o.Path {
		addrs[i] = o.tokens[i].Address
	}

	return addrs
}

// Returns the address of the token which will be bought
func (o *Order) GetBuyToken() *Token {
	return o.tokens[len(o.tokens)-1]
}

// Returns the address of the token which will be sold
func (o *Order) GetSellToken() *Token {
	return o.tokens[0]
}

// Executes the approve transaction if not already approved
// Returns the transaction or nil if the token is already approved
func (o *Order) Approve(c *blockclient.Client, wallet *wallet.Wallet, ammount *big.Int) (*types.Receipt, error) {
	allowance, err := o.GetSellToken().Instance.Allowance(nil, wallet.GetAddress(), o.router.Address)
	if err != nil {
		return nil, err
	}

	// Check if it is already approved
	if allowance.Cmp(o.GetSellToken().Amount) != -1 {
		return nil, nil
	}

	auth, err := o.getTransactOpts(c, wallet)
	if err != nil {
		return nil, err
	}

	tx, err := o.GetSellToken().Instance.Approve(auth, o.router.Address, ammount)
	if err != nil {
		return nil, fmt.Errorf("could not approve token: %w", err)
	}

	return c.GetPendingTxReceipt(context.Background(), tx)
}

//Executes the approve transaction with the exact order sell amount
func (o *Order) ApproveAmount(c *blockclient.Client, wallet *wallet.Wallet) (*types.Receipt, error) {
	return o.Approve(c, wallet, o.GetSellToken().Amount)
}

// Executes the approve transaction with the maximum amount
func (o *Order) ApproveMax(c *blockclient.Client, wallet *wallet.Wallet) (*types.Receipt, error) {
	return o.Approve(c, wallet, abi.MaxUint256)
}

// Executes the swap transaction
func (o *Order) Swap(c *blockclient.Client, wallet *wallet.Wallet, amountOut *big.Int) (*types.Receipt, error) {
	auth, err := o.getTransactOpts(c, wallet)
	if err != nil {
		return nil, err
	}

	// Calculate the amount with the slippage
	amountWithSlippage := new(big.Int).Sub(amountOut, price.MultiplyPercent(amountOut, o.Slippage))

	deadline := big.NewInt(time.Now().Add(time.Minute * SwapDeadlineMin).Unix())
	tx, err := o.router.Instance.SwapExactTokensForTokensSupportingFeeOnTransferTokens(auth, o.GetSellToken().Amount, amountWithSlippage, o.GetAddreses(), wallet.GetAddress(), deadline)
	if err != nil {
		return nil, err
	}

	return c.GetPendingTxReceipt(context.Background(), tx)
}

// Compares the given amount with the wanted amount and returns the transaction or nil
// if no transaction has been made
func (o *Order) CompAndSwap(c *blockclient.Client, wallet *wallet.Wallet, amount *big.Int) (*types.Receipt, error) {

	if o.Type == Limit && amount.Cmp(o.GetBuyToken().Amount) != -1 {
		return o.Swap(c, wallet, o.GetBuyToken().Amount)
	}

	if o.Type == StopLimit && amount.Cmp(o.GetBuyToken().Amount) != 1 {
		return o.Swap(c, wallet, amount)
	}

	return nil, nil
}

// Returns a price channel for receiving a stream of prices and a channel for error.
// Stops only when it receives from the done channel
func (o *Order) GetPriceStream(c bind.ContractBackend, done <-chan bool) (chan *big.Int, chan error) {
	priceChan, errChan := make(chan *big.Int), make(chan error)

	go func() {
		pairInstance, err := getPair(c, o.router.Instance, o.tokens[0].Address, o.tokens[1].Address)
		if err != nil {
			errChan <- err
			return
		}

		for {
			time.Sleep(time.Second)

			syncChan := make(chan *pair.PairSync)
			sub, err := pairInstance.WatchSync(nil, syncChan)
			if err != nil {
				errChan <- err
				continue
			}

			buyAmount, err := o.GetBuyAmount()
			if err != nil {
				errChan <- err
				continue
			}

			priceChan <- buyAmount

		streamLoop:
			for {
				select {
				case <-syncChan:
					buyAmount, err := o.GetBuyAmount()
					if err != nil {
						errChan <- err
						break streamLoop
					}

					priceChan <- buyAmount
				case err := <-sub.Err():
					errChan <- err
					break streamLoop
				case <-done:
					return
				}
			}
		}
	}()

	return priceChan, errChan
}

// Calls get amounts out contract and returns the buy amount
func (o *Order) GetBuyAmount() (*big.Int, error) {
	amountsOut, err := o.router.Instance.GetAmountsOut(nil, o.GetSellToken().Amount, o.GetAddreses())
	if err != nil {
		return nil, err
	}

	return amountsOut[len(amountsOut)-1], nil
}

// Returns the corresponding transaction options
func (o *Order) getTransactOpts(c bind.ContractBackend, wallet *wallet.Wallet) (*bind.TransactOpts, error) {
	auth, err := bind.NewKeyStoreTransactorWithChainID(wallet.KeyStore, wallet.Account, o.chainID)
	if err != nil {
		return nil, err
	}

	auth.Value = big.NewInt(0)
	auth.GasLimit = o.GasLimit

	// Use error group to concurrently get nonce and recommended gas
	wg := new(errgroup.Group)

	//Get Nonce
	var nonce uint64
	wg.Go(func() error {
		var err error
		nonce, err = c.PendingNonceAt(context.Background(), wallet.GetAddress())
		if err != nil {
			return err
		}
		return nil
	})

	// Get recommended gas multiplied by percentage
	var recomGas *big.Int
	wg.Go(func() error {
		var err error
		recomGas, err = c.SuggestGasPrice(context.Background())
		if err != nil {
			return err
		}
		return nil
	})

	if err = wg.Wait(); err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasPrice = price.MultiplyPercent(recomGas, o.GasPriceMult)
	return auth, nil
}

// Returns pancakeswap pair contract instance for the tokens with the given addresses
func getPair(c bind.ContractBackend, r *router.Router, addr1, addr2 common.Address) (*pair.Pair, error) {
	factoryAddress, err := r.Factory(nil)
	if err != nil {
		return nil, err
	}

	factoryInstance, err := factory.NewFactory(factoryAddress, c)
	if err != nil {
		return nil, err
	}

	pairAddress, err := factoryInstance.GetPair(nil, addr1, addr2)
	if err != nil {
		return nil, err
	}

	return pair.NewPair(pairAddress, c)
}

// Returns the token from an address string
func getToken(address string, client bind.ContractBackend) (*Token, error) {
	token := &Token{
		Address: common.HexToAddress(address),
	}

	var err error
	token.Instance, err = bep20_token.NewBep20Token(token.Address, client)
	if err != nil {
		return nil, err
	}

	token.Decimals, err = token.Instance.Decimals(nil)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// Returns the trade from a Token
func getAmount(token *Token, amount string) (*big.Int, error) {
	floatAmount, ok := new(big.Float).SetString(amount)
	if !ok {
		return nil, ErrConv
	}

	return price.AddDecimals(floatAmount, token.Decimals), nil
}
