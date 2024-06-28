package main

import (
	"fmt"
	"os"

	binlog_modifier "github.com/zing22845/go-binlog-modifier"
)

func main() {
	filePath := "./mysql-bin.000004"
	rf, _ := os.Open(filePath)
	// open target
	f, _ := os.OpenFile("./newbinlog.000004", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	// 创建 OnEventFunc
	// 创建 BinlogModifier
	bm := &binlog_modifier.BinlogModifier{
		Reader:           rf,
		WriterAt:         f,
		IsVerifyChecksum: true,
	}
	bm.DisableForeignKeyChecks()

	err := bm.Run()
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		os.Exit(1)
	}
}
