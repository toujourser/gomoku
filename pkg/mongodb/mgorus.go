package mongodb

import (
	"fmt"
	"github.com/toujourser/gomoku/pkg/logger"

	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type hooker struct {
	c *mgo.Collection
}

type M bson.M

func NewHooker(mgoUrl, db, collection string) (*hooker, error) {
	logger.Infof("Connecting to mongodb: %s/%s", mgoUrl, db)
	session, err := mgo.Dial(mgoUrl)
	if err != nil {
		logger.Errorf("Dial failed: %v", err)
		panic(err)
	}
	return &hooker{c: session.DB(db).C(collection)}, nil
}

func NewHookerFromCollection(collection *mgo.Collection) *hooker {
	return &hooker{c: collection}
}

func NewHookerWithAuth(mgoUrl, db, collection, user, pass string) (*hooker, error) {
	session, err := mgo.Dial(mgoUrl)
	if err != nil {
		return nil, err
	}

	if err := session.DB(db).Login(user, pass); err != nil {
		return nil, fmt.Errorf("Failed to login to mongodb: %v", err)
	}

	return &hooker{c: session.DB(db).C(collection)}, nil
}

func NewHookerWithAuthDb(mgoUrl, authdb, db, collection, user, pass string) (*hooker, error) {
	session, err := mgo.Dial(mgoUrl)
	if err != nil {
		return nil, err
	}

	if err := session.DB(authdb).Login(user, pass); err != nil {
		return nil, fmt.Errorf("Failed to login to mongodb: %v", err)
	}

	return &hooker{c: session.DB(db).C(collection)}, nil
}

func (h *hooker) Fire(entry *logrus.Entry) error {
	data := make(logrus.Fields)
	data["Level"] = entry.Level.String()
	data["Time"] = entry.Time
	data["Message"] = entry.Message

	for k, v := range entry.Data {
		if errData, isError := v.(error); logrus.ErrorKey == k && v != nil && isError {
			data[k] = errData.Error()
		} else {
			data[k] = v
		}
	}

	mgoErr := h.c.Insert(M(data))

	if mgoErr != nil {
		return fmt.Errorf("Failed to send log entry to mongodb: %v", mgoErr)
	}

	return nil
}

func (h *hooker) Levels() []logrus.Level {
	return logrus.AllLevels
}
