package flashxroute

import (
	"bytes"
	"encoding/json"
	"math/big"
	"unsafe"

	"github.com/pkg/errors"
)

// ErrRelayErrorResponse means it's a standard Flashbots relay error response - probably a user error rather than JSON or network error
var ErrRelayErrorResponse = errors.New("relay error response")

// Syncing - object with syncing data info
type Syncing struct {
	IsSyncing     bool
	StartingBlock int
	CurrentBlock  int
	HighestBlock  int
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Syncing) UnmarshalJSON(data []byte) error {
	proxy := new(proxySyncing)
	if err := json.Unmarshal(data, proxy); err != nil {
		return err
	}

	proxy.IsSyncing = true
	*s = *(*Syncing)(unsafe.Pointer(proxy))

	return nil
}

// T - input transaction object
type T struct {
	From     string
	To       string
	Gas      int
	GasPrice *big.Int
	Value    *big.Int
	Data     string
	Nonce    int
}

// MarshalJSON implements the json.Unmarshaler interface.
func (t T) MarshalJSON() ([]byte, error) {
	params := map[string]interface{}{
		"from": t.From,
	}
	if t.To != "" {
		params["to"] = t.To
	}
	if t.Gas > 0 {
		params["gas"] = IntToHex(t.Gas)
	}
	if t.GasPrice != nil {
		params["gasPrice"] = BigToHex(*t.GasPrice)
	}
	if t.Value != nil {
		params["value"] = BigToHex(*t.Value)
	}
	if t.Data != "" {
		params["data"] = t.Data
	}
	if t.Nonce > 0 {
		params["nonce"] = IntToHex(t.Nonce)
	}

	return json.Marshal(params)
}

// Transaction - transaction object
type Transaction struct {
	Hash             string
	Nonce            int
	BlockHash        string
	BlockNumber      *int
	TransactionIndex *int
	From             string
	To               string
	Value            big.Int
	Gas              int
	GasPrice         big.Int
	Input            string
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Transaction) UnmarshalJSON(data []byte) error {
	proxy := new(proxyTransaction)
	if err := json.Unmarshal(data, proxy); err != nil {
		return err
	}

	*t = *(*Transaction)(unsafe.Pointer(proxy))

	return nil
}

// Log - log object
type Log struct {
	Removed          bool
	LogIndex         int
	TransactionIndex int
	TransactionHash  string
	BlockNumber      int
	BlockHash        string
	Address          string
	Data             string
	Topics           []string
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (log *Log) UnmarshalJSON(data []byte) error {
	proxy := new(proxyLog)
	if err := json.Unmarshal(data, proxy); err != nil {
		return err
	}

	*log = *(*Log)(unsafe.Pointer(proxy))

	return nil
}

// FilterParams - Filter parameters object
type FilterParams struct {
	FromBlock string     `json:"fromBlock,omitempty"`
	ToBlock   string     `json:"toBlock,omitempty"`
	Address   []string   `json:"address,omitempty"`
	Topics    [][]string `json:"topics,omitempty"`
}

// TransactionReceipt - transaction receipt object
type TransactionReceipt struct {
	TransactionHash   string
	TransactionIndex  int
	BlockHash         string
	BlockNumber       int
	CumulativeGasUsed int
	GasUsed           int
	ContractAddress   string
	Logs              []Log
	LogsBloom         string
	Root              string
	Status            string
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *TransactionReceipt) UnmarshalJSON(data []byte) error {
	proxy := new(proxyTransactionReceipt)
	if err := json.Unmarshal(data, proxy); err != nil {
		return err
	}

	*t = *(*TransactionReceipt)(unsafe.Pointer(proxy))

	return nil
}

// Block - block object
type Block struct {
	Number           int
	Hash             string
	ParentHash       string
	Nonce            string
	Sha3Uncles       string
	LogsBloom        string
	TransactionsRoot string
	StateRoot        string
	Miner            string
	Difficulty       big.Int
	TotalDifficulty  big.Int
	ExtraData        string
	Size             int
	GasLimit         int
	GasUsed          int
	Timestamp        int
	Uncles           []string
	Transactions     []Transaction
}

type proxySyncing struct {
	IsSyncing     bool   `json:"-"`
	StartingBlock hexInt `json:"startingBlock"`
	CurrentBlock  hexInt `json:"currentBlock"`
	HighestBlock  hexInt `json:"highestBlock"`
}

type proxyTransaction struct {
	Hash             string  `json:"hash"`
	Nonce            hexInt  `json:"nonce"`
	BlockHash        string  `json:"blockHash"`
	BlockNumber      *hexInt `json:"blockNumber"`
	TransactionIndex *hexInt `json:"transactionIndex"`
	From             string  `json:"from"`
	To               string  `json:"to"`
	Value            hexBig  `json:"value"`
	Gas              hexInt  `json:"gas"`
	GasPrice         hexBig  `json:"gasPrice"`
	Input            string  `json:"input"`
}

type proxyLog struct {
	Removed          bool     `json:"removed"`
	LogIndex         hexInt   `json:"logIndex"`
	TransactionIndex hexInt   `json:"transactionIndex"`
	TransactionHash  string   `json:"transactionHash"`
	BlockNumber      hexInt   `json:"blockNumber"`
	BlockHash        string   `json:"blockHash"`
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	Topics           []string `json:"topics"`
}

type proxyTransactionReceipt struct {
	TransactionHash   string `json:"transactionHash"`
	TransactionIndex  hexInt `json:"transactionIndex"`
	BlockHash         string `json:"blockHash"`
	BlockNumber       hexInt `json:"blockNumber"`
	CumulativeGasUsed hexInt `json:"cumulativeGasUsed"`
	GasUsed           hexInt `json:"gasUsed"`
	ContractAddress   string `json:"contractAddress,omitempty"`
	Logs              []Log  `json:"logs"`
	LogsBloom         string `json:"logsBloom"`
	Root              string `json:"root"`
	Status            string `json:"status,omitempty"`
}

type hexInt int

func (i *hexInt) UnmarshalJSON(data []byte) error {
	result, err := ParseInt(string(bytes.Trim(data, `"`)))
	*i = hexInt(result)

	return err
}

type hexBig big.Int

func (i *hexBig) UnmarshalJSON(data []byte) error {
	result, err := ParseBigInt(string(bytes.Trim(data, `"`)))
	*i = hexBig(result)

	return err
}

type proxyBlock interface {
	toBlock() Block
}

type proxyBlockWithTransactions struct {
	Number           hexInt             `json:"number"`
	Hash             string             `json:"hash"`
	ParentHash       string             `json:"parentHash"`
	Nonce            string             `json:"nonce"`
	Sha3Uncles       string             `json:"sha3Uncles"`
	LogsBloom        string             `json:"logsBloom"`
	TransactionsRoot string             `json:"transactionsRoot"`
	StateRoot        string             `json:"stateRoot"`
	Miner            string             `json:"miner"`
	Difficulty       hexBig             `json:"difficulty"`
	TotalDifficulty  hexBig             `json:"totalDifficulty"`
	ExtraData        string             `json:"extraData"`
	Size             hexInt             `json:"size"`
	GasLimit         hexInt             `json:"gasLimit"`
	GasUsed          hexInt             `json:"gasUsed"`
	Timestamp        hexInt             `json:"timestamp"`
	Uncles           []string           `json:"uncles"`
	Transactions     []proxyTransaction `json:"transactions"`
}

func (proxy *proxyBlockWithTransactions) toBlock() Block {
	return *(*Block)(unsafe.Pointer(proxy))
}

type proxyBlockWithoutTransactions struct {
	Number           hexInt   `json:"number"`
	Hash             string   `json:"hash"`
	ParentHash       string   `json:"parentHash"`
	Nonce            string   `json:"nonce"`
	Sha3Uncles       string   `json:"sha3Uncles"`
	LogsBloom        string   `json:"logsBloom"`
	TransactionsRoot string   `json:"transactionsRoot"`
	StateRoot        string   `json:"stateRoot"`
	Miner            string   `json:"miner"`
	Difficulty       hexBig   `json:"difficulty"`
	TotalDifficulty  hexBig   `json:"totalDifficulty"`
	ExtraData        string   `json:"extraData"`
	Size             hexInt   `json:"size"`
	GasLimit         hexInt   `json:"gasLimit"`
	GasUsed          hexInt   `json:"gasUsed"`
	Timestamp        hexInt   `json:"timestamp"`
	Uncles           []string `json:"uncles"`
	Transactions     []string `json:"transactions"`
}

func (proxy *proxyBlockWithoutTransactions) toBlock() Block {
	block := Block{
		Number:           int(proxy.Number),
		Hash:             proxy.Hash,
		ParentHash:       proxy.ParentHash,
		Nonce:            proxy.Nonce,
		Sha3Uncles:       proxy.Sha3Uncles,
		LogsBloom:        proxy.LogsBloom,
		TransactionsRoot: proxy.TransactionsRoot,
		StateRoot:        proxy.StateRoot,
		Miner:            proxy.Miner,
		Difficulty:       big.Int(proxy.Difficulty),
		TotalDifficulty:  big.Int(proxy.TotalDifficulty),
		ExtraData:        proxy.ExtraData,
		Size:             int(proxy.Size),
		GasLimit:         int(proxy.GasLimit),
		GasUsed:          int(proxy.GasUsed),
		Timestamp:        int(proxy.Timestamp),
		Uncles:           proxy.Uncles,
	}

	block.Transactions = make([]Transaction, len(proxy.Transactions))
	for i := range proxy.Transactions {
		block.Transactions[i] = Transaction{
			Hash: proxy.Transactions[i],
		}
	}

	return block
}

type RelayErrorResponse struct {
	Error string `json:"error"`
}

type BloxrouteSimulateBundleRequest struct {
	Transaction    []string `json:"transaction"`                  // A list of raw transaction bytes without a 0x prefix.
	BlockNumber      string `json:"block_number"`                 // Block number of a future block to include this bundle in, in hex value.
	StateBlockNumber string `json:"state_block_number,omitempty"` /* [Optional] Block number used as the base state to run a simulation on.
	                                                                 Valid inputs include hex value of block number, or tags like “latest” and “pending”.
                                                                         Default value is “latest”. */
	Timestamp        int64  `json:"timestamp,omitempty"`          // [Optional] Simulation timestamp, an integer in unix epoch format. Default value is None.
}

type BloxrouteBrmSimulateBundleRequest struct {
	TransactionHash  string   `json:"transaction_hash"`             // Trigger private transaction hash.
	Transaction      []string `json:"transaction"`                  // A list of raw transaction bytes without a 0x prefix.
	BlockNumber      string   `json:"block_number"`                 // Block number of a future block to include this bundle in, in hex value.
	StateBlockNumber string   `json:"state_block_number,omitempty"` /* [Optional] Block number used as the base state to run a simulation on.
	                                                                   Valid inputs include hex value of block number, or tags like “latest” and “pending”.
                                                                           Default value is “latest”. */
	Timestamp        int64  `json:"timestamp,omitempty"`            // [Optional] Simulation timestamp, an integer in unix epoch format. Default value is None.
}

type BloxrouteSimulateBundleResult struct {
	GasUsed           int64  `json:"gasUsed"`           // 63197,
	TxHash            string `json:"txHash"`            // "0xe2df005210bdc204a34ff03211606e5d8036740c686e9fe4e266ae91cf4d12df",
	Value             string `json:"value"`             // "0x"
	Error             string `json:"error"`
}

type BloxrouteSimulateBundleResponse struct {
	BundleGasPrice    string                          `json:"bundleGasPrice"`    // "43000001459",
	BundleHash        string                          `json:"bundleHash"`        // "0x2ca9c4d2ba00d8144d8e396a4989374443cb20fb490d800f4f883ad4e1b32158",
	CoinbaseDiff      string                          `json:"coinbaseDiff"`      // "2717471092204423",
	EthSentToCoinbase string                          `json:"ethSentToCoinbase"` // "0",
	GasFees           string                          `json:"gasFees"`           // "2717471092204423",
	Results           []BloxrouteSimulateBundleResult `json:"results"`           // [],
	StateBlockNumber  int64                           `json:"stateBlockNumber"`  // 12960319,
	TotalGasUsed      int64                           `json:"totalGasUsed"`      // 63197
}

type BloxrouteBrmSimulateBundleResponse struct {
	BloxrouteDiff     string                          `json:"bloxrouteDiff"`     // "10000"
	BundleGasPrice    string                          `json:"bundleGasPrice"`    // "43000001459",
	BundleHash        string                          `json:"bundleHash"`        // "0x2ca9c4d2ba00d8144d8e396a4989374443cb20fb490d800f4f883ad4e1b32158",
	CoinbaseDiff      string                          `json:"coinbaseDiff"`      // "2717471092204423",
	EthSentToCoinbase string                          `json:"ethSentToCoinbase"` // "0",
	GasFees           string                          `json:"gasFees"`           // "2717471092204423",
	MinerDiff         string                          `json:"minerDiff"`         // "100000"
	Results           []BloxrouteSimulateBundleResult `json:"results"`           // [],
	SenderDiff        string                          `json:"senderDiff"`        // "50000"
	StateBlockNumber  int64                           `json:"stateBlockNumber"`  // 12960319,
	TotalGasUsed      int64                           `json:"totalGasUsed"`      // 63197
	Status            string                          `json:"status"`            // "good"
}

// SubmitBundle
type BloxrouteSubmitBundleRequest struct {
	Transaction  []string     `json:"transaction"`                   // A list of raw transaction bytes without a 0x prefix.
	BlockNumber  string       `json:"block_number"`                  /* Block number of a future block to include this bundle in, in hex value.
                                                                            For traders who would like more than one block to be targeted, please send multiple requests targeting each specific block. */
	MinTimestamp *uint64      `json:"min_timestamp,omitempty"`       // [Optional] The minimum timestamp that the bundle is valid on, an integer in unix epoch format. Default value is None.
	MaxTimestamp *uint64      `json:"max_timestamp,omitempty"`       // [Optional] The maximum timestamp that the bundle is valid on, an integer in unix epoch format. Default value is None.
	RevertingHashes *[]string `json:"reverting_hashes,omitempty"` /* [Optional] A list of transaction hashes within the bundle that are allowed to revert.
                                                                           Default is empty list: the whole bundle would be excluded if any transaction reverts. */
	Uuid         string       `json:"uuid,omitempty"`                /* [Optional] A unique identifier of the bundle. This field can be used for bundle replacement and bundle cancellation.
                                                                            Some builders like bloxroute and builder0x69 support this field. After receiving a new UUID bundle,
                                                                            the builder would replace the previous bundle that has the same UUID. When the list of transactions is empty in new UUID bundle,
                                                                            the previous bundle associated with the same UUID would be effectively canceled.
                                                                            The response is empty/null instead of bundle hash when UUID is provided in the request. */
	Frontrunning bool         `json:"frontrunning,omitempty"`        /* [Optional, default: True] A boolean flag indicating if the MEV bundle executes frontrunning strategy (e.g. generalized frontrunning,
                                                                            sandwiching). Some block builders and validators may not want to accept frontrunning bundles, which may experience a lower hash power. */
	EffectiveGasPrice string  `json:"effective_gas_price,omitempty"` // [Optional, default: 0] An integer representing current bundle's effective gas price in wei.
	CoinbaseProfit    string  `json:"coinbase_profit"`               // [Optional, default: 0] An integer representing current bundle's coinbase profit in wei.
	MevBuilders  []string                                            /* [Optional, default: bloxroute builder and flashbots builder] A dictionary of MEV builders that should receive the bundle.
                                                                            For each MEV builder, a signature is required. For flashbots builder, please provide the signature used in X-Flashbots-Signature header.
                                                                            For other builders, please provide empty string as signature. 
                                                                            Possible MEV builders are:
                                                                                bloxroute: bloXroute internal builder
                                                                                flashbots: flashbots builder
                                                                                builder0x69: builder0x69​
                                                                                beaverbuild:  beaverbuild.org​
                                                                                all: all builders
                                                                            Traders can refer to List of External Builders page for a full list. */
}

// BackRunMeSubmitBundle
type BloxrouteBrmSubmitBundleRequest struct {
	TransactionHash string   `json:"transaction_hash"`        // Trigger transaction hash  
	Transaction     []string `json:"transaction"`             // A list of raw transaction bytes without a 0x prefix.
	BlockNumber     string   `json:"block_number"`            /* Block number of a future block to include this bundle in, in hex value.
                                                                     For traders who would like more than one block to be targeted, please send multiple requests targeting each specific block. */
	MinTimestamp    *uint64  `json:"min_timestamp,omitempty"` // [Optional] The minimum timestamp that the bundle is valid on, an integer in unix epoch format. Default value is None.
	MaxTimestamp    *uint64  `json:"max_timestamp,omitempty"` // [Optional] The maximum timestamp that the bundle is valid on, an integer in unix epoch format. Default value is None.
}

type BloxrouteSubmitBundleResponse struct {
	BundleHash string `json:"bundleHash"`
}

// SendTransaction
type BloxrouteSendTransactionRequest struct {
	Transaction          string     `json:"transaction"`                  // [Mandatory] Raw transactions bytes without 0x prefix.
	NonceMonitoring      bool       `json:"nonce_monitoring,omitempty"`   /* [Optional, default: False] A boolean flag indicating if Tx Nonce Monitoring should be enabled for the transaction.
                                                                                 This parameter only effects Cloud-API requests.
	                                                                         *Currently only available for users testing the Beta version, but will soon be available to all. */
	BlockchainNetwork    string     `json:""blockchain_network,omitempty` /* [Optional, default: Mainnet] Blockchain network name. Use with Cloud-API when working with BSC.
                                                                                 Available options are: Mainnet for ETH Mainnet, BSC-Mainnet for BSC Mainnet, and Polygon-Mainnet for Polygon Mainnet. */
	ValidatorsOnly       bool       `json:"validators_only,omitempty"`    // [Optional, default: False] Support for semi private transactions in all networks. See section Semi-Private Transaction for more info.
}
