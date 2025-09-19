package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const LOGIN_PAGE = "https://epauth.tepco.co.jp/u/login?state"

const (
	Broker   = "mqtt://core-mosquitto:1883"
	Username = "addons"
	Password = "aiteab5elia9hee9ahp5chaoG1aegohcahzie9iigaewiaPeiquu1lau9Ho5Ooje"
	ClientID = "myenecle-clinet"
)

var mqttClient mqtt.Client

// MQTT 客户端初始化示例
func newMQTTClient() mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(Broker)
	opts.SetUsername(Username)
	opts.SetPassword(Password)
	opts.SetClientID(ClientID)
	opts.SetConnectTimeout(5 * time.Second)

	mqttClinet := mqtt.NewClient(opts)
	if token := mqttClinet.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT connect error: %v", token.Error())
	}
	return mqttClinet
}

func main() {
	var username string
	var password string
	flag.StringVar(&username, "u", "", "-u username")
	flag.StringVar(&password, "p", "", "-p password")
	flag.Parse()

	if username == "" || password == "" {
		log.Fatal("missing USERNAME, PASSWORD")
	}

	log.Println("current OS: ", runtime.GOOS)
	if runtime.GOOS != "darwin" {
		mqttClient = newMQTTClient()
		defer mqttClient.Disconnect(250)
	}

	task(username, password)
}

/**
 * Main task
 */
func task(username, password string) string {
	// 启动浏览器
	var la *launcher.Launcher
	if runtime.GOOS != "darwin" {
		la = launcher.New().
			Bin("/usr/bin/chromium").
			Headless(true)
	}
	if runtime.GOOS == "darwin" {
		la = launcher.New().Headless(true)
	}

	l := la.
		Set("no-sandbox", "").                                 // --no-sandbox
		Set("disable-dev-shm-usage", "").                      // --disable-dev-shm-usage
		Set("disable-gpu", "").                                // --disable-gpu
		Set("lang", "ja").                                     // --lang=ja
		Set("disable-desktop-notifications", "").              // --disable-desktop-notifications
		Set("disable-blink-features", "AutomationControlled"). // --disable-blink-features=AutomationControlled
		Set("ignore-certificate-errors", "").                  // --ignore-certificate-errors
		Set("disable-extensions", "").                         // --disable-extensions
		Set("window-size", "1200,1080").                       // --window-size=420,1080
		Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36").
		MustLaunch()

	browser := rod.New().ControlURL(l).MustConnect()
	defer browser.MustClose()

	// 打开登录页面
	page := browser.
		MustPage(LOGIN_PAGE)

	log.Println("Going to login...")

	page.MustWaitDOMStable()

	go page.EachEvent(func(e *proto.NetworkResponseReceived) {
		body, err := proto.NetworkGetResponseBody{RequestID: e.RequestID}.Call(page)

		url := e.Response.URL
		if err == nil {
			// Get the usage data for this month and put it into MQTT
			if strings.HasPrefix(url, "https://kcx-api.tepco-z.com/kcx/billing/month") {
				json := string(body.Body)
				obj := ParseMonthlyUsage(json)

				log.Printf("Tring to push tepco_this_mon_cost: %.2f JPY\n", obj.BillInfo.UsedInfo.Charge)
				if err := pushEnergySensor("sensor.tepco_this_mon_cost", obj.BillInfo.UsedInfo.Charge, "JPY", "monetary"); err != nil {
					log.Println("Err: ", err)
				}

				log.Printf("Tring to push tepco_this_mon_usage: %.2f kWh\n", obj.BillInfo.UsedInfo.Power)
				if err := pushEnergySensor("sensor.tepco_this_mon_usage", obj.BillInfo.UsedInfo.Power, "kWh", "energy"); err != nil {
					log.Println("Err: ", err)
				}
			}

			// Get the usage data for last month usage and put it into MQTT
			if strings.HasPrefix(url, "https://kcx-api.tepco-z.com/kcx/billing/month-history?contractClass=") {
				json := string(body.Body)
				obj := ParseMonthHistory(json)
				billInfoList := obj.BillInfos
				lastMonth := billInfoList[len(billInfoList)-1]

				log.Printf("Tring to push tepco_last_mon_usage: %.2f JPY\n", lastMonth.DetailInfos[0].UsedCharge)
				if err := pushEnergySensor("sensor.tepco_last_mon_cost", lastMonth.DetailInfos[0].UsedCharge, "JPY", "monetary"); err != nil {
					log.Println("Err: ", err)
				}

				// 燃气费用
				log.Printf("Tring to push tepco_last_mon_cost: %.2f kWh\n", lastMonth.DetailInfos[0].UsedPowerInfo.Power)
				if err := pushEnergySensor("sensor.tepco_last_mon_usage", lastMonth.DetailInfos[0].UsedPowerInfo.Power, "kWh", "energy"); err != nil {
					log.Println("Err: ", err)
				}
			}

			// fmt.Println("Response URL:", url)
			// fmt.Println("Response Body:", string(body.Body))
		}
	})()

	// 填写用户名和密码
	page.MustElement("input[name='username']").MustInput(username)
	page.MustElement("input[name='password']").MustInput(password)
	page.MustWaitIdle()
	// // 提交表单
	page.MustElement(`button[value="default"]`).MustClick()
	page.MustWaitNavigation()

	page.MustWaitDOMStable()
	html, err := page.HTML()
	if err != nil {
		log.Println(err)
		return ""
	}

	// 正则匹配 span 标签内容（去掉内部 icon span）
	re := regexp.MustCompile(`(?s)<span[^>]*id="error-element-password"[^>]*>.*?</span>\s*([^<]+)</span>`)
	matches := re.FindStringSubmatch(html)

	page.MustWaitNavigation()
	page.MustWaitDOMStable()
	if len(matches) > 1 {
		log.Println("Login Fail: ", matches[1])
		return ""
	}
	page.MustWaitNavigation()
	page.MustWaitDOMStable()
	log.Println("Login Successful.")
	return ""
}

// pushEnergySensor 推送一个能源面板可识别的传感器
func pushEnergySensor(entity string, state float64, unit, deviceClass string) error {
	if mqttClient == nil {
		mqttClient = newMQTTClient()
	}

	// 计算 unique_id / device.identifiers
	hash := fmt.Sprintf("%x", md5.Sum([]byte(entity+deviceClass)))
	uniqueID := hash[:8]
	deviceID := hash[:3]

	// 配置 Topic
	// 确保 entity 只包含小写字母、数字和下划线
	safeEntity := strings.ReplaceAll(entity, ".", "_")
	configTopic := fmt.Sprintf("homeassistant/sensor/%s/config", safeEntity)
	stateTopic := fmt.Sprintf("homeassistant/sensor/%s/state", safeEntity)

	// 配置 Payload
	configPayload := fmt.Sprintf(`{
		"device_class": "%s",
		"state_topic": "%s",
		"unit_of_measurement": "%s",
		"value_template": "{{ value_json.state }}",
		"unique_id": "0x%s",
		"device": {
			"identifiers": ["%s"],
			"name": "%s"
		},
		"state_class": "total"
	}`, deviceClass, stateTopic, unit, uniqueID, entity+deviceID, entity)

	// 发布 Discovery 配置
	token := mqttClient.Publish(configTopic, 0, true, configPayload)
	token.WaitTimeout(1 * time.Second)
	if token.Error() != nil {
		return fmt.Errorf("failed to publish config to MQTT: %w", token.Error())
	}

	// 发布状态
	statePayload := fmt.Sprintf(`{"state": %.3f}`, state)
	token = mqttClient.Publish(stateTopic, 0, true, statePayload)
	if token.Error() != nil {
		return fmt.Errorf("failed to publish state to MQTT: %w", token.Error())
	}

	log.Printf("MQTT published entity '%s': state=%.3f %s\n", entity, state, unit)
	return nil
}

// 定义结构体
type MonthlyUsage struct {
	CommonInfo struct {
		Timestamp string `json:"timestamp"`
	} `json:"commonInfo"`

	BillInfo struct {
		UsedMonth            string `json:"usedMonth"`
		BillingStatus        string `json:"billingStatus"`
		ElectricRateCategory string `json:"electricRateCategory"`
		TimezonePrice        string `json:"timezonePrice"`

		MeterInfo struct {
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
		} `json:"meterInfo"`

		UsedInfo struct {
			StartDate string  `json:"startDate"`
			EndDate   string  `json:"endDate"`
			Charge    float64 `json:"charge"`
			Power     float64 `json:"power"`
			Unit      string  `json:"unit"`
		} `json:"usedInfo"`

		PredictionInfo struct {
			Charge string `json:"charge"`
			Power  string `json:"power"`
			Unit   string `json:"unit"`
		} `json:"predictionInfo"`
	} `json:"billInfo"`
}

/**
 * 反序列化每月用量
 */
func ParseMonthlyUsage(jsonStr string) *MonthlyUsage {
	var data MonthlyUsage
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		panic(err)
	}
	return &data
}

type MonthHistory struct {
	CommonInfo struct {
		Timestamp string `json:"timestamp"`
	} `json:"commonInfo"`
	BillInfos []struct {
		UsedMonth     string `json:"usedMonth"`
		BillingStatus string `json:"billingStatus"`
		ContractClass string `json:"contractClass,omitempty"`
		DetailInfos   []struct {
			BillingClass  string  `json:"billingClass"`
			UsedCharge    float64 `json:"usedCharge"`
			UsedPowerInfo struct {
				Power float64 `json:"power"`
				Unit  string  `json:"unit"`
			} `json:"usedPowerInfo"`
			PaymentDateInfo struct {
				PaymentDue      string `json:"paymentDue,omitempty"`
				AccountTransfer string `json:"accountTransfer,omitempty"`
				NextBilling     string `json:"nextBilling,omitempty"`
			} `json:"paymentDateInfo,omitempty"`
		} `json:"detailInfos,omitempty"`
	} `json:"billInfos"`
}

/**
 * 反序列化每月用量
 */
func ParseMonthHistory(jsonStr string) *MonthHistory {
	var data MonthHistory
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		panic(err)
	}
	return &data
}
