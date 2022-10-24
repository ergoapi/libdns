package main

import (
	"github.com/ergoapi/libdns"
	_ "github.com/ergoapi/libdns/alidns"
	_ "github.com/ergoapi/libdns/dnspod"
)

func main() {
	dns, err := libdns.NewDns("alidns", libdns.Option{Secret: "", Key: ""})
	if err != nil {
		panic(err)
	}
	if _, err := dns.GetDomainList(); err != nil {
		panic(err)
	}
}
