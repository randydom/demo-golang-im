package util

import (
	"os/exec"
	"strconv"
	"strings"
	"bytes"
	"os"
	"os/signal"
	"syscall"
	"log"
)


// 执行一个shell命令，不关心返回
func CommandStart(command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	err := cmd.Start()
	return err
}

// 执行shell命令
func Command(command string, argv []string) ([]byte, error) {
	cmd := exec.Command(command, argv...)
	output, err := cmd.CombinedOutput()

	return output, err
}

// 获取进程的pids
func GetPids(command string) []int {
	pids := make([]int, 0)

	cmd := exec.Command("/bin/sh", "-c", command)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return pids
	}

	for {
		line, err := out.ReadString('\n')
		if err != nil {
			break
		}
		tokens := strings.Split(line, " ")

		ft := make([]string, 0)
		for _, t := range tokens {
			if t != "" && t != "\t" {
				ft = append(ft, t)
			}
		}

		pid, err := strconv.Atoi(ft[1])
		if err != nil {
			continue
		}

		pids = append(pids, pid)
	}

	return pids
}

func WaitTerminatingStop() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	WAIT_LOOP:
	for {
		select {
		case <-sc:
			{
				//	app cancelled by user , do clean up work
				log.Println("[IM] Terminating ...")
				break WAIT_LOOP
			}
		}
	}
}