/*
* @Author: Rumple
* @Email: ruipeng.wu@cyclone-robotics.com
* @DateTime: 2021/8/27 15:06
 */

package mongoose

import (
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	DeleteOption struct {
		CollectionName string
		DriverOptions  []*options.DeleteOptions
	}

	FindOneDeleteOption struct {
		CollectionName string
		DriverOptions  []*options.FindOneAndDeleteOptions
	}
)
