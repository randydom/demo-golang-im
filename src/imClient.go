package main

import (
    "log"
    "sync"
    "os"
    "sync/atomic"
    "tcp"
    "tcpImpl"
    "flag"
    "runtime"
    "bufio"
    "util"
    "config"
    "constant"
)

var (
    configPath      = flag.String("path", "E:/data1/www/go/config/127.0.0.1_8899.json", "config path")
    configs         = make(map[string]string)
    serverConnected int32
    stopFlag        int32
)

var (
    userId      int
    lastToken   string
)

func init() {
    flag.Parse()
    configs = config.ParseConfig(*configPath)

    log.Println("[IM] Client Start ...")
    log.Println("[IM] Remote: " + configs["host"] + " : " + configs["port"])
}

func main() {
    // 开启多核模式
    runtime.GOMAXPROCS(runtime.NumCPU())

    remoteAddress := configs["host"] + ":" + configs["port"]

    // create client
    client, clientConn, err := echoClient(remoteAddress)
    if nil != err {
        log.Println(err)
        return
    }

    stopCh := make(chan struct{})

    // process event
    var wg sync.WaitGroup

    wg.Add(1)
    go routineEchoClient(client, &wg, stopCh)

    // input event
    routineInput(clientConn)

    close(stopCh)
    wg.Wait()

    clientConn.Close()

    log.Println("[IM] Client done ...")
}


// echo client routine
func echoClient(remoteAddress string) (*tcp.TCPNetwork, *tcp.Connection, error) {
    var err error
    client := tcp.NewTCPNetwork(1024, tcpImpl.NewStreamProtocol4(), constant.IM_REGISTER_EXT_INFO_LENGTH_LIMIT)
    tcpConn, err := client.Connect(remoteAddress)
    if nil != err {
        return nil, nil, err
    }

    return client, tcpConn, nil
}

func routineEchoClient(client *tcp.TCPNetwork, wg *sync.WaitGroup, stopCh chan struct{}) {
    defer func() {
        log.Println("[IM] Client echo done")
        wg.Done()
    }()

    EVENT_LOOP:
    for {
        select {
        case evt, ok := <-client.GetEventQueue():
            {
                if !ok {
                    return
                }

                switch evt.EventType {
                case tcp.KConnEvent_Connected:
                    {
                        log.Println("Press any thing")
                        atomic.StoreInt32(&serverConnected, 1)
                    }
                case tcp.KConnEvent_Disconnected:
                    {
                        log.Println("Disconnected from server")
                        atomic.StoreInt32(&serverConnected, 0)
                        break EVENT_LOOP
                    }
                case tcp.KConnEvent_Data:
                    {
                        text := string(evt.Data)
                        log.Println(evt.Conn.GetLocalAddress(), ":", text)
                    }
                }
            }
        case <-stopCh:
            {
                return
            }
        }
    }
}

func routineInput(clientConn *tcp.Connection) {
    defer func() {
        log.Println("[IM] Client Input done")
    }()

    reader := bufio.NewReader(os.Stdin)
    for {
        line, _, _ := reader.ReadLine()
        str := string(line)

        if str == "\n" || str == "" {
            continue
        }

        if str == "!quit" {
            return
        }

        if atomic.LoadInt32(&serverConnected) != 1 {
            log.Println("[IM] Not connected")
            continue
        }

        clientConn.Send([]byte(str), 0)
    }
}

// 计算登录token
func getLoginToken(userId int, time int64) string {
    return util.GetToken(configs["login_key"], userId, time)
}