package stream

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"math"
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

	// Put data into instance
	recvRate := recvStatics.GetAll()
	uploadRate := uploadStatics.GetAll()
	var max int64 = 0
	for _, num := range recvRate {
		if num > max {
			max = num
		}
	}
	for _, num := range uploadRate {
		if num > max {
			max = num
		}
	}
	var interval int64 = 1024 * 1024
	if max/(interval) > 10 {
		interval = int64(math.Ceil(float64(max/interval/10))) * interval
	}
	splitNum := int(math.Ceil(float64(max) / float64(interval)))
	max = int64(splitNum) * interval
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(charts.WithYAxisOpts(opts.YAxis{
		Max:         max,
		Type:        "value",
		Scale:       false,
		SplitNumber: splitNum,
		AxisLabel: &opts.AxisLabel{
			Interval: strconv.FormatInt(interval, 10),
			Formatter: opts.FuncOpts(`function(value, index) {
                                if (value == 0) {
                                    value = '0B';
                                } else {
                                    var k = 1024;
                                    var sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
                                    var c = Math.floor(Math.log(value) / Math.log(k));
                                    value = (value / Math.pow(k, c)) + ' ' + sizes[c];
                                }
                                return value;
                            }`),
		},
	}), charts.WithTooltipOpts(opts.Tooltip{
		Show:      true,
		Trigger:   "axis",
		TriggerOn: "",
		Formatter: opts.FuncOpts(`function(value) {
							console.log(value);
                            if (value[0].value == 0) {
                                value[0].value = '0B';
                            } else {
                                var k = 1024;
                                var sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
                                var c = Math.floor(Math.log(value[0].value) / Math.log(k));
                                value[0].value = (value[0].value / Math.pow(k, c)).toPrecision(4) + ' ' + sizes[c];
                            }
                            if (value[1].value == 0) {
                                value[1].value = '0B';
                            } else {
                                var k = 1024;
                                var sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
                                var c = Math.floor(Math.log(value[1].value) / Math.log(k));
                                value[1].value = (value[1].value / Math.pow(k, c)).toPrecision(4) + ' ' + sizes[c];
                            }
                            return value[0].seriesName + ' ' + value[0].value+ '<br/>' + value[1].seriesName + ' ' + value[1].value;
                        }`),
		AxisPointer: nil,
	}),
		charts.WithInitializationOpts(opts.Initialization{PageTitle: "forward监控", Width: "100%", Height: "100vh"}))

	line.SetXAxis(xAxis).
		AddSeries("下行", generateLineItems(recvRate)).
		AddSeries("上行", generateLineItems(uploadRate)).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	line.Render(w)
}
