package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type IPAPI struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	Isp         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
	Query       string  `json:"query"`
}

func GetISP(proxy string) (isp *IPAPI) {
	res, err := http.Get("http://ip-api.com/json/" + proxy)
	if err != nil {
		log.Println("couldn't fetch isp")
		return
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("body error isp")
		return
	}
	if err := json.Unmarshal(body, &isp); err != nil {
		log.Println("json error isp")
		return
	}
	res.Body.Close()
	return
}

func protocolColor(protocol string) string {
	switch strings.ToLower(protocol) {
	case "socks4":
		return "\033[33m"
	case "socks5":
		return "\033[35m"
	case "https":
		return "\033[32m"
	default:
		return "\033[36m"
	}
}

func PrintProxyWithProtocol(proxy string, port int, protocol string) {
	proto := strings.ToUpper(protocol)
	if proto == "" {
		proto = "HTTP"
	}
	color := protocolColor(protocol)
	if config.PrintIps.DisplayIpInfo {
		ipApi := GetISP(proxy)
		if ipApi == nil {
			fmt.Printf("\033[32m[+]\033[39m %s[%s]\033[39m \033[33m%s:%d\033[39m country=unknown city=unknown isp=unknown org=unknown asn=unknown\n", color, proto, proxy, port)
		} else {
			org := ipApi.Org
			if org == "" {
				org = ipApi.Isp
			}
			isp := ipApi.Isp
			if isp == "" {
				isp = org
			}
			asn := ipApi.As
			if asn == "" {
				asn = "unknown"
			}
			location := ipApi.City
			if location == "" {
				location = ipApi.RegionName
			}
			fmt.Printf("\033[32m[+]\033[39m %s[%s]\033[39m \033[33m%s:%d\033[39m country=%s city=%s isp=%s org=%s asn=%s\n",
				color,
				proto,
				proxy,
				port,
				ipApi.Country,
				location,
				isp,
				org,
				asn,
			)
		}
	} else {
		fmt.Printf("\033[32m[+]\033[39m %s[%s]\033[39m \033[33m%s:%d\033[39m\n", color, proto, proxy, port)
	}
}

func PrintProxy(proxy string, port int) {
	PrintProxyWithProtocol(proxy, port, "http")
}
