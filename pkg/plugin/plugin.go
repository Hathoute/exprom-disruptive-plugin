package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/plugin/database"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
)

// Make sure SampleDatasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler      = (*SampleDatasource)(nil)
	_ backend.CheckHealthHandler    = (*SampleDatasource)(nil)
	_ backend.StreamHandler         = (*SampleDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*SampleDatasource)(nil)
)

// NewSampleDatasource creates a new datasource instance.
func NewSampleDatasource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	type JSONDataStruct struct {
		MongodbUrl string
	}
	var jsonData JSONDataStruct
	err := json.Unmarshal(settings.JSONData, &jsonData)
	if err != nil {
		return nil, err
	}

	db, err := database.Connect(jsonData.MongodbUrl)
	if err != nil {
		return nil, errors.New("cannot connect to database: " + err.Error())
	}

	return &SampleDatasource{
		database: db,
	}, nil
}

// SampleDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type SampleDatasource struct {
	database *database.Database
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *SampleDatasource) Dispose() {
	d.database.Close()
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := &backend.DataResponse{}
		// Unmarshal the JSON into our queryModel.
		var qm queryModel
		res.Error = json.Unmarshal(q.JSON, &qm)
		if res.Error != nil {
			response.Responses[q.RefID] = *res
			continue
		}

		switch qm.Entity {
		case "Projects":
			res = d.handleProjectsQuery(req.PluginContext, q, qm)
			break
		case "Devices":
			res = d.handleDevicesQuery(req.PluginContext, q, qm)
			break
		case "Events":
			res = d.handleEventsQuery(req.PluginContext, q, qm)
			break
		default:
			res = &backend.DataResponse{
				Error: errors.New("unknown entity '" + qm.Entity + "'"),
			}
		}

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = *res
	}

	return response, nil
}

func (d *SampleDatasource) handleProjectsQuery(pCtx backend.PluginContext, query backend.DataQuery, qm queryModel) *backend.DataResponse {
	response := &backend.DataResponse{}

	projects, err := d.database.QueryProjects()
	if err != nil {
		response.Error = err
		return response
	}

	// build fields.
	ids := make([]string, len(projects))
	names := make([]string, len(projects))

	for i, project := range projects {
		ids[i] = project.Id
		names[i] = project.DisplayName
	}

	frame := data.NewFrame("response")

	// add fields.
	frame.Fields = append(frame.Fields,
		data.NewField("id", nil, ids),
		data.NewField("name", nil, names),
	)

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

func (d *SampleDatasource) handleDevicesQuery(pCtx backend.PluginContext, query backend.DataQuery, qm queryModel) *backend.DataResponse {
	response := &backend.DataResponse{}

	var projectIdsCsv *string
	var _ *string // TODO: Use reflection to only return requested fields
	if devices, ok := qm.Parameters["projects"]; ok {
		projectIdsCsv = &devices
	}
	if fields, ok := qm.Parameters["fields"]; ok {
		_ = &fields
	}

	devices, err := d.database.QueryDevices(projectIdsCsv)
	if err != nil {
		response.Error = err
		return response
	}

	ids := make([]string, len(devices))
	names := make([]string, len(devices))
	projectIds := make([]string, len(devices))
	deviceTypes := make([]string, len(devices))

	for i, device := range devices {
		ids[i] = device.Id
		names[i] = device.Labels.Name
		projectIds[i] = device.ProjectId
		deviceTypes[i] = device.Type
	}

	frame := data.NewFrame("response")

	// add fields.
	frame.Fields = append(frame.Fields,
		data.NewField("id", nil, ids),
		data.NewField("name", nil, names),
		data.NewField("project_id", nil, projectIds),
		data.NewField("type", nil, deviceTypes),
	)

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

func (d *SampleDatasource) handleEventsQuery(pCtx backend.PluginContext, query backend.DataQuery, qm queryModel) *backend.DataResponse {
	response := &backend.DataResponse{}

	defer func() {
		if r := recover(); r != nil {
			log.DefaultLogger.Error("RECOVER", r)
		}
	}()

	var filter database.Filter
	if filterType, ok := qm.Parameters["filter"]; ok {
		filter.Entity = filterType
	}
	if entities, ok := qm.Parameters[filter.Entity]; ok {
		filter.Value = entities
	}

	projects, err := d.database.QueryEvents(&filter, query.TimeRange)
	if err != nil {
		response.Error = err
		return response
	}

	for _, project := range projects {
		for _, device := range project.Devices {
			frame := deviceToFrame(project.Project, device)

			if qm.WithStreaming {
				channel := live.Channel{
					Scope:     live.ScopeDatasource,
					Namespace: pCtx.DataSourceInstanceSettings.UID,
					Path:      "stream/device/" + device.Device.Id,
				}
				frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
			}

			response.Frames = append(response.Frames, frame)
		}
	}

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *SampleDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (res *backend.CheckHealthResult, _ error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	result := d.database.TestConnection()
	var status = backend.HealthStatusOk
	if !result.Success {
		status = backend.HealthStatusError
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: result.Message,
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (d *SampleDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	status := backend.SubscribeStreamStatusPermissionDenied
	path := strings.Split(req.Path, "/")
	if len(path) == 3 && path[0] == "stream" {
		switch path[1] {
		case "device":
			status = backend.SubscribeStreamStatusOK
		default:
			status = backend.SubscribeStreamStatusNotFound
		}
	}

	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (d *SampleDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "request", req)

	path := strings.Split(req.Path, "/")
	deviceId := path[2]
	lastFetch := time.Now()

	filter := &database.Filter{
		Entity: "devices",
		Value:  deviceId,
	}

	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Info("Context done, finish streaming", "path", req.Path)
			return nil
		case <-time.After(5 * time.Second):
			preFetch := time.Now()
			projects, err := d.database.QueryEvents(filter, backend.TimeRange{
				From: lastFetch,
				To:   preFetch.Add(time.Minute),
			})

			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}

			var frame *data.Frame
			for _, project := range projects {
				for _, device := range project.Devices {
					frame = deviceToFrame(project.Project, device)
				}
			}

			err = sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}

			lastFetch = preFetch
		}
	}
}

// PublishStream is called when a client sends a message to the stream.
func (d *SampleDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}
