package main

import "C"
import (
	"cargo/api"
	"cargo/clib"
	"cargo/config"
	"cargo/decaros"
	"cargo/internal"
	"cargo/msg"
	"cargo/pkg"
	"cargo/screen"
	"cargo/speaker"
	"github.com/getsentry/sentry-go"
	"log"
	"strconv"
	"sync"
	"time"
)

var display *clib.Display

var debug = false

// 自启动至少需要10秒，等待系统组件加载完毕
var sleepTime = 10 * time.Second

// main  -lpos -lm -lpng -lfontconfig -lfreetype -liconv
func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://23d733d9dd54f428ca284df0b2b544dd@sentry.ucorg.com/10",
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer func() {
		if r := recover(); r != nil {
			pkg.Log.Error("main panic:", r)
			sentry.CaptureException(err)
		}
	}()
	if debug {
		sleepTime = 0 * time.Second
	}
	var wg sync.WaitGroup
	conf, err := config.Load("config.yml")
	if err != nil {
		panic("not find config.yml")
	}
	msgChan := make(chan msg.Message)
	display = clib.NewDisplay()

	screen.Set(display)
	display.Show("启动中..", "等待系统组件加载..")
	for i := 0; i < 10; i++ {
		if i%2 != 0 {
			display.Show("启动中..", "等待系统组件加载..")
		} else {
			display.Show("启动中...", "等待系统组件加载...")
		}
		time.Sleep(time.Second)
	}
	display.Show("启动中..", "正在获取配置..")
	emcsConf, err := internal.GetConfig(conf.ServerUrl)
	if err != nil {
		display.LCDRow(err.Error(), 60, 80, clib.DISP_FONT12)
		pkg.Log.Error(err)
		display.Show("错误..", "获取配置错误！！")
		time.Sleep(sleepTime)
		display.Show("正在重试..", "获取配置中..")
		time.Sleep(sleepTime)
		emcsConf, err = internal.GetConfig(conf.ServerUrl)
		if err != nil {
			display.LCDRow(err.Error(), 60, 80, clib.DISP_FONT12)
			pkg.Log.Error(err)
			display.Show("错误..", "获取配置错误！！")
			panic("获取配置错误！")
		}
	}
	//---------设置NTP
	display.Show("启动中..", "正在配置NTP..")
	decaros.SetNTP()
	api.SetConfig(conf.ServerUrl, conf.Sha, emcsConf)
	showNormal(0)
	internal.CheckUpdate(conf.DeviceType, config.Version, emcsConf.EquipmentNo)
	wg.Add(1)
	go func() {
		defer wg.Done()
		clib.StartTcpServer(msgChan)
	}()
	go internal.StartTimer(conf, display)
	go clib.StartC(wg, conf)
	for cMsg := range msgChan {
		pkg.Log.Printf("msg chan-> type=%v,content=%v \n", cMsg.Type, cMsg.Content)
		switch cMsg.Type {
		case msg.IC_CARD:
			resp, err := api.CheckTicket(cMsg.Content, api.ProtoIC)
			parseResp(err, resp)
		case msg.QRCODE:
			// 判断是否是设备操作的二维码
			isDeviceQR := internal.IsDeviceQRString(cMsg.Content)
			if isDeviceQR {
				err = internal.ClearPassedCount()
				if err != nil {
					pkg.Log.Println(err.Error())
				}
				showNormal(0)
				// 清除当天的计数,显示
				continue
			}
			resp, err := api.CheckTicket(cMsg.Content, api.ProtoQr)
			parseResp(err, resp)
		case msg.ID_CARD:
			resp, err := api.CheckTicket(cMsg.Content, api.ProtoID)
			parseResp(err, resp)
		default:
			pkg.Log.Println("undefined type")
		}
	}
	pkg.Log.Print("End......")
}

func parseResp(err error, resp *api.CheckResponse) {
	if err != nil {
		pkg.Log.Error(err)
		pkg.Log.Error(resp)
		speaker.Speaker("验票失败", false)
		screen.Show(err.Error(), false)
		return
	}
	var passedCount = 0
	if resp.RetCode == "0" {
		if resp.MerNotVerTktNum != "" && resp.MerNotVerTktNum != "1" {
			go speaker.SpeakerGroup(resp.MerNotVerTktNum)
			go screen.Show("团票可通行"+resp.MerNotVerTktNum+"人", true)
		} else {
			go screen.Show(resp.RetMsg, true)
			go speaker.Speaker(resp.RetMsg, true)
		}
		go func() {
			passedCount, err = internal.WritePassedCount()
			if err != nil {
				pkg.Log.Println(err)
			}
		}()
	} else {
		pkg.Log.Println("Check ticket Fail!")
		go screen.Show(resp.RetMsg, false)
		go speaker.Speaker(resp.RetMsg, false)
	}
	go func() {
		time.Sleep(3 * time.Second)
		showNormal(passedCount)
	}()
}
func showNormal(pCount int) {
	pkg.Log.Printf("passed count=%d\n", pCount)
	if pCount == 0 {
		passedCount, err := internal.ReadPassedCount()
		if err != nil {
			pkg.Log.Println(err)
		}
		pCount = passedCount
	}
	str := strconv.Itoa(pCount)
	display.ShowNormal(str)
	api.SyncDayCount(int32(pCount))
}
