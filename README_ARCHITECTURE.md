# DialogTree 架构重构说明

## 📋 重构概览

本次重构将 DialogTree 从纯 CLI 应用扩展为支持 Web UI 的完整对话管理系统，核心特性：

- ✅ **会话管理**: Session → Dialog(树状) → Conversation 三层架构
- ✅ **智能上下文**: 短期记忆(N轮对话) + 长期记忆(向量检索)
- ✅ **向量数据库**: 基于 Qdrant 的语义检索
- ✅ **Web API**: RESTful 接口支持前端开发
- ✅ **CLI 兼容**: 重构后仍支持命令行交互

## 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐
│   Web Frontend  │    │   CLI Interface │
│    (Vue.js)     │    │  (Bubbletea)    │
└─────────┬───────┘    └─────────┬───────┘
          │                      │
          ▼                      ▼
┌─────────────────────────────────────────┐
│              Gin Router                 │
│  /api/sessions  /api/dialog  /cli/*     │
└─────────────────┬───────────────────────┘
                  │
          ┌───────▼───────┐
          │  API Layer    │
          │ session_api   │
          │  dialog_api   │
          └───────┬───────┘
                  │
          ┌───────▼───────┐
          │ Service Layer │
          │ dialog_service│
          │vector_service │
          │embedding_svc  │
          └───────┬───────┘
                  │
    ┌─────────────┼─────────────┐
    ▼             ▼             ▼
┌──────────┐  ┌─────────┐  ┌─────────┐
│PostgreSQL│  │  Redis  │  │ Qdrant  │
│(主数据)   │  │ (缓存)   │  │(向量DB) │
└──────────┘  └─────────┘  └─────────┘
```

## 📁 目录结构变化

### 新增文件
```
├── api/
│   ├── session_api/           # 会话管理 API
│   └── dialog_api/            # 对话交互 API
├── service/
│   ├── dialog_service/        # 对话业务逻辑
│   ├── vector_service/        # 向量数据库服务
│   └── embedding_service/     # 文本向量化服务
├── conf/
│   └── conf_vector.go         # 向量数据库配置
└── core/
    └── init_vector.go         # 向量服务初始化
```

### 配置更新
```yaml
# config.yaml 新增配置
ai:
  contextLayers: 3                    # 短期记忆层数
  embeddingModel: "text-embedding-3-small"
vector:
  enable: true
  provider: "qdrant"
  qdrant:
    host: "127.0.0.1"
    port: 6333
    collection: "dialog_memories"
  topK: 5
  similarityThreshold: 0.7
```

## 🚀 部署指南

### 1. 启动 Qdrant
```bash
docker run -p 6333:6333 qdrant/qdrant
```

### 2. 更新依赖
```bash
go mod tidy
```

### 3. 启动服务

**Web 模式:**
```bash
./dialogTree web
```

**CLI 模式:**
```bash
./dialogTree dialog recent    # 进入最近会话
./dialogTree dialog enter     # 选择会话
./dialogTree chitchat         # 快速对话(无状态)
```

## 🌐 API 接口

### 会话管理
- `GET /api/sessions` - 获取会话列表
- `POST /api/sessions` - 创建新会话
- `GET /api/sessions/:id/tree` - 获取对话树
- `DELETE /api/sessions/:id` - 删除会话

### 对话交互
- `POST /api/dialog/chat` - 发起新对话(流式)
- `POST /api/dialog/chat/sync` - 发起新对话(同步)
- `PUT /api/conversations/:id/star` - 标星对话
- `PUT /api/conversations/:id/comment` - 添加评论

## 🧠 智能上下文机制

### 短期记忆
- 从当前节点往上追溯 N 轮对话(可配置)
- 使用对话摘要而非完整内容,节省 token

### 长期记忆  
- 自动向量化所有问答对
- 基于问题相似度检索历史相关内容
- 支持会话级别的记忆隔离

### 上下文构建流程
```
新问题 → embedding → 向量检索(topK) → 组合短期上下文 → 发送给AI
         ↓
    AI回复 → 保存到数据库 → 向量化存储
```

## 🎯 下一步计划

1. **前端开发**: Vue.js 对话树可视化界面
2. **高级功能**: 
   - 会话分支合并
   - 多轮对话模版
   - 对话导出功能
3. **性能优化**: 
   - 向量索引优化
   - 缓存策略完善
   - 流式响应优化

## 💡 使用示例

### Web API 调用
```bash
# 创建会话
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"title": "我的对话", "categoryID": 1}'

# 发起对话
curl -X POST http://localhost:8080/api/dialog/chat/sync \
  -H "Content-Type: application/json" \
  -d '{"content": "你好", "sessionId": 1}'
```

### CLI 使用
```bash
# 快速进入最近会话
./dialogTree dialog

# 选择特定会话
./dialogTree dialog enter
```

重构完成! 🎉 现在你有了一个功能完整的对话管理系统。