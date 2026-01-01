/**
 * YFlow Landing Page - Translations
 */

var translations = {
  zh: {
    // Navigation
    logo_text: "语流",
    nav_features: "特性",
    nav_architecture: "架构",
    nav_cli: "CLI",
    nav_deploy: "部署",
    nav_roadmap: "路线图",
    nav_github: "GitHub",

    // Hero
    badge: "开源 • 自托管 • 现代化",
    hero_title_prefix: "您的",
    hero_title_i18n: "i18n",
    hero_title_suffix: "工作流",
    hero_title_line: "从此与众不同",
    hero_subtitle: "一站式自托管国际化解决方案。CLI 扫描推送、可视化编辑、团队协作、Docker 一键部署 —— 让全球化从未如此简单。",
    btn_deploy: "立即部署",
    btn_cli: "体验 CLI",
    copy_command: "git clone && docker-compose up -d",

    // CLI Demo
    cli_demo_title: "CLI Demo",
    cli_step1_command: "yflow init",
    cli_step1_output: "✓ 配置初始化完成",
    cli_step2_command: "yflow import",
    cli_step2_scanning: "扫描中",
    cli_step2_found: "📁 发现 12 个文件",
    cli_step2_keys: "🔑 解析 156 个 key",
    cli_step2_pushing: "推送中",
    cli_step2_complete: "✓ 上传完成",
    cli_step3_complete: "同步完成!",

    // UI Preview
    ui_title: "翻译管理",
    ui_badge: "中文 (简体)",
    ui_key: "Key",
    ui_english: "English",
    ui_chinese: "简体中文",
    ui_changes: "3 处更新",

    // Features
    section_label_features: "核心特性",
    section_title_features: "Everything you need for i18n",
    section_subtitle_features: "从代码到翻译到部署，全链路解决方案",

    feature_1_title: "多语言矩阵视图",
    feature_1_desc: "直观的表格界面，同时编辑多种语言。所见即所得，翻译进度一目了然。",

    feature_2_title: "CLI 自动化",
    feature_2_desc: "扫描本地文件、自动推送、CI/CD 集成。开发流程零负担。",

    feature_3_title: "团队协作",
    feature_3_desc: "邀请码机制、角色权限管理 (Owner/Editor/Viewer)、操作审计。",

    feature_4_title: "企业级安全",
    feature_4_desc: "JWT 双令牌、API Key 认证、SQL 注入防护、XSS 防护、请求限流。",

    feature_5_title: "Redis 缓存加速",
    feature_5_desc: "高频 API 缓存、分布式支持。响应速度毫秒级。",

    feature_6_title: "完整 Admin UI",
    feature_6_desc: "Vue 3 + Element Plus 管理后台。项目、用户、翻译、邀请码一站式管理。",

    // Architecture
    section_label_arch: "技术架构",
    section_title_arch: "现代技术栈",
    section_subtitle_arch: "每一层都是业界最佳实践",

    arch_backend: "Admin Backend",
    arch_backend_badge: "Go + Gin",
    arch_backend_features: [
      "RESTful API (Swagger 文档)",
      "GORM + MySQL 8.0",
      "Redis 7.2 缓存层",
      "Uber FX 依赖注入",
      "Clean Architecture"
    ],
    arch_backend_stats: ["15+ API 模块", "6 层安全中间件"],

    arch_frontend: "Admin Frontend",
    arch_frontend_badge: "Vue 3 + TS",
    arch_frontend_features: [
      "Composition API",
      "Pinia 状态管理",
      "TanStack Vue Query",
      "Element Plus UI",
      "JWT 认证流程"
    ],
    arch_frontend_stats: ["8+ 功能页面", "RBAC 权限控制"],

    arch_cli: "CLI Tool",
    arch_cli_badge: "Bun + TS",
    arch_cli_features: [
      "Commander 命令行",
      "自动扫描 JSON 文件",
      "扁平化/结构化转换",
      "API Key 认证",
      "CI/CD 友好"
    ],
    arch_cli_stats: ["3 核心命令", "秒级同步"],

    arch_docs: "Documentation",
    arch_docs_badge: "VitePress",
    arch_docs_features: [
      "快速响应的文档站",
      "API 参考文档",
      "部署指南",
      "最佳实践指南",
      "团队协作文档"
    ],
    arch_docs_stats: ["5+ 文档模块", "持续更新"],

    // Architecture Cards
    arch_backend: "Admin Backend",
    arch_backend_badge: "Go + Gin",
    arch_backend_f1: "RESTful API (Swagger 文档)",
    arch_backend_f2: "GORM + MySQL 8.0",
    arch_backend_f3: "Redis 7.2 缓存层",
    arch_backend_f4: "Uber FX 依赖注入",
    arch_backend_f5: "Clean Architecture",
    arch_backend_s1: "15+ API 模块",
    arch_backend_s2: "6 层安全中间件",

    arch_frontend: "Admin Frontend",
    arch_frontend_badge: "Vue 3 + TS",
    arch_frontend_f1: "Composition API",
    arch_frontend_f2: "Pinia 状态管理",
    arch_frontend_f3: "TanStack Vue Query",
    arch_frontend_f4: "Element Plus UI",
    arch_frontend_f5: "JWT 认证流程",
    arch_frontend_s1: "8+ 功能页面",
    arch_frontend_s2: "RBAC 权限控制",

    arch_cli: "CLI Tool",
    arch_cli_badge: "Bun + TS",
    arch_cli_f1: "Commander 命令行",
    arch_cli_f2: "自动扫描 JSON 文件",
    arch_cli_f3: "扁平化/结构化转换",
    arch_cli_f4: "API Key 认证",
    arch_cli_f5: "CI/CD 友好",
    arch_cli_s1: "3 核心命令",
    arch_cli_s2: "秒级同步",

    arch_docs: "Documentation",
    arch_docs_badge: "VitePress",
    arch_docs_f1: "快速响应的文档站",
    arch_docs_f2: "API 参考文档",
    arch_docs_f3: "部署指南",
    arch_docs_f4: "最佳实践指南",
    arch_docs_f5: "团队协作文档",
    arch_docs_s1: "5+ 文档模块",
    arch_docs_s2: "持续更新",

    // CLI Section
    section_label_cli: "CLI Workflow",
    section_title_cli: "从代码到云端，<br>只需三步",

    step_1_number: "01",
    step_1_title: "yflow init",
    step_1_desc: "初始化配置文件，设置项目 ID、API 地址、语言映射。",
    step_1_code: ".i18nrc.json",

    step_2_number: "02",
    step_2_title: "yflow import",
    step_2_desc: "扫描 messagesDir 目录，自动解析 JSON 文件中的翻译 key，推送到服务器。",

    step_3_number: "03",
    step_3_title: "yflow sync",
    step_3_desc: "从服务器拉取最新翻译，保持本地文件与云端同步。",

    terminal_title: "Terminal",
    terminal_init_success: "✓ 配置文件已生成: .i18nrc.json",
    terminal_import_scan: "🔍 扫描 /src/locales...",
    terminal_import_files: "📁 解析 3 个文件",
    terminal_import_keys: "🔑 发现 156 个翻译键",
    terminal_import_pushing: "📤 推送中...",
    terminal_sync_complete: "✓ 同步完成 (耗时 1.2s)",

    // Docker Section
    docker_title: "一键部署，就是这么简单",
    docker_subtitle: "不需要懂 Docker，不需要配置环境。<br>一行命令，全部搞定。",

    step_1_clone: "Step 1: 克隆项目",
    step_1_command: "git clone https://github.com/cerebralatlas/yflow.git",

    step_2_start: "Step 2: 一键启动",
    step_2_magic: "✨ 魔法时刻",
    step_2_command: "docker-compose up -d",
    step_2_note: "自动启动 MySQL、Redis、Backend、Frontend、Docs",

    step_3_use: "Step 3: 开始使用",
    step_3_command: "访问 http://localhost:80",

    services_title: "自动编排的服务",
    service_mysql: "MySQL 8.0",
    service_redis: "Redis 7.2",
    service_backend: "Go Backend",
    service_frontend: "Vue Frontend",
    service_docs: "VitePress Docs",
    service_translation: "LibreTranslate",

    // Roadmap
    section_label_roadmap: "产品路线图",
    section_title_roadmap: "我们的规划",
    section_subtitle_roadmap: "承认不足，持续改进",

    timeline_done: "已完成",
    timeline_v1_title: "v1.0 基础功能",
    timeline_v1_items: [
      "✓ RESTful API 后端 (Go + Gin)",
      "✓ 管理后台 (Vue 3 + Element Plus)",
      "✓ CLI 工具 (Bun + TypeScript)",
      "✓ Docker 部署支持",
      "✓ 基础翻译管理",
      "✓ 机器翻译 (LibreTranslate)"
    ],

    timeline_q1: "Q1 2025",
    timeline_v11_title: "v1.1 格式扩展",
    timeline_v11_items: [
      "🔄 YAML 格式支持",
      "🔄 Gettext (.po/.mo) 支持",
      "🔄 CSV 批量导入导出",
      "🔄 嵌套 key 扁平化优化"
    ],

    timeline_q2: "Q2 2025",
    timeline_v12_title: "v1.2 协作增强",
    timeline_v12_items: [
      "🔜 翻译审核工作流",
      "🔜 翻译记忆库",
      "🔜 团队活动日志"
    ],

    timeline_q3: "Q3 2025",
    timeline_v20_title: "v2.0 生态完善",
    timeline_v1_i1: "✓ RESTful API 后端 (Go + Gin)",
    timeline_v1_i2: "✓ 管理后台 (Vue 3 + Element Plus)",
    timeline_v1_i3: "✓ CLI 工具 (Bun + TypeScript)",
    timeline_v1_i4: "✓ Docker 部署支持",
    timeline_v1_i5: "✓ 基础翻译管理",
    timeline_v11_i1: "🔄 YAML 格式支持",
    timeline_v11_i2: "🔄 Gettext (.po/.mo) 支持",
    timeline_v11_i3: "🔄 CSV 批量导入导出",
    timeline_v11_i4: "🔄 嵌套 key 扁平化优化",
    timeline_v12_i1: "🔜 翻译审核工作流",
    timeline_v12_i3: "🔜 翻译记忆库",
    timeline_v12_i4: "🔜 团队活动日志",
    timeline_v20_i1: "🔜 Webhook 集成",
    timeline_v20_i2: "🔜 VS Code 插件",
    timeline_v20_i3: "🔜 GitHub Action",
    timeline_v20_i4: "🔜 插件系统",
    timeline_v20_items: [
      "🔜 Webhook 集成",
      "🔜 VS Code 插件",
      "🔜 GitHub Action",
      "🔜 插件系统"
    ],

    // Footer
    footer_desc: "强大的自托管 i18n 解决方案",
    footer_product: "产品",
    footer_resources: "资源",
    footer_community: "社区",
    footer_docs: "文档",
    footer_api: "API",
    footer_feedback: "反馈",
    footer_contribute: "贡献",
    footer_license: "License",
    footer_mit: "MIT License • 开源免费",
    footer_love: "Built with ❤️ by Developers, for Developers"
  },
  en: {
    // Navigation
    logo_text: "YFlow",
    nav_features: "Features",
    nav_architecture: "Architecture",
    nav_cli: "CLI",
    nav_deploy: "Deploy",
    nav_roadmap: "Roadmap",
    nav_github: "GitHub",

    // Hero
    badge: "Open Source • Self-Hosted • Modern",
    hero_title_prefix: "Your",
    hero_title_i18n: "i18n",
    hero_title_suffix: "workflow",
    hero_title_line: "reimagined",
    hero_subtitle: "A complete self-hosted internationalization solution. CLI scanning & pushing, visual editing, team collaboration, Docker one-click deployment — globalization has never been easier.",
    btn_deploy: "Deploy Now",
    btn_cli: "Try CLI",
    copy_command: "git clone && docker-compose up -d",

    // CLI Demo
    cli_demo_title: "CLI Demo",
    cli_step1_command: "yflow init",
    cli_step1_output: "✓ Configuration initialized",
    cli_step2_command: "yflow import",
    cli_step2_scanning: "Scanning",
    cli_step2_found: "📁 Found 12 files",
    cli_step2_keys: "🔑 Parsed 156 keys",
    cli_step2_pushing: "Pushing",
    cli_step2_complete: "✓ Upload complete",
    cli_step3_complete: "Sync complete!",

    // UI Preview
    ui_title: "Translation Management",
    ui_badge: "English",
    ui_key: "Key",
    ui_english: "English",
    ui_chinese: "Chinese (Simplified)",
    ui_changes: "3 updates",

    // Features
    section_label_features: "Core Features",
    section_title_features: "Everything you need for i18n",
    section_subtitle_features: "End-to-end solution from code to translation to deployment",

    feature_1_title: "Multi-Language Matrix View",
    feature_1_desc: "Intuitive spreadsheet interface for editing multiple languages at once. WYSIWYG with clear translation progress.",

    feature_2_title: "CLI Automation",
    feature_2_desc: "Scan local files, auto-push, CI/CD integration. Zero-friction development workflow.",

    feature_3_title: "Team Collaboration",
    feature_3_desc: "Invitation codes, role-based permissions (Owner/Editor/Viewer), audit logs.",

    feature_4_title: "Enterprise Security",
    feature_4_desc: "JWT dual tokens, API Key authentication, SQL injection protection, XSS protection, rate limiting.",

    feature_5_title: "Redis Caching",
    feature_5_desc: "High-frequency API caching with distributed support. Millisecond-level response times.",

    feature_6_title: "Complete Admin UI",
    feature_6_desc: "Vue 3 + Element Plus admin dashboard. Projects, users, translations, invitations — all in one place.",

    // Architecture
    section_label_arch: "Tech Stack",
    section_title_arch: "Modern Architecture",
    section_subtitle_arch: "Best practices at every layer",

    arch_backend: "Admin Backend",
    arch_backend_badge: "Go + Gin",
    arch_backend_features: [
      "RESTful API (Swagger docs)",
      "GORM + MySQL 8.0",
      "Redis 7.2 caching layer",
      "Uber FX dependency injection",
      "Clean Architecture"
    ],
    arch_backend_stats: ["15+ API modules", "6 security middleware layers"],

    arch_frontend: "Admin Frontend",
    arch_frontend_badge: "Vue 3 + TS",
    arch_frontend_features: [
      "Composition API",
      "Pinia state management",
      "TanStack Vue Query",
      "Element Plus UI",
      "JWT authentication flow"
    ],
    arch_frontend_stats: ["8+ feature pages", "RBAC access control"],

    arch_cli: "CLI Tool",
    arch_cli_badge: "Bun + TS",
    arch_cli_features: [
      "Commander CLI framework",
      "Auto-scan JSON files",
      "Flat/structured conversion",
      "API Key authentication",
      "CI/CD friendly"
    ],
    arch_cli_stats: ["3 core commands", "Second-level sync"],

    arch_docs: "Documentation",
    arch_docs_badge: "VitePress",
    arch_docs_features: [
      "Fast-loading docs site",
      "API reference docs",
      "Deployment guide",
      "Best practices guide",
      "Team collaboration docs"
    ],
    arch_docs_stats: ["5+ doc modules", "Continuously updated"],

    // Architecture Cards
    arch_backend: "Admin Backend",
    arch_backend_badge: "Go + Gin",
    arch_backend_f1: "RESTful API (Swagger docs)",
    arch_backend_f2: "GORM + MySQL 8.0",
    arch_backend_f3: "Redis 7.2 caching layer",
    arch_backend_f4: "Uber FX dependency injection",
    arch_backend_f5: "Clean Architecture",
    arch_backend_s1: "15+ API modules",
    arch_backend_s2: "6 security middleware layers",

    arch_frontend: "Admin Frontend",
    arch_frontend_badge: "Vue 3 + TS",
    arch_frontend_f1: "Composition API",
    arch_frontend_f2: "Pinia state management",
    arch_frontend_f3: "TanStack Vue Query",
    arch_frontend_f4: "Element Plus UI",
    arch_frontend_f5: "JWT authentication flow",
    arch_frontend_s1: "8+ feature pages",
    arch_frontend_s2: "RBAC access control",

    arch_cli: "CLI Tool",
    arch_cli_badge: "Bun + TS",
    arch_cli_f1: "Commander CLI framework",
    arch_cli_f2: "Auto-scan JSON files",
    arch_cli_f3: "Flat/structured conversion",
    arch_cli_f4: "API Key authentication",
    arch_cli_f5: "CI/CD friendly",
    arch_cli_s1: "3 core commands",
    arch_cli_s2: "Second-level sync",

    arch_docs: "Documentation",
    arch_docs_badge: "VitePress",
    arch_docs_f1: "Fast-loading docs site",
    arch_docs_f2: "API reference docs",
    arch_docs_f3: "Deployment guide",
    arch_docs_f4: "Best practices guide",
    arch_docs_f5: "Team collaboration docs",
    arch_docs_s1: "5+ doc modules",
    arch_docs_s2: "Continuously updated",

    // CLI Section
    section_label_cli: "CLI Workflow",
    section_title_cli: "From code to cloud,<br>in just three steps",

    step_1_number: "01",
    step_1_title: "yflow init",
    step_1_desc: "Initialize configuration file, set project ID, API address, and language mappings.",
    step_1_code: ".i18nrc.json",

    step_2_number: "02",
    step_2_title: "yflow import",
    step_2_desc: "Scan messagesDir directory, automatically parse translation keys from JSON files, and push to server.",

    step_3_number: "03",
    step_3_title: "yflow sync",
    step_3_desc: "Pull latest translations from server, keep local files in sync with cloud.",

    terminal_title: "Terminal",
    terminal_init_success: "✓ Config file generated: .i18nrc.json",
    terminal_import_scan: "🔍 Scanning /src/locales...",
    terminal_import_files: "📁 Parsed 3 files",
    terminal_import_keys: "🔑 Found 156 translation keys",
    terminal_import_pushing: "📤 Pushing...",
    terminal_sync_complete: "✓ Sync complete (1.2s)",

    // Docker Section
    docker_title: "One-click deployment, made simple",
    docker_subtitle: "No Docker knowledge needed, no environment configuration.<br>One command, and you're ready.",

    step_1_clone: "Step 1: Clone repository",
    step_1_command: "git clone https://github.com/cerebralatlas/yflow.git",

    step_2_start: "Step 2: Start services",
    step_2_magic: "✨ Magic moment",
    step_2_command: "docker-compose up -d",
    step_2_note: "Auto-starts MySQL, Redis, Backend, Frontend, Docs",

    step_3_use: "Step 3: Start using",
    step_3_command: "Visit http://localhost:80",

    services_title: "Auto-orchestrated services",
    service_mysql: "MySQL 8.0",
    service_redis: "Redis 7.2",
    service_backend: "Go Backend",
    service_frontend: "Vue Frontend",
    service_docs: "VitePress Docs",
    service_translation: "LibreTranslate",

    // Roadmap
    section_label_roadmap: "Product Roadmap",
    section_title_roadmap: "Our Plans",
    section_subtitle_roadmap: "Acknowledging gaps, continuous improvement",

    timeline_done: "Completed",
    timeline_v1_title: "v1.0 Core Features",
    timeline_v1_items: [
      "✓ RESTful API backend (Go + Gin)",
      "✓ Admin dashboard (Vue 3 + Element Plus)",
      "✓ CLI tool (Bun + TypeScript)",
      "✓ Docker deployment support",
      "✓ Basic translation management",
      "✓ Machine Translation (LibreTranslate)"
    ],

    timeline_q1: "Q1 2025",
    timeline_v11_title: "v1.1 Format Extensions",
    timeline_v11_items: [
      "🔄 YAML format support",
      "🔄 Gettext (.po/.mo) support",
      "🔄 CSV import/export",
      "🔄 Nested key flattening"
    ],

    timeline_q2: "Q2 2025",
    timeline_v12_title: "v1.2 Collaboration Enhancements",
    timeline_v12_items: [
      "🔜 Translation review workflow",
      "🔜 Translation memory",
      "🔜 Team activity logs"
    ],

    timeline_q3: "Q3 2025",
    timeline_v20_title: "v2.0 Ecosystem",
    timeline_v1_i1: "✓ RESTful API backend (Go + Gin)",
    timeline_v1_i2: "✓ Admin dashboard (Vue 3 + Element Plus)",
    timeline_v1_i3: "✓ CLI tool (Bun + TypeScript)",
    timeline_v1_i4: "✓ Docker deployment support",
    timeline_v1_i5: "✓ Basic translation management",
    timeline_v11_i1: "🔄 YAML format support",
    timeline_v11_i2: "🔄 Gettext (.po/.mo) support",
    timeline_v11_i3: "🔄 CSV import/export",
    timeline_v11_i4: "🔄 Nested key flattening",
    timeline_v12_i1: "🔜 Translation review workflow",
    timeline_v12_i3: "🔜 Translation memory",
    timeline_v12_i4: "🔜 Team activity logs",
    timeline_v20_i1: "🔜 Webhook integration",
    timeline_v20_i2: "🔜 VS Code extension",
    timeline_v20_i3: "🔜 GitHub Action",
    timeline_v20_i4: "🔜 Plugin system",
    timeline_v20_items: [
      "🔜 Webhook integration",
      "🔜 VS Code extension",
      "🔜 GitHub Action",
      "🔜 Plugin system"
    ],

    // Footer
    footer_desc: "Powerful self-hosted i18n solution",
    footer_product: "Product",
    footer_resources: "Resources",
    footer_community: "Community",
    footer_docs: "Docs",
    footer_api: "API",
    footer_feedback: "Feedback",
    footer_contribute: "Contribute",
    footer_license: "License",
    footer_mit: "MIT License • Open Source",
    footer_love: "Built with ❤️ by Developers, for Developers"
  }
};

// Make translations available globally
if (typeof window !== 'undefined') {
  window.translations = translations;
  window.translatableElements = {
  // Navigation
  '[data-i18n="nav_features"]': 'nav_features',
  '[data-i18n="nav_architecture"]': 'nav_architecture',
  '[data-i18n="nav_cli"]': 'nav_cli',
  '[data-i18n="nav_deploy"]': 'nav_deploy',
  '[data-i18n="nav_roadmap"]': 'nav_roadmap',
  '[data-i18n="nav_github"]': 'nav_github',

  // Hero
  '[data-i18n="badge"]': 'badge',
  '[data-i18n="hero_title_prefix"]': 'hero_title_prefix',
  '[data-i18n="hero_title_i18n"]': 'hero_title_i18n',
  '[data-i18n="hero_title_suffix"]': 'hero_title_suffix',
  '[data-i18n="hero_title_line"]': 'hero_title_line',
  '[data-i18n="hero_subtitle"]': 'hero_subtitle',
  '[data-i18n="btn_deploy"]': 'btn_deploy',
  '[data-i18n="btn_cli"]': 'btn_cli',

  // CLI Demo
  '[data-i18n="cli_demo_title"]': 'cli_demo_title',
  '[data-i18n="cli_step1_command"]': 'cli_step1_command',
  '[data-i18n="cli_step1_output"]': 'cli_step1_output',
  '[data-i18n="cli_step2_command"]': 'cli_step2_command',
  '[data-i18n="cli_step2_scanning"]': 'cli_step2_scanning',
  '[data-i18n="cli_step2_found"]': 'cli_step2_found',
  '[data-i18n="cli_step2_keys"]': 'cli_step2_keys',
  '[data-i18n="cli_step2_pushing"]': 'cli_step2_pushing',
  '[data-i18n="cli_step2_complete"]': 'cli_step2_complete',
  '[data-i18n="cli_step3_complete"]': 'cli_step3_complete',

  // UI Preview
  '[data-i18n="ui_title"]': 'ui_title',
  '[data-i18n="ui_badge"]': 'ui_badge',
  '[data-i18n="ui_key"]': 'ui_key',
  '[data-i18n="ui_english"]': 'ui_english',
  '[data-i18n="ui_chinese"]': 'ui_chinese',
  '[data-i18n="ui_changes"]': 'ui_changes',

  // Features
  '[data-i18n="section_label_features"]': 'section_label_features',
  '[data-i18n="section_title_features"]': 'section_title_features',
  '[data-i18n="section_subtitle_features"]': 'section_subtitle_features',
  '[data-i18n="feature_1_title"]': 'feature_1_title',
  '[data-i18n="feature_1_desc"]': 'feature_1_desc',
  '[data-i18n="feature_2_title"]': 'feature_2_title',
  '[data-i18n="feature_2_desc"]': 'feature_2_desc',
  '[data-i18n="feature_3_title"]': 'feature_3_title',
  '[data-i18n="feature_3_desc"]': 'feature_3_desc',
  '[data-i18n="feature_4_title"]': 'feature_4_title',
  '[data-i18n="feature_4_desc"]': 'feature_4_desc',
  '[data-i18n="feature_5_title"]': 'feature_5_title',
  '[data-i18n="feature_5_desc"]': 'feature_5_desc',
  '[data-i18n="feature_6_title"]': 'feature_6_title',
  '[data-i18n="feature_6_desc"]': 'feature_6_desc',

  // Architecture
  '[data-i18n="section_label_arch"]': 'section_label_arch',
  '[data-i18n="section_title_arch"]': 'section_title_arch',
  '[data-i18n="section_subtitle_arch"]': 'section_subtitle_arch',

  // CLI Section
  '[data-i18n="section_label_cli"]': 'section_label_cli',
  '[data-i18n="section_title_cli"]': 'section_title_cli',
  '[data-i18n="step_1_number"]': 'step_1_number',
  '[data-i18n="step_1_title"]': 'step_1_title',
  '[data-i18n="step_1_desc"]': 'step_1_desc',
  '[data-i18n="step_1_code"]': 'step_1_code',
  '[data-i18n="step_2_number"]': 'step_2_number',
  '[data-i18n="step_2_title"]': 'step_2_title',
  '[data-i18n="step_2_desc"]': 'step_2_desc',
  '[data-i18n="step_3_number"]': 'step_3_number',
  '[data-i18n="step_3_title"]': 'step_3_title',
  '[data-i18n="step_3_desc"]': 'step_3_desc',
  '[data-i18n="terminal_title"]': 'terminal_title',

  // Docker Section
  '[data-i18n="docker_title"]': 'docker_title',
  '[data-i18n="docker_subtitle"]': 'docker_subtitle',
  '[data-i18n="step_1_clone"]': 'step_1_clone',
  '[data-i18n="step_2_start"]': 'step_2_start',
  '[data-i18n="step_2_magic"]': 'step_2_magic',
  '[data-i18n="step_2_note"]': 'step_2_note',
  '[data-i18n="step_3_use"]': 'step_3_use',
  '[data-i18n="services_title"]': 'services_title',
  '[data-i18n="service_translation"]': 'service_translation',

  // Roadmap
  '[data-i18n="section_label_roadmap"]': 'section_label_roadmap',
  '[data-i18n="section_title_roadmap"]': 'section_title_roadmap',
  '[data-i18n="section_subtitle_roadmap"]': 'section_subtitle_roadmap',
  '[data-i18n="timeline_done"]': 'timeline_done',
  '[data-i18n="timeline_q1"]': 'timeline_q1',
  '[data-i18n="timeline_q2"]': 'timeline_q2',
  '[data-i18n="timeline_q3"]': 'timeline_q3',

  // Footer
  '[data-i18n="footer_desc"]': 'footer_desc',
  '[data-i18n="footer_product"]': 'footer_product',
  '[data-i18n="footer_resources"]': 'footer_resources',
  '[data-i18n="footer_community"]': 'footer_community',
  '[data-i18n="footer_docs"]': 'footer_docs',
  '[data-i18n="footer_api"]': 'footer_api',
  '[data-i18n="footer_feedback"]': 'footer_feedback',
  '[data-i18n="footer_contribute"]': 'footer_contribute',
  '[data-i18n="footer_license"]': 'footer_license',
  '[data-i18n="footer_mit"]': 'footer_mit',
  '[data-i18n="footer_love"]': 'footer_love'
  };
}
