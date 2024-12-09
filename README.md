wechatbot/                # 项目根目录
├── cmd/                  # 主程序入口
│   └── main.go           # 程序入口
├── config/               # 配置相关
│   ├── conf.yaml         # 配置文件
│   └── config.go         # 配置加载逻辑
├── internal/             # 内部逻辑（核心代码）
│   ├── bot/              # 机器人功能逻辑
│   │   ├── bot.go        # 核心机器人逻辑
│   │   ├── handlers/     # 消息处理器
│   │   │   ├── text.go   # 文本消息处理
│   │   │   ├── image.go  # 图片消息处理
│   │   │   └── event.go  # 事件处理
│   │   └── commands/     # 指令处理器
│   │       ├── weather.go # 天气查询指令
│   │       ├── help.go   # 帮助指令
│   │       └── ...       # 其他指令
│   ├── services/         # 业务服务（如天气查询、第三方 API 调用）
│   │   ├── weather.go    # 天气服务
│   │   ├── chatgpt.go    # AI 聊天服务
│   │   └── ...           # 其他服务
│   ├── db/               # 数据库操作
│   │   ├── models/       # 数据模型
│   │   │   └── user.go   # 用户数据模型
│   │   ├── repository/   # 数据操作
│   │   │   ├── user_repo.go # 用户数据操作
│   │   └── db.go         # 数据库连接管理
│   └── utils/            # 工具函数
│       ├── logger.go     # 日志工具
│       ├── http.go       # HTTP 请求工具
│       └── ...           # 其他工具
├── pkg/                  # 第三方扩展或工具库
│   ├── wechat_sdk/       # 微信 SDK 或封装
│   └── ...               # 其他扩展
├── scripts/              # 脚本（如部署脚本）
│   └── migrate_db.sh     # 数据库迁移脚本
├── tests/                # 测试代码
│   ├── bot_test.go       # 机器人核心逻辑测试
│   └── services_test.go  # 服务层测试
├── go.mod                # Go 模块文件
├── go.sum                # Go 依赖文件
└── README.md             # 项目说明文档
