package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
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
	defer d.EstablishConnection().Close()
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
	defer d.EstablishConnection().Close()
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
	defer d.EstablishConnection().Close()
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
	getVlansPtr := flag.Bool("get-iface-vlans", false, "Shows the current interface to vlan mappings")
	getIfacePtr := flag.Bool("get-iface", false, "Shows a specific interface")
	updateVlanptr := flag.Bool("update-iface-vlan", false, "Assign a new VLAN ID to a specified interface")
	ipPtr := flag.String("ip", "", "Target IP")
	usernamePtr := flag.String("username", "", "Device username")
	portPtr := flag.String("iface", "", "Target Interface")
	vlanPtr := flag.String("vlan", "", "VLAN identifyer")
	flag.Parse()
	//if username of ip are not specified the exit the program

	if len(*usernamePtr) == 0 {
		fmt.Println("No username specified, Please specify -username")
		os.Exit(1)
	} else if len(*ipPtr) == 0 {
		fmt.Println("No target specifed, Please specify -ip")
		os.Exit(1)
	}

	d1 := device{*ipPtr, *usernamePtr, "Pa55w0rd1!"}

	if *getVlansPtr {
		d1.GetInterfaces()
	}

	if *getIfacePtr {
		d1.GetInterface(*portPtr)
	}

	if *updateVlanptr {
		if len(*portPtr) == 0 {
			fmt.Println("Error: Wanted to Update Port But No Port Specified, Use -iface")
			os.Exit(1)
		} else if len(*vlanPtr) == 0 {
			fmt.Println("Error: Wanted to Update Port But No VLAN Specified, Use -vlan")
			os.Exit(1)
		}
		d1.UpdateVlan(*portPtr, *vlanPtr)
	}
}
