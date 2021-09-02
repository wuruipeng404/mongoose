/*
* @Author: Rumple
* @Email: ruipeng.wu@cyclone-robotics.com
* @DateTime: 2021/8/24 14:32
 */

package mongoose

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IDocument interface {
	PreCreate()
	PreUpdate()
	PreDelete()
	CollectionName() string
}

type Document struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CreatedAt *time.Time         `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt *time.Time         `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"-"`
}

func (d *Document) PreCreate() {
	d.ID = primitive.NewObjectID()
	d.CreatedAt = Now()
	d.UpdatedAt = Now()
}

func (d *Document) PreUpdate() {
	d.UpdatedAt = Now()
}

func (d *Document) PreDelete() {
	d.DeletedAt = Now()
}

func Now() *time.Time {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &now
}
