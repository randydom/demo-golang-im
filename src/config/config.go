package config

import (
    "util"
    "logger"
    "os"
)

func ParseConfig(path string) map[string]string {
    // 读取JOSN配置文件
    configs, err := util.LoadJsonConfig(path)
    if (err != nil) {
        logger.LogError("Error: %s", err);
        os.Exit(1);
    }

    // 解析配置
    parseConfigs := make(map[string]string)

    // 检测配置文件是否完整
    checkKeys := []string{"host", "port", "system_key", "login_key", "chat_key"}
    for _, key := range checkKeys {
        if _, exists := configs[key]; !exists {
            logger.LogError("constant [%s] not found in constant file:%s", key, path);
        } else {
            parseConfigs[key] = configs[key].(string)
        }
    }

    return parseConfigs
}
