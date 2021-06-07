package jasync

// 异步执行类，提供异步执行的功能，可快速方便的开启异步执行
// 通过NewAsync() 来创建一个新的异步操作对象
// 通过调用 Add 函数来向异步任务列表中添加新的任务
// 通过调用 Run 函数来获取一个接收返回的channel，当返回结果时将会返回一个map[string][]interface{}
// 的结果集，包括每个异步函数所返回的所有的结果
// 通过调用 GetStatus() 来获取任务执行的状态
// 通过调用 GetTotal() 来获取所有任务的数量
import (
	"github.com/chroblert/jgoutils/jconfig"
	"github.com/chroblert/jgoutils/jlog"
	"reflect"
	"sync"
	"time"
)

// 异步执行所需要的数据
type asyncTask struct {
	ReqHandler   reflect.Value
	PrintHandler reflect.Value
	Params       []reflect.Value
	// 结构体中嵌套的结构体无法被正常修改
	// 因而需要采用结构体指针的形式 refer: https://haobook.readthedocs.io/zh_CN/latest/periodical/201611/zhangan.html
	TaskStatus *taskStatus
}

// Async 异步执行对象
type Async struct {
	total     int  // 总共有多少个任务
	count     int  // 需要执行的任务数量
	taskCount int  // 正在执行的任务数量
	tasks     map[string]asyncTask
	mu        *sync.RWMutex

	//210519：获取异步执行结果
	tasksResult map[string][]interface{}
}

// New 创建一个新的异步执行对象
func New() Async {
	return Async{
		tasks: make(map[string]asyncTask),
		mu: new(sync.RWMutex),
		tasksResult: make(map[string][]interface{}),
	}
}

//GetTotal 获取总共的任务数
func (a *Async) GetTotal() int {
	return a.total
}

//GetCount 获取需要执行的任务数量
func (a *Async) GetCount() int {
	return a.count
}

func (a *Async) addTaskCount() {
	a.mu.Lock()
	a.taskCount++
	a.mu.Unlock()
}

func (a *Async) subTaskCount() {
	a.mu.Lock()
	a.taskCount--
	a.mu.Unlock()
}

// 若传进来的值小于1，则使用配置文件中的默认值
func (a *Async) wait(taskMaxLimit int) {
	if taskMaxLimit < 1 {
		taskMaxLimit = jconfig.Conf.AsyncConfig.TaskMaxLimit
	}
	var tmpPreVal int
	tmpPreVal = -1
	for {
		// 如果当前开启的任务数小于配置中设定的最大任务数，则继续开启任务
		a.mu.RLock()
		tmpTaskCount := a.taskCount
		tmpTotal := a.total
		doneTaskCount := a.total-a.count
		a.mu.RUnlock()
		if tmpTaskCount == tmpPreVal{
			continue
		}
		tmpPreVal = tmpTaskCount
		//a.mu.RLock()
		//doneTaskCount := a.total-a.count
		//a.mu.RUnlock()
		if tmpTaskCount < taskMaxLimit {
			break
		}else{
			//jlog.Debugf("达到同时最大任务量限制：taskMaxLimit: %v,taskDoneCount: %v\r\x1b[K",  taskMaxLimit,doneTaskCount)
			jlog.Debugf("达到同时最大任务量限制：taskMaxLimit: %v,taskDoneCount: %v/%v\r",  taskMaxLimit,doneTaskCount,tmpTotal)
		}
	}
}

type taskStatus struct {
	taskStatus  int   // 任务状态 0: queue,1:scheduled,2: doing,3: done
	taskBegTime int64 // 任务开始时间
	taskEndTime int64 // 任务结束时间
}

// 将毫秒级时间戳转换为时间字符串2006-01-02 15:04:05.000
func (a *Async) timeStampToStr(nanotimestamp int64) string {
	if nanotimestamp == 0 {
		return "0"
	}
	timeStr := time.Unix(0, nanotimestamp).Format("2006-01-02 15:04:05.0000")
	return timeStr
}

// GetStatus 获取执行状态
// verbose: 详细模式，显示任务的开始结束时间
// status: 显示指定状态的任务
// taskName: 显示某任务的状态
func (a *Async) GetStatus(taskName string, verbose bool) {
	for k, v := range a.tasks {
		//jlog.Info(k, "Status:", a.getDspByCode(v.TaskStatus.taskStatus), ",Begin:", a.timeStampToStr(v.TaskStatus.taskBegTime), ",End:", a.timeStampToStr(v.TaskStatus.taskEndTime))
		jlog.Debugf("%-5s Status:%-10s, Begin:%.24s ,End:%.24s \n", k, a.getDspByCode(v.TaskStatus.taskStatus), a.timeStampToStr(v.TaskStatus.taskBegTime), a.timeStampToStr(v.TaskStatus.taskEndTime))
	}

}

// GetTasksResult 获取所有任务的执行结果
func (a *Async) GetTasksResult() map[string][]interface{} {
	return a.tasksResult
}

// GetTaskResult 获取任务的某个执行结果
func (a *Async) GetTaskResult(taskName string) []interface{} {
	return a.tasksResult[taskName]
}

// 等待直到全部任务执行完成
func (a *Async) Wait() {
	var tmpPreVal int
	tmpPreVal = -1
	for {
		a.mu.RLock()
		tmpCount := a.count
		tmpTotal := a.total
		//doneTaskCount := a.total-a.count
		a.mu.RUnlock()
		if tmpCount == tmpPreVal{
			continue
		}
		//jlog.Debug("tmpCount:",tmpCount,"tmpPreVal:",tmpPreVal)
		tmpPreVal = tmpCount
		if tmpCount < 1 {
			break
		}
		//time.Sleep(time.Nanosecond * 500)
		jlog.Infof("%d/%d\r", tmpTotal-tmpCount, tmpTotal)
	}
	a.mu.RLock()
	doneTaskCount := a.total-a.count
	a.mu.RUnlock()
	jlog.Infof("%d/%d,所有task执行完毕\n", doneTaskCount, a.total)
}

// 根据code获取对应的状态描述
func (a *Async) getDspByCode(code int) string {
	switch code {
	case 0:
		return "queue"
	case 1:
		return "scheduled"
	case 2:
		return "doing"
	case 3:
		return "done"
	}
	return "error"
}

// Add 添加异步执行任务
// name 任务名，结果返回时也将放在任务名中
// handler 任务执行函数，将需要被执行的函数导入到程序中
// params 任务执行函数所需要的参数
func (a *Async) Add(name string, handler interface{}, printHandler interface{}, params ...interface{}) bool {
	// 用来确保key的唯一性
	a.mu.RLock()
	_, ok := a.tasks[name]
	a.mu.RUnlock()
	if ok {
		return false
	}
	handlerValue := reflect.ValueOf(handler)
	// 判断传入的是否为Func类型
	if handlerValue.Kind() == reflect.Func {
		// 传入了多少个参数
		paramNum := len(params)
		if printHandler != nil && reflect.ValueOf(printHandler).Kind() == reflect.Func {
			a.mu.Lock()
			a.tasks[name] = asyncTask{
				ReqHandler:   handlerValue,
				PrintHandler: reflect.ValueOf(printHandler),
				Params:       make([]reflect.Value, paramNum),
				TaskStatus: &taskStatus{
					taskStatus:  0,
					taskBegTime: 0,
					taskEndTime: 0,
				},
			}
			a.mu.Unlock()
		} else {
			a.mu.Lock()
			a.tasks[name] = asyncTask{
				ReqHandler:   handlerValue,
				PrintHandler: reflect.Value{},
				Params:       make([]reflect.Value, paramNum),
				TaskStatus: &taskStatus{
					taskStatus:  0,
					taskBegTime: 0,
					taskEndTime: 0,
				},
			}
			a.mu.Unlock()
		}
		// 将传入的参数转换成reflect.Value类型
		if paramNum > 0 {
			for k, v := range params {
				a.mu.Lock()
				a.tasks[name].Params[k] = reflect.ValueOf(v)
				a.mu.Unlock()
			}
		}
		a.mu.Lock()
		a.count++
		a.total++
		a.mu.Unlock()
		return true
	}
	return false
}

// Run 任务执行函数，成功时将返回一个用于接受结果的channel
// 在所有异步任务都运行完成时，结果channel将会返回一个map[string][]interface{}的结果。
func (a *Async) Run(taskCountMaxLimit int) (chan map[string][]interface{}, bool) {
	if a.count < 1 {
		return nil, false
	}
	// [任务名]结果切片
	result := make(chan map[string][]interface{})
	chans := make(chan map[string]interface{}, a.count)
	// 开启一个协程，用来接收调用函数的结果
	go func(result chan map[string][]interface{}, chans chan map[string]interface{}) {
		rs := make(map[string][]interface{})
		defer func(rs map[string][]interface{}) {
			result <- rs
		}(rs)
		for {
			if a.count < 1 {
				break
			}
			select {
			// 通过chans接受结果
			case res := <-chans:
				// 如果传入了printHandler,并且reqHandler函数返回参数的个数与printHandler函数形参个数相同
				if a.tasks[res["taskName"].(string)].PrintHandler.IsValid() && a.tasks[res["taskName"].(string)].ReqHandler.Type().NumOut() == a.tasks[res["taskName"].(string)].PrintHandler.Type().NumIn() {
					paramsArg := make([]reflect.Value, len(res["result"].([]interface{})))
					for k, v := range res["result"].([]interface{}) {
						if reflect.ValueOf(v).IsValid() {
							paramsArg[k] = reflect.ValueOf(v)
						} else {
							paramsArg[k] = reflect.Zero(a.tasks[res["taskName"].(string)].PrintHandler.Type().In(k))
						}
					}
					// 调用printHandler
					a.tasks[res["taskName"].(string)].PrintHandler.Call(paramsArg)
				}
				rs[res["taskName"].(string)] = res["result"].([]interface{})
				// 210519: Add 添加每个任务执行的结果
				a.tasksResult[res["taskName"].(string)] = res["result"].([]interface{})
				a.count--

			}
		}
	}(result, chans)
	// 使用协程执行每一个task
	// asyncTaskKey: name,asyncTaskVal:asyncTask
	for asyncTaskKey, asyncTaskVal := range a.tasks {
		// 等待，直到当前开启的任务数小于配置中设定的最大任务数，则继续开启任务
		a.wait(taskCountMaxLimit)
		a.addTaskCount()
		go func(taskName string, routinChans chan map[string]interface{}, task asyncTask) {

			// 全局当前协程数量加一
			jconfig.Conf.GlobalConfig.AddGlobalGoroutinCount()
			// 若全局当前协程数量不小于全局配置中的最大协程数量，则等待
			jconfig.Conf.GlobalConfig.Wait()
			taskResult := make([]interface{}, 0)
			defer func(taskName2 string, resultChans chan map[string]interface{}) {
				// 设置任务的状态为结束
				task.TaskStatus.taskStatus = 3
				// 设置任务结束时间戳，毫秒
				task.TaskStatus.taskEndTime = time.Now().UnixNano()
				// 任务数量减一
				a.subTaskCount()
				// 总实时协程数量减一
				jconfig.Conf.GlobalConfig.SubGlobalGoroutinCount()
				// 通过chans传输结果
				resultChans <- map[string]interface{}{"taskName": taskName2, "result": taskResult}
			}(taskName, routinChans)
			// 设置任务状态为1: scheduled
			task.TaskStatus.taskStatus = 1
			// 设置任务开始时间戳，毫秒
			task.TaskStatus.taskBegTime = time.Now().UnixNano()
			// 调用传入的函数
			values := task.ReqHandler.Call(task.Params)
			// 传入的函数执行的结果保存在values中
			if valuesNum := len(values); valuesNum > 0 {
				resultItems := make([]interface{}, valuesNum)
				// asyncTaskKey:int,asyncTaskVal:value
				for k, v := range values {
					resultItems[k] = v.Interface()
				}
				taskResult = resultItems
				return
			}
		}(asyncTaskKey, chans, asyncTaskVal)
	}

	return result, true
}

// Clean 清空任务队列.
func (a *Async) Clean() {
	a.count = 0
	a.total = 0
	a.tasks = make(map[string]asyncTask)
}
