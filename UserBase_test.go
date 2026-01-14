package main

import (
	"AI_Chat/internal/model"
	"AI_Chat/internal/repository"
	"AI_Chat/pkg/db"
	"AI_Chat/pkg/utils"
	"fmt"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// 1. 初始化日志
	if err := utils.InitLogger(); err != nil {
		// 如果日志初始化失败，只能用标准库打印
		println("Failed to initialize logger:", err.Error())
		os.Exit(1)
	}
	// 确保程序退出时日志缓冲区已刷新
	defer utils.Log.Sync()

	// 2. 初始化配置
	if err := utils.InitConfig(); err != nil {
		utils.Log.Error("初始化配置失败", zap.Error(err))
		return
	}

	// 3. 初始化数据库
	_, err := db.InitDB(utils.Config_Instance.GetMysqlConfig())
	if err != nil {
		utils.Log.Error("初始化Mysql连接失败", zap.Error(err))
		return
	}
	// 假设 InitDB 返回了 *gorm.DB 或 *sql.DB，这里可以处理关闭逻辑
	// sqlDB, _ := database.DB()
	// defer sqlDB.Close()

	err = db.InitRedis(utils.Config_Instance.GetRedisConfig())
	if err != nil {
		utils.Log.Error("初始化Redis连接失败", zap.Error(err))
		return
	}

	// 4. 初始化雪花算法
	if err := utils.InitSnowflake(1); err != nil {
		utils.Log.Error("初始化雪花算法失败", zap.Error(err))
		return
	}
	os.Exit(m.Run())
}
func TestUserBaseRepository_CreateUserBase(t *testing.T) {
	userBaseRepository := repository.NewUserBaseRepository(db.DB)
	userBase := &model.UserBase{
		Username: "test",
		Password: "test",
		Email:    "test@test.com",
	}
	err := userBaseRepository.CreateUserBase(userBase)
	if err != nil {
		t.Errorf("创建用户基本信息失败: %v", err)
	}
}

func TestUserSessionRepository_SetUserSession(t *testing.T) {
	userSessionRepository := repository.NewUserSessionRepository(db.RedisClient)
	userBase := &model.UserBase{
		Username: "test",
		Password: "test",
		Email:    "test@test.com",
	}
	sessionId, err := userSessionRepository.SetUserSession(userBase, gin.CreateTestContextOnly(nil, gin.Default()))
	if err != nil {
		t.Errorf("设置用户会话失败: %v", err)
	}
	fmt.Println(sessionId)
}
