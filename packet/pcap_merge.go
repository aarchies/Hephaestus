package packet

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

// MergePcapStrings 将多个pcap包字符串合并并保存到一个pcap文件中 outputFile:文件根路径（自动按照年月日创建文件夹）return currentPath
func MergePcapStrings(pcapStrings []string, fileName, path string) (string, error) {

	// 创建目录，如果目录不存在
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		fmt.Printf("failed to create directories: %v\n", err)
		return "", fmt.Errorf("failed to create directories: %v", err)
	}

	filename := filepath.Join(path, fileName+".pcap")

	// 创建文件
	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	// 创建pcap writer
	w := pcapgo.NewWriter(f)
	_ = w.WriteFileHeader(65535, layers.LinkTypeEthernet) // 设置文件头部

	for _, pcapStr := range pcapStrings {
		hexStr := hex.EncodeToString([]byte(pcapStr))
		packetData, err := hex.DecodeString(hexStr)
		if err != nil {
			return "", fmt.Errorf("failed to decode pcap string: %v", err)
		}

		if err := w.WritePacket(gopacket.CaptureInfo{
			Timestamp:      time.Now(),
			CaptureLength:  len(packetData),
			Length:         len(packetData),
			InterfaceIndex: 0,
		}, packetData); err != nil {
			return "", fmt.Errorf("failed to write packet: %v", err)
		}
	}

	return filepath.Join(path, fileName+".pcap"), nil
}
