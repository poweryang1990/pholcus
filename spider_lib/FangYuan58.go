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
	// "math"
	//"time"
)

func init() {
	FangYuan58.Register()
}

type HouseSourceSetting struct{
    MaxPage int
    CityCode string
    Areas []string
}
var housesourceSettings=map[string]HouseSourceSetting{
    "成都":HouseSourceSetting{
        MaxPage:70,
        CityCode:"cd",
        Areas:[]string{"jinjiang","chenghua","jinniu","qingyangqu","cdgaoxin","gaoxinxiqu"},
    },
    "武汉":HouseSourceSetting{
        MaxPage:70,
        CityCode:"wh",
        Areas:[]string{"jiangan","jianghan","qiaokou","hanyang","wuchang","whqingshanqu","hongshan","dongxihu","hannan","caidian","jiangxia","huangpo","xinzhouqu","whtkfq"},
    },
    "北京":HouseSourceSetting{
        MaxPage:70,
        CityCode:"bj",
        Areas:[]string{"chaoyang","haidian","dongcheng","xicheng","chongwen","xuanwu","fengtai","tongzhouqu","shijingshan","fangshan","changping","daxing","shunyi","miyun","huairou","yanqing","pinggu","mentougou","bjyanjiao"},
    },
    "杭州":HouseSourceSetting{
        MaxPage:70,
        CityCode:"hz",
        Areas:[]string{"xihuqu","gongshu","jianggan","xiacheng","hzshangcheng","binjiang"},
    },
   
}
var FangYuan58 = &Spider{
	Name:        "58房源抓取",
	Description: "**58房源抓取**",
	Pausetime: 300,
	// Keyin:   KEYIN,
	// Limit:        LIMIT,
	EnableCookie: false,
    Namespace: nil,
    SubNamespace: func(self *Spider, dataCell map[string]interface{}) string {
		return "housesource_58"
	},
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
            for setting := range housesourceSettings {
                //ctx.SetTimer(setting, time.Minute*1, nil)
				ctx.Aid(map[string]interface{}{"setting":setting}, "请求列表")
			}       
		},

		Trunk: map[string]*Rule{

			"请求列表": {
				AidFunc: func(ctx *Context, aid map[string]interface{}) interface{} {
                    
                    key := aid["setting"].(string)
					value := housesourceSettings[key]  
                    for _, area := range value.Areas {
                             
                        for page := 1; page <=value.MaxPage; page++ {  
                            //整租
                            ctx.AddQueue(&request.Request{
                                Url:fmt.Sprintf("http://%s.58.com/%s/zufang/pn%d/",value.CityCode,area,page),
                                Rule: "获取列表",
                                ConnTimeout: -1,
                            })
                            //合租
                             ctx.AddQueue(&request.Request{
                                Url:fmt.Sprintf("http://%s.58.com/%s/hezu/pn%d/",value.CityCode,area,page),
                                Rule: "获取列表",
                                ConnTimeout: -1,
                            })
                        }   
                    }
				return nil
                },
			},

			"获取列表": {
				ParseFunc: func(ctx *Context) {
                    
					ctx.GetDom().
						Find("table.tbimg .img_list a").
						Each(func(i int, s *goquery.Selection) {
							url, _ := s.Attr("href")
                            //logs.Log.Informational("请求房源地址：%s",url)
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
					"城市", "区域","商圈","小区", "地址","出租类型","房屋类型","房间大小","户型","租金","配置","装修","更新时间","楼层","经纪人","联系电话","链家发布","单价",
				},
				ParseFunc: func(ctx *Context) {
					query := ctx.GetDom()
					var 城市, 区域,商圈,小区, 地址,出租类型,房屋类型,房间大小,户型,租金,配置,装修,更新时间,楼层,经纪人,联系电话,链家发布 string
                    var 单价 int
                    城市 = query.Find("#topbar > div > div.bar_left.f16 > h2").Text()
                    
                    租金 =strings.TrimSpace(query.Find("em.house-price").Text())   
                    
                    区域=query.Find("div.xiaoqu a").Eq(0).Text();
                    
                    quyuHref, _:=query.Find("div.xiaoqu a").First().Attr("href")
                    
                    商圈=query.Find("div.xiaoqu a").Eq(1).Text()
                    xiaoquLength:=query.Find("div.xiaoqu a").Length()
                    if xiaoquLength>2 {
                        小区=query.Find("div.xiaoqu a").Eq(2).Text()
                    }else{
                        小区=query.Find("div.xiaoqu").Text()
                    }
                   
                    
                  
                    houseInfoText:=query.Find(".house-type").Text()
                    houseTypeReg:=regexp.MustCompile("([\u4E00-\u9FA5]{0,2}住宅)|公寓|别墅|商住楼|商住两用|其他|(平房/四合院)")
                    houseSizeReg:=regexp.MustCompile("(\\d+)(\\s+)m²")
                    houseHuXinReg:=regexp.MustCompile("(\\d+)室(\\s+)(\\d+)厅(\\s+)(\\d+)卫")
                    houseZhuangXiuReg:=regexp.MustCompile("[\u4E00-\u9FA5]{1,2}装修")
                    houseFloorReg:=regexp.MustCompile("(\\d+)/(\\d+)层")
                    houseRoomReg:=regexp.MustCompile("主卧|次卧|隔断")
                    
                    房屋类型=strings.Replace(houseTypeReg.FindString(houseInfoText)," ","",-1)
                    房屋类型=strings.Replace(房屋类型,"\n","",-1)
                    房间大小=strings.Replace(houseSizeReg.FindString(houseInfoText),"m²","",-1)
                    房间大小=strings.TrimSpace(房间大小)
                    户型=houseHuXinReg.FindString(houseInfoText)
                    装修=houseZhuangXiuReg.FindString(houseInfoText)
                    楼层=houseFloorReg.FindString(houseInfoText)
                      isZhengzu:=strings.Index(quyuHref,"zufang")>=0
                    
                    if isZhengzu {
                         出租类型="整租"
                    }else{
                         出租类型=houseRoomReg.FindString(houseInfoText)
                    }
                    price,_:= strconv.Atoi(租金)   
                    size,_:=strconv.Atoi(房间大小)
                    if size!=0 {
                      单价= price/size  
                    }
                    
                    houseConfigReg:=regexp.MustCompile("暖气|宽带|空调|冰箱|电视|洗衣机|热水器|床|衣柜")
                    addressStr:=query.Find("ul.house-primary-content>li").Eq(3).Find("div.c70").Text()
                    isAddress:=len(houseConfigReg.FindAllString(addressStr,-1))<=0
                    if isAddress {
                         地址 = addressStr
                    }
                   
                    配置=strings.Join(houseConfigReg.FindAllString(query.Find("ul.house-primary-content li.broker-config div>span").Text(),-1),",")  
                    
                    
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
						0:城市,1:区域,2:商圈,3:小区,4: 地址,5:出租类型,6:房屋类型,7:房间大小,8:户型,9:租金,10:配置,11:装修,12:更新时间,13:楼层,14:经纪人,15:联系电话,16:链家发布,17:单价,
					})

					
				},
			},

			
		},
	},
}
