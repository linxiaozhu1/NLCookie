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
			Handle: func(s core.Sender) interface{} {
				//去重
				// var mc = map[string]string{}                       //记录ck对应的clientId
				var mcks = map[*qinglong.QingLong][]qinglong.Env{} //分组记录ck
				for _, ql := range qinglong.GetQLS() {
					tail := fmt.Sprintf("\n		——来自%s", ql.Name)
					envs, err := qinglong.GetEnvs(ql, "JD_COOKIE")
					if err == nil {
						if !ql.AggregatedMode {
							var mc = map[string]bool{}
							nn := []qinglong.Env{}
							for _, env := range envs {
								if env.Status == 0 {
									pt_pin := core.FetchCookieValue(env.Value, "pt_pin")
									pt_pin, _ = url.QueryUnescape(pt_pin)
									if _, ok := mc[pt_pin]; ok {
										if _, err := qinglong.Req(ql, qinglong.PUT, qinglong.ENVS, "/enable", []byte(`["`+env.ID+`"]`)); err == nil {
											s.Reply(fmt.Sprintf("佩琦发现重复🐶京东账号，已隐藏(%s)%s~~~", pt_pin, tail))
										}
										env.Remarks = "假佩琦~~~"
										go qinglong.UdpEnv(ql, env)
									} else {
										mc[pt_pin] = true
										nn = append(nn, env)
									}
								}
							}
							mcks[ql] = nn
						} else {

						}

					}
				}
				//聚合
				//均匀
				return "🐶京东账号迁移完成~~~"
			},
		},
	})
}
