package main

import (
	"fmt"
	"github.com/ugorji/go/codec"
)

type UserStruct struct {
	ID   int32
	Name string
}

var (
	b []byte
	mh codec.MsgpackHandle
)

func main() {
	user := UserStruct{9, "abcd"}
	//关键调用
	mh.StructToArray = true

	enc := codec.NewEncoderBytes(&b, &mh)
	err := enc.Encode(user)
	if err == nil {
		fmt.Println("data:", b)
	} else {
		fmt.Println("err:", err)
	}

	dec := codec.NewDecoderBytes(b, &mh)
	var new_user UserStruct
	err = dec.Decode(&new_user)
	if err == nil {
		fmt.Println("new_user:", new_user)
	} else {
		fmt.Println("err:", err)
	}

}