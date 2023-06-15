package main

import (
	"BlockChain/block"
	"BlockChain/database"
	"BlockChain/web"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func init() {

	logrus.SetLevel(logrus.TraceLevel)

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

}

func main() {

	logFile := "log/log.txt"
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Failed to create logfile" + logFile)
		panic(err)
	}
	defer f.Close()
	logrus.SetOutput(f)

	database.Init()
	block.Init()
	go runWebServer()
	time.Sleep(1e8 * time.Second)

}

// web server
func runWebServer() {
	r := setupRouter()
	httpPort := os.Getenv("PORT")
	r.Use(cors.Default())
	err := r.Run(":" + httpPort)
	logrus.Info("Run client in port 8080！")
	if err != nil {
		logrus.Fatal("ERROR in Run client in port 8080！")
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default())
	// 匹配/api/profile?userid=xxx
	r.GET("/api/profile", web.GetProfile)
	// 匹配/api/mining?userid=xxx&amount=xxx
	r.GET("/api/mining", web.GetMineBlock)

	// 匹配/api/shop/list
	r.GET("/api/shop/list", web.GetShopList)
	// 匹配/api/restaurant/list
	r.GET("/api/restaurant/list", web.GetRestaurantList)

	// 匹配/api/transaction 测试交易
	r.POST("/api/transaction", web.Textcointx)

	// 匹配/api/blockchain/status
	r.GET("/api/blockchain/status", web.GetBlockchainStatus)
	// 匹配/api/blockchain/records?userid=xxx
	r.GET("/api/blockchain/records", web.GetTransactionRecords)

	// 匹配/api/spot/transaction
	r.POST("/api/spot/transaction", web.PostSpotTransaction)

	// 匹配/api/users/sell
	r.POST("/api/users/sell", web.PutOnSell)
	// 匹配/api/users/purchase
	r.GET("/api/users/purchase", web.PurchaseRequest)
	// 匹配/api/users/list
	r.GET("/api/users/list", web.GetUsersSellList)

	// /api/fishing/check?userid=xxx  200 可以，400 不足，提示鱼竿不足，获取鱼竿再来
	r.GET("/api/fishing/check", web.CheckFishing)

	// /api/mining/check?userid=xxx  200 可以，400 不足，提示镐子不足，获取镐子再来
	r.GET("/api/mining/check", web.CheckMining)

	// 匹配/api/fishing?userid=xxx&amount=xxx 捕鱼，无返回值
	r.GET("/api/fishing", web.Fishing)

	// /api/logging/check?userid=xxx  200 可以，400 不足，提示斧子不足，获取斧子再来
	r.GET("/api/logging/check", web.CheckLogging)

	// 匹配/api/logging?userid=xxx&amount=xxx 伐木，无返回值
	r.GET("/api/logging", web.Logging)

	// 匹配/api/register?userid=xxx 200 成功，400 失败
	r.GET("/api/register", web.Register)

	return r
}
