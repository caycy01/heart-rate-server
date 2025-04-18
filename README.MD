# 实时心率监测系统

## 📋 项目概述

基于Go语言开发的实时心率数据监测与共享系统，提供安全可靠的API端点用于数据上报、查询及可视化展示。采用Redis实时存储、GORM关系映射，支持Docker容器化部署。

## 🎖️ 服务器赞助商

[![SponsorName](https://img.shields.io/badge/Powered_by-慈云数据-blue?style=for-the-badge&logo=aws)](https://www.zovps.com/)

![Image](/static/sponsors/ciyun.png)

## 🚀 核心功能

### 用户认证体系

* 基于安全Cookie的认证
* 自动续期机制

### 心率数据管理

* 实时数据上报（支持毫秒级时间戳）
* 历史数据查询（基于Redis有序集合）
* 数据有效性验证（1-250 BPM范围限制）

### 安全共享机制

* UUID加密访问控制
* 临时访问令牌

## 🌐 API端点文档

### 认证相关

| 端点        | 方法   | 描述         | 请求体示例                                         |
|-----------|------|------------|-----------------------------------------------|
| /register | POST | 用户注册       | `{"username":"test","password":"test123456"}` |
| /login    | POST | 用户登录       | 同上                                            |
| /logout   | POST | 退出登录       | 无                                             |
| /uuid     | GET  | 获取当前用户UUID | 需认证                                           |

### 数据操作

| 端点                             | 方法   | 描述           | 参数示例                                                     |
|--------------------------------|------|--------------|----------------------------------------------------------|
| /receive_data                  | POST | 认证用户上报数据     | `{"data":{"heart_rate":72},"measured_at":1711700000000}` |
| /uuid/{uuid}/receive_data      | POST | 通过UUID上报数据   | 同上                                                       |
| /latest-heart-rate             | GET  | 获取最新心率（认证用户） | 无                                                        |
| /uuid/{uuid}/latest-heart-rate | GET  | 获取指定UUID最新数据 | 需URL参数                                                   |

### 可视化端点

| 端点                       | 方法  | 描述        |
|--------------------------|-----|-----------|
| /                        | GET | 主页        |
| /health                  | GET | 健康检查页     |
| /uuid/widget/view/{uuid} | GET | 嵌入式心率展示组件 |

## 🛠️ 安装运行

### 环境要求

* Go 1.24 或更高版本
* Redis
* Docker（可选）

### 本地运行

1. 克隆项目代码：

```sh
    git clone https://github.com/xihan123/heart-rate-server.git
    cd heart-rate-server
```

2. 配置环境变量：

```sh
    cp .env.example .env
    # 编辑 .env 文件以配置您的环境变量
```

3. 安装依赖：

```sh
    go mod download
 ```

4. 启动服务：

```sh
    go run main.go
```

### 使用Docker运行

1. 构建Docker镜像：

```sh
    docker build -t heart-rate-server .
```

2. 运行Docker容器：

```sh
    docker run -d -p 8080:8080 --name heart-rate-server heart-rate-server
```

### 安全注意事项

当部署到外网环境时必须启用HTTPS协议, 安全Cookie机制强制启用Secure属性 [参见SecureCookie实现](internal/middleware/auth.go#L32-L40)

### 环境变量配置

| 变量名              | 描述                                         | 默认值            |
|------------------|--------------------------------------------|----------------|
| SERVER_PORT      | 监听端口                                       | 8080           |
| DB_DSN           | Sqlite数据库路径                                | heartrate.db   |
| REDIS_ADDR       | Redis主机地址                                  | localhost:6379 |
| REDIS_PASSWORD   | Redis密码（如果有）                               | ""             |
| REDIS_DB         | Redis数据库索引                                 | 0              |
| BCRYPT_COST      | Bcrypt加密成本                                 | 10             |
| COOKIE_HASH_KEY  | Cookie加密密钥(64位Hex字符串 openssl rand -hex 64) | ""             |
| COOKIE_BLOCK_KEY | Cookie加密密钥(32位Hex字符串 openssl rand -hex 32) | ""             |

## 示例服务地址

[![SponsorName](https://img.shields.io/badge/Server-香港-blue?style=for-the-badge)](https://public-heart-rate-api.xihan.website/)
[![SponsorName](https://img.shields.io/badge/Server-CloudFlare-blue?style=for-the-badge)](https://public-heart-rate-api.xihan.lat/)

简单注册登录一下就可以用了，不过是基于CF的，国内速度可能很慢

**另外强调一下，这个只是个供测试的接口服务，没有接入任何机器人
这个服务稳定性无法保证，不要用于生产环境！！！**

1. 进入首页点立即开始

![Image](https://github.com/user-attachments/assets/985e63da-0919-4871-8008-7ee77d904305)

2. 点击登录，注册一个账号(不想注册的用test:test123456登录测试也行
)

![Image](https://github.com/user-attachments/assets/06b76b7c-03fb-4115-9681-8abc4908d1f2)

3. - 复制`数据上报接口 (POST)`内容为`https://public-heart-rate-api.xihan.lat/uuid/d7fc02ae-6973-4495-9fea-d7b877ab4329/receive_data`
    - 填入[模块](https://github.com/xihan123/HeartRateHook)时记得要去掉`receive_data`
    - 只需要填入`https://public-heart-rate-api.xihan.lat/uuid/d7fc02ae-6973-4495-9fea-d7b877ab4329/`就行

![Image](https://github.com/user-attachments/assets/7f178fce-1a28-4342-b9eb-e83b1c3c320a)

4. (可选)公开心率查看页面是个可以供直播展示心率用的html

![Image](https://github.com/user-attachments/assets/aaefdb14-baf6-425c-be75-0d9b9ffaaaeb)

## 📄 许可证

此项目基于MIT许可证开源，详情请参阅 [LICENSE](LICENSE) 文件。
