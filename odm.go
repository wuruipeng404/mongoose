/*
* @Author: Rumple
* @Email: ruipeng.wu@cyclone-robotics.com
* @DateTime: 2021/8/24 16:44
 */

package mongoose

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongoose struct {
	client *mongo.Client
	db     *mongo.Database
}

type MOptions struct {
	User     string
	Password string
	Host     string
	Port     int
	DBName   string

	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
}

func Open(opt *MOptions) (*Mongoose, error) {
	var (
		err    error
		dsn    string
		client *mongo.Client
	)

	if opt.User == "" || opt.Password == "" {
		dsn = fmt.Sprintf("mongodb://%s:%d", opt.Host, opt.Port)
	} else {
		dsn = fmt.Sprintf("mongodb://%s:%s@%s:%d", opt.User, opt.Password, opt.Host, opt.Port)
	}

	if opt.ConnectTimeout == 0 {
		opt.ConnectTimeout = 5 * time.Second
	}

	if opt.ServerSelectionTimeout == 0 {
		opt.ServerSelectionTimeout = 10 * time.Second
	}

	if client, err = mongo.Connect(
		context.Background(),
		options.Client().ApplyURI(dsn).SetConnectTimeout(opt.ConnectTimeout),
		options.Client().SetServerSelectionTimeout(opt.ServerSelectionTimeout),
	); err != nil {
		return nil, err
	}

	return &Mongoose{
		client: client,
		db:     client.Database(opt.DBName),
	}, nil
}

func (m *Mongoose) getCollName(filter interface{}) string {
	if doc, ok := filter.(IDocument); !ok {
		return ""
	} else {
		return doc.CollectionName()
	}
}

// InsertOne 插入一条数据
func (m *Mongoose) InsertOne(doc IDocument, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	doc.PreCreate()
	return m.db.Collection(doc.CollectionName()).InsertOne(context.TODO(), doc, opts...)
}

// InsertMany 插入多条数据
func (m *Mongoose) InsertMany(docs []IDocument, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	var data []interface{}

	for _, doc := range docs {
		doc.PreCreate()
		data = append(data, doc)
	}

	return m.db.Collection(docs[0].CollectionName()).InsertMany(context.TODO(), data, opts...)
}

// DeleteOne 删除一条数据
func (m *Mongoose) DeleteOne(filter interface{}, opt *DeleteOption) (*mongo.DeleteResult, error) {
	var collName string

	if collName = m.getCollName(filter); collName == "" {
		collName = opt.CollectionName
	}

	if collName == "" {
		return nil, CollectionNameNotFound
	}

	return m.db.Collection(collName).DeleteOne(context.TODO(), ConvertFilter(filter), opt.DriverOptions...)
}

// DeleteMany 删除多条数据
func (m *Mongoose) DeleteMany(filter interface{}, opt *DeleteOption) (*mongo.DeleteResult, error) {
	var collName string

	if collName = m.getCollName(filter); collName == "" {
		collName = opt.CollectionName
	}

	if collName == "" {
		return nil, CollectionNameNotFound
	}

	return m.db.Collection(collName).DeleteMany(context.TODO(), ConvertFilter(filter), opt.DriverOptions...)
}

// UpdateByID 通过ID更新 支持 string 或 objectId
func (m *Mongoose) UpdateByID(id interface{}, update IDocument, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var (
		err error
		oid primitive.ObjectID
	)
	if oid, err = ConvertId(id); err != nil {
		return nil, err
	}
	update.PreUpdate()
	return m.db.Collection(update.CollectionName()).UpdateByID(context.TODO(), oid, update, opts...)
}

// UpdateOne 更新一条数据
func (m *Mongoose) UpdateOne(filter interface{}, update IDocument, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	update.PreUpdate()
	return m.db.Collection(update.CollectionName()).UpdateOne(context.TODO(), ConvertFilter(filter), update, opts...)
}

// UpdateMany 更新多条数据
func (m *Mongoose) UpdateMany(filter interface{}, update IDocument, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	update.PreUpdate()
	return m.db.Collection(update.CollectionName()).UpdateMany(context.TODO(), ConvertFilter(filter), update, opts...)
}

// FindOne 查找一条数据
func (m *Mongoose) FindOne(filter interface{}, result IDocument, opts ...*options.FindOneOptions) error {
	return m.db.Collection(result.CollectionName()).FindOne(context.TODO(), ConvertFilter(filter), opts...).Decode(result)
}

// FindByID 通过id查找数据
func (m *Mongoose) FindByID(id interface{}, result IDocument, opts ...*options.FindOneOptions) (err error) {
	var oid primitive.ObjectID

	if oid, err = ConvertId(id); err != nil {
		return err
	}
	return m.FindOne(IdFilter(oid), result, opts...)
}

func (m *Mongoose) FindUnDeleteByID(id interface{}, result IDocument, opts ...*options.FindOneOptions) (err error) {
	var oid primitive.ObjectID

	if oid, err = ConvertId(id); err != nil {
		return err
	}
	return m.FindOne(UnDeletedFilterByID(oid), result, opts...)
}

// FindUndeleteByFilter 查找未删除的所有数据
func (m *Mongoose) FindUndeleteByFilter(filter interface{}, results []IDocument, opts ...*options.FindOptions) (err error) {
	return m.Find(CombAndFilters(ConvertFilter(filter), UndeleteFilter()), results, opts...)
}

// FindOneUndeleteByFilter 查找一条未删除的数据
func (m *Mongoose) FindOneUndeleteByFilter(filter interface{}, result IDocument, opts ...*options.FindOneOptions) (err error) {
	return m.FindOne(CombAndFilters(ConvertFilter(filter), UndeleteFilter()), result, opts...)
}

// Find 基础查找
func (m *Mongoose) Find(filter interface{}, results []IDocument, opts ...*options.FindOptions) (err error) {
	var cursor *mongo.Cursor

	if cursor, err = m.db.Collection(results[0].CollectionName()).Find(context.TODO(), ConvertFilter(filter),
		opts...); err != nil {
		return
	}

	defer func() {
		_ = cursor.Close(context.TODO())
	}()

	return cursor.All(context.TODO(), results)
}
