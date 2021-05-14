package jasync

// 异步执行类，提供异步执行的功能，可快速方便的开启异步执行
// 通过NewAsync() 来创建一个新的异步操作对象
// 通过调用 Add 函数来向异步任务列表中添加新的任务
// 通过调用 Run 函数来获取一个接收返回的channel，当返回结果时将会返回一个map[string][]interface{}
// 的结果集，包括每个异步函数所返回的所有的结果
// 通过调用 GetStatus() 来获取任务执行的状态
// 通过调用 GetTotal() 来获取所有任务的数量
import (
	"github.com/chroblert/JC-GoUtils/jlog"
	"reflect"
	"time"
)
// 异步执行所需要的数据
type asyncTask struct {
	ReqHandler reflect.Value
	PrintHandler reflect.Value
	Params     []reflect.Value
	// 结构体中嵌套的结构体无法被正常修改
	// 因而需要采用结构体指针的形式 refer: https://haobook.readthedocs.io/zh_CN/latest/periodical/201611/zhangan.html
	TaskStatus *taskStatus
}

// Async 异步执行对象
type Async struct {
	total int
	count int
	tasks map[string]asyncTask
}

// New 创建一个新的异步执行对象
func New() Async {
	return Async{tasks: make(map[string]asyncTask)}
}

//GetTotal 获取总共的任务数
func(a *Async) GetTotal() int {
	return a.total
}

type taskStatus struct{
	taskStatus int  // 任务状态 0: queue,1:scheduled,2: doing,3: done
	taskBegTime int64 // 任务开始时间
	taskEndTime int64 // 任务结束时间
}

// 将毫秒级时间戳转换为时间字符串2006-01-02 15:04:05.000
func (a *Async)timeStampToStr(nanotimestamp int64) string {
	if nanotimestamp == 0{
		return "0"
	}
	timeStr := time.Unix(0,nanotimestamp).Format("2006-01-02 15:04:05.0000")
	return timeStr
}


// GetStatus 获取执行状态
// verbose: 详细模式，显示任务的开始结束时间
// status: 显示指定状态的任务
// taskName: 显示某任务的状态
func (a *Async) GetStatus(taskName string,verbose bool){
	for k,v := range a.tasks{
		jlog.Info(k,"Status:",a.getDspByCode(v.TaskStatus.taskStatus),",Begin:",a.timeStampToStr(v.TaskStatus.taskBegTime),",End:",a.timeStampToStr(v.TaskStatus.taskEndTime))
	}

}

// 等待直到全部任务执行完成
func (a *Async) Wait(){
	for a.count > 0{
	}
}

// 根据code获取对应的状态描述
func (a *Async) getDspByCode(code int) string{
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
func (a *Async) Add(name string, handler interface{},printHandler interface{}, params ...interface{}) bool {
	// 用来确保key的唯一性
	if _, e := a.tasks[name]; e {
		return false
	}
	handlerValue := reflect.ValueOf(handler)
	// 判断传入的是否为Func类型
	if handlerValue.Kind() == reflect.Func {
		// 传入了多少个参数
		paramNum := len(params)
		if printHandler != nil && reflect.ValueOf(printHandler).Kind() == reflect.Func {
			a.tasks[name] = asyncTask{
				ReqHandler: handlerValue,
				PrintHandler: reflect.ValueOf(printHandler),
				Params:     make([]reflect.Value, paramNum),
				TaskStatus: &taskStatus{
					taskStatus:  0,
					taskBegTime: 0,
					taskEndTime: 0,
				},
			}
		}else{
			a.tasks[name] = asyncTask{
				ReqHandler: handlerValue,
				PrintHandler: reflect.Value{},
				Params:     make([]reflect.Value, paramNum),
				TaskStatus: &taskStatus{
					taskStatus:  0,
					taskBegTime: 0,
					taskEndTime: 0,
				},
			}
		}
		// 将传入的参数转换成reflect.Value类型
		if paramNum > 0 {
			for k, v := range params {
				a.tasks[name].Params[k] = reflect.ValueOf(v)
			}
		}
		a.count++
		a.total++
		return true
	}

	return false
}

// Run 任务执行函数，成功时将返回一个用于接受结果的channel
// 在所有异步任务都运行完成时，结果channel将会返回一个map[string][]interface{}的结果。
func (a *Async) Run() (chan map[string][]interface{}, bool) {
	if a.count < 1 {
		return nil, false
	}
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
				if a.tasks[res["taskName"].(string)].PrintHandler.IsValid() && a.tasks[res["taskName"].(string)].ReqHandler.Type().NumOut() == a.tasks[res["taskName"].(string)].PrintHandler.Type().NumIn(){
					paramsArg := make([]reflect.Value,len(res["result"].([]interface{})))
					for k,v := range res["result"].([]interface{}){
						if reflect.ValueOf(v).IsValid(){
							paramsArg[k] = reflect.ValueOf(v)
						}else{
							paramsArg[k] = reflect.Zero(a.tasks[res["taskName"].(string)].PrintHandler.Type().In(k))
						}
					}
					// 调用printHandler
					a.tasks[res["taskName"].(string)].PrintHandler.Call(paramsArg)
				}
				rs[res["taskName"].(string)] = res["result"].([]interface{})
				a.count--

			}
		}
	}(result, chans)
	// 使用协程执行每一个task
	// asyncTaskKey: name,asyncTaskVal:asyncTask
	for asyncTaskKey, asyncTaskVal := range a.tasks {
		go func(taskName string, routinChans chan map[string]interface{}, task asyncTask) {
			taskResult := make([]interface{}, 0)
			defer func(taskName2 string, resultChans chan map[string]interface{}) {
				// 设置任务的状态为结束
				task.TaskStatus.taskStatus = 3
				// 设置任务结束时间戳，毫秒
				task.TaskStatus.taskEndTime = time.Now().UnixNano()
				// 通过chans传输结果
				resultChans <- map[string]interface{}{"taskName": taskName2, "result": taskResult}
			}(taskName, routinChans)
			// 设置任务状态为1: scheduled
			task.TaskStatus.taskStatus = 1
			// 设置任务开始时间戳，毫秒
			task.TaskStatus.taskBegTime = time.Now().UnixNano()
			// 调用传入的函数
			values := task.ReqHandler.Call(task.Params)
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