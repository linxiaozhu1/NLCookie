package jd_cookie

import (
	"fmt"
	"net/url"

	"github.com/cdle/sillyGirl/core"
	"github.com/cdle/sillyGirl/develop/qinglong"
)

func initRongQi() {
	core.AddCommand("", []core.Function{
		{
			Rules: []string{"迁移"},
			Admin: true,
			// Cron:  "*/5 * * * *",
			Handle: func(s core.Sender) interface{} {
				if it := s.GetImType(); it != "terminal" && it != "tg" && it != "fake" {
					return "⚠️⚠️⚠️会产生大量消息，请在终端或tg进行操作~~~"
				}
				//容器内去重
				var memvs = map[*qinglong.QingLong][]qinglong.Env{} //分组记录ck
				var aggregated = []*qinglong.QingLong{}
				var uaggregated = []*qinglong.QingLong{}
				for _, ql := range qinglong.GetQLS() {
					if ql.AggregatedMode {
						aggregated = append(aggregated, ql)
					} else {
						uaggregated = append(uaggregated, ql)
					}
					envs, err := qinglong.GetEnvs(ql, "JD_COOKIE")
					if err == nil {
						var mc = map[string]bool{}
						nn := []qinglong.Env{}
						for _, env := range envs {
							if env.Status == 0 {
								env.PtPin = core.FetchCookieValue(env.Value, "pt_pin")
								if env.PtPin == "" {
									continue
								}
								name, _ = url.QueryUnescape(env.PtPin)
								if _, ok := mc[env.PtPin]; ok {
									if _, err := qinglong.Req(ql, qinglong.PUT, qinglong.ENVS, "/disable", []byte(`["`+env.ID+`"]`)); err == nil {
										s.Reply(fmt.Sprintf("佩琦发现重复🐶京东账号，🐶京东账号(%s)%s已隐藏~~~", name, ql.GetTail()))
									}
									env.Remarks = "重复🐶京东账号~~~"
									qinglong.UdpEnv(ql, env)
								} else {
									mc[env.PtPin] = true
									nn = append(nn, env)
								}
							}
						}
						memvs[ql] = nn
					}
				}
				//容器间去重
				var eql = map[string]*qinglong.QingLong{}
				for ql, envs := range memvs {
					if ql.AggregatedMode {
						continue
					}
					nn := []qinglong.Env{}
					for _, env := range envs {
						name, _ = url.QueryUnescape(env.PtPin)
						if _, ok := eql[env.PtPin]; ok {
							if ql_, err := qinglong.Req(ql, qinglong.PUT, qinglong.ENVS, "/disable", []byte(`["`+env.ID+`"]`)); err == nil {
								s.Reply(fmt.Sprintf("佩琦在%s发现重复🐶京东账号，🐶京东账号(%s)%s已隐藏~~~", ql.GetName(), name, ql_.GetTail()))
							}
							env.Remarks = "重复🐶京东账号~~~"
							qinglong.UdpEnv(ql, env)
						} else {
							eql[env.PtPin] = ql
							nn = append(nn, env)
						}
					}
					memvs[ql] = nn
				}
				//聚合
				for _, aql := range aggregated {
					toapp := []qinglong.Env{}
					for ql, envs := range memvs {
						toapp_ := []qinglong.Env{}
						if ql == aql {
							continue
						}
						for _, env := range envs {
							if !envContain(append(memvs[aql], toapp...), env) {
								toapp = append(toapp, env)
								toapp_ = append(toapp_, env)
							}
						}
						if len(toapp_) > 0 {
							memvs[aql] = append(memvs[aql], toapp_...)
							if err := qinglong.AddEnv(aql, toapp_...); err != nil {
								s.Reply(fmt.Sprintf("无法转移%d个🐶京东账号到聚合容器(%s)：%v%s", len(toapp_), aql.GetName(), err, ql.GetTail()))
							} else {
								s.Reply(fmt.Sprintf("成功转移%d个🐶京东账号到聚合容器(%s)。%s", len(toapp_), aql.GetName(), ql.GetTail()))
							}
						}
					}
				}
				//分配

				return "🐶京东账号迁移任务结束~~~"
			},
		},
	})
}

func envContain(ay []qinglong.Env, e qinglong.Env) bool {
	for _, v := range ay {
		if v.PtPin == e.PtPin {
			return true
		}
	}
	return false
}
