package plugin

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/plugin/database"
	"time"
)

func getValueFromEvent(event *database.Event) (float64, bool) {
	switch event.Type {
	case "temperature":
		return event.Data.Temperature.Value, false
	case "objectPresent":
		if event.Data.ObjectPresent.State == "NOT_PRESENT" {
			return 0, false
		}
		return 1, false
	default:
		// Unknown event type, skip
		return 0, true
	}
}

func deviceToFrame(project *database.Project, device *database.DeviceWithEvents) *data.Frame {
	frame := data.NewFrame(project.Id)

	times := make([]time.Time, 0)
	values := make([]float64, 0)

	for _, ev := range device.Events {
		value, skip := getValueFromEvent(ev)
		if skip {
			continue
		}

		times = append(times, ev.Timestamp)
		values = append(values, value)
	}

	valueField := data.NewField("Value", data.Labels{"device": device.Device.Labels.Name}, values)
	timeField := data.NewField("Time", nil, times)

	// populate fields with metric values
	frame.Fields = append(frame.Fields, valueField, timeField)

	return frame
}
