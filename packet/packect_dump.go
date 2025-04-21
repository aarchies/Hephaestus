package packet

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Interface struct {
	Name        string //设备名称
	Description string //设备描述信息
	Flags       uint32
	Addresses   []InterfaceAddress //网口的地址信息列表
}

// InterfaceAddress describes an address associated with an Interface.
// Currently, it's IPv4/6 specific.
type InterfaceAddress struct {
	IP        net.IP
	Netmask   net.IPMask // Netmask may be nil if we were unable to retrieve it.
	Broadaddr net.IP     // Broadcast address for this IP may be nil
	P2P       net.IP     // P2P destination address for this IP may be nil
}

var (
	//device       string = "\\Device\\NPF_{5ED349BF-98AF-40C4-8E60-4388DA2518D9}"
	snapshot_len int32 = 1024
	promiscuous  bool  = false
	err          error
	timeout      time.Duration = 30 * time.Second
	handle       *pcap.Handle
	PacketCh     = make(chan []byte, 1000)
)

func OpenOffline(filePath string) chan []byte {
	handle, err = pcap.OpenOffline(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	var ch = make(chan []byte, 1024)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		ch <- packet.Data()
	}
	return ch
}

func OpenLive(isDebug bool) {

	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}
	// 打印设备信息
	fmt.Println("Devices found:")
	for _, device := range devices {
		fmt.Println("\nName: ", device.Name)
		fmt.Println("Description: ", device.Description)
		fmt.Println("Devices addresses: ", device.Description)

		go func(d string) {
			handle, err = pcap.OpenLive(d, snapshot_len, promiscuous, timeout)
			if err != nil {
				log.Fatal(err)
			}
			defer handle.Close()

			packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
			for packet := range packetSource.Packets() {
				if isDebug {
					fmt.Println("-----------------------------------------------", device.Name)
					printPacketInfo(packet)
					fmt.Println("_______________________________________________")
				}
				PacketCh <- packet.Data()
			}
		}(device.Name)
	}
}

func printPacketInfo(packet gopacket.Packet) {
	// Let's see if the packet is an ethernet packet
	// 判断数据包是否为以太网数据包，可解析出源mac地址、目的mac地址、以太网类型（如ip类型）等
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		fmt.Println("Ethernet layer detected.")
		ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
		fmt.Println("Source MAC: ", ethernetPacket.SrcMAC)
		fmt.Println("Destination MAC: ", ethernetPacket.DstMAC)
		// Ethernet type is typically IPv4 but could be ARP or other
		fmt.Println("Ethernet type: ", ethernetPacket.EthernetType)
		fmt.Println()
	}
	// Let's see if the packet is IP (even though the ether type told us)
	// 判断数据包是否为IP数据包，可解析出源ip、目的ip、协议号等
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		fmt.Println("IPv4 layer detected.")
		ip, _ := ipLayer.(*layers.IPv4)
		fmt.Printf("From %s to %s\n", ip.SrcIP, ip.DstIP)
		fmt.Println("Protocol: ", ip.Protocol)
		fmt.Println()
	}
	// Let's see if the packet is TCP
	// 判断数据包是否为TCP数据包，可解析源端口、目的端口、seq序列号、tcp标志位等
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		fmt.Println("TCP layer detected.")
		tcp, _ := tcpLayer.(*layers.TCP)
		// TCP layer variables:
		// SrcPort, DstPort, Seq, Ack, DataOffset, Window, Checksum, Urgent
		// Bool flags: FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS
		fmt.Printf("From port %d to %d\n", tcp.SrcPort, tcp.DstPort)
		fmt.Println("Sequence number: ", tcp.Seq)
		fmt.Println()
	}
	// Iterate over all layers, printing out each layer type
	fmt.Println("All packet layers:")
	for _, layer := range packet.Layers() {
		fmt.Println("- ", layer.LayerType())
	}
	// Check for errors
	// 判断layer是否存在错误
	if err := packet.ErrorLayer(); err != nil {
		fmt.Println("Error decoding some part of the packet:", err)
	}
}
