# 在线考试审批系统

一个基于Go语言开发的在线考试审批系统，支持教师创建试卷、学生参加考试、教师评分等功能。

## 技术栈

- **后端语言**: Go (Golang)
- **Web框架**: Gin
- **ORM框架**: GORM
- **数据库**: MySQL/PostgreSQL (通过GORM支持)
- **项目架构**: MVC模式

## 系统架构

系统采用经典的MVC架构设计：

- **Models**: 数据模型层，定义系统核心数据结构
- **Controllers**: 控制器层，处理HTTP请求和响应
- **Services**: 业务逻辑层，实现核心业务功能
- **Repositories**: 数据访问层，处理数据库操作

## 主要功能

### 用户管理
- 多角色支持（学生、教师、管理员）
- 用户认证和授权
- 密码管理
- 用户信息管理

### 试卷管理
- 教师创建试卷
- 试卷发布和分发
- 试卷状态管理（草稿、已发布、已审批等）
- 试卷删除和更新

### 考试系统
- 学生参加考试
- 在线提交答案
- 考试时间管理
- 考试状态跟踪

### 评分系统
- 教师评分功能
- 评分记录管理
- 评语系统
- 成绩统计

### 管理员功能
- 用户管理
- 系统监控
- 数据统计
- 权限控制

## 项目结构

```
exam-approval-system/
├── configs/         # 配置文件
├── controllers/     # 控制器
├── models/         # 数据模型
├── repositories/   # 数据访问层
├── services/       # 业务逻辑层
├── templates/      # 前端模板
└── main.go         # 程序入口
```

## 安装和运行

1. 克隆项目
```bash
git clone https://github.com/yourusername/exam-approval-system.git
```

2. 安装依赖
```bash
go mod download
```

3. 配置数据库
- 创建数据库
- 修改配置文件中的数据库连接信息

4. 运行项目
```bash
go run main.go
```

## API文档

### 用户相关API
- POST /login - 用户登录
- POST /register - 用户注册
- GET /dashboard - 获取仪表板信息

### 试卷相关API
- POST /papers - 创建试卷
- GET /papers - 获取试卷列表
- PUT /papers/:id - 更新试卷
- DELETE /papers/:id - 删除试卷

### 考试相关API
- POST /exams/:id/submit - 提交考试答案
- GET /exams/:id - 获取考试详情
- GET /exams/:id/result - 获取考试结果

### 评分相关API
- POST /grade - 教师评分
- GET /grades - 获取评分列表
