package repository

import (
	"reflect"

	"dix975.com/basic/logger"
	"dix975.com/basic/pageable"
	"dix975.com/database"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	mongo         *db.DB
	configuration *db.MongoServerConfig
)

type Repository interface {
}

func Init(mongoDBConfig *db.MongoServerConfig, mongoDatabase *db.DB) {

	configuration = mongoDBConfig
	mongo = mongoDatabase
}

func WithDB(f func(db *mgo.Database)) {

	copy := mongo.Session.Copy()
	defer copy.Close()
	f(copy.DB(configuration.DatabaseName))
}

func List(result interface{}, collection string) error {

	return NextPage(result, collection, &pageable.Pageable{
		Page: 0,
		Size: 0})

}

func NextPage(result interface{}, collection string, pageable *pageable.Pageable) error {

	return NextPageWithQuery(bson.M{}, result, collection, pageable)

}

func NextPageWithQuery(query bson.M, result interface{}, collection string, pageable *pageable.Pageable) error {

	var err error

	pageIndex := pageable.Page - 1

	logger.Trace.Println("Page index : ", pageIndex)

	skip := pageable.Size * pageIndex

	WithDB(func(db *mgo.Database) {

		query := db.C(collection).Find(query)

		if pageable.Page > 0 {
			logger.Trace.Println("Setting skip to : ", skip)
			query.Skip(skip)
		}

		if pageable.Size > 0 {
			logger.Trace.Println("Setting limite to : ", pageable.Size)
			query.Limit(pageable.Size)
		}

		pageable.ApplySort(query)
		err = query.All(result)

		resultSize := reflect.Indirect(reflect.ValueOf(result)).Len()
		logger.Debug.Printf("Setting page current count [%d]\n", resultSize)
		pageable.CurrentCount = resultSize

	})

	return err
}
