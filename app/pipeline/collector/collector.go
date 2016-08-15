// 结果收集与输出
package collector

import (
	"runtime"
	"time"

	"github.com/henrylee2cn/pholcus/app/pipeline/collector/data"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

// 结果收集与输出
type Collector struct {
	*spider.Spider                    //绑定的采集规则
	*DockerQueue                      //分批输出结果的缓存块队列
	DataChan       chan data.DataCell //文本数据收集通道
	FileChan       chan data.FileCell //文件收集通道
	ctrl           chan bool          //长度为零时退出并输出
	outType        string             //输出方式
	timing         time.Time          //上次输出完成的时间点
	outCount       [4]uint            //[文本输出开始，文本输出结束，文件输出开始，文件输出结束]
	sum            [4]uint64          //收集的数据总数[上次输出后文本总数，本次输出后文本总数，上次输出后文件总数，本次输出后文件总数]，非并发安全
	// size     [2]uint64 //数据总输出流量统计[文本，文件]，文本暂时未统计
}

func NewCollector() *Collector {
	self := &Collector{
		DataChan:    make(chan data.DataCell, config.DATA_CHAN_CAP),
		FileChan:    make(chan data.FileCell, 512),
		DockerQueue: NewDockerQueue(),
		ctrl:        make(chan bool, 1),
	}
	return self
}

func (self *Collector) Init(sp *spider.Spider) {
	self.Spider = sp
	self.outType = cache.Task.OutType
	self.DataChan = make(chan data.DataCell, config.DATA_CHAN_CAP)
	self.FileChan = make(chan data.FileCell, 512)
	self.DockerQueue = NewDockerQueue()
	self.ctrl = make(chan bool, 1)
	self.sum = [4]uint64{}
	// self.size = [2]uint64{}
	self.outCount = [4]uint{}
	self.timing = cache.StartTime
}

func (self *Collector) CollectData(dataCell data.DataCell) {
	self.DataChan <- dataCell
}

func (self *Collector) CollectFile(fileCell data.FileCell) {
	self.FileChan <- fileCell
}

// 是否已发出停止命令
func (self *Collector) beStopping() bool {
	return len(self.ctrl) == 0
}

// 停止
func (self *Collector) Stop() {
	<-self.ctrl
}

// 启动数据收集/输出管道
func (self *Collector) Start() {
	// 标记程序已启动
	self.ctrl <- true

	// 启动输出协程
	go func() {

		// 只有当收到退出通知并且通道内无数据时，才退出循环
		for !(self.beStopping() && len(self.DataChan) == 0 && len(self.FileChan) == 0) {
			select {
			case data := <-self.DataChan:
				// 追加数据
				self.Dockers[self.Curr] = append(self.Dockers[self.Curr], data)

				// 未达到设定的分批量时，仅缓存
				if len(self.Dockers[self.Curr]) < cache.Task.DockerCap {
					continue
				}

				// 执行输出
				self.outputData()

				// 更换一个空Docker用于curDocker
				self.DockerQueue.Change()

			case file := <-self.FileChan:
				go self.outputFile(file)

			default:
				runtime.Gosched()
			}
		}

		// 将剩余收集到但未输出的数据输出
		self.outputData()

		// 等待所有输出完成
		for (self.outCount[0] > self.outCount[1]) || (self.outCount[2] > self.outCount[3]) || len(self.FileChan) > 0 {
			runtime.Gosched()
		}

		// 返回报告
		self.Report()
	}()
}

// 获取文本数据总量
func (self *Collector) dataSum() uint64 {
	return self.sum[1]
}

// 更新文本数据总量
func (self *Collector) addDataSum(add uint64) {
	self.sum[0] = self.sum[1]
	self.sum[1] += add
}

// 获取文件数据总量
func (self *Collector) fileSum() uint64 {
	return self.sum[3]
}

// 更新文件数据总量
func (self *Collector) addFileSum(add uint64) {
	self.sum[2] = self.sum[3]
	self.sum[3] += add
}

// // 获取文本输出流量
// func (self *Collector) dataSize() uint64 {
// 	return self.size[0]
// }

// // 更新文本输出流量记录
// func (self *Collector) addDataSize(add uint64) {
// 	self.size[0] += add
// }

// // 获取文件输出流量
// func (self *Collector) fileSize() uint64 {
// 	return self.size[1]
// }

// // 更新文本输出流量记录
// func (self *Collector) addFileSize(add uint64) {
// 	self.size[1] += add
// }

// 返回报告
func (self *Collector) Report() {
	cache.ReportChan <- &cache.Report{
		SpiderName: self.Spider.GetName(),
		Keyin:      self.GetKeyin(),
		DataNum:    self.dataSum(),
		FileNum:    self.fileSum(),
		// DataSize:   self.dataSize(),
		// FileSize: self.fileSize(),
		Time: time.Since(cache.StartTime),
	}
}
