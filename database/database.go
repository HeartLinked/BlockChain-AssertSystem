package database

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

type Mongo struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var (
	//url = "mongodb+srv://xfydemx:LFYmdb1213-@cluster0.ivvl0ib.mongodb.net/?retryWrites=true&w=majority"
	url = "mongodb://xfydemx:mongodb@81.68.140.151/?authSource=admin"
	Mgo = &Mongo{}
)

func Init() *mongo.Client {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI(url).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logrus.Fatal("FAILED to connect MongoDB!")
	}
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		logrus.Fatal("FAILED to connect MongoDB!")
	} else {
		logrus.Info("Successfully connected MongoDB and pinged.")
	}

	Mgo.Client = client

	return client
}

func GenerateRandomString(length int) string {
	randomBytes := make([]byte, length)
	rand.Read(randomBytes)
	randomString := base64.URLEncoding.EncodeToString(randomBytes)
	return randomString[:length]
}
