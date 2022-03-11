package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type AddressRequest struct {
	AddressList []string `json:"addressList"`
}
type AddressEntity struct {
	Address      string
	Overview     string
	TokenTracker string
}

func main() {
	GetLabelMainCollector()
}
func GetLabelMainCollector() {
	f, err := os.Create("etherscan-tag.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	c := colly.NewCollector(
		colly.CacheDir("./cache4"),
	)
	// Before making a request print "Visiting ..."
	c.OnRequest(func(req *colly.Request) {
		req.Headers.Set("authority", "etherscan.io")
		req.Headers.Set("cache-control", "max-age=0")
		req.Headers.Set("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"99\", \"Google Chrome\";v=\"99\"")
		req.Headers.Set("sec-ch-ua-mobile", "?0")
		req.Headers.Set("sec-ch-ua-platform", "\"macOS\"")
		req.Headers.Set("dnt", "1")
		req.Headers.Set("upgrade-insecure-requests", "1")
		req.Headers.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36")
		req.Headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		req.Headers.Set("sec-fetch-site", "same-origin")
		req.Headers.Set("sec-fetch-mode", "navigate")
		req.Headers.Set("sec-fetch-user", "?1")
		req.Headers.Set("sec-fetch-dest", "document")
		req.Headers.Set("referer", "https://etherscan.io/")
		req.Headers.Set("accept-language", "zh-CN,zh;q=0.9")
		req.Headers.Set("cookie", "amp_fef1e8=49632fbe-150b-46a4-ac9d-77bff59d5d8bR...1fqdc9jnk.1fqddjf7h.7.2.9; _ga=GA1.2.961539908.1642397094; __stripe_mid=3e4ef33a-25aa-46e9-91bc-ac43b3a2352967627b; _ga_0JZ9C3M56S=GS1.1.1643298644.2.0.1643298644.0; etherscan_cookieconsent=True; CultureInfo=en; _gid=GA1.2.2147003388.1646809956; ASP.NET_SessionId=qfftxyqeigggnaqjtgvyqb2n; cf_clearance=u3Lbplegav1xg9X625dDHa3X4zhBbPouLdarU7YSmAQ-1646980947-0-250; __cflb=0H28vPcoRrcznZcNZSuFrvaNdHwh858argkNWk9AiWg; etherscan_userid=wangshuaishuai; etherscan_pwd=4792:Qdxb:Hz3J/4v9FEQRy0RSTsgx5vuUeQeAztw2lgr1FI3sRU4=; etherscan_autologin=True; _gat_gtag_UA_46998878_6=1; __cf_bm=pMSZaPOhm9reN1kiBqqnay51r47zzzJ6eLgEjH1eIWI-1646990430-0-AVfDcIp6y3Puxp7KSNIcvCB8ktPsgRQCC8gnwzonIaCH0haXq4eW9SJKdv4ZBE7+H/c482x8GqDxbL+LlXE0KIqYlZxM1MoBO/Cz01GlEA9a/ll9sou0dOuHcit2nGb72w==")
	})
	c.SetRequestTimeout(30 * time.Second)
	entity := AddressEntity{}
	file := "address.json"
	address := readAddress(file)
	file2 := "address1.json"
	address2 := readAddress(file2)
	addressMap := make(map[string]int)
	for i := 0; i < len(address); i++ {
		addressMap[address[i]] = 0
	}
	for i := 0; i < len(address2); i++ {
		addressMap[address2[i]] = 1
	}
	c.OnRequest(func(r *colly.Request) {
		entity = AddressEntity{}
		entity.Address = strings.Split(r.URL.Path, "/")[2]
	})
	c.OnResponse(func(r *colly.Response) {
		//fmt.Println("response", r.Request.URL.Path)
	})
	list := make([]AddressEntity, 0, 200)
	c.OnError(func(r *colly.Response, e error) {
		log.Fatal(fmt.Sprintf("%s LabelMain OnError:%v", r.Request.URL.String(), e))
	})
	c.OnHTML("#ContentPlaceHolder1_divSummary", func(e *colly.HTMLElement) {
		fmt.Println(fmt.Sprintf("\"%s\",", entity.Address))
		overview := e.ChildText(".u-label.u-label--secondary.text-dark span")
		if overview != "" {
			entity.Overview = overview
		}
		tokenTracker := e.ChildText("#ContentPlaceHolder1_tr_tokeninfo .row.align-items-center .col-md-8 a")
		if tokenTracker != "" {
			entity.TokenTracker = tokenTracker
		}
		if overview != "" || tokenTracker != "" {
			list = append(list, entity)
		}
		addressMap[entity.Address] = 1
	})
	//c.Visit(fmt.Sprintf("https://etherscan.io/address/%s", address[0]))
	for k, v := range addressMap {
		if v == 0 {
			c.Visit(fmt.Sprintf("https://etherscan.io/address/%s", k))
		}
	}
	c.Wait()
	f.WriteString("\xEF\xBB\xBF") // 写入一个UTF-8 BOM

	w := csv.NewWriter(f) //创建一个新的写入文件流

	var data = make([][]string, len(list)+1)
	data[0] = []string{"address", "Overview", "TokenTracker"}
	for i := 0; i < len(list); i++ {
		data[i+1] = []string{list[i].Address, list[i].Overview, list[i].TokenTracker}
	}
	w.WriteAll(data)
	w.Flush()
}

func readAddress(filename string) []string {

	filePtr, err := os.Open(filename)
	if err != nil {
		fmt.Println("Open file failed [Err:%s]", err.Error())
		return nil
	}
	defer filePtr.Close()

	var list AddressRequest

	// 创建json解码器
	decoder := json.NewDecoder(filePtr)
	err = decoder.Decode(&list)
	if err != nil {
		fmt.Println("Decoder failed", err.Error())

	} else {
		fmt.Println("Decoder success")
		//fmt.Println(list)
	}
	return list.AddressList
}
