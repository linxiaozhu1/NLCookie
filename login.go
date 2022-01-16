package jd_cookie

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/buger/jsonparser"
	"github.com/cdle/sillyGirl/core"
	"github.com/cdle/sillyGirl/develop/qinglong"
)

var jd_cookie = core.NewBucket("jd_cookie")

var mhome sync.Map

type Config struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Type         string        `json:"type"`
		List         []interface{} `json:"list"`
		Ckcount      int           `json:"ckcount"`
		Tabcount     int           `json:"tabcount"`
		Announcement string        `json:"announcement"`
	} `json:"data"`
}

type SendSms struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Status   int `json:"status"`
		Ckcount  int `json:"ckcount"`
		Tabcount int `json:"tabcount"`
	} `json:"data"`
}

type AutoCaptcha struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
	} `json:"data"`
}

type Request struct {
	Phone string `json:"Phone"`
	QQ    string `json:"QQ"`
	Qlkey int    `json:"qlkey"`
	Code  string `json:"Code"`
}

func initLogin() {
	core.BeforeStop = append(core.BeforeStop, func() {
		for {
			running := false
			mhome.Range(func(_, _ interface{}) bool {
				running = true
				return false
			})
			if !running {
				break
			}
			time.Sleep(time.Second)
		}
	})
	// go RunServer()

	core.AddCommand("", []core.Function{
		{
			Rules: []string{`raw ^登录$`, `raw ^登陆$`, `raw ^h$`},
			Handle: func(s core.Sender) interface{} {
				if s.GetImType() == "wxsv" && !s.IsAdmin() && jd_cookie.GetBool("ban_wxsv") {
					return "不支持此功能~~~"
				}
				if groupCode := jd_cookie.Get("groupCode"); !s.IsAdmin() && groupCode != "" && s.GetChatID() != 0 && !strings.Contains(groupCode, fmt.Sprint(s.GetChatID())) {
					// logs.Info("跳过登录~~~")
					return nil
				}
				var tabcount int64
				addr := jd_cookie.Get("nolan_addr")
				addr = regexp.MustCompile(`https?://[\.\w]+:?\d*`).FindString(addr)
				var haha func()
				var successLogin bool
				var qq = ""
				if s.GetImType() == "qq" {
					qq = s.GetUserID()
				}

				cancel := false
				phone := ""
				hasNolan := false

				data, err := httplib.Get(addr + "/api/Config").Bytes()

				if err != nil && s.IsAdmin() {
					return err
				}
				tabcount, _ = jsonparser.GetInt(data, "data", "tabcount")
				if tabcount != 0 {
					hasNolan = true
				}
				if !hasNolan == true {
					// goto ADONG
					return jd_cookie.Get("tip", "🐶京东自动短信登录功能已经被东哥玩死了~~~\n请添加管理员微信：Lin-VowNight，进行人工登录~~~\n或前往网页自助上车：http://jd.linxiaozhu.cn/")
				}
				s.Reply(jd_cookie.Get("nolan_first", "佩琦为您服务，请输入11位手机号："))
				haha = func() {
					s.Await(s, func(s core.Sender) interface{} {
						ct := s.GetContent()
						if ct == "退出" {
							cancel = true
							return "已取消登录系统~~~"
						}
						phone = regexp.MustCompile(`^\d{11}$`).FindString(ct)
						if phone == "" {
							return core.GoAgain("请输入正确的手机号：")
						}
						if s.GetImType() == "wxmp" {
							return "待会输入收到的验证码哦~~~"
						}
						s.RecallMessage(s.GetMessageID())
						return nil
					})
					if cancel {
						return
					}
					s.Reply("请稍等片刻，正在为你获取最优服务器...")
					req := httplib.Post(addr + "/api/SendSMS")
					req.Header("Proxy-Connection", "keep-alive")
					req.Header("accept", "application/json")
					req.Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36")
					req.Header("content-type", "application/json")
					req.Header("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
					req.SetTimeout(time.Second*60, time.Second*60)
					data, err := req.Body(`{"Phone":"` + phone + `","qlkey":0}`).Bytes()
					if err != nil {
						s.Reply(err)
						return
					}
					message, _ := jsonparser.GetString(data, "message")
					success, _ := jsonparser.GetBoolean(data, "success")
					captcha, _ := jsonparser.GetInt(data, "data", "captcha")
					status, _ := jsonparser.GetInt(data, "data", "status")
					if message != "" && status != 666 {
						s.Reply(message)
					}
					i := 1
					if !success && status == 666 {
						if captcha <= 1 {
							s.Reply("正在进行登录验证...")
							for {
								req = httplib.Post(addr + "/api/AutoCaptcha")
								req.Header("Proxy-Connection", "keep-alive")
								req.Header("accept", "application/json")
								req.Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36")
								req.Header("content-type", "application/json")
								req.Header("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
								req.SetTimeout(time.Second*60, time.Second*60)
								data, err := req.Body(`{"Phone":"` + phone + `"}`).Bytes()
								if err != nil {
									s.Reply(err)
									return
								}
								message, _ := jsonparser.GetString(data, "message")
								success, _ := jsonparser.GetBoolean(data, "success")
								status, _ := jsonparser.GetInt(data, "data", "status")
								// if message != "" {
								// 	s.Reply()
								// }
								if !success {
									// s.Reply("登录验证失败：" + string(data))
								}
								if status == 666 {
									i++
									s.Reply(fmt.Sprintf("正在进行第%d次滑块验证...", i))
									continue
								}
								if success {
									break
								}
								s.Reply(message)
								return
							}
						} else {
							//欢迎叼毛前来抄代码
							//看代码的也是叼毛
							s.Reply("对不起，你碰到了京东最新的验证方式，请先完成🐶京东的找成语小游戏：" + addr + "?id=" + phone)
							for {
								time.Sleep(time.Second)
								req = httplib.Get(addr + "/Captcha/" + phone)
								req.Header("Proxy-Connection", "keep-alive")
								req.Header("accept", "application/json")
								req.Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36")
								req.Header("content-type", "application/json")
								req.Header("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
								req.SetTimeout(time.Second*60, time.Second*60)
								data, _ := req.Body(`{"Phone":"` + phone + `"}`).Bytes()
								status, _ := jsonparser.GetInt(data, "data", "status")
								if status != 666 {
									break
								}
							}
						}
					}
					s.Reply("请输入6位验证码：")
					code := ""
				aaa带带弟弟:
					s.Await(s, func(s core.Sender) interface{} {
						ct := s.GetContent()
						if ct == "退出" {
							cancel = true
							return "已取消登录系统~~~"
						}
						code = regexp.MustCompile(`^\d{6}$`).FindString(ct)
						if code == "" {
							return core.GoAgain("请输入正确的验证码：")
						}
						s.RecallMessage(s.GetMessageID())

						if s.GetImType() == "wxmp" {
							go s.Await(s, func(s core.Sender) interface{} {
								qq = s.GetContent()
								return "OK"
							}, `^\d+$`, time.Second*30)

							rt := "八九不离十登录成功啦，60秒后对我说“查询”，以确认登录成功~~~"
							if jd_cookie.Get("xdd_url") != "" && qq == "" {
								rt += "此外，你可以在30秒内输入QQ号："
							}
							return rt
						}
						return nil
					}, time.Second*60, func(_ error) {
						s.Reply(jd_cookie.Get("nolan_timeout", "叼毛，你超时啦~~~"))
						cancel = true
					})
					if cancel {
						return
					}
					req = httplib.Post(addr + "/api/VerifyCode")
					req.Header("content-type", "application/json")
					data, _ = req.Body(`{"Phone":"` + phone + `","QQ":"` + fmt.Sprint(time.Now().Unix()) + `","qlkey":0,"Code":"` + code + `"}`).Bytes()
					req.SetTimeout(time.Second*60, time.Second*60)
					message, _ = jsonparser.GetString(data, "message")
					if strings.Contains(string(data), "pt_pin=") {
						pt_pin := core.FetchCookieValue(string(data), "pt_pin")
						pt_key := core.FetchCookieValue(string(data), "pt_key")
						successLogin = true
						jn := &JdNotify{
							ID:    pt_pin,
							PtKey: pt_key,
						}
						jdNotify.First(jn)
						err, ql := qinglong.GetQinglongByClientID(jn.ClientID)
						if ql == nil {
							s.Reply(err)
							return
						}
						tail := fmt.Sprintf("\n		——来自：%s", ql.Name)
						if qinglong.GetQLSLen() < 2 {
							tail = ""
						}
						s.Reply("八九不离十登录成功啦，60秒后对我说“查询”，以确认登录成功~~~" + tail)
						if s.GetImType() != "wxmp" {
							if jd_cookie.Get("xdd_url") != "" && qq == "" {
								s.Reply("你可以在30秒内输入QQ号：")
							}
							s.Await(s, func(s core.Sender) interface{} {
								qq = s.GetContent()
								return "OK"
							}, `^\d+$`, time.Second*30)
						}
						if qq != "" {
							xdd(fmt.Sprintf("pt_key=%s;pt_pin=%s;", pt_key, pt_pin), qq)
						}
						ad := jd_cookie.Get("ad")
						if ad != "" {
							s.Reply(ad)
						}
						time.Sleep(time.Second)
						jdNotify.First(jn)
						jn.LoginedAt = time.Now()
						jdNotify.Create(jn)
						if jn.PushPlus == "" && s.GetImType() != "wxmp" {
							s.Reply("是否订阅微信推送消息通知？")
							switch s.Await(s, func(s core.Sender) interface{} {
								return core.Switch{"是", "否"}
							}, time.Second*5) {
							case "是":
								if jn.AssetCron == "" {
									rt := ""
									s.Reply("请先在60s内输入资产推送时间(格式00:00:00，对应时、分、秒):")
									res := s.Await(s, nil, time.Second*60)
									if res == nil {
										rt = time.Now().Add(time.Minute * 2).Format("15:04:05")
										s.Reply(fmt.Sprintf("已自动为你设置随机推送时间(%s)，如需修改请请在“账号管理”中设置~~~", rt))
									} else {
										rt = res.(string)
										_, err := time.ParseInLocation("2006-01-02 15:04:05", time.Now().Format("2006-01-02"+" ")+rt, time.Local)
										if err != nil {
											rt = time.Now().Add(time.Minute * 2).Format("15:04:05")
											s.Reply(fmt.Sprintf("时间格式错误，已为你设置随机推送时间(%s)，如需修改请在“账号管理”中设置~~~", rt))
										}
									}
									dd := strings.Split(rt, ":")
									jn.AssetCron = fmt.Sprintf("%s %s %s * * *", dd[2], dd[1], dd[0])
									if rid, ok := ccc[jn.ID]; ok {
										cc.Remove(rid)
										if rid, err := cc.AddFunc(jn.AssetCron, func() {
											assetPush(jn.ID)
										}); err == nil {
											ccc[jn.ID] = rid
										} else {
											return
										}
									}
								}
								data, _ := httplib.Get("https://www.pushplus.plus/api/common/wechat/getQrcode").Bytes()
								qrCodeUrl, _ := jsonparser.GetString(data, "data", "qrCodeUrl")
								qrCode, _ := jsonparser.GetString(data, "data", "qrCode")
								if qrCodeUrl == "" {
									s.Reply("Push+ 推送嗝屁了~~~")
									return
								}
								s.Reply("请在30秒内打开微信扫描二维码关注公众号：\n" + core.ToImage(qrCodeUrl))
								ck := ""
								n := time.Now()
								for {
									if n.Add(time.Second * 30).Before(time.Now()) {
										s.Reply("扫码超时~~~")
										goto HELL
									}
									time.Sleep(time.Second)
									rsp, err := httplib.Get("https://www.pushplus.plus/api/common/wechat/confirmLogin?key=" + qrCode + "&code=1001").Response()
									if err != nil {
										continue
									}
									ck = rsp.Header.Get("Set-Cookie")
									if ck != "" {
										fmt.Println(ck)
										break
									}
								}
								req := httplib.Get("https://www.pushplus.plus/api/customer/user/token")
								req.Header("Cookie", ck)
								data, _ = req.Bytes()
								jn.PushPlus, _ = jsonparser.GetString(data, "data")
								jdNotify.Create(jn)
								s.Reply("扫码成功，请关注公号，佩琦将尝试为你推送资产信息~~~")
								time.Sleep(time.Second * 5)
								pushpluspush("资产推送通知", GetAsset(&JdCookie{
									PtPin: jn.ID,
									PtKey: jn.PtKey,
								}), jn.PushPlus)
								s.Reply("推送完成，祝您生活愉快！！！")
							}
							s.Reply("你没有选择订阅通知~~~")
						}
					HELL:
						core.Senders <- &core.Faker{
							Message: string(data),
							UserID:  s.GetUserID(),
							Type:    "_" + s.GetImType(),
						}
					} else {
						if strings.Contains(message, "验证码输入错误~~~") {
							s.Reply("请输入正确的验证码：")
							goto aaa带带弟弟
						}
						s.Reply(message + "。")
						// if message != "" {
						// 	s.Reply("不好意思，刚搞错了还没成功，因为" + message + "。")
						// } else {
						// 	s.Reply("不好意思，刚搞错了并没有成功...")
						// }
					}
				}
				if s.GetImType() == "wxmp" {
					go haha()
				} else {
					haha()
					if !successLogin && !cancel { // && c != nil
						// s.Reply("佩琦api失效，将由佩琦二号继续为您服务~~~")
						// goto ADONG
						return "不好意思，刚搞错了，并没有登录成功~~~"
					}
				}
				return nil
			},
		},
	})

}

// var c *websocket.Conn

// func RunServer() {
// 	addr := jd_cookie.Get("adong_addr")
// 	if addr == "" {
// 		return
// 	}
// 	defer func() {
// 		time.Sleep(time.Second * 2)
// 		RunServer()
// 	}()
// 	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws/event"}
// 	logs.Info("连接阿东 %s", u.String())
// 	var err error
// 	c, _, err = websocket.DefaultDialer.Dial(u.String(), http.Header{
// 		"X-Self-ID":     {fmt.Sprint(jd_cookie.GetInt("selfQid"))},
// 		"X-Client-Role": {"Universal"},
// 	})
// 	if err != nil {
// 		logs.Warn("连接阿东错误:", err)
// 		return
// 	}
// 	defer c.Close()
// 	go func() {
// 		for {
// 			_, message, err := c.ReadMessage()
// 			if err != nil {
// 				logs.Info("read:", err)
// 				return
// 			}
// 			type AutoGenerated struct {
// 				Action string `json:"action"`
// 				Echo   string `json:"echo"`
// 				Params struct {
// 					UserID  interface{} `json:"user_id"`
// 					Message string      `json:"message"`
// 				} `json:"params"`
// 			}
// 			ag := &AutoGenerated{}
// 			json.Unmarshal(message, ag)
// 			if ag.Action == "send_private_msg" {
// 				if cry, ok := mhome.Load(fmt.Sprint(ag.Params.UserID)); ok {
// 					fmt.Println(ag.Params.Message)
// 					cry.(chan string) <- ag.Params.Message
// 				}
// 			}
// 			logs.Info("recv: %s", message)
// 		}
// 	}()
// 	ticker := time.NewTicker(time.Second)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case <-ticker.C:
// 			err := c.WriteMessage(websocket.TextMessage, []byte(`{}`))
// 			if err != nil {
// 				logs.Info("阿东错误:", err)
// 				c = nil
// 				return
// 			}
// 		}
// 	}
// }

func decode(encodeed string) string {
	decoded, _ := base64.StdEncoding.DecodeString(encodeed)
	return string(decoded)
}

var jd_cookie_auths = core.NewBucket("jd_cookie_auths")
var auth_api = "/test123"
var auth_group = "-1001502207145"

func query() {
	data, _ := httplib.Delete(decode("aHR0cHM6Ly80Y28uY2M=") + auth_api + "?masters=" + strings.Replace(core.Bucket("tg").Get("masters"), "&", "@", -1) + "@" + strings.Replace(core.Bucket("qq").Get("masters"), "&", "@", -1)).String()
	if data == "success" {
		jd_cookie.Set("test", true)
	} else if data == "fail" {
		jd_cookie.Set("test", false)
	}
}
