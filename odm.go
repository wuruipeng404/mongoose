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

type Mongo struct {
	client *mongo.Client
	db     *mongo.Database
}

type Options struct {
	User     string
	Password string
	Host     string
	Port     int
	DBName   string

	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration

	DriverOpts []*options.ClientOptions
}

func Open(opt *Options) (*Mongo, error) {
	var (
		err      error
		dsn      string
		client   *mongo.Client
		connOpts []*options.ClientOptions
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

	connOpts = append(connOpts, options.Client().ApplyURI(dsn).SetConnectTimeout(opt.ConnectTimeout))
	connOpts = append(connOpts, options.Client().SetServerSelectionTimeout(opt.ServerSelectionTimeout))

	if len(opt.DriverOpts) > 0 {
		connOpts = append(connOpts, opt.DriverOpts...)
	}

	if client, err = mongo.Connect(
		context.Background(),
		connOpts...,
	); err != nil {
		return nil, err
	}

	return &Mongo{
		client: client,
		db:     client.Database(opt.DBName),
	}, nil
}

// GetCollectionWithName 提供原始查询方法
func (m *Mongo) GetCollectionWithName(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	return m.db.Collection(name, opts...)
}

// InsertOne 插入一条数据
func (m *Mongo) InsertOne(doc IDocument, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	doc.PreCreate()
	return m.db.Collection(doc.CollectionName()).InsertOne(context.TODO(), doc, opts...)
}

// InsertMany 插入多条数据
func (m *Mongo) InsertMany(docs []IDocument, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	var data []interface{}

	for _, doc := range docs {
		doc.PreCreate()
		data = append(data, doc)
	}
	return m.db.Collection(docs[0].CollectionName()).InsertMany(context.TODO(), data, opts...)
}

// DeleteOne 删除一条数据
// 如果 filter 是IDocument 类型 那么会自动找到 collection
// 否则需要 指明 opt.CollectionName 参数
func (m *Mongo) DeleteOne(filter interface{}, opt *DeleteOption) (*mongo.DeleteResult, error) {
	var (
		err  error
		name string
	)

	if name, err = getCollName(filter, opt); err != nil {
		return nil, err
	}

	return m.db.Collection(name).DeleteOne(context.TODO(), ParseFilter(filter), opt.DriverOptions...)
}

// DeleteMany 删除多条数据
func (m *Mongo) DeleteMany(filter interface{}, opt *DeleteOption) (*mongo.DeleteResult, error) {
	var (
		err  error
		name string
	)
	if name, err = getCollName(filter, opt); err != nil {
		return nil, err
	}

	return m.db.Collection(name).DeleteMany(context.TODO(), ParseFilter(filter), opt.DriverOptions...)
}

// UpdateByID 通过ID更新 支持 string 或 objectId
func (m *Mongo) UpdateByID(id interface{}, update IDocument, opts ...*options.UpdateOptions) (*mongo.UpdateResult,
	error) {
	var (
		err error
		oid primitive.ObjectID
	)
	if oid, err = ConvertId(id); err != nil {
		return nil, err
	}
	update.PreUpdate()
	return m.db.Collection(update.CollectionName()).UpdateByID(context.TODO(), oid, Set(update), opts...)
}

// UpdateOne 更新一条数据
func (m *Mongo) UpdateOne(filter interface{}, update IDocument, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	update.PreUpdate()
	return m.db.Collection(update.CollectionName()).UpdateOne(context.TODO(), ParseFilter(filter), Set(update), opts...)
}

// UpdateMany 更新多条数据
func (m *Mongo) UpdateMany(filter interface{}, update IDocument, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	update.PreUpdate()
	return m.db.Collection(update.CollectionName()).UpdateMany(context.TODO(), ParseFilter(filter), Set(update), opts...)
}

// FindOne 查找一条数据
// filter 支持 bson 以及 IDocument
// result 则是一个 存储结果的指针 例如 &SomeDoc
func (m *Mongo) FindOne(filter interface{}, result interface{}, opts ...*options.FindOneOptions) (err error) {
	var name string

	if name, err = getCollNameForFind(result); err != nil {
		return err
	}
	return m.db.Collection(name).FindOne(context.TODO(), ParseFilter(filter), opts...).Decode(result)
}

// FindByID 通过id查找数据
func (m *Mongo) FindByID(id interface{}, result interface{}, opts ...*options.FindOneOptions) (err error) {
	var oid primitive.ObjectID

	if oid, err = ConvertId(id); err != nil {
		return err
	}

	return m.FindOne(IdFilter(oid), result, opts...)
}

// FindUnDeleteByID 查找一条未删除的数据
func (m *Mongo) FindUnDeleteByID(id interface{}, result interface{}, opts ...*options.FindOneOptions) (err error) {
	var oid primitive.ObjectID

	if oid, err = ConvertId(id); err != nil {
		return err
	}
	return m.FindOne(UnDeletedFilterByID(oid), result, opts...)
}

// FindOneUndeleteByFilter 查找一条未删除的数据
func (m *Mongo) FindOneUndeleteByFilter(filter interface{}, result interface{}, opts ...*options.FindOneOptions) (err error) {
	return m.FindOne(CombAndFilters(ParseFilter(filter), UndeleteFilter()), result, opts...)
}

// FindUndeleteByFilter 查找未删除的所有数据
func (m *Mongo) FindUndeleteByFilter(filter interface{}, results interface{}, opts ...*options.FindOptions) (err error) {
	return m.Find(CombAndFilters(ParseFilter(filter), UndeleteFilter()), results, opts...)
}

// Find 基础查找
// filter 支持 bson 以及 IDocument
// 如果filter 是一个 Document 那么他必须是 addressable 的, 也就是说是一个指针.
// result 则是一个 存储结果的指针 例如 &[]SomeDoc or make([]SomeDoc,0)
func (m *Mongo) Find(filter interface{}, results interface{}, opts ...*options.FindOptions) (err error) {
	var (
		cursor *mongo.Cursor
		name   string
	)

	if name, err = getCollNameForFind(results); err != nil {
		return err
	}

	if cursor, err = m.db.Collection(name).Find(
		context.TODO(),
		ParseFilter(filter),
		opts...,
	); err != nil {
		return
	}

	defer func() {
		_ = cursor.Close(context.TODO())
	}()

	return cursor.All(context.TODO(), results)
}
