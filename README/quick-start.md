# 快速开始（部署指南）

本项目支持使用 **Docker Compose** 在本地一键启动，适合学习和调试分布式微服务架构。

## 环境准备

在开始之前，请确保本地已安装以下工具：

- Git
- Docker
- Docker Compose（推荐 v2 及以上版本）

## 启动步骤

首先配置**docker-compose.yml**文件的**ai**service的**environment**参数**MODEL_NAME**和**DASHSCOPE_API_KEY**，并保存。

```bash
# 1. 克隆仓库
git clone https://github.com/MoScenix/ai-code.git
cd ai-code

# 2. 启动所有服务
docker compose up -d

# 3. 查看服务状态

docker compose ps

# 4. 访问前端服务

http://localhost:8080

# 5. 关闭所有服务

docker compose down

# 6. 删除所有数据

docker compose down --vo

```

## 修改密码

在**docker-compose.yml**文件中,可以修改服务的**environment**参数，修改密码。


#### 分布式参数

对于**consul**作为注册中心的项目需要改**consul_address**参数，这样就可以部署在多个服务器上。

在**docker-compose.yml**文件保存所需的**server**就行。