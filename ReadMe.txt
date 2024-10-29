/back_end
│
├── /cmd                 # 应用入口
│   └── /your_app       # 应用的主要程序
│       └── main.go     # main 函数，启动服务器
│
├── /config              # 配置文件
│   ├── config.go       # 配置结构体和加载配置的逻辑
│   └── config.yaml     # 应用的配置文件
│
├── /internal            # 内部业务逻辑，不应被外部导入
│   ├── /handler         # 处理 HTTP 请求的处理器
│   │   ├── user.go     # 用户相关的处理函数
│   │   └── post.go     # 文章相关的处理函数
│   │
│   ├── /middleware      # 中间件
│   │   ├── auth.go     # 身份验证中间件
│   │   └── logger.go   # 日志记录中间件
│   │
│   ├── /model           # 数据模型
│   │   ├── user.go     # 用户模型
│   │   └── post.go     # 文章模型
│   │
│   ├── /repository      # 数据访问层
│   │   ├── user_repo.go # 用户数据访问逻辑
│   │   └── post_repo.go # 文章数据访问逻辑
│   │
│   └── /service         # 业务逻辑层
│       ├── user_service.go # 用户业务逻辑
│       └── post_service.go # 文章业务逻辑
│
├── /pkg                 # 公共库
│   └── util.go         # 工具函数
│
├── /scripts             # 脚本文件
│   └── migrate.go      # 数据库迁移脚本
│
├── /test                # 测试文件
│   ├── handler_test.go  # 处理器测试
│   └── service_test.go  # 服务层测试
│
├── go.mod               # Go 模块文件
└── go.sum               # Go 依赖文件
文件夹和文件说明
/cmd: 存放应用程序的入口文件。每个应用都可以有自己的子目录，例如 your_app，里面包含 main.go 文件，用于启动应用。

/config: 存放应用的配置相关文件，包括 Go 代码和 YAML 配置文件。

/internal: 存放应用的内部业务逻辑，不应被外部包引用。

/handler: 存放处理 HTTP 请求的函数，通常会根据功能模块划分。
/middleware: 存放中间件，处理请求和响应的逻辑，比如身份验证、日志记录等。
/model: 定义数据模型，通常是与数据库交互的结构体。
/repository: 数据访问层，负责与数据库交互的逻辑。
/service: 业务逻辑层，处理核心业务逻辑，协调不同的模型和数据访问层。
/pkg: 存放公共库和工具函数，其他项目可以引用。

/scripts: 存放一些管理脚本，比如数据库迁移脚本。

/test: 存放测试文件，按照功能模块进行划分，便于测试代码的管理和维护。

go.mod 和 go.sum: Go 的模块管理文件，记录项目依赖的库和版本信息。