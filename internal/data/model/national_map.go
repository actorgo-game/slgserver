package model

type NationalMap struct {
	MId   int  `bson:"mid" json:"mid"`
	X     int  `bson:"x" json:"x"`
	Y     int  `bson:"y" json:"y"`
	Type  int8 `bson:"type" json:"type"`
	Level int8 `bson:"level" json:"level"`
}
