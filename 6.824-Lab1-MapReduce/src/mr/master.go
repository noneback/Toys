package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Master struct {
	// Your definitions here.
	nReduce      int
	taskQueue    chan Task
	tasksContext []TaskContext
	lock         sync.Mutex
	files        []string
	phase        PhaseKind
	done         bool
	workerID     int
}

// inititalize the tasksContext
func (m *Master) initTask(phase PhaseKind) {

	m.phase = phase
	switch phase {
	case MAP:
		m.tasksContext = make([]TaskContext, len(m.files))
		log.Println("map task init...")
	case REDUCE:
		m.tasksContext = make([]TaskContext, m.nReduce)
		log.Println("Reduce task init")
	default:
		panic("initTask:No such phase")
	}

	for i := 0; i < len(m.tasksContext); i++ {
		m.tasksContext[i].state = IDEL
	}
}

// do schedule
func (m *Master) scheduling() {
	for !m.done {
		go m.schedule()
		time.Sleep(SCHEDULE_INTERVAL)
	}
}

func (m *Master) schedule() {
	//	fmt.Println("scheduling...")
	m.lock.Lock()
	defer m.lock.Unlock()
	allComplete := true
	// fmt.Println(len(m.taskQueue))
	for index, context := range m.tasksContext {
		switch context.state {
		case IDEL:
			allComplete = false
			m.addTask2Queue(index)
			//fmt.Println("pick ", m.files[index], len(m.taskQueue))
		case READY:
			allComplete = false
		case RUNNING:
			allComplete = false
			// if err := m.timeout(&context); err != nil {
			// 	log.Println(err)
			// 	m.addTask2Queue(index)
			// }
			if ok := m.timeout(&context); !ok {
				//log.Printf("Task %d timeout,restart\n", context.t.ID)
				m.addTask2Queue(index)
			}
		case COMPLETE:
		case FAILED:
			log.Println("Task ", context.t.ID, " failed,restart")
			m.addTask2Queue(index)
		default:
			log.Fatalln("error schedule")
		}
	}

	if allComplete {
		if m.phase == MAP {
			m.phase = REDUCE
			//init all reduce task
			m.initTask(REDUCE)
		} else {
			m.done = true
		}
	}

}

//RPC handlers for the worker to call.
func (m *Master) RegTask(args *RegTaskArgs, reply *RegTaskReply) error {

	if m.done {
		reply.HasT = false
		return nil
	}

	t := <-m.taskQueue
	context := &m.tasksContext[t.ID]
	context.state = RUNNING
	context.startTime = time.Now()
	context.workerID = args.WorkerID
	reply.T = t
	reply.HasT = true
	return nil

}

func (m *Master) RegWorker(args *RegWorkerArgs, reply *RegWorkerReply) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	log.Println("worker", m.workerID, "has been registed")
	reply.ID = m.workerID
	reply.NMap = len(m.files)
	reply.NReduce = m.nReduce
	m.workerID++
	return nil
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

//RegTask for worker to assign a task

// ReportTask worker to report their work
func (m *Master) ReportTask(args *ReportTaskArgs, reply *ReportTaskReply) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	switch args.State {
	case RUNNING:
		m.tasksContext[args.TaskID].state = RUNNING
		log.Println("task", args.TaskID, " running")
	case FAILED:
		m.tasksContext[args.TaskID].state = FAILED
		log.Println("task", args.TaskID, " failed")
	case COMPLETE:
		m.tasksContext[args.TaskID].state = COMPLETE
		log.Println("task", args.TaskID, "finished by", args.WorkerID)
	default:
		panic("error : wrong state")
	}

	return nil
}

// checkout whether timeout
func (m *Master) timeout(context *TaskContext) bool {
	curTime := time.Now()
	interval := curTime.Sub(context.startTime)
	//fmt.Println(interval, "  ", MAX_PROCESSING_TIME)
	if interval > MAX_PROCESSING_TIME {
		return false
	}
	return true
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	return m.done
}

//generate a task and put it in queue
func (m *Master) addTask2Queue(taskID int) {
	m.tasksContext[taskID].state = READY
	t := Task{
		Filename: "",
		ID:       taskID,
		Phase:    m.phase,
	}

	if m.phase == MAP {
		t.Filename = m.files[taskID]
	}
	m.taskQueue <- t
}

// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	var n int

	if len(files) > nReduce {
		n = len(files)
	} else {
		n = nReduce
	}

	contexts := make([]TaskContext, len(files))

	m := Master{
		taskQueue:    make(chan Task, n),
		lock:         sync.Mutex{},
		tasksContext: contexts,
		done:         false,
		files:        files,
		nReduce:      nReduce,
		workerID:     0,
	}

	// Your code here.
	m.server()
	m.initTask(MAP)
	go m.scheduling()

	return &m
}
