package web

import (
	"BlockChain/block"
	"BlockChain/commodity"
	"BlockChain/database"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

// GetMineBlock 匹配/api/mining?userid=xxx&amount=xxx
func GetMineBlock(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	from := c.Query("userid")
	amount, err := strconv.Atoi(c.DefaultQuery("amount", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount value"})
		return
	} else if from != "" {
		newBlock := block.MineBlock(from, amount)
		c.JSON(http.StatusCreated, newBlock)
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Userid value"})
		return
	}
}

// GetProfile 匹配/api/profile?userid=xxx
func GetProfile(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	userid := c.Query("userid")
	C, _ := commodity.GetPersonalInfo(userid)
	balance := block.GetBalance(userid)
	profile := commodity.Profile{Commodity: C, Balance: balance}
	c.JSON(http.StatusOK, profile)
}

// GetShopList 匹配/api/shop/list
func GetShopList(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	C, err := commodity.GetShopList()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			response := gin.H{
				"message": "BadRequest: Shop list is not exist",
			}
			c.JSON(http.StatusNotFound, response)
		} else {
			logrus.Error("MgoDB Database error: ", err)
			response := gin.H{
				"message": "Internal Server Error",
			}
			c.JSON(http.StatusInternalServerError, response)
		}
	} else {
		c.JSON(http.StatusOK, C)
	}
}

// GetRestaurantList 匹配/api/restaurant/list
func GetRestaurantList(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	C, err := commodity.GetRestaurantList()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			response := gin.H{
				"message": "BadRequest: Restaurant list is not exist",
			}
			c.JSON(http.StatusNotFound, response)
		} else {
			logrus.Error("MgoDB Database error: ", err)
			response := gin.H{
				"message": "Internal Server Error",
			}
			c.JSON(http.StatusInternalServerError, response)
		}
	} else {
		c.JSON(http.StatusOK, C)
	}
}

// Textcointx 匹配/api/transaction
func Textcointx(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	from, ok := c.Request.Form["from"]
	if !ok || len(from) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing from value"})
		return
	}
	to, ok := c.Request.Form["to"]
	if !ok || len(to) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing to value"})
		return
	}
	amount, ok := c.Request.Form["amount"]
	amountInt, err := strconv.Atoi(amount[0])
	if !ok || len(amount) == 0 || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing amount value"})
		return
	}
	timeNow := time.Now().String()
	t := block.NewTransaction(timeNow, from[0], to[0], amountInt)
	b := block.TXBlock(timeNow, t)
	c.JSON(http.StatusCreated, b)
}

// GetBlockchainStatus 匹配/api/blockchain/status
func GetBlockchainStatus(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	blocks := block.FindAllBlocks()
	c.JSON(http.StatusOK, blocks)
}

// GetTransactionRecords 匹配/api/blockchain/records?userid=xxx
func GetTransactionRecords(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	userid := c.Query("userid")
	if userid != "" {
		records := block.FindAllTransactionRecords(userid)
		if records == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Userid has no transaction records"})
			return
		}
		c.JSON(http.StatusOK, records)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Userid value"})
		return
	}
}

// PostSpotTransaction 匹配/api/spot/transaction
func PostSpotTransaction(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	var err1, err2, err3, err4 error
	C := commodity.Commodity{}
	C.UserID = c.DefaultPostForm("userid", "user001")
	C.Diamond, err1 = strconv.Atoi(c.DefaultPostForm("diamond", "0"))
	C.Axe, err2 = strconv.Atoi(c.DefaultPostForm("axe", "0"))
	C.Pickaxe, err3 = strconv.Atoi(c.DefaultPostForm("pickaxe", "0"))
	C.Fishingrod, err4 = strconv.Atoi(c.DefaultPostForm("fishingrod", "0"))
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	C.Beer, err1 = strconv.Atoi(c.DefaultPostForm("beer", "0"))
	C.Soda, err2 = strconv.Atoi(c.DefaultPostForm("soda", "0"))
	C.Hamburger, err3 = strconv.Atoi(c.DefaultPostForm("hamburger", "0"))
	C.Cola, err4 = strconv.Atoi(c.DefaultPostForm("cola", "0"))
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	amount := C.Diamond*500 + C.Axe*30 + C.Pickaxe*50 + C.Fishingrod*70
	amount += C.Beer*7 + C.Soda*3 + C.Hamburger*10 + C.Cola*3

	if amount > block.GetBalance(C.UserID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Balance is not enough"})
		return
	} else {
		commodity.PostTransaction(C)
		timeNow := time.Now().String()
		t := block.NewTransaction(timeNow, C.UserID, "shop", amount)
		block.TXBlock(timeNow, t)
		balance := block.GetBalance(C.UserID)
		Como, _ := commodity.GetPersonalInfo(C.UserID)
		profile := commodity.Profile{Commodity: Como, Balance: balance}
		c.JSON(http.StatusOK, profile)
	}

}

var cnt int

type Sell struct {
	ID        string `json:"id" bson:"id"`
	User      string `json:"user" form:"user" binding:"required" bson:"user" `
	Commodity string `json:"commodity" form:"commodity" binding:"required" bson:"commodity"`
	Amount    int    `json:"amount" form:"amount" binding:"required" bson:"amount"`
	Price     int    `json:"price" form:"price" binding:"required" bson:"price"`
}

func PutOnSell(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	var sell Sell
	if err := c.ShouldBind(&sell); err == nil {
		//spew.Dump(sell)
	}
	client := database.Mgo.Client
	collection1 := client.Database("BlockChain").Collection("Commodity")
	result := commodity.Commodity{}
	filter := bson.D{{"userid", sell.User}}
	err := collection1.FindOne(context.TODO(), filter).Decode(&result)

	switch sell.Commodity {
	case "diamond":
		if result.Diamond < sell.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Diamond is not enough"})
			return
		} else {
			result.Diamond -= sell.Amount
		}
	case "axe":
		if result.Axe < sell.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Axe is not enough"})
			return
		} else {
			result.Axe -= sell.Amount
		}
	case "pickaxe":
		if result.Pickaxe < sell.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pickaxe is not enough"})
			return
		} else {
			result.Pickaxe -= sell.Amount
		}
	case "fishingrod":
		if result.Fishingrod < sell.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Fishingrod is not enough"})
			return
		} else {
			result.Fishingrod -= sell.Amount
		}
	case "beer":
		if result.Beer < sell.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Beer is not enough"})
			return
		} else {
			result.Beer -= sell.Amount
		}
	case "soda":
		if result.Soda < sell.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Soda is not enough"})
			return
		} else {
			result.Soda -= sell.Amount
		}
	case "hamburger":
		if result.Hamburger < sell.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Hamburger is not enough"})
			return
		} else {
			result.Hamburger -= sell.Amount
		}
	case "cola":
		if result.Cola < sell.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cola is not enough"})
			return
		} else {
			result.Cola -= sell.Amount
		}
	case "fish":
		if result.Fish < sell.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Fish is not enough"})
			return
		} else {
			result.Fish -= sell.Amount
		}
	case "log":
		if result.Log < sell.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Log is not enough"})
			return
		} else {
			result.Log -= sell.Amount
		}

	}

	filter = bson.D{{"userid", sell.User}}
	update := bson.D{{"$set", bson.D{{"diamond", result.Diamond}, {"axe", result.Axe},
		{"pickaxe", result.Pickaxe}, {"fishingrod", result.Fishingrod}, {"beer", result.Beer},
		{"soda", result.Soda}, {"hamburger", result.Hamburger}, {"cola", result.Cola}, {"fish", result.Fish}, {"log", result.Log}}}}
	_, err = collection1.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logrus.Error("MgoDB: Update Commodity data error: ", err)
	} else {
		logrus.Info("MgoDB: Update Commodity data success")
	}

	sell.ID = database.GenerateRandomString(20)
	collection := client.Database("BlockChain").Collection("UsersSell")
	_, err = collection.InsertOne(context.TODO(), sell)
	if err != nil {
		logrus.Error("MgoDB: Insert BlockChain data error: ", err)
	} else {
		logrus.Info("MgoDB: Insert BlockChain data success")
	}
	c.JSON(http.StatusOK, gin.H{"id": sell.ID})
}

func PurchaseRequest(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	userid := c.Query("userid")
	ID := c.Query("id")
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("UsersSell")
	result := Sell{}
	filter := bson.D{{"id", ID}}
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		logrus.Error("MgoDB: FindOne BlockChain data error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	balance := block.GetBalance(userid)
	if balance < result.Price*result.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Balance is not enough"})
		return
	} else {
		collection1 := client.Database("BlockChain").Collection("Commodity")
		result2 := commodity.Commodity{}
		filter := bson.D{{"userid", userid}}
		err := collection1.FindOne(context.TODO(), filter).Decode(&result2)
		if err != nil {
			logrus.Error("MgoDB: FindOne Commodity data error: ", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User"})
			return
		}
		switch result.Commodity {
		case "diamond":
			result2.Diamond += result.Amount
		case "axe":
			result2.Axe += result.Amount
		case "pickaxe":
			result2.Pickaxe += result.Amount
		case "fishingrod":
			result2.Fishingrod += result.Amount
		case "beer":
			result2.Beer += result.Amount
		case "soda":
			result2.Soda += result.Amount
		case "hamburger":
			result2.Hamburger += result.Amount
		case "cola":
			result2.Cola += result.Amount
		case "fish":
			result2.Fish += result.Amount
		case "log":
			result2.Log += result.Amount
		}
		filter = bson.D{{"userid", userid}}
		update := bson.D{{"$set", bson.D{{"diamond", result2.Diamond},
			{"axe", result2.Axe}, {"pickaxe", result2.Pickaxe},
			{"fishingrod", result2.Fishingrod}, {"beer", result2.Beer},
			{"soda", result2.Soda}, {"hamburger", result2.Hamburger}, {"cola", result2.Cola}, {"fish", result2.Fish}, {"log", result2.Log}}}}
		_, err = collection1.UpdateOne(context.Background(), filter, update)
		if err != nil {
			logrus.Error("MgoDB: Update Commodity data error: ", err)
		} else {
			logrus.Info("MgoDB: Update Commodity data success")
		}
		timeNow := time.Now().String()
		t := block.NewTransaction(timeNow, userid, result.User, result.Price)
		_ = block.TXBlock(timeNow, t)

		collection := client.Database("BlockChain").Collection("UsersSell")
		filter3 := bson.D{{"id", ID}}
		collection.DeleteOne(context.Background(), filter3)
		c.JSON(http.StatusOK, "Purchase success")
	}

}

func GetUsersSellList(c *gin.Context) {
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("UsersSell")
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		logrus.Error("MgoDB: Find BlockChain data error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var results []Sell
	if err = cursor.All(context.Background(), &results); err != nil {
		logrus.Error("MgoDB: All BlockChain data error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	c.JSON(http.StatusOK, results)
}

func Fishing(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	userid := c.Query("userid")
	amount, _ := strconv.Atoi(c.Query("amount"))
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Commodity")
	result := commodity.Commodity{}
	filter := bson.D{{"userid", userid}}
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		logrus.Error("MgoDB: FindOne Commodity data error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User"})
		return
	}

	result.Fish += amount
	filter = bson.D{{"userid", userid}}
	update := bson.D{{"$set", bson.D{{"fish", result.Fish}}}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logrus.Error("MgoDB: Update Commodity data error: ", err)
	} else {
		logrus.Info("MgoDB: Update Commodity data success")
	}
	c.JSON(http.StatusOK, "Fishing success")
}

func Logging(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	userid := c.Query("userid")
	amount, _ := strconv.Atoi(c.Query("amount"))
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Commodity")
	result := commodity.Commodity{}
	filter := bson.D{{"userid", userid}}
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		logrus.Error("MgoDB: FindOne Commodity data error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User"})
		return
	}

	result.Log += amount
	filter = bson.D{{"userid", userid}}
	update := bson.D{{"$set", bson.D{{"log", result.Log}}}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logrus.Error("MgoDB: Update Commodity data error: ", err)
	} else {
		logrus.Info("MgoDB: Update Commodity data success")
	}
	c.JSON(http.StatusOK, "Logging success")
}

func CheckFishing(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	userid := c.Query("userid")
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Commodity")
	result := commodity.Commodity{}
	filter := bson.D{{"userid", userid}}
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		logrus.Error("MgoDB: FindOne Commodity data error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User"})
		return
	}
	if result.Fishingrod <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fishingrod is not enough"})
		return
	} else {
		c.JSON(http.StatusOK, "Fishingrod is enough")
	}
}

func CheckMining(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	userid := c.Query("userid")
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Commodity")
	result := commodity.Commodity{}
	filter := bson.D{{"userid", userid}}
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		logrus.Error("MgoDB: FindOne Commodity data error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User"})
		return
	}
	if result.Pickaxe <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pickaxe is not enough"})
		return
	} else {
		c.JSON(http.StatusOK, "Pickaxe is enough")
	}
}

func CheckLogging(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	userid := c.Query("userid")
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Commodity")
	result := commodity.Commodity{}
	filter := bson.D{{"userid", userid}}
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		logrus.Error("MgoDB: FindOne Commodity data error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User"})
		return
	}
	if result.Axe <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Axe is not enough"})
		return
	} else {
		c.JSON(http.StatusOK, "Axe is enough")
	}
}

func Register(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	userid := c.Query("userid")
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Commodity")
	result := commodity.Commodity{}
	filter := bson.D{{"userid", userid}}
	err := collection.FindOne(context.Background(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			result.Pickaxe = 1
			result.UserID = userid
			_, err := collection.InsertOne(context.Background(), result)
			if err != nil {
				logrus.Error("MgoDB: Insert Commodity data error: ", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User"})
			} else {
				logrus.Info("MgoDB: Insert Commodity data success")
			}
			c.JSON(http.StatusOK, "Register success")
		} else {
			logrus.Error("MgoDB: Query BlockChain data error: ", err)
		}
	} else if result.Axe <= 0 && result.Fishingrod <= 0 && result.Pickaxe <= 0 && result.Log <= 0 &&
		result.Fish <= 0 && result.Hamburger <= 0 && result.Diamond <= 0 && result.Beer <= 0 &&
		result.Soda <= 0 && result.Cola <= 0 {

		result.Pickaxe = 1
		filter = bson.D{{"userid", userid}}

		update := bson.D{{"$set", bson.D{{"pickaxe", result.Pickaxe}}}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			logrus.Error("MgoDB: Update Commodity data error: ", err)
		} else {
			logrus.Info("MgoDB: Update Commodity data success")
		}
		c.JSON(http.StatusOK, "Register success: Axe")
	} else {
		c.JSON(http.StatusOK, "Register success")
	}
}
