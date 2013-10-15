package leafveingo

import (
	"fmt"
	"mime/multipart"
	"net/url"
	"reflect"
	"testing"
)

type User struct {
	Name    string
	Sex     int
	Type    *UserType
	Arrays  *[]UserArray
	File    multipart.FileHeader
	Files   []multipart.FileHeader
	FilePrt *multipart.FileHeader
	Files2  []*multipart.FileHeader
}

type UserArray struct {
	ArrayUUID string
	Strs      []string
}

type UserType struct {
	TypeName string
	TypeUUID string
	Tag      *TypeTag
}

type TypeTag struct {
	TagName string
}

func TestSetStructFieldValue(t *testing.T) {
	leafvein := &sfLeafvein{}

	//	基本设值测试
	// var params struct{ Users User }
	// fmt.Println("基本设值测试")
	// leafvein.setStructFieldValue(reflect.ValueOf(&params), "users.Name", "user-slowfei_1")
	// leafvein.setStructFieldValue(reflect.ValueOf(&params), "users.Type.TypeName", "type-slowfei_1")
	// leafvein.setStructFieldValue(reflect.ValueOf(&params), "users.Type.TypeUUID", "UUID-slowfei_1")
	// leafvein.setStructFieldValue(reflect.ValueOf(&params), "users.Type.Tag.TagName", "tag-slowfei_1")
	// fmt.Println(params.Users.Name)
	// fmt.Println(params.Users.Type.TypeName)
	// fmt.Println(params.Users.Type.TypeUUID)
	// fmt.Println(params.Users.Type.Tag.TagName)

	//	集合设值测试
	var params struct {
		Users  []User
		TagStr []string
	}
	keys := url.Values{}
	keys.Add("users[0].name", "sl_name_1")
	keys.Add("users[1].name", "sl_name_2")
	keys.Add("users[2].name", "sl_name_3")
	keys.Add("users[0].type.typeName", "sl_0_type_name")
	keys.Add("users[0].type.typeUUID", "sl_0_type_UUID")
	keys.Add("users[0].type.tag.tagName", "sl_0_type_tagName")
	keys.Add("users[1].type.typeName", "sl_1_type_name")
	keys.Add("users[1].type.typeUUID", "sl_1_type_UUID")
	keys.Add("users[1].type.tag.tagName", "sl_1_type_tagName")
	keys.Add("users[2].arrays[0].strs[0]", "sl_array_2_0_strs_0")
	keys.Add("users[0].arrays[0].strs[0]", "sl_array_0_0_strs_0")
	keys.Add("users[0].arrays[1].arrayUUID", "sl_array_0_1")
	keys.Add("users[0].arrays[2].arrayUUID", "sl_array_0_2")
	keys.Add("users[0].arrays[2].strs[0]", "sl_array_0_2_strs_0")
	keys.Add("users[0].arrays[2].strs[1]", "sl_array_0_2_strs_1")
	keys.Add("users[1].arrays[0].arrayUUID", "sl_array_1_0")
	keys.Add("users[1].arrays[1].arrayUUID", "sl_array_1_1")
	keys.Add("users[0].Files[0]", "")
	keys.Add("tagStr[0]", "tag_0")
	keys.Add("tagStr[1]", "tag_1")
	keys.Add("tagStr[2]", "tag_2")
	keys.Add("users.name", "sf_name_nil")

	params = leafvein.newStructPtr(reflect.TypeOf(params), keys).Elem().Interface().(struct {
		Users  []User
		TagStr []string
	})
	// leafvein.setStructFieldValue(reflect.ValueOf(&params), "users[0].name", "sl_name_1")
	// fmt.Println(params.Users[0].Name)
	for k, v := range keys {
		leafvein.setStructFieldValue(reflect.ValueOf(&params), k, v[0])
	}

	leafvein.setStructFieldValue(reflect.ValueOf(&params), "users[0].File", multipart.FileHeader{Filename: "fileName-sf"})
	leafvein.setStructFieldValue(reflect.ValueOf(&params), "users[0].Files[0]", multipart.FileHeader{Filename: "fileName_array-sf"})
	leafvein.setStructFieldValue(reflect.ValueOf(&params), "users[0].FilePrt", multipart.FileHeader{Filename: "fileName_prt-sf"})
	files2 := make([]*multipart.FileHeader, 1, 1)
	files2[0] = &multipart.FileHeader{Filename: "fileName_prt-sf"}
	leafvein.setStructFieldValue(reflect.ValueOf(&params), "users[0].Files2", files2)

	fmt.Println("----------------------------")
	fmt.Println("Users-Count:", len(params.Users))
	fmt.Println("tagStr-Count:", len(params.TagStr))
	fmt.Println(" ")
	for _, v := range params.Users {
		fmt.Println("file.Filename:", v.File.Filename)
		fmt.Println("files:", v.Files)
		for _, v2 := range v.Files2 {
			fmt.Println("files2:", v2)
		}
		fmt.Println("fileprt.Filename:", v.FilePrt.Filename)
		fmt.Println("Name:", v.Name)
		fmt.Println("TypeName:", v.Type.TypeName)
		fmt.Println("TypeUUID:", v.Type.TypeUUID)
		fmt.Println("TagName:", v.Type.Tag.TagName)
		fmt.Println("   arrayNum:", len(*(v.Arrays)))
		for _, v2 := range *(v.Arrays) {
			fmt.Println("      ArrayUUID:", v2.ArrayUUID)
			fmt.Println("           array.strs:num", len(v2.Strs))
			for _, v3 := range v2.Strs {
				fmt.Println("               Strs:", v3)
			}

		}

		fmt.Println("...")
	}
	fmt.Println("----------------------------")
	for _, v := range params.TagStr {
		fmt.Println(v)
		fmt.Println("...")
	}
}
