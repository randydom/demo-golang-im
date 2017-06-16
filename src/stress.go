// 聊天服务器压力测试工具
package main

import (
    "flag"
    "fmt"
    "runtime"
    "time"
    "os"
    "github.com/golang/net"
)

// 用来接受host和port参数
var (
    open     = flag.Int("o", 3000, "open count")
    host     = flag.String("host", "172.17.73.107", "im server host")
    port     = flag.String("port", "9090", "im server port")
    step     = flag.Int("s", 300, "send step")
    count    = flag.Int("c", 300, "send message count")
)

// 随机发言的内容
var randMsgArr = [45]string{
    "大家好，我是机器人",
    "你好，请问你是机器人么",
    "你才是机器人，你全家都是机器人",
    "Do one thing at a time, and do well.",
    "Never forget to say &ldquo;thanks&rdquo;.",
    "Keep on going never give up.",
    "Whatever is worth doing is worth doing well.",
    "Believe in yourself.",
    "I can because i think i can.",
    "Action speak louder than words.",
    "Never say die.",
    "Never put off what you can do today until tomorrow.",
    "The best preparation for tomorrow is doing your best today.",
    "You cannot improve your past, but you can improve your future. Once time is wasted, life is wasted.",
    "Knowlegde can change your fate and English can accomplish your future.",
    "Don't aim for success if you want it; just do what you love and believe in, and it will come naturally.",
    "Jack of all trades and master of none.",
    "Judge not from appearances.",
    "Justice has long arms.",
    "Keep good men company and you shall be of the number.",
    "Kill two birds with one stone.",
    "Kings go mad, and the people suffer for it.",
    "Kings have long arms.",
    "Knowledge is power.",
    "Knowledge makes humble, ignorance makes proud.",
    "Learn and live.",
    "Learning makes a good man better and ill man worse.",
    "Learn not and know not.",
    "Learn to walk before you run.",
    "Let bygones be bygones.",
    "Let sleeping dogs lie.",
    "Let the cat out of the bag.",
    "Lies have short legs.",
    "Life is but a span.",
    "Life is half spent before we know what it is.",
    "Life is not all roses.",
    "Life without a friend is death.",
    "Like a rat in a hole.",
    "Like author, like book.",
    "Like father, like son.",
    "Like for like.",
    "Like knows like.",
    "Like mother, like daughter.",
    "Like teacher, like pupil.",
    "Like tree, like fruit.",
}

func init() {
    // 解析参数
    flag.Parse()
    fmt.Println("im server, host:" + *host + ":" + *port + ", count:", *count, ", step:", *step)
}

func main() {
    // 开启多核模式
    runtime.GOMAXPROCS(runtime.NumCPU())

    // 启动客户端
    for i := 1; i <= *open; i++ {
        go startClient(i)
        time.Sleep(time.Second / 1000)
    }

    os.Exit(0)
}

func startClient(i int) {
    // connect im server
    var tcpAddr *net.TCPAddr
    tcpAddr, _ = net.ResolveTCPAddr("tcp", *host+":"+*port)
    conn, _ := net.DialTCP("tcp", nil, tcpAddr)

    defer func() {
        // 捕获异常
        if err := recover(); err != nil {
            fmt.Println("tcpPipe defer recover error:", err)
        }

        conn.Close()
    }()

    //// login im server
    //loginBody := make(map[string]interface{})
    //loginBody["userId"] = strconv.Itoa(userId)
    //loginBody["platformId"] = platformId
    //loginBody["platformName"] = platformName
    //loginBody["time"] = common.GetTime()
    //protocal.Send(conn, config.IM_LOGIN, config.IM_FROM_TYPE_AI, loginBody)
    //
    //// send a message step sencond
    //for {
    //    // 随机休息时间、防止脚本启动消息时，同时发送的消息过多
    //    rand.Seed(time.Now().UnixNano() - int64(userId*userId*userId))
    //    randSecond := int(rand.Int31n(int32(*step)))
    //    for j := 1; j <= randSecond+1; j++ {
    //        time.Sleep(time.Second)
    //    }
    //
    //    messageBody := make(map[string]interface{})
    //
    //    // 获取随机消息
    //    rand.Seed(time.Now().UnixNano() - int64(userId*userId*userId))
    //    randKey := int(rand.Int31n(45))
    //    messageBody["msg"] = "[" + strconv.Itoa(totalMsg) + "]:" + randMsgArr[randKey]
    //
    //    // 发送消息
    //    protocal.Send(conn, config.IM_CHAT_BORADCAST, config.IM_FROM_TYPE_AI, messageBody)
    //
    //    totalMsg++
    //    if totalMsg%100 == 0 {
    //        fmt.Println("["+common.GetTimestamp()+"]"+"totalMsg:", totalMsg)
    //    }
    //
    //}
}