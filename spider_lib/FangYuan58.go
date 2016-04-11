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
    //"regexp"
	"strconv"
	"strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
)

func init() {
	FangYuan58.Register()
}

var FangYuan58 = &Spider{
	Name:        "成都58房源抓取",
	Description: "**成都58房源抓取**",
	// Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			ctx.AddQueue(&request.Request{
				Url:  "http://cd.58.com/chuzu/",
				Rule: "请求列表",
				Temp: map[string]interface{}{"p": 1},
			})
           
		},

		Trunk: map[string]*Rule{

			"请求列表": {
				ParseFunc: func(ctx *Context) {
					var curr int
					ctx.GetTemp("p", &curr)
					if ctx.GetDom().Find(".pager strong span").Text() != strconv.Itoa(curr) {
						return
					}
					ctx.AddQueue(&request.Request{
						Url:         "http://cd.58.com/chuzu/pn" + strconv.Itoa(curr+1),
						Rule:        "请求列表",
						Temp:        map[string]interface{}{"p": curr + 1},
						ConnTimeout: -1,
					})

					// 用指定规则解析响应流
					ctx.Parse("获取列表")
				},
			},

			"获取列表": {
				ParseFunc: func(ctx *Context) {
					ctx.GetDom().
						Find("table.tbimg .img_list a").
						Each(func(i int, s *goquery.Selection) {
							url, _ := s.Attr("href")
							ctx.AddQueue(&request.Request{
								Url:         url,
								Rule:        "输出结果",
								ConnTimeout: -1,
							})
						})
				},
			},

			"输出结果": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"标题", "区域","商圈","小区", "地址","出租类型","房屋类型","房间大小","户型","租金","配置","装修","更新时间","楼层","经纪人","联系电话","链家发布","单价",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					var 标题, 区域,商圈,小区, 地址,出租类型,房屋类型,房间大小,户型,租金,配置,装修,更新时间,楼层,经纪人,联系电话,链家发布 string
                    标题 = query.Find(".house-title h1").Text()
                    //digitsRegexp:= regexp.MustCompile("*.58.com/(.*?)x.shtml")
                    
                    区域=query.Find("div.xiaoqu a").Eq(0).Text();
                    
                    quyuHref, _:=query.Find("div.xiaoqu a").First().Attr("href")
                    
                    商圈=query.Find("div.xiaoqu a").Eq(1).Text()
                    xiaoquLength:=query.Find("div.xiaoqu a").Length()
                    if xiaoquLength>2 {
                        小区=query.Find("div.xiaoqu a").Eq(2).Text()
                    }else{
                        小区=query.Find("div.xiaoqu").Text()
                    }
                    
                    地址 = query.Find("ul.house-primary-content>li").Eq(3).Find("div.c70").Text()
                    
                    isZhengzu:=strings.Index(quyuHref,"zufang")>=0
                    
                    if isZhengzu {
                         出租类型="整租"
                    }else{
                         出租类型="合租"
                    }
                  
                    fangwuInfoList:= strings.Split(query.Find(".house-type").Text(),"层")
                    fangwuInfoList1:= strings.Split(fangwuInfoList[0],"-")
                    fangwuInfoList2:= strings.Split(fangwuInfoList[1],"-")
                    if isZhengzu {
                        房屋类型=strings.TrimSpace(fangwuInfoList2[2])
                        房间大小=strings.TrimSpace(strings.Replace(fangwuInfoList1[1],"m²","",-1))
                        户型=strings.TrimSpace(fangwuInfoList1[0])
                        装修=strings.TrimSpace(fangwuInfoList2[0])
                        楼层=strings.TrimSpace(fangwuInfoList[2])
                    }else{
                        房屋类型=strings.TrimSpace(fangwuInfoList2[2])
                        房间大小=strings.TrimSpace(strings.Replace(fangwuInfoList1[2],"m²","",-1))
                        户型=strings.TrimSpace(fangwuInfoList1[1])
                        装修=strings.TrimSpace(fangwuInfoList2[0])
                        楼层=strings.TrimSpace(fangwuInfoList1[2])
                    }
                   
					租金 = query.Find("em.house-price").Text()  
                   
                    price,_:= strconv.Atoi(租金)   
                    size,_:=strconv.Atoi(房间大小)
                    单价:= price/size 
                       
                    query.Find("ul.house-primary-content>li").Last().Find("div>span").Each(func (i int, s *goquery.Selection)  {
                        if i==0 {
                            配置=配置+s.Text()
                        }else{
                            配置=配置+","+s.Text()
                        }
                    })
                    
                    更新时间=strings.Replace(query.Find("div.house-title div.title-right-info span").Eq(0).Text(),"更新时间：","",-1)
                    经纪人=query.Find(".tel.cfff span").Eq(1).Text()
                    联系电话=query.Find(".tel.cfff span").Eq(0).Text()
                    isLianJia:=strings.Index(query.Find("div.broker-info-wrap").Text(),"链家")>=0
                    if isLianJia {
                        链家发布="是"
                    }else{
                        链家发布="否"
                    }

					// 结果输出方式一（推荐）
					ctx.Output(map[int]interface{}{
						0:标题,1:区域,2:商圈,3:小区,4: 地址,5:出租类型,6:房屋类型,7:房间大小,8:户型,9:租金,10:配置,11:装修,12:更新时间,13:楼层,14:经纪人,15:联系电话,16:链家发布,17:单价,
					})

					
				},
			},

			
		},
	},
}
