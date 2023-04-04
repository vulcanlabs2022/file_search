package db

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	MgoCli, err = mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		panic("connect mongo error" + err.Error())
	}
	err = MgoCli.Ping(context.TODO(), nil)
	if err != nil {
		panic("ping mongo error" + err.Error())
	}
	collection = MgoCli.Database("terminus").Collection("conversation")
	err = InsertSingleConversation(Message{
		ConversationId: "123",
		MessageId:      "123",
		Prompt:         "123",
		Text:           "123",
		StartTime:      0,
		Model:          "123",
	})
	if err != nil {
		panic(err)
	}
}
