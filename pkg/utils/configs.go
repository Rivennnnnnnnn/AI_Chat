package utils

import (
	"errors"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

type MysqlConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	MaxIdleConns int
	MaxOpenConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type Config struct {
	Mysql MysqlConfig
	Redis RedisConfig
}

func (c *Config) GetMysqlConfig() MysqlConfig {
	return c.Mysql
}
func (c *Config) SetMysqlConfig(mysqlConfig MysqlConfig) {
	c.Mysql = mysqlConfig
}
func (c *Config) GetRedisConfig() RedisConfig {
	return c.Redis
}
func (c *Config) SetRedisConfig(redisConfig RedisConfig) {
	c.Redis = redisConfig
}
// 全局变量声明
var config_names []string = []string{
	"mysql",
	"redis",
}
var config_names_flag map[string]bool = map[string]bool{}

var Config_Instance Config

func InitConfig() error {
	// 获取当前函数所在文件的路径 (pkg/utils/configs.go)
	_, filename, _, _ := runtime.Caller(0)

	// 计算出项目根目录或 configs 目录的绝对路径
	// filename 是 d:/Code/GoProject/AI_Chat/pkg/utils/configs.go
	// 我们需要向上退两级到 AI_Chat，再进到 configs
	root := filepath.Join(filepath.Dir(filename), "../../configs")
	viper.AddConfigPath(root)
	for _, config := range config_names {
		config_names_flag[config] = false
	}
	for _, config_name := range config_names {
		viper.SetConfigName(config_name)
		viper.SetConfigType("yaml")
		err := viper.ReadInConfig()
		if err != nil {
			return errors.New("必要配置文件错误:" + err.Error() + ",请检查配置文件是否正确\n")
		}
		config_names_flag[config_name] = true
		switch config_name {
		case "mysql":
			viper.SetDefault("mysql.max_idle_conns", 10)
			viper.SetDefault("mysql.max_open_conns", 100)
			mysqlConfig := MysqlConfig{
				Host:         viper.GetString("mysql.host"),
				Port:         viper.GetInt("mysql.port"),
				User:         viper.GetString("mysql.user"),
				Password:     viper.GetString("mysql.password"),
				DBName:       viper.GetString("mysql.dbname"),
				MaxIdleConns: viper.GetInt("mysql.max_idle_conns"),
				MaxOpenConns: viper.GetInt("mysql.max_open_conns"),
			}
			if mysqlConfig.Host == "" || mysqlConfig.Port == 0 || mysqlConfig.User == "" || mysqlConfig.Password == "" || mysqlConfig.DBName == "" {
				return errors.New("mysql配置文件错误: 请检查配置文件是否正确,必要配置项（host,port,user,password,dbname）不能为空\n")
			}
			Config_Instance.SetMysqlConfig(mysqlConfig)
		case "redis":
			viper.SetDefault("redis.db", 0)
			redisConfig := RedisConfig{
				Host:     viper.GetString("redis.host"),
				Port:     viper.GetInt("redis.port"),
				Password: viper.GetString("redis.password"),
				DB:       viper.GetInt("redis.db"),
			}
			if redisConfig.Host == "" || redisConfig.Port == 0 {
				return errors.New("redis配置文件错误: 请检查配置文件是否正确,必要配置项（host,port）不能为空\n")
			}
			Config_Instance.SetRedisConfig(redisConfig)	
		}
	}
	return nil
}
