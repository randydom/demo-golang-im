package main

import (
    "log"
    "sync"
    "tcp"
    "tcpImpl"
    "flag"
    "runtime"
    "config"
    "constant"
    "util"
)

var (
    configPath      = flag.String("path", "E:/data1/www/go/config/127.0.0.1_8899.json", "config path")
    configs         = make(map[string]string)
)

func init() {
    flag.Parse()
    configs = config.ParseConfig(*configPath)

    log.Println("[IM] Server Start ...")
    log.Println("[IM] Remote: " + configs["host"] + " : " + configs["port"])
}

func main() {
    // 开启多核模式
    runtime.GOMAXPROCS(runtime.NumCPU())

    remoteAddress := configs["host"] + ":" + configs["port"]

    // create server
    server, err := echoServer(remoteAddress)
    if nil != err {
        log.Println(err)
        return
    }

    stopCh := make(chan struct{})

    // process event
    var wg sync.WaitGroup
    wg.Add(1)
    go routineEchoServer(server, &wg, stopCh)

    // wait
    util.WaitTerminatingStop()

    close(stopCh)
    wg.Wait()

    server.Shutdown()

    log.Println("[IM] Server done ...")
}

// echo server routine
func echoServer(remoteAddress string) (*tcp.TCPNetwork, error) {
    var err error
    server := tcp.NewTCPNetwork(1024, tcpImpl.NewStreamProtocol4(), constant.IM_REGISTER_EXT_INFO_LENGTH_LIMIT)
    err = server.Listen(remoteAddress)
    if nil != err {
        return nil, err
    }

    return server, nil
}

func routineEchoServer(server *tcp.TCPNetwork, wg *sync.WaitGroup, stopCh chan struct{}) {
    defer func() {
        log.Println("[IM] Server echo done")
        wg.Done()
    }()
    for {
        select {
        case evt, ok := <-server.GetEventQueue():
            {
                if !ok {
                    return
                }

                switch evt.EventType {
                case tcp.KConnEvent_Connected:
                    {
                        log.Println("[IM] Client ", evt.Conn.GetRemoteAddress(), " connected")
                    }
                case tcp.KConnEvent_Disconnected:
                    {
                        log.Println("[IM] Client ", evt.Conn.GetRemoteAddress(), " disconnected")
                    }
                case tcp.KConnEvent_Close:
                    {
                        log.Println("[IM] Client ", evt.Conn.GetRemoteAddress(), " closed")
                    }
                case tcp.KConnEvent_Data:
                    {
                        text := string(evt.Data)
                        log.Println(evt.Conn.GetRemoteAddress(), ":", text)
                        evt.Conn.Send(evt.Data, 0)
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