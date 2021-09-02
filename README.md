### mongoose document

### Example

```golang
package main

import (
	"log"
	"github.com/wuruipeng404/mongoose"
	"go.mongodb.org/mongo-driver/bson"
)

var odm *mongoose.Mongo

// get mongo client
func init() {

	var err error

	if odm, err = mongoose.Open(&mongoose.Options{
		User:       "user",
		Password:   "password",
		Host:       "localhost",
		Port:       27017,
		DBName:     "your-db",
		DriverOpts: nil, // and you can add driver client options
	}); err != nil {
		log.Fatalf("connect mongoose failed:%s", err)
	}
}

// define your schema
type YourSchema struct {
	mongoose.Document `bson:",inline"`
	FieldA            string     `bson:"field_a,omitempty"`
	FieldB            int        `bson:"field_b,omitempty"`
	Son               *SubSchema `bson:"son,omitempty"`
	*SubSchema        `bson:",inline"` // inline field
}

type SubSchema struct {
	FieldC string `bson:"field_c,omitempty"`
	FieldD string `bson:"field_d,omitempty"`
}

// CollectionName impl mongoose.IDocument interface
func (*YourSchema) CollectionName() string {
	return "your_collection"
}

func Create() {
	// will auto add create time for now , and also you can set your time 
	// and all create method will auto find collection name
	odm.InsertOne(&YourSchema{
		FieldA: "test",
		FieldB: 3,
		SubSchema: &SubSchema{
			FieldC: "111",
			FieldD: "222",
		},
	})

	// if your want to use driver method
	odm.GetCollectionWithName("your collection").InsertOne(bson.M{})
}

func HaveFilter() {
	// update find delete
	// id support string (primitive.ObjectID.hex()) and ObjectID
	var result YourSchema
	odm.FindByID(id, &result)

	// filter support bson and IDocument
	// This is equivalent
	var result2 []YourSchema
	odm.Find(YourSchema{FieldA: "zhangsan"}, &result2)
	odm.Find(bson.M{"field_a": "zhangsan"}, &result2)
	// this.is sugar
	odm.Find(mongoose.Eq("field_a", "zhangsan"), &result2)
}

// Special the delete method have a little bit different
func Special() {

	// when your filter is bson or not a IDocument
	// you need add option CollectionName

	option := mongoose.DeleteOption{
		CollectionName: "your collection",
		DriverOptions:  nil, // driver delete options []
	}
	odm.Delete(mongoose.Eq("field_a", "aaa"), option)
}

```
