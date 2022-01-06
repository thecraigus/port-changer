package main

import (
	"encoding/xml"
	"fmt"
	"sort"

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

	fmt.Println("Interface Details")
	conn := d.EstablishConnection

	res, err := conn().Get(netconf.WithNetconfFilter(InterfaceFilter))
	if err != nil {
		fmt.Println("Error: ", err)
	}
	bs := []byte(res.Result)
	var output RpcReply
	xerr := xml.Unmarshal(bs, &output)

	if xerr != nil {
		fmt.Println("Err XML, ", xerr)

	}
	InterfaceData := output.Data.System.IntfItems.PhysItems.PhysIfList
	sort.SliceStable(InterfaceData, func(i, j int) bool { return InterfaceData[i].ID < InterfaceData[j].ID })
	fmt.Printf("%v\n", InterfaceData)

	for _, v := range InterfaceData {
		fmt.Printf("%v		VLAN: %v \n", v.ID, v.PhysItems.AccessVlan)
	}
}

func (d device) GetInterface(port_id string) {
	InterfaceFilter := "" +
		"<System xmlns=\"http://cisco.com/ns/yang/cisco-nx-os-device\">\n" +
		"<intf-items>\n" +
		"<phys-items>\n" +
		"<PhysIf-list>\n" +
		"<id>" + port_id + "</id>\n" +
		"<phys-items>\n" +
		"<accessVlan></accessVlan>\n" +
		"</phys-items>\n" +
		"</PhysIf-list>\n" +
		"</phys-items>\n" +
		"</intf-items>\n" +
		"</System>"
	fmt.Println("Interface Details")
	conn := d.EstablishConnection
	res, err := conn().Get(netconf.WithNetconfFilter(InterfaceFilter))
	if err != nil {
		fmt.Println("Error: ", err)
	}
	bs := []byte(res.Result)
	var output RpcReply
	xerr := xml.Unmarshal(bs, &output)

	if xerr != nil {
		fmt.Println("Err XML, ", xerr)

	}
	InterfaceData := output.Data.System.IntfItems.PhysItems.PhysIfList
	if len(InterfaceData) != 0 {
		for _, v := range InterfaceData {
			fmt.Printf("%v		VLAN: %v \n", v.ID, v.PhysItems.AccessVlan)
		}
	} else {
		fmt.Println("No Such Interface, Please Check Input")
	}
}

func main() {
	d1 := device{"192.168.137.101", "admin", "Pa55w0rd1!"}
	d1.GetInterfaces()
	d1.GetInterface("eth1/199")

}
