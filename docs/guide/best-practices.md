# 翻译管理最佳实践

本文档提供在 i18n-flow 中高效管理翻译的最佳实践，涵盖项目组织、键命名、同步策略和团队协作等方面。

## 1. 项目组织结构

良好的项目组织结构是翻译管理的基础。一个清晰的结构可以帮助团队快速定位和管理翻译资源。

### 推荐目录结构

```
src/
├── locales/
│   ├── en/
│   │   ├── common.json      # 通用文本（按钮、提示语等）
│   │   ├── auth.json        # 认证相关（登录、注册、密码重置）
│   │   ├── dashboard.json   # 仪表板相关
│   │   └── errors.json      # 错误消息
│   ├── zh-CN/
│   │   ├── common.json
│   │   ├── auth.json
│   │   ├── dashboard.json
│   │   └── errors.json
│   └── ja-JP/
│       ├── common.json
│       ├── auth.json
│       ├── dashboard.json
│       └── errors.json
└── ...
```

### 模块化分组原则

将翻译按功能模块分组可以提高可维护性和可发现性：

| 模块 | 包含内容 |
|------|----------|
| `common` | 通用按钮、导航、提示语 |
| `auth` | 登录、注册、密码相关 |
| `dashboard` | 仪表板、统计、图表 |
| `errors` | 错误消息、验证提示 |
| `forms` | 表单标签、验证消息 |
| `modals` | 弹窗标题、确认按钮 |

### 配置文件示例

```json
// .i18nrc.json
{
  "messagesDir": "./src/locales",
  "projectId": 1,
  "apiUrl": "http://localhost:8080/api",
  "apiKey": "${I18N_API_KEY}",
  "languageMapping": {
    "zh_CN": "zh",
    "zh_TW": "tw"
  }
}
```

## 2. 键命名规范

一致的键命名规范可以提高翻译的可读性和可维护性。

### 点分隔命名法（推荐）

使用点号分隔的层级命名，反映翻译键的所属模块和用途：

```
# 推荐格式
auth.login.title              // 认证模块 > 登录页面 > 标题
auth.login.button.submit      // 认证模块 > 登录页面 > 按钮 > 提交
auth.login.error.invalid      // 认证模块 > 登录页面 > 错误 > 无效
common.button.cancel          // 通用 > 按钮 > 取消
common.error.network          // 通用 > 错误 > 网络
common.error.validation       // 通用 > 错误 > 验证

# 不推荐格式
loginTitle                    // 含义不清晰
LoginButtonSubmit             // 冗长，不好维护
ERROR_NETWORK                 // 全大写，不一致
user-login-page-title         // 连字符，不如点号清晰
```

### 命名空间约定

使用统一的命名空间前缀来组织翻译键：

| 命名空间 | 用途 | 示例 |
|----------|------|------|
| `common` | 全局通用文本 | `common.button.save`, `common.confirm` |
| `auth` | 认证相关 | `auth.login.title`, `auth.register.button` |
| `dashboard` | 仪表板相关 | `dashboard.title`, `dashboard.stats.users` |
| `errors` | 错误消息 | `errors.notFound`, `errors.permission` |
| `validation` | 表单验证 | `validation.email.required`, `validation.password.min` |

### 避免常见错误

```json
// ❌ 错误示例
{
  "Submit": "提交",              // 使用英文作为键
  "BUTTON_1": "确定",            // 含义不清晰
  "msg1": "密码错误",            // 命名无意义
  "UserNameLabel": "用户名"     // 混合大小写
}

// ✅ 正确示例
{
  "common.button.submit": "提交",
  "auth.login.error.password": "密码错误",
  "validation.username.required": "用户名不能为空"
}
```

## 3. 翻译键管理

### 使用描述和上下文

为翻译键添加描述可以帮助译者理解翻译的用途和上下文：

```json
{
  "auth.login.button.submit": {
    "value": "登录",
    "description": "登录表单的提交按钮，显示在用户名和密码输入框下方"
  },
  "common.confirm.delete": {
    "value": "确认删除",
    "description": "删除确认弹窗中的确认按钮，用于确认删除操作"
  }
}
```

### 变量占位符

在翻译中使用变量占位符可以创建动态翻译：

```json
{
  "greeting": "Hello, {name}!",
  "items_count": "{count} items",
  "welcome_message": "Welcome, {username}! You have {notificationCount} new notifications.",
  "date_format": "{year}年{month}月{day}日"
}
```

在前端代码中使用：

```typescript
// Vue 3 示例
const t = useI18n()
t('greeting', { name: 'John' })  // "Hello, John!"

// 带多个变量
t('welcome_message', {
  username: 'John',
  notificationCount: 5
})  // "Welcome, John! You have 5 new notifications."
```

### 复数形式处理

处理不同语言的复数规则：

```json
{
  "items_count": {
    "one": "{count} item",
    "other": "{count} items"
  },
  "messages_unread": {
    "zero": "No messages",
    "one": "{count} unread message",
    "other": "{count} unread messages"
  }
}
```

## 4. 同步策略

### 导入流程

使用 `import` 命令将本地翻译导入到 i18n-flow 后端：

```bash
# 标准导入
i18n-flow import

# 预览导入内容（不实际执行）
i18n-flow import --dry-run

# 指定配置文件
i18n-flow import --config .i18nrc.json
```

导入流程说明：

```
1. 扫描 messagesDir 目录下的所有 JSON 文件
2. 读取每种语言的翻译键值对
3. 分批推送到后端（每批 50 个键）
4. 显示导入结果统计（新增/更新/失败）
```

### 同步流程

使用 `sync` 命令从后端同步翻译到本地：

```bash
# 标准同步（保留本地修改）
i18n-flow sync

# 预览同步差异
i18n-flow sync --dry-run

# 强制覆盖所有本地翻译
i18n-flow sync --force

# 指定配置文件
i18n-flow sync --config .i18nrc.json
```

### 冲突解决策略

| 场景 | 处理方式 | 命令 |
|------|----------|------|
| 本地新增 key | 推送到后端 | `i18n-flow import` |
| 后端新增 key | 同步到本地 | `i18n-flow sync` |
| 双方修改同一 key | 默认保留本地 | `i18n-flow sync` |
| 需要覆盖本地 | 强制覆盖 | `i18n-flow sync --force` |

### 推荐的同步工作流

```bash
# 1. 开始工作前同步最新翻译
i18n-flow sync

# 2. 翻译工作...

# 3. 预览更改
i18n-flow import --dry-run

# 4. 确认无误后导入
i18n-flow import

# 5. 同步回本地确保一致
i18n-flow sync
```

### 环境变量覆盖

在 CI/CD 环境中使用环境变量：

```bash
# 设置环境变量
export I18N_API_URL="https://api.i18n.example.com/api"
export I18N_API_KEY="${PRODUCTION_API_KEY}"
export I18N_PROJECT_ID="123"

# 运行同步
i18n-flow sync
```

## 5. 团队协作

### 角色分工

| 角色 | 职责 | 权限 |
|------|------|------|
| **Owner** | 项目负责人 | 完整管理权限 |
| **Editor** | 翻译人员 | 编辑翻译、导入导出 |
| **Viewer** | 审核人员 | 只读、审核反馈 |

### 推荐团队结构

```
项目 A
├── @owner: 技术主管
│   ├── @editor: 翻译团队 Lead
│   │   ├── @editor: 翻译人员 A
│   │   └── @editor: 翻译人员 B
│   └── @viewer: 产品经理（审核）
```

### 审核流程

1. **翻译阶段**：Editor 完成翻译并标记为"待审核"
2. **审核阶段**：Viewer 检查翻译质量，添加反馈或通过
3. **发布阶段**：Owner 确认后同步到生产环境

### 沟通最佳实践

- **使用描述字段**：为每个翻译键添加详细描述
- **添加上下文注释**：解释翻译的使用场景
- **定期同步**：每天同步翻译，避免冲突
- **使用 issue 跟踪**：翻译相关问题使用项目 issue 跟踪

## 6. 语言配置

### 默认语言选择

选择一种语言作为"源语言"（通常是需要翻译最多的语言）：

| 选择依据 | 建议 |
|----------|------|
| 开发团队语言 | 优先选择团队最熟悉的语言 |
| 内容丰富度 | 选择内容最完整的语言 |
| 英语作为通用选择 | 英语通常作为默认语言 |

### 语言代码映射

处理本地文件命名与后端语言代码不一致的情况：

```json
{
  "languageMapping": {
    "zh_CN": "zh",       // 本地 zh_CN.json → 后端 zh
    "zh_TW": "tw",       // 本地 zh_TW.json → 后端 tw
    "en_US": "en",       // 本地 en_US.json → 后端 en
    "pt_BR": "pt"        // 本地 pt_BR.json → 后端 pt
  }
}
```

### 区域变体处理

| 场景 | 处理方式 |
|------|----------|
| 简体中文 vs 繁体中文 | 使用 `zh-CN` 和 `zh-TW` |
| 美式英语 vs 英式英语 | 使用 `en-US` 和 `en-GB` |
| 西班牙语不同地区 | 使用 `es-ES`, `es-MX`, `es-AR` |

## 7. 质量控制

### 翻译一致性

保持相同术语的一致翻译：

```json
{
  "common.button.save": "保存",
  "common.button.submit": "提交",
  "common.button.confirm": "确认",
  "common.button.cancel": "取消"
}
```

避免在同一项目中混用同义词：

```json
// ❌ 不一致
{
  "user.title": "用户",
  "account.label": "账户",
  "customer.name": "客户"
}

// ✅ 一致
{
  "user.title": "用户",
  "user.account.label": "账户",
  "customer.info": "客户信息"
}
```

### 长度限制

考虑 UI 布局限制：

```json
{
  "auth.login.button.submit": "登录",           // 短文本
  "auth.login.welcome": "欢迎回来",             // 适中
  "common.error.network": "网络连接失败，请检查网络设置"  // 较长
}
```

### 特殊字符处理

| 类型 | 示例 | 注意事项 |
|------|------|----------|
| HTML 标签 | `<b>重要</b>` | 确保目标语言也支持 |
| 换行符 | `第一行\n第二行` | 使用 `\n` 而非实际换行 |
| 引号 | `"请输入\"用户名\""` | 注意转义 |
| 格式化 | `{name} 的评分：{score}` | 保持变量顺序 |

## 8. 常见问题

### 键冲突

当本地和远程存在相同 key 但值不同时：

```bash
# 查看冲突
i18n-flow sync --dry-run

# 解决方式
# 1. 使用本地值（不执行同步）
# 2. 使用远程值（先 sync 再编辑）
# 3. 手动合并后导入
```

### 缺失翻译

检测和修复缺失的翻译：

```bash
# 同步时查看缺失
i18n-flow sync --dry-run

# 常见原因
# 1. 新增 key 未导入
# 2. 语言文件不同步
# 3. 键名拼写错误
```

### 格式问题

JSON 格式验证：

```bash
# 使用 CLI 检查
i18n-flow import --dry-run

# 或手动验证
node -e "JSON.parse(require('fs').readFileSync('en/common.json'))"
```

### 编码问题

确保文件使用 UTF-8 编码：

```bash
# 检查编码
file -I en/common.json

# 转换编码（如果需要）
iconv -f GBK -t UTF-8 zh-CN/original.json > zh-CN/converted.json
```

## 9. 集成示例

### Vue 3 集成

```typescript
// src/plugins/i18n.ts
import { createI18n } from 'vue-i18n'
import messages from '@/locales/en/common.json'

export default createI18n({
  legacy: false,
  locale: localStorage.getItem('locale') || 'en',
  fallbackLocale: 'en',
  messages: {
    en: messages,
    zh: loadLocaleMessages('zh-CN'),
    ja: loadLocaleMessages('ja-JP')
  }
})

function loadLocaleMessages(locale: string) {
  return require(`@/locales/${locale}/common.json`)
}
```

```vue
<!-- 使用示例 -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <h1>{{ t('auth.login.title') }}</h1>
  <button>{{ t('common.button.submit') }}</button>
</template>
```

### Next.js 集成

```typescript
// src/i18n.ts
import 'server-only'

const dictionaries = {
  en: () => import('@/locales/en.json').then(module => module.default),
  zh: () => import('@/locales/zh-CN.json').then(module => module.default),
  ja: () => import('@/locales/ja-JP.json').then(module => module.default)
}

export const getDictionary = async (locale: string) => {
  return dictionaries[locale as keyof typeof dictionaries]?.() ?? dictionaries.en()
}
```

```typescript
// app/[lang]/page.tsx
import { getDictionary } from '@/i18n'

export default async function HomePage({ params }: { params: { lang: string } }) {
  const dict = await getDictionary(params.lang)

  return (
    <h1>{dict.auth.login.title}</h1>
    <button>{dict.common.button.submit}</button>
  )
}
```

### 框架无关使用

```typescript
// 简单的翻译函数
import translations from '@/locales/en.json'

function t(key: string, params?: Record<string, string>): string {
  const keys = key.split('.')
  let value: any = translations

  for (const k of keys) {
    value = value?.[k]
    if (value === undefined) return key
  }

  if (typeof value === 'string' && params) {
    return Object.entries(params).reduce(
      (str, [k, v]) => str.replace(`{${k}}`, v),
      value
    )
  }

  return value
}

// 使用
t('auth.login.title')  // "登录"
t('greeting', { name: 'John' })  // "Hello, John!"
```

## 10. 性能优化

### 大项目建议

对于翻译数量超过 1000 条的项目：

1. **按需加载**：只加载当前语言和当前页面需要的翻译
2. **缓存策略**：使用 Vue Query 或类似库缓存翻译数据
3. **增量同步**：只同步变更的翻译，而不是全量同步

### 示例：按模块加载

```json
// common.json
{
  "button": {
    "save": "保存",
    "cancel": "取消"
  }
}

// auth.json
{
  "login": {
    "title": "登录",
    "button": "登录"
  }
}
```

```typescript
// 动态加载模块
const modules = ['common', 'auth', 'dashboard', 'errors']

async function loadMessages(locale: string, module: string) {
  const data = await fetch(`/locales/${locale}/${module}.json`)
  return data.json()
}
```

## 相关文档

- [翻译管理](/guide/translation-guide) - 基本翻译操作
- [CLI 使用指南](/guide/cli-guide) - 命令行工具详细用法
- [团队协作](/guide/team-collaboration) - 团队角色和权限
- [API 概述](/api/overview) - API 端点参考
