package tz

import (
	"time"
	_ "time/tzdata"

	"github.com/ringsaturn/tzf"
	tzfrel "github.com/ringsaturn/tzf-rel"
	"github.com/ringsaturn/tzf/pb"
	"google.golang.org/protobuf/proto"
)

var finder *tzf.Finder

func InitializeTimezone() {
	input := &pb.Timezones{}

	// Lite data, about 11MB
	//dataFile := tzfrel.LiteData

	// Full data, about 83.5MB
	dataFile := tzfrel.FullData

	if err := proto.Unmarshal(dataFile, input); err != nil {
		panic(err)
	}
	var err error
	finder, err = tzf.NewFinderFromPB(input)
	if err != nil {
		panic(err)
	}
}

func SearchTimezone(lat, lng float64) string {
	if finder == nil {
		InitializeTimezone()
	}
	return finder.GetTimezoneName(lng, lat)
}

func GetTimezone(lat, lng float64) *time.Location {
	if finder == nil {
		InitializeTimezone()
	}

	tz, _ := finder.GetTimezoneLoc(lng, lat)
	return tz
}
