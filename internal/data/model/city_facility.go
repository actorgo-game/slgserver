package model

type Facility struct {
	Name         string `bson:"name" json:"name"`
	PrivateLevel int8   `bson:"level" json:"level"`
	Type         int8   `bson:"type" json:"type"`
	UpTime       int64  `bson:"up_time" json:"upTime"`
}

type CityFacility struct {
	Id         int        `bson:"id" json:"id"`
	ServerId   int        `bson:"server_id" json:"serverId"`
	RId        int        `bson:"rid" json:"rid"`
	CityId     int        `bson:"cityId" json:"cityId"`
	Facilities []Facility `bson:"facilities" json:"facilities"`
}

func (CityFacility) CollectionName() string {
	return "city_facilities"
}
