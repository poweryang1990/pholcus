package spider_lib

// 基础包
import (
	"github.com/PuerkitoBio/goquery"                        //DOM解析
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
    //"github.com/henrylee2cn/pholcus/logs"               //信息输出
	. "github.com/henrylee2cn/pholcus/app/spider"           //必需
	// . "github.com/henrylee2cn/pholcus/app/spider/common"          //选用

	// net包
	// "net/http" //设置http.Header
	// "net/url"

	// 编码包
	// "encoding/xml"
	// "encoding/json"

	// 字符串处理包
    
	"strings"
	// 其他包
	"fmt"
	// "math"
	//"time"
	"regexp"
)

func init() {
	ProxyIpDownload.Register()
}

var ProxyIpDownload = &Spider{
	Name:        "代理IP爬取",
	Description: "**抓区后使用：使用时 复制 Url列粘贴到proxy.lib中  重新 以管理员方式运行爬虫程序",
	Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
    Namespace: func(self *Spider) string {
		return "proxyip"
	},
    SubNamespace: func(self *Spider, dataCell map[string]interface{}) string {
        return "http"
	},
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
           ctx.AddQueue(&request.Request{
				Url:  "http://www.proxy-ip.cn/other/1/1",
				Rule: "请求列表",
				Temp: map[string]interface{}{"p": 1},
			})   
		},

		Trunk: map[string]*Rule{

			"请求列表": {
				ParseFunc: func(ctx *Context)  {
                    
                   var curr int
					ctx.GetTemp("p", &curr)
                    if curr<40 {
                         ctx.AddQueue(&request.Request{
                            Url:fmt.Sprintf("http://www.proxy-ip.cn/other/1/%d",curr+1),
                            Rule: "请求列表",
                            ConnTimeout: -1,
                           	Temp: map[string]interface{}{"p":curr+1},
                        }) 
                    }
                    ctx.Parse("获取列表")
                },
			},

			"获取列表": {
                ItemFields: []string{
					"Url","IP", "Port","Address","Type",
				},
				ParseFunc: func(ctx *Context) {          
					ctx.GetDom().
						Find("table.proxy_table tr").
						Each(func(i int, s *goquery.Selection) {
                            if i==0 {
                                return
                            }
							var Url,IP,Port,Address,Type string
                            IP=strings.TrimSpace(s.Find("td").Eq(0).Text())
                            Port=strings.TrimSpace(s.Find("td").Eq(1).Text())
                            Address=strings.TrimSpace(s.Find("td").Eq(2).Text())
                            Type=strings.TrimSpace(s.Find("td").Eq(3).Text())
                            switch Type {
                                case "HTTP":
                                    Url="http://"+IP+":"+Port
                                 case "HTTPS":
                                    Url="https://"+IP+":"+Port
                            }
                            spaceReg:=regexp.MustCompile("\\s+")
                            Url=spaceReg.ReplaceAllString(Url,"")
                            ctx.Output(map[int]interface{}{
                                0:Url,1:IP,2:Port,3: Address,4:Type,
                            })
						})
				},
			},

			
		},
	},
}
