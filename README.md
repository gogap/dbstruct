# dbstruct

Convert database table to golang struct, and produce the sturct instance

```
package main

import (
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gogap/dbstruct"
)

func nameMapper(name string) (newName string) {
	secs := strings.Split(name, "_")

	for i := 0; i < len(secs); i++ {
		secs[i] = strings.Title(secs[i])
	}

	return strings.Join(secs, "")
}

func main() {
	s, err := dbstruct.New(
		dbstruct.Driver("mysql"),
		dbstruct.DSN("root:password@tcp(localhost:3306)/test?charset=utf8&parseTime=True&loc=Local"),
		dbstruct.NameMapper(nameMapper),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	dbTable, err := s.Describe("users")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(dbTable.NewStruct())
}
```