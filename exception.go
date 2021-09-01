/*
* @Author: Rumple
* @Email: ruipeng.wu@cyclone-robotics.com
* @DateTime: 2021/8/27 14:40
 */

package mongoose

import (
	"errors"
)

var (
	CollectionNameNotFound = errors.New("when filter is not IDocument, then need CollectionName option")
	InvalidDocument        = errors.New("invalid document impl IDocument interface at least")
)
