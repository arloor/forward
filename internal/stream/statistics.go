package stream

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"net/http"
	"strconv"
	"time"
)

const size = 120

var (
	xAxis                 = make([]string, size, size)
	recvStatics     *Ring = New(size)
	uploadStatics   *Ring = New(size)
	lastRecvBytes   int64 = 0
	lastUploadBytes int64 = 0
)

func init() {
	for i := 0; i < cap(recvStatics.buf); i++ {
		xAxis[i] = strconv.Itoa(i)
	}
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			recvStatics.Put(recvBytes - lastRecvBytes)
			lastRecvBytes = recvBytes
			uploadStatics.Put(uploadBytes - lastUploadBytes)
			lastUploadBytes = uploadBytes
		}
	}()
}

// generate random data for line chart
func generateLineItems(data []int64) []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < len(data); i++ {
		items = append(items, opts.LineData{Value: data[i]})
	}
	return items
}

func LineGraph(w http.ResponseWriter, _ *http.Request) {
	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{PageTitle: "forward监控", Theme: types.ThemeRoma, Width: "100%", Height: "100vh"}))

	// Put data into instance
	line.SetXAxis(xAxis).
		AddSeries("下行", generateLineItems(recvStatics.GetAll())).
		AddSeries("上行", generateLineItems(uploadStatics.GetAll())).
		SetGlobalOptions(charts.WithTooltipOpts(opts.Tooltip{
			Show:        true,
			Trigger:     "axis",
			TriggerOn:   "",
			Formatter:   "{a0} {c0} B/s<br/>{a1} {c1} B/s",
			AxisPointer: nil,
		})).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	line.Render(w)
}
