package main

import (
	"encoding/xml"
	"fmt"

	"github.com/scrapli/scrapligo/driver/base"
	"github.com/scrapli/scrapligo/netconf"
)

type device struct {
	ip       string
	username string
	password string
}

type RpcReply struct {
	XMLName   xml.Name `xml:"rpc-reply"`
	Text      string   `xml:",chardata"`
	Xmlns     string   `xml:"xmlns,attr"`
	MessageID string   `xml:"message-id,attr"`
	Data      struct {
		Text   string `xml:",chardata"`
		System struct {
			Text      string `xml:",chardata"`
			Xmlns     string `xml:"xmlns,attr"`
			IntfItems struct {
				Text      string `xml:",chardata"`
				PhysItems struct {
					Text       string `xml:",chardata"`
					PhysIfList []struct {
						Text      string `xml:",chardata"`
						ID        string `xml:"id"`
						PhysItems struct {
							Text       string `xml:",chardata"`
							AccessVlan string `xml:"accessVlan"`
						} `xml:"phys-items"`
					} `xml:"PhysIf-list"`
				} `xml:"phys-items"`
			} `xml:"intf-items"`
		} `xml:"System"`
	} `xml:"data"`
}

func (d device) EstablishConnection() *netconf.Driver {
	fmt.Println("Establishing Connection")
	conn, err := netconf.NewNetconfDriver(
		d.ip,
		base.WithPort(830),
		base.WithAuthStrictKey(false),
		base.WithAuthUsername(d.username),
		base.WithAuthPassword(d.password),
	)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println(conn)
	conn.Open()
	return conn
}

func (d device) GetInterfaces() {
	InterfaceFilter := `<System xmlns="http://cisco.com/ns/yang/cisco-nx-os-device">
	<intf-items>
	<phys-items>
	<PhysIf-list>
	<id></id>
	<phys-items>
	<accessVlan></accessVlan>
	</phys-items>
	</PhysIf-list>
	</phys-items>
	</intf-items>
	</System>`
	fmt.Println("Getting Interfaces")
	conn := d.EstablishConnection

	res, err := conn().Get(netconf.WithNetconfFilter(InterfaceFilter))
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println(res.Result)
	bs := []byte(res.Result)
	var output RpcReply
	xerr := xml.Unmarshal(bs, &output)

	if xerr != nil {
		fmt.Println("Err XML, ", xerr)

	}
	fmt.Println(output.Data.System.IntfItems.PhysItems.PhysIfList)
}

func main() {
	d1 := device{"192.168.137.101", "admin", "Pa55w0rd1!"}
	d1.GetInterfaces()

}
