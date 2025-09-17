package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
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

	mqttClient = newMQTTClient()
	defer mqttClient.Disconnect(250)

	task(username, password)
}

/**
 *
 */
func task(username, password string) string {
	// 启动浏览器
	// 启动浏览器
	l := launcher.New().
		Bin("/usr/bin/chromium").
		Headless(true).                                        // headless 模式
		Set("no-sandbox", "").                                 // --no-sandbox
		Set("disable-dev-shm-usage", "").                      // --disable-dev-shm-usage
		Set("disable-gpu", "").                                // --disable-gpu
		Set("lang", "ja").                                     // --lang=ja
		Set("disable-desktop-notifications", "").              // --disable-desktop-notifications
		Set("disable-blink-features", "AutomationControlled"). // --disable-blink-features=AutomationControlled
		Set("ignore-certificate-errors", "").                  // --ignore-certificate-errors
		Set("disable-extensions", "").                         // --disable-extensions
		Set("window-size", "1920,1080").                       // --window-size=1920,1080
		Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36").
		MustLaunch()

	browser := rod.New().ControlURL(l).MustConnect()
	defer browser.MustClose()

	// 打开登录页面
	page := browser.
		MustPage(LOGIN_PAGE)

	page.MustWaitDOMStable()

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

	// initlize http cline
	client := &http.Client{}

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

	log.Printf("Info: close popup")

	// close ads script
	exeCloseAdsScript(page)

	lastMonPowerDataList := exeGetLastestMonCostAndUsageScript(page)
	log.Printf("Last month cost money: %d 円, usage: %d kWh\n", lastMonPowerDataList[0], lastMonPowerDataList[1])

	thisMonPowerDataList := exeGetThisMonCostAndUsageScript(page)
	log.Printf("This month cost money: %d 円, usage: %d kWh\n", thisMonPowerDataList[0], thisMonPowerDataList[1])

	//
	log.Println("Tring to push tepco_last_mon_usage")
	if err := pushEnergySensor(client, "sensor.tepco_last_mon_cost", float64(lastMonPowerDataList[0]), "JPY", "monetary"); err != nil {
		log.Println("Err: ", err)
	}

	// 燃气费用
	log.Println("Tring to push tepco_last_mon_cost")
	if err := pushEnergySensor(client, "sensor.tepco_last_mon_usage", float64(lastMonPowerDataList[1]), "kWh", "energy"); err != nil {
		log.Println("Err: ", err)
	}

	//
	log.Println("Tring to push tepco_this_mon_cost")
	if err := pushEnergySensor(client, "sensor.tepco_this_mon_cost", float64(thisMonPowerDataList[0]), "JPY", "monetary"); err != nil {
		log.Println("Err: ", err)
	}

	log.Println("Tring to push tepco_this_mon_usage")
	if err := pushEnergySensor(client, "sensor.tepco_this_mon_usage", float64(thisMonPowerDataList[1]), "kWh", "energy"); err != nil {
		log.Println("Err: ", err)
	}

	return ""
}

// 执行关闭广告的脚本
func exeCloseAdsScript(page *rod.Page) {
	// close ads script
	javascript := rod.Eval(`
		() => {
		 	const list = ["close_icon", "close_about", "close_button", "btn_close"];
		 	list.forEach((el) => {
				const nodeList = document.getElementsByClassName(el);
				Array.from(nodeList).forEach((node) => {
					node.click();
				})
			})
			return void 0;
		}`)

	for closeAttempts := 0; ; closeAttempts++ {
		log.Println("Tring to close ads")
		if _, err := page.Evaluate(javascript); err != nil {
			log.Printf("error: %v\n", err)
			break
		}
		time.Sleep(300 * time.Millisecond)
		if closeAttempts > 20 {
			break
		}
	}
}

// 获取最近一个月的电费和使用量
func exeGetLastestMonCostAndUsageScript(page *rod.Page) [2]int {
	log.Println("Execute get lastest Month Cost script...")
	// close ads script
	javascript := rod.Eval(`
		() => {
		 	document.querySelector("#gaclick_top_graph_yen").click();
			const btnList = Array.from(document.querySelector("ul.month_list").childNodes).filter(e => e.nodeName !== "#comment");
			const target = btnList[btnList.length - 2];
			target.click();
			const cost = document.querySelector("p.price.selected_month").innerText;
			document.querySelector("#gaclick_top_graph_kwh").click();
			const usage = document.querySelector("p.price.selected_month").innerText;
			return usage + "|" + cost
		}`)

	log.Println("Tring to excute javascript")
	cost, err := page.Evaluate(javascript)
	if err != nil {
		log.Printf("error: %v\n", err)
		return [2]int{0, 0}
	}
	time.Sleep(300 * time.Millisecond)

	result := cost.Value.Str()
	list := strings.Split(result, "|")
	return [2]int{extractNumberInt(list[0]), extractNumberInt(list[1])}
}

// 获取本月的电费
func exeGetThisMonCostAndUsageScript(page *rod.Page) [2]int {
	log.Println("Execute get this Month cost script...")
	javascript := rod.Eval(`
		() => {
		 	document.querySelector("#gaclick_top_graph_yen").click();
			const btnList = Array.from(document.querySelector("ul.month_list").childNodes).filter(e => e.nodeName !== "#comment");
			const target = btnList[btnList.length - 1];
			target.click();
			const cost = document.querySelector("p.price.selected_month").innerText;
			document.querySelector("#gaclick_top_graph_kwh").click();
			const usage = document.querySelector("p.price.selected_month").innerText;
			return usage + "|" + cost
		}`)

	log.Println("Tring to excute javascript")
	cost, err := page.Evaluate(javascript)
	if err != nil {
		log.Printf("error: %v\n", err)
		return [2]int{0, 0}
	}
	time.Sleep(300 * time.Millisecond)

	result := cost.Value.Str()
	list := strings.Split(result, "|")
	return [2]int{extractNumberInt(list[0]), extractNumberInt(list[1])}
}

func extractNumber(s string) string {
	re := regexp.MustCompile(`[\d,]+`)
	match := re.FindString(s)
	// 去掉逗号
	return strings.ReplaceAll(match, ",", "")
}

func extractNumberInt(s string) int {
	str := extractNumber(s)
	n, _ := strconv.Atoi(str)
	return n
}

// pushEnergySensor 推送一个能源面板可识别的传感器
func pushEnergySensor(client *http.Client, entity string, state float64, unit, deviceClass string) error {
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
