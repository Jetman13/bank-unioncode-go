package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

type BankInfo struct {
	Bank     string //银行名称
	Province string
	City     string
	BankName string //支行名称
	UnionNo  string //联行号
}

type UnionCode interface {
	fetchAllBranch() []BankInfo
}

func main() {
	fileDir := "/tmp/unioncode.csv"
	t1 := time.Now() // get current time
	f, err := os.Create(fileDir)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM

	var hebeiChannel = hebeiChannel{}
	allBank := hebeiChannel.fetchAllBranch()

	csvHead := []string{"省份", "城市", "银行名称", "支行名称", "联行号"}
	wr := csv.NewWriter(f)
	wr.Write(csvHead)

	for _, bankInfo := range allBank {
		wr.Write([]string{bankInfo.Province, bankInfo.City, bankInfo.Bank, bankInfo.BankName, bankInfo.UnionNo})
	}
	wr.Flush()
	elapsed := time.Since(t1)
	fmt.Println("App elapsed: ", elapsed)
}
