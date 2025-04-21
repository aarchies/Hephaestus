package utils

import (
	"fmt"
	"net"
	"sync"

	"github.com/golang/groupcache/lru"
	"github.com/ip2location/ip2location-go/v9"
)

type GeoIP struct {
	Db      *ip2location.DB
	Cache   *lru.Cache
	mutex   sync.Mutex
	AreaMap map[string]string
}

type Location struct {
	Country          string  `json:"country"`
	Province         string  `json:"province"`
	City             string  `json:"city"`
	PostalCode       string  `json:"postal_code"`
	Latitude         float32 `json:"latitude"`
	Longitude        float32 `json:"longitude"`
	IsAnonymousProxy bool    `json:"is_anonymous_proxy"`
}

func NewGeoIP(dbPath string) (*GeoIP, error) {
	db, err := ip2location.OpenDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开 IP2Location 数据库文件: %v", err)
	}

	return &GeoIP{
		Db:    db,
		Cache: lru.New(100000),
	}, nil
}

func (g *GeoIP) GetLocation(ipStr string) (*Location, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if val, ok := g.Cache.Get(ipStr); ok {
		location := val.(*Location)
		return location, nil
	}

	location, err := g.GetLocationV3(ipStr)
	if err != nil {
		return nil, err
	}

	g.Cache.Add(ipStr, location)
	return location, nil
}

func (g *GeoIP) GetLocationV3(ipStr string) (*Location, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil || ip.IsPrivate() || ip.IsLoopback() || ip.IsMulticast() || ip.IsUnspecified() {
		return nil, fmt.Errorf("无效的 IP 地址: %s", ipStr)
	}

	record, err := g.Db.Get_all(ipStr)
	if err != nil {
		return nil, fmt.Errorf("无法查询地理位置信息: %v", err)
	}
	country := record.Country_long
	if value, ok := g.AreaMap[country]; ok {
		country = value
	}
	region := record.Region
	if value, ok := g.AreaMap[record.Region]; ok {
		region = value
	}
	city := record.City
	if value, ok := g.AreaMap[record.City]; ok {
		city = value
	}
	location := &Location{
		Country:    country,
		Province:   region,
		City:       city,
		PostalCode: record.Zipcode,
		Latitude:   record.Latitude,
		Longitude:  record.Longitude,
	}

	return location, nil
}

func (g *GeoIP) Close() {
	g.Db.Close()
}
