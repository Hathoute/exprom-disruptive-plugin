package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"

	"errors"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"strconv"
)

type TestResult struct {
	Success bool
	Message string
}

const databaseName = "disruptiveBackup"

var appContext = context.TODO()

func Connect(mongoUrl string) (*Database, error) {
	clientOptions := options.Client().ApplyURI(mongoUrl)
	client, err := mongo.Connect(appContext, clientOptions)

	return &Database{
		client: client,
		open:   true,
	}, err
}

func (db *Database) IsConnected() bool {
	return db != nil && db.open
}

func (db *Database) Close() error {
	if !db.IsConnected() {
		return nil
	}

	err := db.client.Disconnect(appContext)
	if err != nil {
		return err
	}

	db.open = false
	return nil
}

func (db *Database) TestConnection() (result *TestResult) {
	log.DefaultLogger.Info("TestConnection called")
	defer func() {
		if r := recover(); r != nil {
			result = new(TestResult)
			result.Success = false
			result.Message = "Database error"
		}
	}()

	var commandResult bson.M
	command := bson.D{{"serverStatus", 1}}
	err := db.client.Database("databaseName").RunCommand(appContext, command).Decode(&commandResult)

	if err != nil {
		db.Close()
		return &TestResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &TestResult{
		Success: true,
		Message: fmt.Sprintf("OK: MongoDB version: %+v", commandResult["version"]),
	}
}

func (db *Database) QueryProjects() ([]*Project, error) {
	return db.queryProjects(nil)
}

func (db *Database) queryProjects(projectIdsCsv *string) ([]*Project, error) {
	log.DefaultLogger.Info("queryProjects called")
	if !db.IsConnected() {
		return nil, errors.New("not connected to any database")
	}

	collection := db.client.Database(databaseName).Collection("projects")
	filter := bson.D{{}}
	if projectIdsCsv != nil {
		filter = append(filter, bson.E{
			Key:   "project_id",
			Value: bson.D{{"$in", toBsonArray(strings.Split(*projectIdsCsv, ","))}},
		})
	}
	cursor, err := collection.Find(appContext, bson.D{{}})

	if err != nil {
		return nil, err
	}

	projects := make([]*Project, 0)
	for cursor.Next(appContext) {
		var project Project
		err := cursor.Decode(&project)
		if err != nil {
			return nil, err
		}

		projects = append(projects, &project)
	}

	log.DefaultLogger.Info("Found " + strconv.Itoa(len(projects)) + " projects.")
	return projects, nil
}

func (db *Database) QueryDevices(projectIdsCsv *string) ([]*Device, error) {
	return db.queryDevices(projectIdsCsv, nil)
}

func (db *Database) queryDevices(projectIdsCsv *string, deviceIdsCsv *string) ([]*Device, error) {
	log.DefaultLogger.Info("queryDevices called")
	if !db.IsConnected() {
		return nil, errors.New("not connected to any database")
	}

	collection := db.client.Database(databaseName).Collection("devices")
	filter := bson.D{{}}
	if projectIdsCsv != nil {
		filter = append(filter, bson.E{
			Key:   "project_id",
			Value: bson.D{{"$in", toBsonArray(strings.Split(*projectIdsCsv, ","))}},
		})
	}
	if deviceIdsCsv != nil {
		filter = append(filter, bson.E{
			Key:   "_id",
			Value: bson.D{{"$in", toBsonArray(strings.Split(*deviceIdsCsv, ","))}},
		})
	}

	cursor, err := collection.Find(appContext, filter)

	if err != nil {
		return nil, err
	}

	devices := make([]*Device, 0)
	for cursor.Next(appContext) {
		var device Device
		err := cursor.Decode(&device)
		if err != nil {
			return nil, err
		}

		devices = append(devices, &device)
	}

	return devices, nil
}

func (db *Database) QueryEvents(filter *Filter, timerange backend.TimeRange) ([]*ProjectWithDevices, error) {
	log.DefaultLogger.Info("QueryEvents called")

	if !db.IsConnected() {
		return nil, errors.New("not connected to any database")
	}

	collection := db.client.Database(databaseName).Collection("events")
	filterName := "project_id"
	if filter.Entity == "devices" {
		filterName = "device_id"
	}

	log.DefaultLogger.Info("Timerange: " + timerange.From.Format(time.RFC3339Nano))

	mongoFilter := bson.D{
		{filterName, bson.D{
			{"$in", *toBsonArray(strings.Split(filter.Value, ","))},
		}},
		{"eventType", bson.D{
			{"$ne", "networkStatus"},
		}},
		{"timestamp", bson.D{
			{"$gte", timerange.From.Format(time.RFC3339Nano)},
			{"$lt", timerange.To.Format(time.RFC3339Nano)},
		}},
	}

	findOptions := options.Find()
	// Sort by timestamp asc
	findOptions.SetSort(bson.D{{"timestamp", 1}})

	cursor, err := collection.Find(appContext, mongoFilter, findOptions)

	if err != nil {
		return nil, err
	}

	eventsByDeviceId := map[string][]*Event{}
	for cursor.Next(appContext) {
		var event Event
		err := cursor.Decode(&event)
		if err != nil {
			return nil, err
		}

		list, ok := eventsByDeviceId[event.DeviceId]
		if !ok {
			list = make([]*Event, 0)
		}

		list = append(list, &event)
		eventsByDeviceId[event.DeviceId] = list
	}

	// Find all devices
	deviceIdsCsv := ""
	separator := ""
	for id, _ := range eventsByDeviceId {
		deviceIdsCsv += separator + id
		separator = ","
	}

	devices, err := db.queryDevices(nil, &deviceIdsCsv)
	if err != nil {
		return nil, err
	}

	var projects []*ProjectWithDevices
	projectById := map[string]*ProjectWithDevices{}
	for _, device := range devices {
		deviceWithEvents := &DeviceWithEvents{
			Device: device,
			Events: eventsByDeviceId[device.Id],
		}

		project, ok := projectById[device.ProjectId]
		if !ok {
			project = &ProjectWithDevices{
				Project: &Project{
					Id:               device.ProjectId,
					DisplayName:      device.ProjectId,
					OrganizationName: "Org",
				},
				Devices: make([]*DeviceWithEvents, 0),
			}
			projectById[device.ProjectId] = project
			projects = append(projects, project)
		}

		project.Devices = append(project.Devices, deviceWithEvents)
	}

	return projects, nil
}

func toBsonArray(arr []string) *bson.A {
	bsonA := bson.A{}
	for _, item := range arr {
		bsonA = append(bsonA, item)
	}

	return &bsonA
}
