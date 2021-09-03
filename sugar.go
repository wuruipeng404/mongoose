/*
* @Author: Rumple
* @Email: ruipeng.wu@cyclone-robotics.com
* @DateTime: 2021/8/24 15:54
 */

package mongoose

import (
	"reflect"
	"strings"
	"unsafe"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// var (
// 	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
// 	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
// )
//
// func ToSnakeCase(str string) string {
// 	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
// 	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
// 	return strings.ToLower(snake)
// }

// ParseFilter 检查filter并进行转化. 转化为 driver 支持的格式
// 如果本身就是 bson.M 或 bson.D 系列则不进行转化
// 如果是结构体或者结构体指针则进行转化 (支持复杂结构体)
func ParseFilter(filter interface{}) interface{} {
	var (
		finallyF bson.M
		refType  = reflect.TypeOf(filter).String()
	)

	if refType == "primitive.M" || refType == "primitive.D" {
		return filter
	}
	finallyF = ConvertFilter(filter, "")
	return finallyF
}

// ConvertFilter convert Struct or Ptr to bson.M
func ConvertFilter(v interface{}, fatherTag string) bson.M {
	var (
		result   = bson.M{}
		rv       = reflect.ValueOf(v)
		rt       = reflect.TypeOf(v)
		iterRv   reflect.Value
		numField int
	)

	if rv.IsZero() {
		return result
	}

	// 获取 field 数量
	if rv.Kind() == reflect.Ptr {
		numField = rv.Elem().NumField()
		iterRv = rv.Elem()
	} else {
		numField = rv.NumField()
		iterRv = reflect.New(rv.Type()).Elem()
		iterRv.Set(rv)
	}

	for i := 0; i < numField; i++ {
		var (
			nextFatherTag  string
			currentBsonTag string
			bsonTag        string
			currentValue   reflect.Value
			currentField   reflect.StructField
			ignoreZero     = true
		)

		if rv.Kind() == reflect.Ptr {
			currentField = rt.Elem().Field(i)
		} else {
			currentField = rt.Field(i)
		}

		tmp := iterRv.Field(i)
		currentValue = reflect.NewAt(tmp.Type(), unsafe.Pointer(tmp.UnsafeAddr())).Elem()

		kind := currentValue.Kind()
		rValue := currentValue.Interface()

		switch kind {
		case reflect.Interface, reflect.Func, reflect.Chan, reflect.Invalid, reflect.UnsafePointer:
			continue

		default:
			bsonTag = currentField.Tag.Get("bson")

			// 获取 下一级处理的tag名称
			if bsonTag == "" {
				currentBsonTag = currentField.Name
			} else if bsonTag == "-" {
				continue
			} else if strings.Contains(bsonTag, "inline") {
				currentBsonTag = ""
			} else {
				if strings.Contains(bsonTag, ",") {
					currentBsonTag = strings.Split(bsonTag, ",")[0]
				} else {
					currentBsonTag = bsonTag
				}
			}

			if fatherTag == "" {
				nextFatherTag = currentBsonTag
			} else {
				nextFatherTag = fatherTag + "." + currentBsonTag
			}

			if !strings.Contains(bsonTag, "omitempty") {
				ignoreZero = false
			}

			// 有子结构的继续进行 字段遍历
			if ((kind == reflect.Ptr && getElemType(rValue).Kind() == reflect.Struct) || kind == reflect.Struct) &&
				getElemType(rValue).String() != "time.Time" {
				// 递归处理
				for nk, nv := range ConvertFilter(currentValue.Interface(), nextFatherTag) {
					result[nk] = nv
				}
				// 没有子结构的则进行取值
			} else if currentValue.Kind() == reflect.Slice {

				for sk, sv := range ConvertSliceFilter(currentValue, currentBsonTag) {
					result[sk] = sv
				}
			} else if kind == reflect.Map {

				for iter := currentValue.MapRange(); iter.Next(); {
					k := iter.Key().Interface().(string)
					vl := iter.Value().Interface()
					result[nextFatherTag+"."+k] = vl
				}

			} else {
				if currentValue.IsZero() && ignoreZero {
					continue
				}

				result[nextFatherTag] = currentValue.Interface()
			}
		}

	}
	return result
}

func ConvertSliceFilter(value reflect.Value, tag string) bson.M {
	var (
		result    = bson.M{}
		in        []interface{}
		subResult []bson.M
	)

	if value.Len() == 0 {
		return result
	}

label:
	for i := 0; i < value.Len(); i++ {
		erv := value.Index(i)
		// ert := erv.Type()

		// 判断slice的元素类型
		switch erv.Kind() {
		// 结构体需要继续处理
		case reflect.Struct, reflect.Ptr:
			if erv.Type().String() == "time.Time" || erv.Type().String() == "*time.Time" {
				in = append(in, erv.Interface())
				continue
			}

			if r := ConvertFilter(erv.Interface(), ""); len(r) != 0 {
				subResult = append(subResult, r)
			}

		// 因为所有元素类型都应该是一样的 这里有一个 是如下类型 直接 break 掉
		case reflect.Interface, reflect.Func, reflect.Chan, reflect.Array, reflect.Invalid, reflect.UnsafePointer:
			break label

		case reflect.Map:
			// todo: support
			break label
		// 基础数据类型
		default:
			in = append(in, erv.Interface())
		}
	}

	if len(in) > 0 {
		return In(tag, in)
	}

	if len(subResult) > 0 {

		ba := bson.A{}
		for _, sub := range subResult {

			tmpBm := bson.M{}
			for k, v := range sub {
				tmpBm[tag+"."+k] = v
			}
			ba = append(ba, tmpBm)
		}

		result["$or"] = ba
	}

	return result
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

func Set(doc interface{}) bson.M {
	return bson.M{"$set": doc}
}

// get collection by struct
// func (m *Mongo) coll(sc interface{}) *mongo.Collection {
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

func getElemType(a interface{}) reflect.Type {
	for t := reflect.TypeOf(a); ; {
		switch t.Kind() {
		case reflect.Ptr, reflect.Slice:
			t = t.Elem()
		default:
			return t
		}
	}
}

func getCollNameForFind(findRes interface{}) (string, error) {
	if n := reflect.New(getElemType(findRes)).MethodByName("CollectionName"); n.IsValid() {
		name := n.Call(nil)[0].Interface().(string)
		if name == "" {
			return "", CollectionNameNotFound
		}
		return name, nil
	} else {
		return "", InvalidDocument
	}
}

func getCollNameForOpt(filter, opt interface{}) (string, error) {
	var (
		ok  bool
		doc IDocument
	)

	if doc, ok = filter.(IDocument); ok {
		return doc.CollectionName(), nil
	}

	ref := reflect.ValueOf(opt)

	if !ref.IsValid() {
		return "", CollectionNameNotFound
	}

	if fn := ref.Elem().FieldByName("CollectionName"); fn.IsZero() {
		return "", CollectionNameNotFound
	} else {
		return fn.String(), nil
	}
}
