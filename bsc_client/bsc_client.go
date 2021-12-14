package bsc_client

import (
	"context"
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const tries int = 10

var (
	ErrTimeout       = errors.New("timed out of attempts")
	ErrReceiptStatus = errors.New("the receipt has a status of 0")
)

// An ethereum client that implements the ContractBackend interface
// https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractBackend
// and reconnects automatically on failure
type Client struct {
	client *ethclient.Client
}

// Dial connects to a client to the given URL
func Dial(rawurl string) (*Client, error) {
	c, err := ethclient.Dial(rawurl)
	if err != nil {
		return nil, err
	}

	return &Client{client: c}, nil
}

// Implements the ContractCaller interface https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractCaller
func (c *Client) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		data, err := c.client.CodeAt(ctx, contract, blockNumber)
		if err == nil {
			return data, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return nil, ErrTimeout
}

// Implements the ContractCaller interface https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractCaller
func (c *Client) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		data, err := c.client.CallContract(ctx, call, blockNumber)
		if err == nil {
			return data, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return nil, ErrTimeout
}

// Implements the ContractFilterer interface https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractFilterer
func (c *Client) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		logs, err := c.client.FilterLogs(ctx, query)
		if err == nil {
			return logs, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return nil, ErrTimeout
}

// Implements the ContractFilterer interface https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractFilterer
func (c *Client) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		sub, err := c.client.SubscribeFilterLogs(ctx, query, ch)
		if err == nil {
			return sub, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return nil, ErrTimeout
}

// Implements the ContractTransactor interface https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractTransactor
func (c *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		header, err := c.client.HeaderByNumber(ctx, number)
		if err == nil {
			return header, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return nil, ErrTimeout
}

// Implements the ContractTransactor interface https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractTransactor
func (c *Client) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		data, err := c.client.PendingCodeAt(ctx, account)
		if err == nil {
			return data, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return nil, ErrTimeout
}

// Implements the ContractTransactor interface https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractTransactor
func (c *Client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		nonce, err := c.client.PendingNonceAt(ctx, account)
		if err == nil {
			return nonce, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return 0, ErrTimeout
}

// Implements the ContractTransactor interface https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractTransactor
func (c *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		gasPrice, err := c.client.SuggestGasPrice(ctx)
		if err == nil {
			return gasPrice, err
		}
		log.Println(err)
		triesRemaining--
	}
	return nil, ErrTimeout
}

// Implements the ContractTransactor interface https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractTransactor
func (c *Client) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		gasTipCap, err := c.client.SuggestGasTipCap(ctx)
		if err == nil {
			return gasTipCap, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return nil, ErrTimeout
}

// Implements the ContractTransactor interface https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractTransactor
func (c *Client) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		gas, err := c.client.EstimateGas(ctx, call)
		if err == nil {
			return gas, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return 0, ErrTimeout
}

// Implements the ContractTransactor interface https://pkg.go.dev/github.com/ethereum/go-ethereum@v1.10.12/accounts/abi/bind#ContractTransactor
func (c *Client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	triesRemaining := tries
	for triesRemaining > 0 {
		err := c.client.SendTransaction(ctx, tx)
		if err == nil {
			return nil
		}
		log.Println(err)
		triesRemaining--
	}
	return ErrTimeout
}

func (c *Client) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		sub, err := c.client.SubscribeNewHead(ctx, ch)
		if err == nil {
			return sub, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return nil, ErrTimeout
}

func (c *Client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		tx, isPending, err := c.client.TransactionByHash(ctx, hash)
		if err == nil {
			return tx, isPending, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return nil, false, ErrTimeout
}

func (c *Client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	triesRemaining := tries
	for triesRemaining > 0 {
		receipt, err := c.client.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}
		log.Println(err)
		triesRemaining--
	}
	return nil, ErrTimeout
}

// Blocks until the transaction is mined and returns it's receipt.
// It returns an error if the receipt status is 0
func (c *Client) GetPendingTxReceipt(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	headerChan := make(chan *types.Header)
	sub, err := c.SubscribeNewHead(ctx, headerChan)
	if err != nil {
		return nil, err
	}

	triesRemaining := tries
	for triesRemaining > 0 {
		select {
		// When a new block is mined
		case <-headerChan:
			_, isPending, err := c.TransactionByHash(ctx, tx.Hash())
			if err != nil {
				return nil, err
			}

			if isPending {
				continue
			}

			receipt, err := c.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				return nil, err
			}

			if receipt.Status == 0 {
				return nil, ErrReceiptStatus
			}

			return receipt, nil
		case err = <-sub.Err():
			log.Println(err)
			triesRemaining--
		}
	}

	return nil, err
}
