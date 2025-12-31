# 翻译管理

学习如何在 i18n-flow 中管理和编辑翻译。

## 翻译矩阵

翻译矩阵以表格形式展示所有翻译键在不同语言下的值。

### 访问翻译矩阵

1. 进入「翻译」页面
2. 选择需要管理的项目
3. 查看翻译矩阵

### 矩阵结构

| Key | en (默认) | zh-CN | ja-JP |
|-----|-----------|-------|-------|
| greeting | Hello | 你好 | こんにちは |
| user.name | Name | 姓名 | 名前 |

## 添加翻译

### 方式一：手动添加

1. 在翻译矩阵底部点击「添加键」
2. 输入翻译键名称（如 `user.login.button`）
3. 为每种语言输入翻译值

### 方式二：批量导入

支持 JSON 和 CSV 格式导入：

```json
// translations.json
{
  "greeting": "Hello",
  "user.name": "Name",
  "user.email": "Email"
}
```

导入步骤：
1. 点击「导入」按钮
2. 选择导入文件
3. 选择目标语言
4. 确认导入

### 方式三：使用 CLI

```bash
# 扫描本地文件并导入
i18n-flow import
```

## 编辑翻译

1. 在矩阵中直接点击对应单元格
2. 编辑翻译值
3. 自动保存

## 导出翻译

### 导出为 JSON

1. 点击「导出」按钮
2. 选择导出格式（JSON）
3. 选择要导出的语言
4. 下载文件

### 导出为 CSV

适用于 Excel 打开或翻译服务处理：

```csv
key,en,zh-CN,ja-JP
greeting,Hello,你好,こんにちは
user.name,Name,姓名,名前
```

## 翻译最佳实践

### 命名规范

推荐使用点分隔的层级命名：

```
// 推荐
user.login.title
user.login.button.submit
common.error.network

// 不推荐
userLoginTitle
USER_LOGIN_TITLE
```

### 描述说明

为重要翻译键添加描述：

```json
{
  "key": "user.login.button.submit",
  "description": "登录表单的提交按钮文字"
}
```

### 变量占位符

支持在翻译中使用变量：

```json
{
  "greeting": "Hello, {name}!",
  "items_count": "{count} items"
}
```

## 下一步

- [翻译管理最佳实践 →](/guide/best-practices)
- [使用 CLI 同步翻译 →](/guide/cli-guide)
- [团队协作翻译 →](/guide/team-collaboration)
