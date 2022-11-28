package flashxroute

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

// RpcError - ethereum error
type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err RpcError) Error() string {
	return fmt.Sprintf("Error %d (%s)", err.Code, err.Message)
}

type rpcResponse struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *RpcError       `json:"error"`
}

type rpcRequest struct {
	ID      int         `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// FlashXRoute - Ethereum rpc client
type FlashXRoute struct {
	url     string
	client  httpClient
	log     logger
	Debug   bool
	Headers map[string]string // Additional headers to send with the request
	Timeout time.Duration
}

// New create new rpc client with given url
func New(url string, options ...func(rpc *FlashXRoute)) *FlashXRoute {
	rpc := &FlashXRoute{
		url:     url,
		client:  http.DefaultClient,
		log:     log.New(os.Stderr, "", log.LstdFlags),
		Headers: make(map[string]string),
		Timeout: 30 * time.Second,
	}
	for _, option := range options {
		option(rpc)
	}

	return rpc
}

// NewFlashXRoute create new rpc client with given url
func NewFlashXRoute(url string, options ...func(rpc *FlashXRoute)) *FlashXRoute {
	return New(url, options...)
}

func (rpc *FlashXRoute) call(method string, target interface{}, params ...interface{}) error {
	result, err := rpc.Call(method, params...)
	if err != nil {
		return err
	}

	if target == nil {
		return nil
	}

	return json.Unmarshal(result, target)
}

// URL returns client url
func (rpc *FlashXRoute) URL() string {
	return rpc.url
}

// Call returns raw response of method call
func (rpc *FlashXRoute) Call(method string, params ...interface{}) (json.RawMessage, error) {
	request := rpcRequest{
		ID:      1,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params[0],
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", rpc.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	for k, v := range rpc.Headers {
		req.Header.Add(k, v)
	}
	httpClient := &http.Client{
		Timeout: rpc.Timeout,
	}

	response, err := httpClient.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if rpc.Debug {
		rpc.log.Println(fmt.Sprintf("%s\nRequest: %s\nResponse: %s\n", method, body, data))
	}

	resp := new(rpcResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, *resp.Error
	}

	return resp.Result, nil
}

// CallWithBloxrouteAuthHeader is like Call but also signs the request
func (rpc *FlashXRoute) CallWithBloxrouteAuthHeader(method string, authHeader string, params ...interface{}) (json.RawMessage, error) {
	request := rpcRequest{
		ID:      1,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params[0],
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", rpc.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", authHeader)
	for k, v := range rpc.Headers {
		req.Header.Add(k, v)
	}
	httpClient := &http.Client{
		Timeout: rpc.Timeout,
	}

	response, err := httpClient.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if rpc.Debug {
		rpc.log.Println(fmt.Sprintf("%s\nRequest: %s\nAuthHeader: %s\nResponse: %s\n", method, body, authHeader, data))
	}

	// On error, response looks like this instead of JSON-RPC: {"error":"block param must be a hex int"}
	errorResp := new(RelayErrorResponse)
	if err := json.Unmarshal(data, errorResp); err == nil && errorResp.Error != "" {
		// relay returned an error
		return nil, fmt.Errorf("%w: %s", ErrRelayErrorResponse, errorResp.Error)
	}

	resp := new(rpcResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("%w: %s", ErrRelayErrorResponse, (*resp).Error.Message)
	}

	return resp.Result, nil
}

// RawCall returns raw response of method call (Deprecated)
func (rpc *FlashXRoute) RawCall(method string, params ...interface{}) (json.RawMessage, error) {
	return rpc.Call(method, params...)
}

// Web3ClientVersion returns the current client version.
func (rpc *FlashXRoute) Web3ClientVersion() (string, error) {
	var clientVersion string

	err := rpc.call("web3_clientVersion", &clientVersion)
	return clientVersion, err
}

// Web3Sha3 returns Keccak-256 (not the standardized SHA3-256) of the given data.
func (rpc *FlashXRoute) Web3Sha3(data []byte) (string, error) {
	var hash string

	err := rpc.call("web3_sha3", &hash, fmt.Sprintf("0x%x", data))
	return hash, err
}

// NetVersion returns the current network protocol version.
func (rpc *FlashXRoute) NetVersion() (string, error) {
	var version string

	err := rpc.call("net_version", &version)
	return version, err
}

// NetListening returns true if client is actively listening for network connections.
func (rpc *FlashXRoute) NetListening() (bool, error) {
	var listening bool

	err := rpc.call("net_listening", &listening)
	return listening, err
}

// NetPeerCount returns number of peers currently connected to the client.
func (rpc *FlashXRoute) NetPeerCount() (int, error) {
	var response string
	if err := rpc.call("net_peerCount", &response); err != nil {
		return 0, err
	}

	return ParseInt(response)
}

// EthProtocolVersion returns the current ethereum protocol version.
func (rpc *FlashXRoute) EthProtocolVersion() (string, error) {
	var protocolVersion string

	err := rpc.call("eth_protocolVersion", &protocolVersion)
	return protocolVersion, err
}

// EthSyncing returns an object with data about the sync status or false.
func (rpc *FlashXRoute) EthSyncing() (*Syncing, error) {
	result, err := rpc.RawCall("eth_syncing")
	if err != nil {
		return nil, err
	}
	syncing := new(Syncing)
	if bytes.Equal(result, []byte("false")) {
		return syncing, nil
	}
	err = json.Unmarshal(result, syncing)
	return syncing, err
}

// EthCoinbase returns the client coinbase address
func (rpc *FlashXRoute) EthCoinbase() (string, error) {
	var address string

	err := rpc.call("eth_coinbase", &address)
	return address, err
}

// EthMining returns true if client is actively mining new blocks.
func (rpc *FlashXRoute) EthMining() (bool, error) {
	var mining bool

	err := rpc.call("eth_mining", &mining)
	return mining, err
}

// EthHashrate returns the number of hashes per second that the node is mining with.
func (rpc *FlashXRoute) EthHashrate() (int, error) {
	var response string

	if err := rpc.call("eth_hashrate", &response); err != nil {
		return 0, err
	}

	return ParseInt(response)
}

// EthGasPrice returns the current price per gas in wei.
func (rpc *FlashXRoute) EthGasPrice() (big.Int, error) {
	var response string
	if err := rpc.call("eth_gasPrice", &response); err != nil {
		return big.Int{}, err
	}

	return ParseBigInt(response)
}

// EthAccounts returns a list of addresses owned by client.
func (rpc *FlashXRoute) EthAccounts() ([]string, error) {
	accounts := []string{}

	err := rpc.call("eth_accounts", &accounts)
	return accounts, err
}

// EthBlockNumber returns the number of most recent block.
func (rpc *FlashXRoute) EthBlockNumber() (int, error) {
	var response string
	if err := rpc.call("eth_blockNumber", &response); err != nil {
		return 0, err
	}

	return ParseInt(response)
}

// EthGetBalance returns the balance of the account of given address in wei.
func (rpc *FlashXRoute) EthGetBalance(address, block string) (big.Int, error) {
	var response string
	if err := rpc.call("eth_getBalance", &response, address, block); err != nil {
		return big.Int{}, err
	}

	return ParseBigInt(response)
}

// EthGetStorageAt returns the value from a storage position at a given address.
func (rpc *FlashXRoute) EthGetStorageAt(data string, position int, tag string) (string, error) {
	var result string

	err := rpc.call("eth_getStorageAt", &result, data, IntToHex(position), tag)
	return result, err
}

// EthGetTransactionCount returns the number of transactions sent from an address.
func (rpc *FlashXRoute) EthGetTransactionCount(address, block string) (int, error) {
	var response string

	if err := rpc.call("eth_getTransactionCount", &response, address, block); err != nil {
		return 0, err
	}

	return ParseInt(response)
}

// EthGetBlockTransactionCountByHash returns the number of transactions in a block from a block matching the given block hash.
func (rpc *FlashXRoute) EthGetBlockTransactionCountByHash(hash string) (int, error) {
	var response string

	if err := rpc.call("eth_getBlockTransactionCountByHash", &response, hash); err != nil {
		return 0, err
	}

	return ParseInt(response)
}

// EthGetBlockTransactionCountByNumber returns the number of transactions in a block from a block matching the given block
func (rpc *FlashXRoute) EthGetBlockTransactionCountByNumber(number int) (int, error) {
	var response string

	if err := rpc.call("eth_getBlockTransactionCountByNumber", &response, IntToHex(number)); err != nil {
		return 0, err
	}

	return ParseInt(response)
}

// EthGetUncleCountByBlockHash returns the number of uncles in a block from a block matching the given block hash.
func (rpc *FlashXRoute) EthGetUncleCountByBlockHash(hash string) (int, error) {
	var response string

	if err := rpc.call("eth_getUncleCountByBlockHash", &response, hash); err != nil {
		return 0, err
	}

	return ParseInt(response)
}

// EthGetUncleCountByBlockNumber returns the number of uncles in a block from a block matching the given block number.
func (rpc *FlashXRoute) EthGetUncleCountByBlockNumber(number int) (int, error) {
	var response string

	if err := rpc.call("eth_getUncleCountByBlockNumber", &response, IntToHex(number)); err != nil {
		return 0, err
	}

	return ParseInt(response)
}

// EthGetCode returns code at a given address.
func (rpc *FlashXRoute) EthGetCode(address, block string) (string, error) {
	var code string

	err := rpc.call("eth_getCode", &code, address, block)
	return code, err
}

// EthSign signs data with a given address.
// Calculates an Ethereum specific signature with: sign(keccak256("\x19Ethereum Signed Message:\n" + len(message) + message)))
func (rpc *FlashXRoute) EthSign(address, data string) (string, error) {
	var signature string

	err := rpc.call("eth_sign", &signature, address, data)
	return signature, err
}

// EthSendTransaction creates new message call transaction or a contract creation, if the data field contains code.
func (rpc *FlashXRoute) EthSendTransaction(transaction T) (string, error) {
	var hash string

	err := rpc.call("eth_sendTransaction", &hash, transaction)
	return hash, err
}

// EthSendRawTransaction creates new message call transaction or a contract creation for signed transactions.
func (rpc *FlashXRoute) EthSendRawTransaction(data string) (string, error) {
	var hash string

	err := rpc.call("eth_sendRawTransaction", &hash, data)
	return hash, err
}

// EthCall executes a new message call immediately without creating a transaction on the block chain.
func (rpc *FlashXRoute) EthCall(transaction T, tag string) (string, error) {
	var data string

	err := rpc.call("eth_call", &data, transaction, tag)
	return data, err
}

// EthEstimateGas makes a call or transaction, which won't be added to the blockchain and returns the used gas, which can be used for estimating the used gas.
func (rpc *FlashXRoute) EthEstimateGas(transaction T) (int, error) {
	var response string

	err := rpc.call("eth_estimateGas", &response, transaction)
	if err != nil {
		return 0, err
	}

	return ParseInt(response)
}

func (rpc *FlashXRoute) getBlock(method string, withTransactions bool, params ...interface{}) (*Block, error) {
	result, err := rpc.RawCall(method, params...)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(result, []byte("null")) {
		return nil, nil
	}

	var response proxyBlock
	if withTransactions {
		response = new(proxyBlockWithTransactions)
	} else {
		response = new(proxyBlockWithoutTransactions)
	}

	err = json.Unmarshal(result, response)
	if err != nil {
		return nil, err
	}

	block := response.toBlock()
	return &block, nil
}

// EthGetBlockByHash returns information about a block by hash.
func (rpc *FlashXRoute) EthGetBlockByHash(hash string, withTransactions bool) (*Block, error) {
	return rpc.getBlock("eth_getBlockByHash", withTransactions, hash, withTransactions)
}

// EthGetBlockByNumber returns information about a block by block number.
func (rpc *FlashXRoute) EthGetBlockByNumber(number int, withTransactions bool) (*Block, error) {
	return rpc.getBlock("eth_getBlockByNumber", withTransactions, IntToHex(number), withTransactions)
}

func (rpc *FlashXRoute) getTransaction(method string, params ...interface{}) (*Transaction, error) {
	transaction := new(Transaction)

	err := rpc.call(method, transaction, params...)
	return transaction, err
}

// EthGetTransactionByHash returns the information about a transaction requested by transaction hash.
func (rpc *FlashXRoute) EthGetTransactionByHash(hash string) (*Transaction, error) {
	return rpc.getTransaction("eth_getTransactionByHash", hash)
}

// EthGetTransactionByBlockHashAndIndex returns information about a transaction by block hash and transaction index position.
func (rpc *FlashXRoute) EthGetTransactionByBlockHashAndIndex(blockHash string, transactionIndex int) (*Transaction, error) {
	return rpc.getTransaction("eth_getTransactionByBlockHashAndIndex", blockHash, IntToHex(transactionIndex))
}

// EthGetTransactionByBlockNumberAndIndex returns information about a transaction by block number and transaction index position.
func (rpc *FlashXRoute) EthGetTransactionByBlockNumberAndIndex(blockNumber, transactionIndex int) (*Transaction, error) {
	return rpc.getTransaction("eth_getTransactionByBlockNumberAndIndex", IntToHex(blockNumber), IntToHex(transactionIndex))
}

// EthGetTransactionReceipt returns the receipt of a transaction by transaction hash.
// Note That the receipt is not available for pending transactions.
func (rpc *FlashXRoute) EthGetTransactionReceipt(hash string) (*TransactionReceipt, error) {
	transactionReceipt := new(TransactionReceipt)

	err := rpc.call("eth_getTransactionReceipt", transactionReceipt, hash)
	if err != nil {
		return nil, err
	}

	return transactionReceipt, nil
}

// EthGetCompilers returns a list of available compilers in the client.
func (rpc *FlashXRoute) EthGetCompilers() ([]string, error) {
	compilers := []string{}

	err := rpc.call("eth_getCompilers", &compilers)
	return compilers, err
}

// EthNewFilter creates a new filter object.
func (rpc *FlashXRoute) EthNewFilter(params FilterParams) (string, error) {
	var filterID string
	err := rpc.call("eth_newFilter", &filterID, params)
	return filterID, err
}

// EthNewBlockFilter creates a filter in the node, to notify when a new block arrives.
// To check if the state has changed, call EthGetFilterChanges.
func (rpc *FlashXRoute) EthNewBlockFilter() (string, error) {
	var filterID string
	err := rpc.call("eth_newBlockFilter", &filterID)
	return filterID, err
}

// EthNewPendingTransactionFilter creates a filter in the node, to notify when new pending transactions arrive.
// To check if the state has changed, call EthGetFilterChanges.
func (rpc *FlashXRoute) EthNewPendingTransactionFilter() (string, error) {
	var filterID string
	err := rpc.call("eth_newPendingTransactionFilter", &filterID)
	return filterID, err
}

// EthUninstallFilter uninstalls a filter with given id.
func (rpc *FlashXRoute) EthUninstallFilter(filterID string) (bool, error) {
	var res bool
	err := rpc.call("eth_uninstallFilter", &res, filterID)
	return res, err
}

// EthGetFilterChanges polling method for a filter, which returns an array of logs which occurred since last poll.
func (rpc *FlashXRoute) EthGetFilterChanges(filterID string) ([]Log, error) {
	var logs = []Log{}
	err := rpc.call("eth_getFilterChanges", &logs, filterID)
	return logs, err
}

// EthGetFilterLogs returns an array of all logs matching filter with given id.
func (rpc *FlashXRoute) EthGetFilterLogs(filterID string) ([]Log, error) {
	var logs = []Log{}
	err := rpc.call("eth_getFilterLogs", &logs, filterID)
	return logs, err
}

// EthGetLogs returns an array of all logs matching a given filter object.
func (rpc *FlashXRoute) EthGetLogs(params FilterParams) ([]Log, error) {
	var logs = []Log{}
	err := rpc.call("eth_getLogs", &logs, params)
	return logs, err
}

// Eth1 returns 1 ethereum value (10^18 wei)
func (rpc *FlashXRoute) Eth1() *big.Int {
	return Eth1()
}

// Eth1 returns 1 ethereum value (10^18 wei)
func Eth1() *big.Int {
	return big.NewInt(1000000000000000000)
}

// https://docs.bloxroute.com/apis/mev-solution/bundle-simulation
func (rpc *FlashXRoute) BloxrouteSimulateBundle(authHeader string, params BloxrouteSimulateBundleRequest) (res BloxrouteSimulateBundleResponse, err error) {
	rawMsg, err := rpc.CallWithBloxrouteAuthHeader("blxr_simulate_bundle", authHeader, params)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(rawMsg, &res)
	return res, err
}


// https://docs.bloxroute.com/apis/mev-solution/arb-only-bundle-simulation
func (rpc *FlashXRoute) BloxrouteBrmSimulateBundle(authHeader string, params BloxrouteSimulateBundleRequest) (res BloxrouteSimulateBundleResponse, err error) {
	rawMsg, err := rpc.CallWithBloxrouteAuthHeader("simulate_arb_only_bundle", authHeader, params)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(rawMsg, &res)
	return res, err
}

// https://docs.bloxroute.com/apis/mev-solution/bundle-submission
func (rpc *FlashXRoute) BloxrouteSubmitBundle(authHeader string, params BloxrouteSubmitBundleRequest) (res BloxrouteSubmitBundleResponse, err error) {
	rawMsg, err := rpc.CallWithBloxrouteAuthHeader("blxr_submit_bundle", authHeader, params)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(rawMsg, &res)
	return res, err
}

// https://docs.bloxroute.com/apis/mev-solution/arb-only-bundle-submission
func (rpc *FlashXRoute) BloxrouteBrmSubmitBundle(authHeader string, params BloxrouteSubmitBundleRequest) (res BloxrouteSubmitBundleResponse, err error) {
	rawMsg, err := rpc.CallWithBloxrouteAuthHeader("submit_arb_only_bundle", authHeader, params)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(rawMsg, &res)
	return res, err
}

// Simulate a full Ethereum block. numTx is the maximum number of tx to include, used for troubleshooting (default: 0 - all transactions)
func (rpc *FlashXRoute) BloxrouteSimulateBlock(authHeader string, block *types.Block, maxTx int) (res BloxrouteSimulateBundleResponse, err error) {
	if rpc.Debug {
		fmt.Printf("Simulating block %s 0x%x %s \t %d tx \t timestamp: %d\n", block.Number(), block.Number(), block.Header().Hash(), len(block.Transactions()), block.Header().Time)
	}

	txs := make([]string, 0)
	for _, tx := range block.Transactions() {
		// fmt.Println("tx", i, tx.Hash(), "type", tx.Type())
		from, fromErr := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		txIsFromCoinbase := fromErr == nil && from == block.Coinbase()
		if txIsFromCoinbase {
			if rpc.Debug {
				fmt.Printf("- skip tx from coinbase: %s\n", tx.Hash())
			}
			continue
		}

		to := tx.To()
		txIsToCoinbase := to != nil && *to == block.Coinbase()
		if txIsToCoinbase {
			if rpc.Debug {
				fmt.Printf("- skip tx to coinbase: %s\n", tx.Hash())
			}
			continue
		}

		rlp := TxToRlp(tx)

		// Might need to strip beginning bytes
		if rlp[:2] == "b9" {
			rlp = rlp[6:]
		} else if rlp[:2] == "b8" {
			rlp = rlp[4:]
		}

		// callBundle expects a 0x prefix
		rlp = "0x" + rlp
		txs = append(txs, rlp)

		if maxTx > 0 && len(txs) == maxTx {
			break
		}
	}

	if rpc.Debug {
		fmt.Printf("sending %d tx for simulation to %s...\n", len(txs), rpc.url)
	}

	params := BloxrouteSimulateBundleRequest{
		Transaction:      txs,
		BlockNumber:      fmt.Sprintf("0x%x", block.Number()),
		StateBlockNumber: block.ParentHash().Hex(),
	}

	res, err = rpc.BloxrouteSimulateBundle(authHeader, params)
	return res, err
}

// This endpoint allows you to send a single transaction that will be distributed faster using the BDN.
func (rpc *FlashXRoute) BloxrouteSendTransaction(authHeader string, params BloxrouteSendTransactionRequest) (txHash string, err error) {
	rawMsg, err := rpc.CallWithBloxrouteAuthHeader("blxr_tx", authHeader, params)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(rawMsg, &txHash)
	return txHash, err
}
