package utils

import (
	"log"
	"testing"
)

//http://ip-api.com/json/?fields=status,message,continent,continentCode,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,mobile

func TestGet(t *testing.T) {
	url := "http://ip-api.com/json/"
	var data = struct {
		Fields []string `json:"fields"`
	}{
		Fields: []string{
			"status", "message", "continent", "continentCode", "country", "countryCode", "region",
			"regionName", "city", "zip", "lat", "lon", "timezone", "isp", "mobile",
		},
	}
	ret, _, err := Get(url, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	log.Print(ret)
}
