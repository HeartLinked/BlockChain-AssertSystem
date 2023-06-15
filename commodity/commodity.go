package commodity

import (
	"BlockChain/database"
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	_ "go.mongodb.org/mongo-driver/mongo/options"
)

type Commodity struct {
	UserID     string `bson:"userid" json:"userid"`
	Diamond    int    `bson:"diamond" json:"diamond"`
	Axe        int    `bson:"axe" json:"axe"`
	Pickaxe    int    `bson:"pickaxe" json:"pickaxe"`
	Fishingrod int    `bson:"fishingrod" json:"fishingrod"`
	Beer       int    `bson:"beer" json:"beer"`
	Soda       int    `bson:"soda" json:"soda"`
	Hamburger  int    `bson:"hamburger" json:"hamburger"`
	Cola       int    `bson:"cola" json:"cola"`
	Fish       int    `bson:"fish" json:"fish"`
	Log        int    `bson:"log" json:"log"`
}

type Profile struct {
	Commodity Commodity `bson:"commodity" json:"commodity"`
	Balance   int       `bson:"balance" json:"balance"`
}

type Shop struct {
	Diamond    Info `bson:"diamond" json:"diamond"`
	Axe        Info `bson:"axe" json:"axe"`
	Pickaxe    Info `bson:"pickaxe" json:"pickaxe"`
	Fishingrod Info `bson:"fishingrod" json:"fishingrod"`
}

type Restaurant struct {
	Beer      Info `bson:"beer" json:"beer"`
	Soda      Info `bson:"soda" json:"soda"`
	Hamburger Info `bson:"hamburger" json:"hamburger"`
	Cola      Info `bson:"cola" json:"cola"`
}

type Info struct {
	Stock int `bson:"stock" json:"stock"`
	Price int `bson:"price" json:"price"`
}

func GetShopList() (Shop, error) {
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Shop")
	result := Shop{}
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&result)
	return result, err
}

func GetRestaurantList() (Restaurant, error) {
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Restaurant")
	result := Restaurant{}
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&result)
	return result, err
}

func GetPersonalInfo(userid string) (Commodity, error) {
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Commodity")
	result := Commodity{}
	filter := bson.D{{"userid", userid}}
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	result.UserID = userid
	return result, err
}

func PostTransaction(C Commodity) {
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Commodity")
	result := Commodity{}
	filter := bson.D{{"userid", C.UserID}}
	err := collection.FindOne(context.TODO(), filter).Decode(&result)

	tmp := Commodity{
		UserID:     C.UserID,
		Diamond:    result.Diamond + C.Diamond,
		Axe:        result.Axe + C.Axe,
		Pickaxe:    result.Pickaxe + C.Pickaxe,
		Fishingrod: result.Fishingrod + C.Fishingrod,
		Beer:       result.Beer + C.Beer,
		Soda:       result.Soda + C.Soda,
		Hamburger:  result.Hamburger + C.Hamburger,
		Cola:       result.Cola + C.Cola,
		Fish:       result.Fish + C.Fish,
		Log:        result.Log + C.Log,
	}

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			_, err2 := collection.InsertOne(context.Background(), tmp)
			if err2 != nil {
				logrus.Error("MgoDB: Insert Commodity data error: ", err2)
			}
		} else {
			logrus.Error("MgoDB: Query BlockChain data error: ", err)
		}
	} else {
		filter2 := bson.D{{"userid", C.UserID}}
		update := bson.D{{"$set", bson.D{
			{"diamond", tmp.Diamond},
			{"axe", tmp.Axe},
			{"pickaxe", tmp.Pickaxe},
			{"fishingrod", tmp.Fishingrod},
			{"beer", tmp.Beer},
			{"soda", tmp.Soda},
			{"hamburger", tmp.Hamburger},
			{"cola", tmp.Cola},
		}}}
		_, err = collection.UpdateOne(context.Background(), filter2, update)
		if err != nil {
			logrus.Error("MgoDB: Update Commodity data error: ", err)
		} else {
			logrus.Info("MgoDB: Update Commodity transaction success")
		}
	}
	ChangeStock(C)
}

func ChangeStock(C Commodity) {
	client := database.Mgo.Client
	collection := client.Database("BlockChain").Collection("Shop")
	result := Shop{}
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&result)
	if err != nil {
		logrus.Error("MgoDB: Query Shop data error: ", err)
	}
	result2 := Shop{
		Diamond: Info{
			Stock: result.Diamond.Stock - C.Diamond, Price: result.Diamond.Price,
		},
		Axe: Info{
			Stock: result.Axe.Stock - C.Axe, Price: result.Axe.Price,
		},
		Pickaxe: Info{
			Stock: result.Pickaxe.Stock - C.Pickaxe, Price: result.Pickaxe.Price,
		},
		Fishingrod: Info{
			Stock: result.Fishingrod.Stock - C.Fishingrod, Price: result.Fishingrod.Price,
		},
	}
	_, err = collection.DeleteOne(context.Background(), bson.M{})
	if err != nil {
		logrus.Error("MgoDB: Delete Shop data error: ", err)
	}
	_, err = collection.InsertOne(context.Background(), result2)
	if err != nil {
		logrus.Error("MgoDB: Insert Shop data error: ", err)
	}

	collection = client.Database("BlockChain").Collection("Restaurant")
	result4 := Restaurant{}
	err = collection.FindOne(context.TODO(), bson.M{}).Decode(&result4)
	if err != nil {
		logrus.Error("MgoDB: Query Restaurant data error: ", err)
	}
	result3 := Restaurant{
		Beer: Info{
			Stock: result4.Beer.Stock - C.Beer, Price: result4.Beer.Price,
		},
		Soda: Info{
			Stock: result4.Soda.Stock - C.Soda, Price: result4.Soda.Price,
		},
		Hamburger: Info{
			Stock: result4.Hamburger.Stock - C.Hamburger, Price: result4.Hamburger.Price,
		},
		Cola: Info{
			Stock: result4.Cola.Stock - C.Cola, Price: result4.Cola.Price,
		},
	}
	_, err = collection.DeleteOne(context.Background(), bson.M{})
	if err != nil {
		logrus.Error("MgoDB: Delete Restaurant data error: ", err)
	}
	_, err = collection.InsertOne(context.Background(), result3)
	if err != nil {
		logrus.Error("MgoDB: Insert Restaurant data error: ", err)
	}
}
