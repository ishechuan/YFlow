# CLI 架构

了解 Bun CLI 工具的技术实现和架构设计。

## 技术栈

- **Bun 1.0+** - 运行时
- **TypeScript** - 开发语言
- **Commander** - CLI 框架
- **Axios** - HTTP 客户端

## 目录结构

```
cli/src/
├── commands/          # 命令实现
│   ├── init.ts        # 初始化命令
│   ├── sync.ts        # 同步命令
│   └── import.ts      # 导入命令
├── api.ts             # API 客户端
├── config.ts          # 配置管理
├── scanner.ts         # 文件扫描器
├── types.ts           # 类型定义
├── index.ts           # CLI 入口
└── language-mapping.ts # 语言映射
```

## 核心模块

### CLI 入口

```typescript
// index.ts
import { Command } from 'commander'
import { initCommand } from './commands/init'
import { syncCommand } from './commands/sync'
import { importCommand } from './commands/import'

const program = new Command()

program
  .name('yflow')
  .description('i18n management CLI tool')
  .version('1.0.0')

program.addCommand(initCommand)
program.addCommand(syncCommand)
program.addCommand(importCommand)

program.parse()
```

### 配置管理

```typescript
// config.ts
import fs from 'fs'
import path from 'path'

interface Config {
  messagesDir: string
  projectId: number
  apiUrl: string
  apiKey: string
}

const CONFIG_FILE = '.i18nrc.json'

export function loadConfig(): Config {
  const configPath = findConfigFile()
  if (!configPath) {
    throw new Error('Config file not found. Run "yflow init" first.')
  }
  const config = JSON.parse(fs.readFileSync(configPath, 'utf-8'))
  return validateConfig(config)
}

function findConfigFile(): string | null {
  let dir = process.cwd()
  while (dir !== '/') {
    const configPath = path.join(dir, CONFIG_FILE)
    if (fs.existsSync(configPath)) {
      return configPath
    }
    dir = path.dirname(dir)
  }
  return null
}
```

### API 客户端

```typescript
// api.ts
import axios from 'axios'
import { loadConfig } from './config'

const client = axios.create({
  timeout: 30000
})

client.interceptors.request.use((config) => {
  const configData = loadConfig()
  config.headers['X-API-Key'] = configData.apiKey
  return config
})

export async function getTranslations(projectId: number) {
  const { data } = await client.get('/cli/translations', {
    params: { project_id: projectId }
  })
  return data
}

export async function pushTranslations(
  projectId: number,
  translations: Record<string, Record<string, string>>
) {
  const { data } = await client.post('/cli/keys', {
    project_id: projectId,
    translations
  })
  return data
}
```

### 文件扫描器

```typescript
// scanner.ts
import fs from 'fs'
import path from 'path'
import { loadConfig } from './config'

export interface TranslationFile {
  language: string
  content: Record<string, string>
}

export function scanTranslationFiles(): TranslationFile[] {
  const config = loadConfig()
  const files = fs.readdirSync(config.messagesDir)

  return files
    .filter((file) => file.endsWith('.json'))
    .map((file) => {
      const language = file.replace('.json', '')
      const content = JSON.parse(
        fs.readFileSync(path.join(config.messagesDir, file), 'utf-8')
      )
      return { language, content }
    })
}

export function flattenTranslations(
  files: TranslationFile[]
): Record<string, Record<string, string>> {
  const result: Record<string, Record<string, string>> = {}

  for (const file of files) {
    for (const [key, value] of Object.entries(file.content)) {
      if (!result[key]) {
        result[key] = {}
      }
      result[key][file.language] = value
    }
  }

  return result
}
```

## 命令实现

### sync 命令

```typescript
// commands/sync.ts
import { Command } from 'commander'
import { loadConfig } from '../config'
import { getTranslations } from '../api'
import fs from 'fs'
import path from 'path'

export const syncCommand = new Command()
  .name('sync')
  .description('Sync translations from remote to local')
  .option('--lang <language>', 'Specific language to sync')
  .action(async (options) => {
    const config = loadConfig()
    console.log(`Syncing translations for project ${config.projectId}...`)

    const translations = await getTranslations(config.projectId)

    // Write to local files
    for (const [language, values] of Object.entries(translations)) {
      const filePath = path.join(config.messagesDir, `${language}.json`)
      fs.writeFileSync(filePath, JSON.stringify(values, null, 2))
      console.log(`  Updated ${language}.json`)
    }

    console.log('Sync completed!')
  })
```

### import 命令

```typescript
// commands/import.ts
import { Command } from 'commander'
import { loadConfig } from '../config'
import { pushTranslations } from '../api'
import { scanTranslationFiles, flattenTranslations } from '../scanner'

export const importCommand = new Command()
  .name('import')
  .description('Import local translations to remote')
  .action(async () => {
    const config = loadConfig()
    console.log('Scanning local translation files...')

    const files = scanTranslationFiles()
    const translations = flattenTranslations(files)

    console.log(`Found ${Object.keys(translations).length} translation keys`)

    await pushTranslations(config.projectId, translations)

    console.log('Import completed!')
  })
```

## 工作流程

### sync 命令流程

```
yflow sync
       │
       ▼
加载配置文件 (.i18nrc.json)
       │
       ▼
请求 /api/cli/translations
       │
       ▼
接收翻译数据
       │
       ▼
写入本地 JSON 文件
       │
       ▼
完成
```

### import 命令流程

```
yflow import
       │
       ▼
扫描 messagesDir 目录
       │
       ▼
读取所有 JSON 文件
       │
       ▼
扁平化翻译数据
       │
       ▼
POST 到 /api/cli/keys
       │
       ▼
完成
```

## 下一步

- [CLI 使用指南 →](/guide/cli-guide)
- [部署指南 →](/deployment/docker)
