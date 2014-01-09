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
//
//  Create on 2013-8-16
//  Update on 2013-10-23
//  Email  slowfei@foxmail.com
//  Home   http://www.slowfei.com

//	leafveingo web form的参数解析封装结构
package leafveingo

import (
	// "fmt"
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

//	设置参数值
//
//	@param sutValue 	字段的反射对象
//	@param fieldName 	字段名
//	@param value	 	需要设置的值
//
func (lv *sfLeafvein) setParamValue(fieldValue reflect.Value, fieldName string, value interface{}) {
	err := SFReflectUtil.SetBaseTypeValue(fieldValue, value)
	if nil != err {
		lvLog.Error("%s set value error: %s", fieldName, err.Error())
	}
}

//	过滤参数封装一些系统的struct
//	在有些设置结构参数时，系统的struct不必要再递归分析
//	@return true 属于过滤字段 false不是过滤的字段
func (lv *sfLeafvein) filterParamPackStructType(valueType reflect.Type) bool {
	result := true
	// strings.Index(fieldValue.Type().String(), "multipart.FileHeader")
	switch valueType.String() {
	case "[]multipart.FileHeader", "[]*multipart.FileHeader", "*multipart.FileHeader", "multipart.FileHeader":
	case "":
		lvLog.Error("can not read type.String() : %v", valueType)
	default:
		result = false
	}
	return result
}

//	针对字段进行设值
//
//	@param fieldValue		字段反射对象
//	@param fieldName		字段名，以便递归寻找下个设值字段
//	@param fieldSplitName	字段名分割集合，以"."进行分割，主要是为了递归子字段进行拼接传递
//	@param value			设置值
//
func (lv *sfLeafvein) setFieldValue(fieldValue reflect.Value, fieldName string, fieldSplitName []string, value interface{}) {

	if fieldValue.IsValid() {

		//	为递归下一个参数做准备，保留后面的参数名(tag.TagName)
		isRec := false
		joinLaterFieldName := ""
		if 1 < len(fieldSplitName) {
			joinLaterFieldName = strings.Join(fieldSplitName[1:], ".")
			isRec = true
			//	进入这里表明还需要进行一次字段查询，所以需要进行递归操作，直到截取到最后一位的参数名标识(TagName)
		}

		switch fieldValue.Kind() {
		case reflect.Ptr:

			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}

			//	指针与非指针区分开，主要是在进行参数设值的时候对应与设置值相同的类型，减免指针的过多操作。
			switch fieldValue.Elem().Kind() {
			case reflect.Struct:
				if isRec && !lv.filterParamPackStructType(fieldValue.Type().Elem()) {
					lv.setStructFieldValue(fieldValue, joinLaterFieldName, value)
				} else {
					lv.setParamValue(fieldValue, fieldName, value)
				}

			case reflect.Slice:
				//	如果属于切片类型，传递fieldName主要是在操作一遍集合元素的赋值，因为不是结构类型无需再向下查找
				lv.setStructFieldValue(fieldValue, fieldName, value)
			default:
				lv.setParamValue(fieldValue, fieldName, value)
			}

		case reflect.Struct:

			if isRec && !lv.filterParamPackStructType(fieldValue.Type()) {
				lv.setStructFieldValue(fieldValue, joinLaterFieldName, value)
			} else {
				//	如果检测的是系统或则非用户定义的struct就可以直接赋值了，赋值那里已经做了匹配类型才进行赋值的处理
				lv.setParamValue(fieldValue, fieldName, value)
			}
		case reflect.Slice:
			lv.setStructFieldValue(fieldValue, fieldName, value)
		default:
			lv.setParamValue(fieldValue, fieldName, value)
		}

	}
}

//	根据field的字段名称设置struct字段属性值
//	字段名称可为"type.tag.TagName", 以"."作为子字段的名称分割
//
//	@param sutValue		被设值的结构反射类型
//	@param fieldName	字段名(type.tag.TagName) or (tagnName)
//	@param value		设值值
func (lv *sfLeafvein) setStructFieldValue(sutValue reflect.Value, fieldName string, value interface{}) {
	fieldName = strings.TrimSpace(fieldName)
	sutValueElem := reflect.Indirect(sutValue)

	if len(fieldName) == 0 {
		return
	}

	//	由于考虑到key的值可能为(type.tag.TagName)，嵌套的赋值，所以需要进行"."的分割，每次获取slice的第一项
	fieldsName := strings.Split(fieldName, ".")
	childFieldName := strings.Title(fieldsName[0])

	if sutValueElem.Kind() == reflect.Slice {
		//	集合字段的设值操作

		strIndex := _arrayTagRex.FindString(childFieldName)
		if 0 >= len(strIndex) {
			//	判断如果是数组类型，但是设值的字段名称不是数组的标识则跳过(数组标识Users[0])
			//	由于有可能是需要直接设置数组参数，但是字段名称中(childFieldName = "Files")中未包含"[\d]"的标识
			//	所以尝试直接设值
			lv.setParamValue(sutValueElem, fieldName, value)
			return
		}

		intIndex, e := strconv.Atoi(strIndex[1 : len(strIndex)-1])
		if nil == e && intIndex < sutValueElem.Len() {
			fieldValue := sutValueElem.Index(intIndex)

			lv.setFieldValue(fieldValue, fieldName, fieldsName, value)
		}

	} else {

		//	判断是否为数组标识，如果是的话就删除，例如：Users[0] 删除[0] = Users
		//	这样便于FieldByName查找到相应的字段信息
		reIndex := _arrayTagRex.FindStringIndex(childFieldName)
		if 0 < len(reIndex) {
			childFieldName = childFieldName[:reIndex[0]]
		}

		//	查找字段
		fieldValue := sutValueElem.FieldByName(childFieldName)

		lv.setFieldValue(fieldValue, fieldName, fieldsName, value)
	}
}

//	根据url和form参数信息创建一个struct体，如果存在集合结构会根据form参数信息make分配切片大小。
//
//	@param structType 需要操作的的结构
//	@param urlValues  url或form请求参数(主要为了操作slice的创建元素)
//	@return	返回创建好的函数反射对象信息（指针类型的）
//
func (lv *sfLeafvein) newStructPtr(structType reflect.Type, urlValues url.Values) reflect.Value {
	//	考虑到结构内包含指针，如果不进行new的话直接(.)会爆出空指针异常，所以这里需要遍历每个对象

	if structType.Kind() != reflect.Struct {
		return reflect.Zero(structType)
	}

	structValue := reflect.New(structType)

	if 0 == len(urlValues) {
		return structValue
	}

	//	递归函数，用于遍历filed，根据类型初始化指针与数组
	var findFiledFunc func(structV reflect.Value, layerLevel int, joinFieldName string)
	//	设置slice的大小
	var setSliceSizeFunc func(value reflect.Value, fieldName string, layerLevel int, joinFieldName string)

	//  记录当前数组操作的层的级别和数组的下标
	crtArrayLayerLlIndex := make(map[int]int)

	setSliceSizeFunc = func(value reflect.Value, fieldName string, layerLevel int, joinFieldName string) {

		if 0 < len(urlValues) && reflect.Slice == reflect.Indirect(value).Kind() {
			//	如果是slice类型则根据form的name与递归连接的字段名进行比较来统计需要make的slice size

			//	存储相同的切片标识
			sliceSizeMap := make(map[string]bool)

			for k, _ := range urlValues {

				//	由于joinFieldName 是根据结构的字段名进行拼接的，所以在操作form的name(参数名)时需要去除"[\d]"集合的标识来进行比较
				paramName := _arrayTagRex.ReplaceAllString(k, "")

				//	根据递归级别分割paramName,以便与joinFieldName进行比较。
				//	因为有可能paramName的级别大于joinFieldName
				index := -1
				levelIndex := -1 // 由于级别是从0开始，所以-1
				for i := 0; i < len(paramName); i++ {
					if paramName[i] == '.' {
						levelIndex++
					}
					if levelIndex == layerLevel {
						index = i
						break
					}
				}

				if -1 < index {
					paramName = paramName[:index]
				}

				//	比较并存储集合标识的参数名
				if 0 != len(paramName) && strings.ToLower(paramName) == strings.ToLower(joinFieldName) {
					//	处理相同的集合标识

					keySplits := strings.Split(k, ".")

					if layerLevel < len(keySplits) {

						//	验证上级的数组元素是否为当前记录的操作下标
						//	这里处理的可能性是集合下的元素为集合类型oneUser.hobbys[1].names[0]
						//	如果不进行处理当参数字段为oneUser.hobbys[0].names[0]的时候就会出现相同的子集合个数
						if vIndex, ok := crtArrayLayerLlIndex[layerLevel-1]; ok {
							layerFieldKey := keySplits[layerLevel-1]
							strIndex := _arrayTagRex.FindString(layerFieldKey)
							if 0 == len(strIndex) {
								continue
							}
							intIndex, e := strconv.Atoi(strIndex[1 : len(strIndex)-1])
							if nil != e || intIndex != vIndex {
								continue
							}
						}

						sliceSizeMap[keySplits[layerLevel]] = true
					}
				}
			}

			sliceSize := len(sliceSizeMap)
			if 0 < sliceSize {
				valueElem := reflect.Indirect(value)
				valueElem.Set(reflect.MakeSlice(valueElem.Type(), sliceSize, sliceSize))

				//	继续执行集合子元素的遍历
				//	reflect.Struct == valueElem.Index(0).Kind() &&
				if !lv.filterParamPackStructType(valueElem.Type()) {
					for j := 0; j < sliceSize; j++ {
						childIndex := valueElem.Index(j)

						if reflect.Ptr == childIndex.Kind() && childIndex.IsNil() {
							//	初始化子元素指针的操作
							childIndex.Set(reflect.New(childIndex.Type().Elem()))
						}
						//	记录当前操作层级的集合下标，以便在递归的时候能够识别出当前操作下标进行判断
						crtArrayLayerLlIndex[layerLevel] = j
						findFiledFunc(childIndex, layerLevel+1, joinFieldName)
					}
					delete(crtArrayLayerLlIndex, layerLevel)
				}

			}

		}

	}

	findFiledFunc = func(structV reflect.Value, layerLevel int, joinFieldName string) {
		structV = reflect.Indirect(structV)
		if reflect.Struct != structV.Kind() {
			return
		}

		if 0 != len(joinFieldName) {
			joinFieldName += "."
		}

		//	遍历结构字段寻找需要初始化的slice
		filedCount := structV.NumField()
		for i := 0; i < filedCount; i++ {
			childField := structV.Field(i)
			childFieldType := childField.Type()
			childFieldName := structV.Type().Field(i).Name

			if !childField.CanSet() {
				continue
			}

			if reflect.Ptr == childFieldType.Kind() && childField.IsNil() {
				//	初始化子结构指针的操作
				childField.Set(reflect.New(childField.Type().Elem()))
			}

			switch reflect.Indirect(childField).Kind() {
			case reflect.Struct:
				if !lv.filterParamPackStructType(childField.Type()) {
					findFiledFunc(childField, layerLevel+1, joinFieldName+childFieldName)
				}
			case reflect.Slice:
				setSliceSizeFunc(childField, childFieldName, layerLevel, joinFieldName+childFieldName)
			}
		}

	}
	//	执行调用
	findFiledFunc(structValue, 0, "")

	return structValue
}
