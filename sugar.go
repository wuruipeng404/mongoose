/*
* @Author: Rumple
* @Email: ruipeng.wu@cyclone-robotics.com
* @DateTime: 2021/8/24 15:54
 */

package mongoose

import (
	"reflect"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// ConvertFilter 检查filter并进行转化. 转化为driver 支持的格式
// 如果本身就是 bson.M 或 bson.D 系列则不进行转化
// 如果是结构体或者结构体指针则进行转化 (支持复杂结构体)
func ConvertFilter(filter interface{}) interface{} {
	var (
		finallyF bson.M
	)
	// todo:.......
	return finallyF
}

// SimpleStructToDoc 没有嵌套的结构体可以使用此方法进行转化
func SimpleStructToDoc(v interface{}) (doc *bsoncore.Document, err error) {
	var data []byte

	if data, err = bson.Marshal(v); err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

func ConvertId(id interface{}) (oid primitive.ObjectID, err error) {
	if reflect.TypeOf(id) != reflect.TypeOf(primitive.NilObjectID) {
		if oid, err = primitive.ObjectIDFromHex(id.(string)); err != nil {
			return
		}
	} else {
		oid = id.(primitive.ObjectID)
	}
	return
}

func CombAndFilters(filters ...interface{}) bson.M {
	return CombineFilters("$and", filters...)
}

func CombOrFilters(filters ...interface{}) bson.M {
	return CombineFilters("$or", filters...)
}

// CombineFilters 用来组合多个bson.M filter
// operator :  $and  $or ...
func CombineFilters(operator string, filters ...interface{}) bson.M {
	var cb bson.A

	if len(filters) == 0 {
		return bson.M{}
	}

	for _, f := range filters {
		cb = append(cb, f)
	}

	return bson.M{operator: cb}
}

func UndeleteFilter() bson.M {
	return Eq("deleted_at", nil)
}

func IdFilter(id primitive.ObjectID) bson.M {
	return bson.M{"_id": id}
}

func UnDeletedFilterByID(id primitive.ObjectID) bson.M {
	return CombAndFilters(UndeleteFilter(), IdFilter(id))
}

func Eq(field string, value interface{}) bson.M {
	return bson.M{field: bson.M{"$eq": value}}
}

func Ne(field string, value interface{}) bson.M {
	return bson.M{field: bson.M{"$ne": value}}
}

func Gt(field string, value interface{}) bson.M {
	return bson.M{field: bson.M{"$gt": value}}
}

func Gte(field string, value interface{}) bson.M {
	return bson.M{field: bson.M{"$gte": value}}
}

func In(field string, value interface{}) bson.M {
	return bson.M{field: bson.M{"$in": value}}
}

func Lt(field string, value interface{}) bson.M {
	return bson.M{field: bson.M{"$lt": value}}
}

func Lte(field string, value interface{}) bson.M {
	return bson.M{field: bson.M{"$lte": value}}
}

func Nin(field string, value interface{}) bson.M {
	return bson.M{field: bson.M{"$nin": value}}
}

// get collection by struct
// func (m *Mongoose) coll(sc interface{}) *mongo.Collection {
// 	var (
// 		name string
// 		rt   = reflect.TypeOf(sc)
// 		rv   = reflect.ValueOf(sc)
// 	)
//
// 	collM := rv.MethodByName("CollectionName")
//
// 	if collM.IsValid() {
// 		name = collM.Call([]reflect.Value{})[0].Interface().(string)
// 	} else {
// 		if rt.Kind() == reflect.Ptr {
// 			name = rt.Elem().Name()
// 		} else {
// 			name = rt.Name()
// 		}
// 	}
//
// 	name = toSnakeCase(name)
// 	return m.db.Collection(name)
// }
