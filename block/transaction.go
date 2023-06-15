package block

import (
	"BlockChain/database"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

// Transaction 由交易 ID，输入和输出构成
type Transaction struct {
	ID   string     `bson:"id"`
	Vin  []TXInput  `bson:"vin"`
	Vout []TXOutput `bson:"vout"`
}

// TXInput 包含 3 部分
// Txid: 一个交易输入引用了之前一笔交易的一个输出, ID 表明是之前哪笔交易
// Vout: 一笔交易可能有多个输出，Vout 为输出的索引
// ScriptSig: 提供解锁输出 Txid:Vout 的数据
type TXInput struct {
	Txid      string `bson:"txid"`
	Vout      int    `bson:"vout"`
	ScriptSig string `bson:"scriptSig"`
}

// TXOutput 包含两部分
// Value: 有多少币，就是存储在 Value 里面
// ScriptPubKey: 对输出进行锁定: （目前版本的实现）货币的拥有者，也就是地址
// 在当前实现中，ScriptPubKey 将仅用一个字符串来代替
type TXOutput struct {
	Value        int    `bson:"value"`
	ScriptPubKey string `bson:"scriptPubKey"`
	Ifused       bool   `bson:"ifused"`
}

type Record struct {
	Timestamp string `bson:"timestamp" json:"timestamp"`
	From      string `bson:"from" json:"from"`
	To        string `bson:"to" json:"to"`
	Amount    int    `bson:"amount" json:"amount"`
	Txid      string `bson:"txid" json:"txid"`
}

// CanUnlockOutputWith 这里的 unlockingData 可以理解为地址
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// NewCoinbaseTX coinbase 交易只有一个输出，没有输入。
// 在我们的实现中，它表现为 Txid 为空，Vout 等于 -1。
// 并且，在当前实现中，coinbase 交易也没有在 ScriptSig 中存储脚本，
// 而只是存储了一个任意的字符串 data。
func NewCoinbaseTX(to string, amount int) Transaction {

	txin := TXInput{"", -1, "genesis"}
	// subsidy = 10, 第一个coinbase交易的奖励是10个币
	txout := TXOutput{amount, to, false}
	tx := Transaction{"", []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return tx
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hex.EncodeToString(hash[:])
}

func NewTransaction(timestamp string, from, to string, amount int) Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	// 1. 找到足够的钱
	// 2. 创建输入
	// 3. 创建输出
	// 4. 创建交易
	acc, validOutputs := FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}
	// Build a list of inputs
	for txid, outs := range validOutputs {

		for _, out := range outs {
			input := TXInput{txid, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TXOutput{amount, to, false})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from, false}) // a change
	}

	tx := Transaction{"", inputs, outputs}
	tx.SetID()

	InsertRecord(timestamp, from, to, amount, tx.ID)
	return tx
}

// FindSpendableOutputs 为用户 address 找到足够 amount 的输出
func FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0

	Blocks := FindAllBlocks()
	logrus.Info("FindSpendableOutputs: ", address, amount)
Work:
	for _, block := range Blocks {
		tx := block.Transactions
		tx2 := block.Transactions
		flag := false
		for id, out := range tx.Vout {
			if out.ScriptPubKey == address && out.Ifused == false {
				unspentOutputs[tx.ID] = append(unspentOutputs[tx.ID], id)
				accumulated += out.Value
				flag = true
				tx2.Vout[id].Ifused = true

				if accumulated >= amount {
					if flag {
						hash := block.Hash
						updateUsed(tx2, hash)
					}
					break Work
				}
			}
		}
		if flag {
			hash := block.Hash
			updateUsed(tx2, hash)
		}
	}
	return accumulated, unspentOutputs
}

func updateUsed(tx Transaction, hash string) {
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Block")
	filter := bson.D{{"hash", hash}}
	update := bson.D{{"$set", bson.D{{"transaction", tx}}}}
	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		logrus.Info("MgoDB: find block error: ", err)
	}
}

func GetBalance(Userid string) int {
	balance := 0
	Blocks := FindAllBlocks()
	for _, block := range Blocks {
		tx := block.Transactions
		for _, out := range tx.Vout {
			if out.ScriptPubKey == Userid && out.Ifused == false {
				balance += out.Value
			}
		}
	}
	return balance
}

func FindAllTransactionRecords(Userid string) []Record {
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Transaction")
	filter := bson.D{{"$or", bson.A{bson.D{{"from", Userid}}, bson.D{{"to", Userid}}}}}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		logrus.Info("MgoDB: find record error: ", err)
	}
	var results []Record
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Info("MgoDB: find record error: ", err)
	}
	for _, result := range results {
		cursor.Decode(&result)
	}
	return results
}

func InsertRecord(timestamp string, from, to string, amount int, txid string) {
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Transaction")
	record := Record{timestamp, from, to, amount, txid}
	_, err := collection.InsertOne(context.Background(), record)
	if err != nil {
		logrus.Info("MgoDB: insert record error: ", err)
	}
}
