package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

/**
河北银行网站联行号渠道
*/

var provinces = [][]string{
	{"1000", "北京市"},
	{"1100", "天津市"},
	{"1210", "河北"},
	{"1610", "山西"},
	{"1910", "内蒙古"},
	{"2210", "辽宁"},
	{"2410", "吉林"},
	{"2610", "黑龙江"},
	{"2900", "上海市"},
	{"3010", "江苏"},
	{"3310", "浙江"},
	{"3610", "安徽"},
	{"3910", "福建"},
	{"4210", "江西"},
	{"4510", "山东"},
	{"4910", "河南"},
	{"5210", "湖北"},
	{"5510", "湖南"},
	{"5810", "广东"},
	{"6110", "广西"},
	{"6410", "海南"},
	{"6510", "四川"},
	{"6530", "重庆市"},
	{"7010", "贵州"},
	{"7310", "云南"},
	{"7700", "西藏"},
	{"7910", "陕西"},
	{"8210", "甘肃"},
	{"8510", "青海"},
	{"8710", "宁夏"},
	{"8810", "新疆"},
}

var bankNames = [][]string{
	{"104", "中国银行"},
	{"102", "中国工商银行"},
	{"103", "中国农业银行"},
	{"105", "中国建设银行"},
	{"301", "交通银行"},
	{"308", "招商银行"},
	{"304", "华夏银行"},
	{"318", "渤海银行"},
	{"306", "广发银行"},
	{"307", "平安银行"},
	{"302", "中信银行"},
	{"303", "中国光大银行"},
	{"305", "中国民生银行"},
	{"309", "兴业银行"},
	{"310", "上海浦东发展银行"},
	{"403", "中国邮政储蓄银行"},
	{"-1", "其他银行"},
}

type KColl struct {
	Id     string  `xml:"id,attr"`
	Append bool    `xml:"append,attr"`
	Fields []Field `xml:"field"`
	IColls []IColl `xml:"iColl"`
}

type IColl struct {
	Id     string  `xml:"id,attr"`
	Append bool    `xml:"append,attr"`
	KColls []KColl `xml:"kColl"`
}

type Field struct {
	Id    string `xml:"id,attr"`
	Value string `xml:"value,attr"`
}

type city struct {
	no   string
	name string
}

type hebeiChannel struct {
	//bankCh chan
}

func (h hebeiChannel) fetchAllBranch() []BankInfo {
	bankCh := make(chan []BankInfo, 10)
	allBank := make([]BankInfo, 0, 10000)
	for _, pro := range provinces {
		citys := fetchCityNos(pro[0])
		for _, city := range citys {
			for _, bankname := range bankNames {
				go fetchCityBranch(pro[1], city.no, city.name, bankname[0], bankname[1], bankCh)
				//allBank = append(allBank,bankInfos...)
			}
		}
	}
	for {
		select {
		case value := <-bankCh:
			allBank = append(allBank, value...)
		case <-time.After(20 * time.Second):
			fmt.Println("timeout 20s ,爬取完成")
			fmt.Printf("总共获取全国银行%d家\n", len(allBank))
			return allBank

		}
		//if value,ok:=<-bankCh;ok {
		//	allBank = append(allBank,value...)
		//} else {
		//	break;
		//}
	}
}

func fetchCityNos(province string) []city {
	//生成client 参数为默认
	client := &http.Client{}

	url := "https://www.hebbank.com/corporbank/cityQueryAjax.do?provinceCode=" + province
	resp, err := client.Post(url, "text/xml; charset=UTF-8", nil)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	result := KColl{}
	if err = xml.Unmarshal([]byte(body), &result); err != nil {
		panic(err)
	}

	ans := make([]city, 0)

	for _, citys := range result.IColls {
		if citys.Id != "iCityInfo" {
			continue
		}
		for _, bank := range citys.KColls {
			city := city{}
			for _, field := range bank.Fields {
				if field.Id == "cityCode" {
					//爬虫cityNo需要去除最后两位字符
					city.no = field.Value[0 : len(field.Value)-2]
				} else if field.Id == "cityName" {
					city.name = field.Value
				}
			}
			ans = append(ans, city)
		}
	}

	//fmt.Printf("北京总共有%d家中国银行\n",len(ans))
	//for _,bank := range ans {
	//	fmt.Printf("%s,%s,%s,%s,%s\n",bank.Bank,bank.Province,bank.City,bank.BankName,bank.UnionNo)
	//}
	return ans
}

func fetchCityBranch(provinceName, cityNo, cityName, bankNo, bankName string, ch chan []BankInfo) {
	//生成client 参数为默认
	client := &http.Client{}

	url := "https://www.hebbank.com/corporbank/bankQueryAjax.do?bankType=" + bankNo + "&cityCode=" + cityNo + "&_="
	resp, err := client.Post(url, "text/xml; charset=UTF-8", nil)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	result := KColl{}
	if err = xml.Unmarshal([]byte(body), &result); err != nil {
		panic(err)
	}

	ans := make([]BankInfo, 0)

	for _, banks := range result.IColls {
		if banks.Id != "iBankInfo" {
			continue
		}
		for _, bank := range banks.KColls {
			bankInfo := BankInfo{Province: provinceName, City: cityName, Bank: bankName}
			for _, field := range bank.Fields {
				if field.Id == "unionBankNo" {
					bankInfo.UnionNo = field.Value
				} else if field.Id == "bankName" {
					bankInfo.BankName = field.Value
				}
			}
			ans = append(ans, bankInfo)
		}
	}

	fmt.Printf("%s的%s总共有%d家%s\n", provinceName, cityName, len(ans), bankName)
	//for _,bank := range ans {
	//	fmt.Printf("%s,%s,%s,%s,%s\n",bank.Bank,bank.Province,bank.City,bank.BankName,bank.UnionNo)
	//}
	ch <- ans
	//return ans
}
