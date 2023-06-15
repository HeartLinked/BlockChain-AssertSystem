package block

import (
	"BlockChain/database"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Index        int         `bson:"index"`
	Timestamp    string      `bson:"timestamp"`
	Hash         string      `bson:"hash"`
	PrevHash     string      `bson:"prevHash"`
	Nonce        int         `bson:"nonce"`
	Transactions Transaction `bson:"transaction"`
}

func Init() {
	mutex.Lock()
	defer mutex.Unlock()

	// 如果必要的话初始化区块链
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Block")
	result := Block{}
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			initBlockChain()
		} else {
			logrus.Error("MgoDB: Query BlockChain data error: ", err)
		}
	}

}

func findLastBlock() (string, int) {

	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Block")
	filter := bson.D{{}}
	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		logrus.Error("MgoDB: Count BlockChain data error: ", err)
	}
	LastIndex := int(count) - 1
	result := Block{}
	err = collection.FindOne(context.TODO(), bson.D{{"index", LastIndex}}).Decode(&result)
	if err != nil {
		logrus.Error("MgoDB: Query BlockChain data error: ", err)
	}
	return result.Hash, LastIndex
}

var mutex = &sync.Mutex{}

func AppendBlock(newBlock Block) {
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Block")
	result := Block{}
	LastHash, _ := findLastBlock()
	err := collection.FindOne(context.TODO(), bson.D{{"hash", LastHash}}).Decode(&result)
	if err != nil {
		logrus.Error("MgoDB: Query BlockChain data error: ", err)
	}

	_, err = collection.InsertOne(context.TODO(), newBlock)
	if err != nil {
		logrus.Error("MgoDB: Insert BlockChain data error: ", err)
	} else {
		logrus.Info("MgoDB: Insert BlockChain data success")
	}
}

// make sure block is valid by checking index, and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateBlockHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// SHA256 hashing
func calculateBlockHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + block.PrevHash
	t, _ := json.Marshal(block.Transactions)
	record += strconv.Itoa(block.Nonce) + string(t)
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block) Block {

	var newBlock Block
	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = time.Now().String()
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateBlockHash(newBlock)

	return newBlock
}

func initBlockChain() {

	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Block")
	logrus.Info("MgoDB: No BlockChain data, init BlockChain")

	// 第一个 coinbase 交易
	cbAddress := NewCoinbaseTX("genesis coinbaseTX", 50)

	genesisBlock := Block{}
	genesisBlock = Block{0, time.Now().String(),
		calculateBlockHash(genesisBlock), "", 0, cbAddress}

	_, err := collection.InsertOne(context.TODO(), genesisBlock)
	if err != nil {
		logrus.Error("MgoDB: Init genesis Block error: ", err)
	} else {
		logrus.Info("MgoDB: Init genesis Block success")
	}
}

func proofOfWork(b Block) (bool, string) {
	for {
		h := sha256.New()
		h.Write([]byte(strconv.Itoa(b.Nonce) + b.PrevHash + b.Timestamp))
		hashed := h.Sum(nil)
		hash := hex.EncodeToString(hashed)
		if strings.HasPrefix(hash, "0000") {
			return true, hash
		}
		b.Nonce++
	}

}

func MineBlock(Userid string, amount int) Block {
	var newBlock Block
	a, b := findLastBlock()
	//fmt.Println("LastBlock index: ", b, "LastBlock hash: ", a)
	newBlock.Index = b + 1
	newBlock.Timestamp = time.Now().String()
	newBlock.PrevHash = a
	newBlock.Nonce = 0
	flag := false
	for {
		newBlock.Hash = calculateBlockHash(newBlock)
		st := []string{"0"}
		for _, v := range st {
			if strings.HasPrefix(newBlock.Hash, v) {
				flag = true
				break
			}
		}
		if flag == false {
			newBlock.Nonce++
		} else {
			break
		}
	}
	//fmt.Println("NewBlock index: ", newBlock.Index, "NewBlock hash: ", newBlock.Hash)
	newBlock.Transactions = NewCoinbaseTX(Userid, amount)
	InsertRecord(newBlock.Timestamp, "genesis", Userid, amount, newBlock.Transactions.ID)
	AppendBlock(newBlock)
	return newBlock
}

func TXBlock(timestamp string, transaction Transaction) Block {
	var newBlock Block
	a, b := findLastBlock()
	newBlock.Index = b + 1
	newBlock.Timestamp = timestamp
	newBlock.PrevHash = a
	newBlock.Nonce = 0
	newBlock.Transactions = transaction
	newBlock.Hash = calculateBlockHash(newBlock)
	//spew.Dump(newBlock)
	AppendBlock(newBlock)
	return newBlock
}

func FindAllBlocks() []Block {
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Block")
	filter := bson.D{{"index", bson.D{{"$gt", 0}}}}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		logrus.Info("MgoDB: find blocks error: ", err)
	}

	var results []Block
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Info("MgoDB: find transactions error: ", err)
	}
	for _, result := range results {
		cursor.Decode(&result)
	}
	return results
}
