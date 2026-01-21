package app

import (
	"AI_Chat/internal/handler"
	"AI_Chat/internal/memory"
	"AI_Chat/internal/middleware"
	"AI_Chat/internal/model"
	"AI_Chat/internal/repository"
	"AI_Chat/pkg/db"
	"AI_Chat/pkg/utils"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type app struct {
	router             *gin.Engine
	authHandler        *handler.AuthHandler
	testHandler        *handler.TestHandler
	chatHandler        *handler.ChatHandler
	personaHandler     *handler.PersonaHandler
	memoryHandler      *handler.MemoryHandler
	privateInterceptor []gin.HandlerFunc
}

var App app

func InitRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.GinLogger())
	root := router.Group("/api/v1")
	{
		public := root.Group("/")
		auth := public.Group("/auth")
		{
			auth.POST("/register", App.authHandler.Register)
			auth.POST("/login", App.authHandler.Login)
			auth.POST("/logout", App.authHandler.Logout)
		}

		//需要鉴权的接口
		private := root.Group("/", App.privateInterceptor...)
		{
			private.POST("/test", App.testHandler.Test)
			chatGroup := private.Group("/ai")
			{
				chatGroup.POST("/create-conversation", App.chatHandler.CreateConversation)
				chatGroup.POST("/chat-with-persona", App.chatHandler.ChatWithPersona)
				chatGroup.GET("/conversations", App.chatHandler.GetConversations)
				chatGroup.POST("/conversation-messages", App.chatHandler.GetConversationMessages)
			}
			personaGroup := private.Group("/persona")
			{
				personaGroup.POST("/create", App.personaHandler.CreatePersona)
				personaGroup.GET("/list", App.personaHandler.GetPersonas)
				
				// 记忆管理路由
				memoryGroup := personaGroup.Group("/:personaId/memory")
				{
					memoryGroup.POST("/create", App.memoryHandler.CreateMemory)
					memoryGroup.GET("/list", App.memoryHandler.GetMemories)
					memoryGroup.PUT("/:memoryId", App.memoryHandler.UpdateMemory)
					memoryGroup.DELETE("/:memoryId", App.memoryHandler.DeleteMemory)
				}
			}
		}

	}
	return router
}
func Init() {
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
	// 数据库迁移
	db.DB.AutoMigrate(&model.Memory{})
	// 假设 InitDB 返回了 *gorm.DB 或 *sql.DB，这里可以处理关闭逻辑
	// sqlDB, _ := database.DB()
	// defer sqlDB.Close()
	err = db.InitRedis(utils.Config_Instance.GetRedisConfig())
	if err != nil {
		utils.Log.Error("初始化Redis连接失败", zap.Error(err))
		return
	}
	err = db.InitMilvus(utils.Config_Instance.GetMilvusConfig())
	if err != nil {
		utils.Log.Error("初始化Milvus连接失败", zap.Error(err))
		return
	}
	// 4. 初始化雪花算法
	if err := utils.InitSnowflake(1); err != nil {
		utils.Log.Error("初始化雪花算法失败", zap.Error(err))
		return
	}
	//初始化Handler

	userBaseRepository := repository.NewUserBaseRepository(db.DB)
	userSessionRepository := repository.NewUserSessionRepository(db.RedisClient)
	conversationRepository := repository.NewConversationRepository(db.DB)
	personaRepository := repository.NewPersonaRepository(db.DB)
	memoryRepository := repository.NewMemoryRepository(db.DB)

	milvusConfig := utils.Config_Instance.GetMilvusConfig()
	milvusStore := memory.NewMilvusStore(db.MilvusClient, memory.MilvusStoreConfig{
		Collection: milvusConfig.Collection,
		Dimension:  milvusConfig.Dimension,
		MetricType: milvusConfig.MetricType,
	})
	memoryService := memory.NewMemoryService(memoryRepository, db.RedisClient, milvusStore)

	authHandler := handler.NewAuthHandler(userBaseRepository, userSessionRepository)
	testHandler := handler.NewTestHandler()
	chatHandler := handler.NewChatHandler(conversationRepository, personaRepository, memoryService)
	personaHandler := handler.NewPersonaHandler(personaRepository)
	memoryHandler := handler.NewMemoryHandler(memoryRepository, personaRepository, memoryService)
	
	App.authHandler = authHandler
	App.testHandler = testHandler
	App.chatHandler = chatHandler
	App.personaHandler = personaHandler
	App.memoryHandler = memoryHandler
	//初始化Interceptor

	privateInterceptor := []gin.HandlerFunc{
		middleware.Auth(userSessionRepository, userBaseRepository),
	}
	App.privateInterceptor = privateInterceptor

}
func Run() {
	Init()
	App.router = InitRouter()
	App.router.Run(":8001")
}
