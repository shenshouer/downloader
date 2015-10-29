package main
import (
	"bytes"
	"fmt"
)

func main() {
	testdata := []byte("1234567890qwertyuiopasdfghjklzxcvbnm")
	buf := bytes.NewBuffer(testdata)
	fmt.Println(string(buf.Bytes()[:]))

	testdata2 := []byte{}
	if n, err := buf.Write(testdata2); err != nil{
		fmt.Println(err)
	}else{
		fmt.Println("==", n, len(testdata2), testdata2[:])
	}

	fmt.Println(string(buf.Bytes()[:]))
	fmt.Println(len(testdata2), string(testdata2[:]))
	fmt.Println(len(testdata), string(testdata[:]))
}