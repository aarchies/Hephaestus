package localization

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"
)

func TestNewGeoIP(t *testing.T) {
	filePath := path.Join("/home/iiot/static", "IP2LOCATION-LITE-DB11.BIN")
	filePath2 := path.Join("/home/iiot/static", "area.json")

	LocationDb, err := NewGeoIP(filePath)
	if err != nil {
		log.Fatal("locationDb error %w", err)

	}

	file, err := os.Open(filePath2)
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return
	}
	defer file.Close()

	// 创建一个 map 存储解析后的 JSON 数据
	//translations := make(map[string]map[string]string)

	// 使用 json 包解码 JSON 文件到 map 中
	decoder := json.NewDecoder(file)
	m := make(map[string]string)
	if err := decoder.Decode(&m); err != nil {
		fmt.Println("ProtocolKeyMap Error decoding JSON:", err)
		return
	}
	LocationDb.AreaMap = m

}

func TestGeoIP_GetLocation(t *testing.T) {
	filePath := path.Join("/home/iiot/static", "IP2LOCATION-LITE-DB11.BIN")
	filePath2 := path.Join("/home/iiot/static", "area.json")

	LocationDb, err := NewGeoIP(filePath)
	if err != nil {
		log.Fatal("locationDb error %w", err)

	}

	file, err := os.Open(filePath2)
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return
	}
	defer file.Close()

	// 创建一个 map 存储解析后的 JSON 数据
	//translations := make(map[string]map[string]string)

	// 使用 json 包解码 JSON 文件到 map 中
	decoder := json.NewDecoder(file)
	m := make(map[string]string)
	if err := decoder.Decode(&m); err != nil {
		fmt.Println("ProtocolKeyMap Error decoding JSON:", err)
		return
	}
	LocationDb.AreaMap = m
	srcLocation, errSrc := LocationDb.GetLocation("192.16.1.1")
	if errSrc != nil {
		fmt.Println("GetLocation Error decoding JSON:", errSrc)
	}
	if errSrc == nil {
		fmt.Println(srcLocation.City)
		fmt.Println(srcLocation.Country, "Country")
		fmt.Println(srcLocation.Province, "Province")
		fmt.Println(srcLocation.Latitude)
		fmt.Println(srcLocation.Longitude)
		fmt.Println(srcLocation)

	}
}

func TestGetProcessTime(t *testing.T) {
	filePath := path.Join("/home/iiot/static", "IP2LOCATION-LITE-DB11.BIN")
	filePath2 := path.Join("/home/iiot/static", "area.json")

	LocationDb, err := NewGeoIP(filePath)
	if err != nil {
		log.Fatal("locationDb error %w", err)

	}

	file, err := os.Open(filePath2)
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return
	}
	defer file.Close()

	// 创建一个 map 存储解析后的 JSON 数据
	//translations := make(map[string]map[string]string)

	// 使用 json 包解码 JSON 文件到 map 中
	decoder := json.NewDecoder(file)
	m := make(map[string]string)
	if err := decoder.Decode(&m); err != nil {
		fmt.Println("ProtocolKeyMap Error decoding JSON:", err)
		return
	}
	LocationDb.AreaMap = m
	a := time.Now()
	for i := 0; i < 10000; i++ {
		rand.Seed(time.Now().UnixNano()) // 初始化随机数生成器
		ip := fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
		LocationDb.GetLocation(ip)

	}
	fmt.Println(time.Since(a).Seconds())

}
