package stream

import (
	"encoding/json"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"math"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

const size = 300

var (
	xAxis                              = make([]string, size, size)
	recvStatics     *Ring              = New(size)
	uploadStatics   *Ring              = New(size)
	lastRecvBytes   int64              = 0
	lastUploadBytes int64              = 0
	tmpl            *template.Template = template.New("monitor")
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

	tmpl.Parse(`                    <!DOCTYPE html>
                    <html lang="en">
                    <head>
                        <meta charset="UTF-8">
                        <title>{{.title}}</title>
                        <meta http-equiv="refresh" content="3">
                        <script src="{{.echarts_url}}"></script>
                    </head>
                    <body style="margin: 0;height:100%;">
                    <div id="main" style="width: 100%;height: 100vh;"></div>
                    <script type="text/javascript">
                        // 基于准备好的dom，初始化echarts实例
                        var myChart = echarts.init(document.getElementById('main'));
                        // 指定图表的配置项和数据
                        var option = {
                            title: {
                        text: '{{.title}}'
                    },
                    tooltip: {
                        trigger: 'axis',
                        formatter: function(value) {
                            //这里的value[0].value就是我需要每次显示在图上的数据
                            if (value[0].value <= 0) {
                                value[0].value = '0B';
                            } else {
                                var k = 1024;
                                var sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
                                //这里是取自然对数，也就是log（k）（value[0].value），求出以k为底的多少次方是value[0].value
                                var c = Math.floor(Math.log(value[0].value) / Math.log(k));
                                value[0].value = (value[0].value / Math.pow(k, c)).toPrecision(4) + ' ' + sizes[c];
                            }
                            if (value[1].value <= 0) {
                                value[1].value = '0B';
                            } else {
                                var k = 1024;
                                var sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
                                //这里是取自然对数，也就是log（k）（value[0].value），求出以k为底的多少次方是value[0].value
                                var c = Math.floor(Math.log(value[1].value) / Math.log(k));
                                value[1].value = (value[1].value / Math.pow(k, c)).toPrecision(4) + ' ' + sizes[c];
                            }
                            //这里的value[0].name就是每次显示的name
                            return value[0].name + "<br/>" + "上行网速: " + value[0].value+ "<br/>" + "下行网速: " + value[1].value;
                        }
                    },
                    legend: {
                        data: {{.legends}}
                    },
                    toolbox: {
                        feature: {
                            mark: {
                                show: true
                            },
                            dataView: {
                                show: true,
                                readOnly: false
                            },
                            magicType: {
                                show: true,
                                type: ['line', 'bar']
                            },
                            restore: {
                                show: true
                            },
                            saveAsImage: {
                                show: true
                            }
                        }
                    },
                    xAxis: {
                        type: 'category',
                        boundaryGap: false,
                        data: {{.scales}}
                    },
                    yAxis: {
                        type: "value",
                        max: function(value) {
                            var k = 1024;
                            var c = Math.floor(Math.log(value.max) / Math.log(k));
                            interval = Math.pow(k, c);
                            return Math.ceil(value.max / interval) * interval;
                        },
                        interval: {{.interval}},
                        axisLabel: {
                            formatter: function(value, index) {
                                if (value <= 0) {
                                    value = '0B';
                                } else {
                                    var k = 1024;
                                    var sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
                                    //这里是取自然对数，也就是log（k）（value），求出以k为底的多少次方是value
                                    var c = Math.floor(Math.log(value) / Math.log(k));
                                    value = (value / Math.pow(k, c)) + ' ' + sizes[c];
                                }
                                //这里的value[0].name就是每次显示的name
                                return value;
                            }
                        },
                    },
                    series: [        {
                        itemStyle:{
							color: '#ff8000',
                        },
                        "data": {{.seriesUp}},
                        "markLine": {
                            "data": [{
                                "type": "average",
                                "name": "平均值"
                            }],
                            "label": {
                                formatter: function(value) {
                                    if (value.value <= 0) {
                                        value = '0B';
                                    } else {
                                        var k = 1024;
                                        var sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
                                        //这里是取自然对数，也就是log（k）（value），求出以k为底的多少次方是value
                                        var c = Math.floor(Math.log(value.value) / Math.log(k));
                                        value = (value.value / Math.pow(k, c)).toPrecision(4) + ' ' + sizes[c];
                                    }
                                    //这里的value[0].name就是每次显示的name
                                    return value;
                                }
                            }
                        },
                        "markPoint": {
                            "data": [{
                                "type": "max",
                                "name": "最大值"
                            }],
                            symbol: "roundRect",
                            symbolSize: [70, 30],
                            "label": {
                                formatter: function(value) {
                                    if (value.value <= 0) {
                                        value = '0B';
                                    } else {
                                        var k = 1024;
                                        var sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
                                        //这里是取自然对数，也就是log（k）（value），求出以k为底的多少次方是value
                                        var c = Math.floor(Math.log(value.value) / Math.log(k));
                                        value = (value.value / Math.pow(k, c)).toPrecision(4) + ' ' + sizes[c];
                                    }
                                    //这里的value[0].name就是每次显示的name
                                    return value;
                                }
                            }
                        },
                        "name": "上行网速",
                        "smooth": false,
                        "type": "line"
                    },
                    {
                        itemStyle:{
                            color: '#2eb82e',
                        },
                        "data": {{.seriesDown}},
                        "markLine": {
                            "data": [{
                                "type": "average",
                                "name": "平均值"
                            }],
                            "label": {
                                formatter: function(value) {
                                    if (value.value <= 0) {
                                        value = '0B';
                                    } else {
                                        var k = 1024;
                                        var sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
                                        //这里是取自然对数，也就是log（k）（value），求出以k为底的多少次方是value
                                        var c = Math.floor(Math.log(value.value) / Math.log(k));
                                        value = (value.value / Math.pow(k, c)).toPrecision(4) + ' ' + sizes[c];
                                    }
                                    //这里的value[0].name就是每次显示的name
                                    return value;
                                }
                            }
                        },
                        "markPoint": {
                             "data": [{
                                 "type": "max",
                                 "name": "最大值"
                             }],
                             symbol: "roundRect",
                             symbolSize: [70, 30],
                             "label": {
                                 formatter: function(value) {
                                     if (value.value <= 0) {
                                         value = '0B';
                                     } else {
                                         var k = 1024;
                                         var sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
                                          //这里是取自然对数，也就是log（k）（value），求出以k为底的多少次方是value
                                         var c = Math.floor(Math.log(value.value) / Math.log(k));
                                         value = (value.value / Math.pow(k, c)).toPrecision(4) + ' ' + sizes[c];
                                     }
                                     //这里的value[0].name就是每次显示的name
                                     return value;
                                 }
                             }
                         },
                        "name": "下行网速",
                        "smooth": false,
                        "type": "line"
                    }],
                    animation: false,
                    animationDuration: 5
                };
                // 使用刚指定的配置项和数据显示图表。
                myChart.setOption(option);
            </script>
            </body>
            </html>
`)
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

func ServeLine(w http.ResponseWriter, _ *http.Request) {
	legends, _ := json.Marshal([]string{"上行网速", "下行网速"})
	scales, _ := json.Marshal(xAxis)
	uploadRate := uploadStatics.GetAll()
	seriesUp, _ := json.Marshal(uploadRate)
	recvRate := recvStatics.GetAll()
	seriesDown, _ := json.Marshal(recvRate)
	interval := int64(1024 * 1024)
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
	if max/(interval) > 10 {
		interval = int64(math.Ceil(float64(max/interval/10))) * interval
	}

	param := make(map[string]interface{})
	param["legends"] = string(legends)
	param["scales"] = string(scales)
	param["seriesUp"] = string(seriesUp)
	param["seriesDown"] = string(seriesDown)
	param["interval"] = strconv.FormatInt(interval, 10)
	param["title"] = "实时网速"
	param["echarts_url"] = "https://www.arloor.com/echarts.min.js"
	tmpl.Execute(w, param)
}
