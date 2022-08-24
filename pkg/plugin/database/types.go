package database

import (
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Database struct {
	client *mongo.Client
	open   bool
}

type Filter struct {
	Entity string
	Value  string
}

type Project struct {
	Id               string `bson:"_id"`
	DisplayName      string `bson:"displayName"`
	OrganizationName string `bson:"organizationDisplayName"`
}

type Device struct {
	Id     string `bson:"_id"`
	Labels struct {
		Name string `bson:"name"`
	} `bson:"labels"`
	ProjectId     string    `bson:"project_id"`
	Type          string    `bson:"type"`
	LastEventTime time.Time `bson:"lastEventTime"`
}

type Event struct {
	Id        string    `bson:"_id"`
	DeviceId  string    `bson:"device_id"`
	ProjectId string    `bson:"project_id"`
	Type      string    `bson:"eventType"`
	Timestamp time.Time `bson:"timestamp"`
	Data      struct {
		Temperature struct {
			Value float64 `bson:"value"`
		} `bson:"temperature"`
		ObjectPresent struct {
			State string `bson:"state"`
		} `bson:"objectPresent"`
		NetworkStatus struct {
			SignalStrength  int32 `bson:"signalStrength"`
			RSSI            int32 `bson:"rssi"`
			CloudConnectors []struct {
				Id string `bson:"id"`
			} `bson:"cloudConnectors"`
		} `bson:"networkStatus"`
	} `bson:"data"`
}

type DeviceWithEvents struct {
	Device *Device
	Events []*Event
}

type ProjectWithDevices struct {
	Project *Project
	Devices []*DeviceWithEvents
}
