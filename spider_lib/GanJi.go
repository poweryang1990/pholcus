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
    "regexp"
	"strconv"
	"strings"
	// 其他包
	"fmt"
	//"math"
	//"time"
)

func init() {
	GanJi.Register()
}

type HouseSourceGanJi struct{
    MaxPage int
    CityCode string
    Areas []string
}
var houseSourceSetGanJi=map[string]HouseSourceGanJi{
    "成都":HouseSourceGanJi{
        MaxPage:1,
        CityCode:"cd",
        Areas:[]string{"wuhou","qingyang","jinniu","jinjiang","chenghua","gaoxin","shuangliu","wenjiang","pixian","gaoxinxiqu","longquanyi","xindu","qingbaijiang","jintang","dujiangyan","pengzhou","xinjin","chongzhou","dayi","qionglai","pujiang","qita"},
    },
    "武汉":HouseSourceGanJi{
        MaxPage:100,
        CityCode:"wh",
        Areas:[]string{"jianghan","jiangan","qiaokou","wuchang","hongshan","qingshan","hanyang","jingjijishukaifaqu","dongxihu","caidian","huangbei","xinzhou","jiangxia","hannan"},
    },
    "北京":HouseSourceGanJi{
        MaxPage:100,
        CityCode:"bj",
        Areas:[]string{"haidian","chaoyang","dongcheng","xicheng","chongwen","xuanwu","fengtai","shijingshan","changping","tongzhou","daxing","shunyi","fangshan","miyun","mentougou","huairou","pinggu","yanqing","yanjiao","beijingzhoubian"},
    },
    "杭州":HouseSourceGanJi{
        MaxPage:100,
        CityCode:"hz",
        Areas:[]string{"gongshu","xihu","shangcheng","xiacheng","jianggan","binjiang","xiaoshan","yuhang","linan","fuyang","tonglu","jiande","chunan"},
    },
   
}
var GanJi = &Spider{
	Name:        "赶集房源抓取",
	Description: "**赶集房源抓取**",
	Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
    Namespace: func(self *Spider) string {
		return "housesource"//表名
	},
    SubNamespace: func(self *Spider, dataCell map[string]interface{}) string {
        return "ganji"
		//return dataCell["Data"].(map[string]interface{})["城市"].(string)//根据数据内容来划分 用来才拆分多个表 不能返回 "" 可以返回 nil(默认)
	},
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
            for setting := range houseSourceSetGanJi {
                //ctx.SetTimer(setting, time.Minute*1, nil)
				ctx.Aid(map[string]interface{}{"setting":setting}, "请求列表")
			}       
		},

		Trunk: map[string]*Rule{

			"请求列表": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
                    key := aid["setting"].(string)
					value := houseSourceSetGanJi[key]  
                    for _, area := range value.Areas {
                             
                        for page := 1; page <=value.MaxPage; page++ {  
                            //整租
                            ctx.AddQueue(&request.Request{
                                Url:fmt.Sprintf("http://%s.ganji.com/fang1/%s/m1o%d/",value.CityCode,area,page),
                                Rule: "获取列表",
                                ConnTimeout: -1,
                                Reloadable: true,
                                Temp:map[string]interface{}{"site":fmt.Sprintf("http://%s.ganji.com",value.CityCode)},
                            })
                             //合租
                            ctx.AddQueue(&request.Request{
                                Url:fmt.Sprintf("http://%s.ganji.com/fang3/%s/a3o%d/",value.CityCode,area,page),
                                Rule: "获取列表",
                                ConnTimeout: -1,
                                Reloadable: true,
                                Temp: map[string]interface{}{"site":fmt.Sprintf("http://%s.ganji.com",value.CityCode)},
                            })
                        }   
                    }
				return nil
                },
			},

			"获取列表": {
				ParseFunc: func(ctx *Context) {
					ctx.GetDom().
					Find(".listBox ul.list-style1 li.list-img a.img-box").
					Each(func(i int, s *goquery.Selection) {
						url, _ := s.Attr("href")
                        var site string
					    ctx.GetTemp("site", &site)
                        httpUrlReg:=regexp.MustCompile("https?://(.*?)+")
                        
                        if !httpUrlReg.MatchString(url) {
                            url=site+url
                        }
                        ctx.AddQueue(&request.Request{
                            Url:url,
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
                    
                    query.Find(".basic-info-ul li").Each(func(i int,  s *goquery.Selection) {
						dt := s.Find(".fc-gray9").Text()  //这里需要去掉中间的空格i标签

                        houseTypeReg:=regexp.MustCompile("(\\d+)室(\\d+)厅(\\d+)卫")
                        //houseSizeReg:=regexp.MustCompile("(\\d+)(\\s+)㎡")

						switch dt {
						case "租金：":  
							租金 = strings.TrimSpace(s.Find("b").Text())
                            
                        case "户型：":  
							houseTypeInfo := s.Text()
                            户型= houseTypeReg.FindString(houseTypeInfo)
                            出租类型 = strings.Split(houseTypeInfo, "-")[1]
                            房间大小= strings.TrimSpace(strings.Replace(strings.Split(houseTypeInfo, "-")[2],"㎡","",-1))
                            房间大小 = strings.Split(房间大小,".")[0]
                            
                        case "概况：":  
							houseTypeInfo := s.Text()
                            房屋类型= strings.Split(houseTypeInfo, "-")[1]
                            装修 = strings.Split(houseTypeInfo, "-")[2]
                            
                        case "楼层：":  
							楼层 = strings.TrimSpace(strings.Split(s.Text(), "：")[1])+"层"
                                 
                        case "小区：":  
							 小区 = strings.TrimSpace(s.Find("div.spc-cont a").Eq(0).Text())
                             
                        case "位置：":  
							 城市 = s.Find("a").Eq(0).Text()
                             区域 = s.Find("a").Eq(1).Text()
                             商圈 = s.Find("a").Eq(2).Text()
                             
                        case "配置：":  
							 配置 = strings.TrimSpace(strings.TrimSpace(s.Find("p").Text()))
                               
                        case "":  
							 地址 = s.Find("span.addr-area").Text()
                            
						}
					})
                    经纪人 = query.Find("div.basic-info-contact div.contact-person.tel-number.clearfix span i").Eq(0).Text()
                    联系电话 = strings.TrimSpace(query.Find("div.basic-info-contact div.contact-telphone.clearfix span").Eq(1).Find("em").Text())
                    更新时间 = query.Find("ul.title-info-l.clearfix li").Eq(0).Find(".f10.pr-5").Text()
                    

                    isLianJia:=strings.Index(query.Find("p.company-name").Text(),"链家")>=0
                    if isLianJia {
                        链家发布="是"
                    }else{
                        链家发布="否"
                    }
                    
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
