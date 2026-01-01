# 机器翻译

学习如何使用 YFlow 的机器翻译功能自动填充翻译。

## 概述

YFlow 内置机器翻译功能，基于开源的 [LibreTranslate](https://libretranslate.com/) 引擎。该功能可以帮助你快速填充项目中的缺失翻译，支持多种语言之间的互译。

## 前提条件

在使用机器翻译之前，你需要确保：

1. 已部署 LibreTranslate 服务（见[部署指南](/deployment/docker)）
2. 已在环境变量中配置翻译服务地址

### 环境变量配置

在 `admin-backend/.env` 文件中配置：

```env
# LibreTranslate 服务地址
LIBRE_TRANSLATE_URL=http://localhost:5000

# LibreTranslate API Key（可选，如服务需要）
LIBRE_TRANSLATE_API_KEY=
```

## 使用机器翻译

### 通过 Web 界面

1. 进入「翻译」页面
2. 选择需要翻译的项目
3. 点击「机器翻译」按钮
4. 选择源语言和目标语言
5. 点击「开始翻译」

系统将自动扫描项目中的缺失翻译，并调用机器翻译服务进行填充。

### 支持的语言

机器翻译支持以下语言：

| 语言代码 | 语言名称 |
|---------|---------|
| en | 英语 |
| zh | 中文 |
| zh-TW | 繁体中文 |
| ja | 日语 |
| ko | 韩语 |
| fr | 法语 |
| de | 德语 |
| es | 西班牙语 |
| it | 意大利语 |
| ru | 俄语 |
| ar | 阿拉伯语 |
| pt | 葡萄牙语 |
| nl | 荷兰语 |
| pl | 波兰语 |
| tr | 土耳其语 |
| vi | 越南语 |
| th | 泰语 |
| hi | 印地语 |
| id | 印尼语 |
| ... | 更多语言 |

## 批量翻译

### 界面操作

1. 打开机器翻译对话框
2. 选择「源语言」（可选，如不选则自动检测）
3. 选择「目标语言」
4. 系统将显示预计翻译数量
5. 点击确认开始翻译

### 翻译进度

翻译过程中，你会看到：
- 总共需要翻译的条数
- 已完成的条数
- 失败的条数（如有）

## 注意事项

### 翻译质量

机器翻译提供的是基础翻译，建议：
- 重要文案请人工校对
- 保留原文中的变量占位符（如 `{name}`）
- 检查文化差异和本地化适配

### 速率限制

LibreTranslate 服务有速率限制，YFlow 已配置：
- 每批最多翻译 10 条
- 批次间隔 100ms

### 成本考虑

- 自托管 LibreTranslate 无需付费
- 但需要服务器资源运行翻译服务

## 故障排除

### 服务不可用

如果机器翻译按钮不可用，检查：

1. LibreTranslate 服务是否运行：
   ```bash
   docker compose ps | grep libretranslate
   ```

2. 服务健康状态：
   ```bash
   curl http://localhost:5000/languages
   ```

3. 后端配置是否正确：
   ```env
   LIBRE_TRANSLATE_URL=http://localhost:5000
   ```

### 翻译失败

翻译失败可能原因：
- 网络连接问题
- 源文本过长（建议单条不超过 500 字符）
- 不支持的语言组合

## 部署 LibreTranslate

### Docker 部署

YFlow 的 Docker Compose 已包含 LibreTranslate 服务：

```bash
cd admin-backend
docker compose up -d libretranslate
```

### 独立部署

如需独立部署 LibreTranslate：

```bash
docker run -d \
  --name libretranslate \
  -p 5000:5000 \
  -e LT_TITLE="My Translation" \
  -e LT_DISABLE_RATE_LIMIT=1 \
  libretranslate/libretranslate
```

### 生产环境配置

```yaml
# docker-compose.yml
services:
  libretranslate:
    image: libretranslate/libretranslate:latest
    restart: unless-stopped
    ports:
      - "5000:5000"
    environment:
      - LT_TITLE=YFlow Translation
      - LT_DISABLE_RATE_LIMIT=1
      - LT_REQUIRE_API_KEY=false
      # 生产环境建议启用 API Key
      # - LT_REQUIRE_API_KEY=true
      # - LT_API_KEY=your-api-key
    volumes:
      - libretranslate_data:/var/lib/LibreTranslate

volumes:
  libretranslate_data:
    driver: local
```

## 下一步

- [翻译管理最佳实践 →](/guide/best-practices)
- [团队协作翻译 →](/guide/team-collaboration)
- [CLI 使用 →](/guide/cli-guide)
