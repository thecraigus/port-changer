package main

import (
	"encoding/xml"
	"flag"
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

type RpcReplyError struct {
	XMLName   xml.Name `xml:"rpc-reply"`
	Text      string   `xml:",chardata"`
	Xmlns     string   `xml:"xmlns,attr"`
	MessageID string   `xml:"message-id,attr"`
	RpcError  struct {
		Text          string `xml:",chardata"`
		ErrorType     string `xml:"error-type"`
		ErrorTag      string `xml:"error-tag"`
		ErrorSeverity string `xml:"error-severity"`
		ErrorMessage  struct {
			Text string `xml:",chardata"`
			Lang string `xml:"lang,attr"`
		} `xml:"error-message"`
		ErrorPath string `xml:"error-path"`
	} `xml:"rpc-error"`
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

func (d device) UpdateVlan(p string, v string) {

	changevlan := `
	<config>
	<System xmlns="http://cisco.com/ns/yang/cisco-nx-os-device">
	<intf-items>
	<phys-items>
	<PhysIf-list>
	<id>` + p + `</id>
	<accessVlan>vlan-` + v + `</accessVlan>
	</PhysIf-list>
	</phys-items>
	</intf-items>
	</System>
	</config>`

	fmt.Println("Updating Access VLAN")
	conn := d.EstablishConnection
	res, err := conn().EditConfig("running", changevlan)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	bs := []byte(res.Result)
	var output RpcReplyError
	xml.Unmarshal(bs, &output)
	if len(output.RpcError.ErrorMessage.Text) > 0 {
		fmt.Println("Unable to update VLAN, Error Message Below: ")
		fmt.Println(output.RpcError.ErrorMessage.Text)
	} else {
		fmt.Println("Interface Updated!")

		d.GetInterface(p)

	}
}

func main() {
	d1 := device{"192.168.137.101", "admin", "Pa55w0rd1!"}
	getVlansPtr := flag.Bool("list", false, "Use this flag to retrieve a list of vlan assignmens")
	ipPtr := flag.String("ip", "", "The IP of the device")
	usernamePtr := flag.String("username", "", "The IP of the device")
	portPtr := flag.String("iface", "", "The IP of the device")
	vlanPtr := flag.String("vlan", "", "The IP of the device")
	flag.Parse()

	fmt.Printf("%v", *getVlansPtr)
	if *getVlansPtr {
		d1.GetInterfaces()
	}

	fmt.Println("Your command line argumanet is: ", *ipPtr)
	fmt.Println("Your command line argumanet is: ", *usernamePtr)
	fmt.Println("Your command line argumanet is: ", *portPtr)
	//
	_ = vlanPtr

	d1.GetInterface("eth1/1671")
	d1.UpdateVlan("eth1/17", "11")

}
