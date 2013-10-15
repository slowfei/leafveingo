//	Copyright 2013 slowfei And The Contributors All rights reserved.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

//	leafveingo web form的参数解析封装结构
//
//	email	slowfei@foxmail.com
//	createTime 	2013-8-16
//	updateTime	2013-10-9
package leafveingo

import (
	"github.com/slowfei/gosfcore/log"
	"github.com/slowfei/gosfcore/utils/reflect"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	//	判断数组标识的正则
	_arrayTagRex = regexp.MustCompile("\\[\\d*\\]")
)

//	过滤参数封装一些系统的struct
//	在有些设置结构参数时，系统的struct不必要再递归分析
//	@return true 属于过滤字段 false不是过滤的字段
func (lv *sfLeafvein) filterParamPackStructType(valueType reflect.Type) bool {
	result := true
	// strings.Index(fieldValue.Type().String(), "multipart.FileHeader")
	switch valueType.String() {
	case "[]multipart.FileHeader", "[]*multipart.FileHeader", "*multipart.FileHeader", "multipart.FileHeader":
	case "":
		SFLog.Error("can not read type.String() : %v", valueType)
	default:
		result = false
	}
	return result
}

//	根据field的字段名称设置struct字段属性值
//	字段名称可为"type.tag.TagName", 以"."作为子字段的名称分割
func (lv *sfLeafvein) setStructFieldValue(sutValue reflect.Value, fieldName string, value interface{}) {
	isSlice := false
	fieldName = strings.TrimSpace(fieldName)
	sutValueElem := sutValue.Elem()

	if len(fieldName) == 0 {
		return
	}
	if sutValue.Kind() != reflect.Ptr || sutValueElem.Kind() != reflect.Struct {
		// 判断数组
		if sutValue.Kind() == reflect.Ptr && sutValueElem.Kind() == reflect.Slice {
			isSlice = true
		} else {
			return
		}
	}

	//	由于考虑到key的值可能为(type.tag.TagName)，嵌套的赋值，所以需要进行"."的分割，每次获取slice的第一项
	fieldsName := strings.Split(fieldName, ".")
	childFieldName := strings.Title(fieldsName[0])

	if isSlice {
		strIndex := _arrayTagRex.FindString(childFieldName)
		if 0 >= len(strIndex) {
			//	判断如果是数组类型，但是设值的字段名称不是数组的标识则跳过(数组标识Users[0])
			//	由于有可能是需要直接设置数组参数，但是字段名称中(childFieldName = "Files")中未包含"[\d]"的标识
			//	所以尝试直接设值
			SFReflectUtil.SetBaseTypeValue(sutValueElem, value)
			return
		}

		intIndex, e := strconv.Atoi(strIndex[1 : len(strIndex)-1])
		if nil == e && intIndex < sutValueElem.Len() {
			joinLaterFieldName := strings.Join(fieldsName[1:], ".")
			fieldValue := sutValueElem.Index(intIndex)

			switch fieldValue.Kind() {
			case reflect.Ptr:

				switch fieldValue.Elem().Kind() {
				case reflect.Struct:
					if !lv.filterParamPackStructType(fieldValue.Type().Elem()) {
						lv.setStructFieldValue(fieldValue, joinLaterFieldName, value)
					} else {
						SFReflectUtil.SetBaseTypeValue(fieldValue, value)
					}
				default:
					SFReflectUtil.SetBaseTypeValue(fieldValue, value)
				}

			case reflect.Struct:
				if !lv.filterParamPackStructType(fieldValue.Type()) {
					lv.setStructFieldValue(fieldValue.Addr(), joinLaterFieldName, value)
				} else {
					//	如果检测的是系统或则非用户定义的struct就可以直接赋值了，赋值那里已经做了匹配类型才进行赋值的处理
					SFReflectUtil.SetBaseTypeValue(fieldValue.Addr(), value)
				}
			default:
				SFReflectUtil.SetBaseTypeValue(fieldValue.Addr(), value)
			}
		}

	} else {

		//	判断是否为数组标识，如果是的话就删除，例如：Users[0] 删除[0]
		reIndex := _arrayTagRex.FindStringIndex(childFieldName)
		if 0 < len(reIndex) {
			childFieldName = childFieldName[:reIndex[0]]
		}

		//	查找字段
		fieldValue := sutValueElem.FieldByName(childFieldName)

		if fieldValue.IsValid() {

			//	为递归下一个参数做准备，保留后面的参数名(tag.TagName)
			isRec := false
			joinLaterFieldName := ""
			if 1 < len(fieldsName) {
				joinLaterFieldName = strings.Join(fieldsName[1:], ".")
				isRec = true
			}

			switch fieldValue.Kind() {
			case reflect.Ptr:

				if fieldValue.IsNil() {
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
				}

				switch fieldValue.Elem().Kind() {
				case reflect.Struct:
					if isRec && !lv.filterParamPackStructType(fieldValue.Type().Elem()) {
						lv.setStructFieldValue(fieldValue, joinLaterFieldName, value)
					} else {
						SFReflectUtil.SetBaseTypeValue(fieldValue, value)
					}
				case reflect.Slice:
					lv.setStructFieldValue(fieldValue, fieldName, value)
				default:
					SFReflectUtil.SetBaseTypeValue(fieldValue, value)
				}

			case reflect.Struct:

				if isRec && !lv.filterParamPackStructType(fieldValue.Type()) {
					lv.setStructFieldValue(fieldValue.Addr(), joinLaterFieldName, value)
				} else {
					//	如果检测的是系统或则非用户定义的struct就可以直接赋值了，赋值那里已经做了匹配类型才进行赋值的处理
					SFReflectUtil.SetBaseTypeValue(fieldValue.Addr(), value)
				}

			case reflect.Slice:

				lv.setStructFieldValue(fieldValue.Addr(), fieldName, value)
			default:

				SFReflectUtil.SetBaseTypeValue(fieldValue.Addr(), value)
			}

		}
	}
}

//	根据url和form参数信息创建一个struct体，主要为了指针的字段，因为如果指针没有初始化直接User.Type就会出现空的错误
//	此方法就是防止这个错误发生，就算是没有参数值设置，也会新建一个指针地址。
//	还有的就是设置Slice的数量，以便再设置参数值的时候可以根据[1]中括号的下标进行对应的设置。
func (lv *sfLeafvein) newStructPtr(structType reflect.Type, urlValues url.Values) reflect.Value {
	structValue := reflect.New(structType)
	//	考虑到结构内包含指针，如果不进行new的话直接(.)会爆出空指针异常，所以这里需要遍历每个对象
	if 0 == len(urlValues) {
		return structValue
	}

	//	用于存储属于字段名的集合，例如：users[0]:true; users[1]:trye;...
	//	后期还需要将数组标识的"[\d]"去除合并成数组需要生成的数量
	arrayFields := make(map[string]bool)

	for key, _ := range urlValues {
		reIndex := _arrayTagRex.FindAllStringIndex(key, -1)
		if 0 < len(reIndex) {
			endSlice := reIndex[len(reIndex)-1]
			subIndex := endSlice[len(endSlice)-1]
			arrayFields[key[:subIndex]] = true
		}
	}

	//	递归函数，用于遍历filed，根据类型初始化指针与数组
	var findFiledFunc func(structV reflect.Value, layerLevel int)
	//	设置slice的大小
	var setSliceSizeFunc func(value reflect.Value, fieldName string, layerLevel int)

	//	记录当前数组操作的层的级别和数组的下标
	crtArrayLayerLlIndex := make(map[int]int)

	setSliceSizeFunc = func(value reflect.Value, fieldName string, layerLevel int) {
		if 0 < len(arrayFields) {
			sliceSizeMap := make(map[string]bool)
			for k, _ := range arrayFields {

				keySplits := strings.Split(k, ".")
				keySplitsCount := len(keySplits)
				if layerLevel < keySplitsCount {
					//	层级别的意思是("users[0].array[0].uuid") 根据递归层次调用的不同根据"."分割获取相应的字段名称
					fieldTag := keySplits[layerLevel]

					isContinue := false
					//	由于每次进行数组循环设值的时候会存储一次操作的下标和层级，所以用次来判断("users[0].array[0].uuid")父级的层级是否属于当前字段的级别
					//	例如：当前操作的是(array[0]的子字段uuid)就需要匹配前面的(users[0].array[0])是否是相同的。
					for kLayerL, vIndex := range crtArrayLayerLlIndex {
						layerFieldKey := keySplits[kLayerL]
						strIndex := _arrayTagRex.FindString(layerFieldKey)
						intIndex, e := strconv.Atoi(strIndex[1 : len(strIndex)-1])
						if nil != e || intIndex != vIndex {
							isContinue = true
						}
					}
					if isContinue {
						continue
					}

					//	去除参数key的"[\d]"进行比较 例如：Users[0] 保留"Users"跟 childFieldName进行对比，匹配正确才加入sliceSizeMap
					tagIndex := _arrayTagRex.FindStringIndex(fieldTag)
					if 0 < len(tagIndex) {
						tagMate := strings.Title(fieldTag[:tagIndex[0]])
						if tagMate == fieldName {
							sliceSizeMap[fieldTag] = true
						}
					}

				}
			}

			sliceSize := len(sliceSizeMap)
			if 0 < sliceSize {
				valueElem := value.Elem()
				valueElem.Set(reflect.MakeSlice(valueElem.Type(), sliceSize, sliceSize))
				//	如果不是struct就没有必要执行findFiledFunc()进行字段的查询了
				if reflect.Struct == valueElem.Index(0).Kind() && !lv.filterParamPackStructType(valueElem.Type()) {
					for j := 0; j < sliceSize; j++ {
						crtArrayLayerLlIndex[layerLevel] = j
						findFiledFunc(valueElem.Index(j).Addr(), layerLevel+1)
					}
					delete(crtArrayLayerLlIndex, layerLevel)
				}

			}

		}
	}

	findFiledFunc = func(structV reflect.Value, layerLevel int) {
		if structV.Kind() != reflect.Ptr || structV.Elem().Kind() != reflect.Struct {
			return
		}

		structVElem := structV.Elem()
		filedCount := structVElem.NumField()
		for i := 0; i < filedCount; i++ {
			childField := structVElem.Field(i)
			childFieldSF := structVElem.Type().Field(i)
			childFieldName := childFieldSF.Name

			if !childField.CanSet() {
				continue
			}

			switch childField.Kind() {
			case reflect.Ptr:
				//	由于当前方法创建的是一个新的struct 所以不存在指针IsNil()的字段，所以直接初始化
				childField.Set(reflect.New(childField.Type().Elem()))

				switch childField.Type().Elem().Kind() {
				case reflect.Array:
					//	由于array是一个固定的数值集合，所以不进行处理
				case reflect.Slice:
					setSliceSizeFunc(childField, childFieldName, layerLevel)
				default:
					if childField.Elem().Kind() == reflect.Struct && !lv.filterParamPackStructType(childField.Type().Elem()) {
						findFiledFunc(childField, layerLevel+1)
					}
				}
			case reflect.Struct:
				if !lv.filterParamPackStructType(childField.Type()) {
					findFiledFunc(childField.Addr(), layerLevel+1)
				}
			case reflect.Array:
				//	由于array是一个固定的数值集合，所以不进行处理
			case reflect.Slice:
				setSliceSizeFunc(childField.Addr(), childFieldName, layerLevel)
			}
		}

	}
	//	执行调用
	findFiledFunc(structValue, 0)

	return structValue
}
