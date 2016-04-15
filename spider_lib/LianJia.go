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
	"fmt"
	"math"
	//"time"
	"regexp"
)

func init() {
	LianJia.Register()
}
func Round(f float64, places int) (float64) {
    shift := math.Pow(10, float64(places))
    return math.Floor(f * shift + .5) / shift
}
type HouseSourceVar struct{
    MaxPage int
    CityCode string
    Areas []string
}
var houseSourceSet=map[string]HouseSourceVar{
    "成都":HouseSourceVar{
        MaxPage:100,
        CityCode:"cd",
        Areas:[]string{"jinjiang","qingyang","wuhou","gaoxinnan","chenghua","jinniu","tianfuxinqu","shuangliu","wenjiang","pixian","longanyi","xindou"},
    },
    "武汉":HouseSourceVar{
        MaxPage:100,
        CityCode:"wh",
        Areas:[]string{"jiangan","jianghan","qiaokou","dongxihu","wuchang","qingshan","hongshan","hanyang"},
    },
    "北京":HouseSourceVar{
        MaxPage:100,
        CityCode:"bj",
        Areas:[]string{"dongcheng","xicheng","chaoyang","haidian","fengtai","shijingshan","tongzhou","changping","daxing","yizhuangkaifaqu","shunyi","fangshan","mentougou","pinggu","huairou","miyun","yanqing","yanjiao"},
    },
    "杭州":HouseSourceVar{
        MaxPage:100,
        CityCode:"hz",
        Areas:[]string{"jiande","xihu","xiacheng","jianggan","gongshu","shangcheng","binjiang","yuhang","xiaoshan","linan"},
    },
   
}
var LianJia = &Spider{
	Name:        "链家房源抓取",
	Description: "**链家房源抓取**",
	Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
    Namespace: func(self *Spider) string {
		return "housesource"//表名
	},
    SubNamespace: func(self *Spider, dataCell map[string]interface{}) string {
        return "lianjia"
		//return dataCell["Data"].(map[string]interface{})["城市"].(string)//根据数据内容来划分 用来才拆分多个表 不能返回 "" 可以返回 nil(默认)
	},
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
            for setting := range houseSourceSet {
                //ctx.SetTimer(setting, time.Minute*1, nil)
				ctx.Aid(map[string]interface{}{"setting":setting}, "请求列表")
			}       
		},

		Trunk: map[string]*Rule{

			"请求列表": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
                    
                    key := aid["setting"].(string)
					value := houseSourceSet[key]  
                    for _, area := range value.Areas {
                             
                        for page := 1; page <=value.MaxPage; page++ {  
                            //整租
                            ctx.AddQueue(&request.Request{
                                Url:fmt.Sprintf("http://%s.lianjia.com/zufang/%s/pg%d/",value.CityCode,area,page),
                                Rule: "获取列表",
                                ConnTimeout: -1,
								Reloadable: true,
								Temp: map[string]interface{}{"site":fmt.Sprintf("http://%s.lianjia.com",value.CityCode)},
                            })
                            
                        }   
                    }
				return nil
                },
			},

			"获取列表": {
				ParseFunc: func(ctx *Context) {
                    
					ctx.GetDom().
						Find("ul.house-lst .pic-panel a").
						Each(func(i int, s *goquery.Selection) {
							url, _ := s.Attr("href")
							var site string
                            ctx.GetTemp("site", &site)
                            httpUrlReg:=regexp.MustCompile("https?://(.*?)+")
                            
                            if !httpUrlReg.MatchString(url) {
                                url=site+url
                            }
							ctx.AddQueue(&request.Request{
								Url: url,
								Rule: "输出结果",
								ConnTimeout: -1,
								Priority: 1,
							})
						})
				},
			},

			"输出结果": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"城市", "区域","商圈","小区", "地址","出租类型","房屋类型","房间大小","户型","租金","配置","装修","更新时间","楼层","经纪人","联系电话","链家发布","单价",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					var 城市, 区域,商圈,小区, 地址,出租类型,房屋类型,房间大小,户型,租金,配置,装修,更新时间,楼层,经纪人,联系电话,链家发布 string
                    var 单价 float64
                    
                     城市 = strings.Replace(query.Find(".fl.l-txt a").Eq(1).Text(),"租房","",-1)
                     区域 = strings.Replace(query.Find(".fl.l-txt a").Eq(2).Text(),"租房","",-1)
                     商圈 = strings.Replace(query.Find(".fl.l-txt a").Eq(3).Text(),"租房","",-1)
                   
                    // 地址 = query.Find("#lj-common-bread > div.container > div.fl l-txt").Text()
                     出租类型 = "整租"
                    // 房屋类型 = query.Find("#lj-common-bread > div.container > div.fl l-txt").Text()
                    // 配置 = query.Find("#lj-common-bread > div.container > div.fl l-txt").Text()
                     装修 = query.Find(".decoration-ex span").Text()
                  
                 
                    
                    query.Find(".info-box.left dl").Each(func(i int,  s *goquery.Selection) {
						dt := s.Find("dt").Text()
                        
						switch dt {
						case "租金：":  
							租金 = s.Find(".ft-num").Text()
                            房间大小=strings.Replace(s.Find(".em-text i").Text(),"㎡","",-1)
                            房间大小=strings.Replace(房间大小,"/","",-1)
                            房间大小=strings.TrimSpace(房间大小)
                            
                        case "户型：":  
							户型 = s.Find("dd").Text()
                            
                        case "楼层：":  
							楼层 = s.Find("dd").Text()
                                
                        case "小区：":  
							小区 = s.Find(".zone-name.laisuzhou").Text()
                               
                        case "更新：":  
							更新时间 = s.Find("dd").Text()
                            更新时间=strings.Replace(更新时间,"年","-",-1)
                            更新时间=strings.Replace(更新时间,"月","-",-1)
                            更新时间=strings.Replace(更新时间,"日","-",-1)       
						}
					})
                    经纪人 = query.Find(".p-del.right .p-01 a").Eq(0).Text()
                    联系电话 = query.Find(".contact-panel .ft-num").Text()
                    链家发布 = "是"
                    
                    price,_:= strconv.ParseFloat(租金,64)   
                    size,_:=strconv.ParseFloat(房间大小,64)
                    if size!=0 {
                     单价= Round(price/size,2)
                    }
                   

					// 结果输出方式一（推荐）
					ctx.Output(map[int]interface{}{
						0:城市,1:区域,2:商圈,3:小区,4: 地址,5:出租类型,6:房屋类型,7:房间大小,8:户型,9:租金,10:配置,11:装修,12:更新时间,13:楼层,14:经纪人,15:联系电话,16:链家发布,17:单价,
					})

					
				},
			},

			
		},
	},
}
