import { defineConfig } from 'vitepress'

export default defineConfig({
  title: "i18n-flow",
  description: "a powerful and open source i18n solution for self hosted",
  themeConfig: {
    nav: [
      { text: '首页', link: '/' },
      {
        text: '指南',
        items: [
          { text: '快速开始', link: '/guide/quick-start' },
          { text: '使用指南', link: '/guide/getting-started' },
          { text: '最佳实践', link: '/guide/best-practices' },
          { text: 'CLI 使用', link: '/guide/cli-guide' },
          { text: '团队协作', link: '/guide/team-collaboration' }
        ]
      },
      {
        text: '参考',
        items: [
          { text: 'API 参考', link: '/api/overview' },
          { text: '架构说明', link: '/architecture/overview' }
        ]
      },
      { text: '部署', link: '/deployment/docker' }
    ],

    sidebar: {
      '/guide/': [
        {
          text: '使用指南',
          items: [
            { text: '快速开始', link: '/guide/quick-start' },
            { text: '入门介绍', link: '/guide/getting-started' },
            { text: '最佳实践', link: '/guide/best-practices' },
            { text: '项目管理', link: '/guide/project-guide' },
            { text: '翻译管理', link: '/guide/translation-guide' },
            { text: 'CLI 使用', link: '/guide/cli-guide' },
            { text: '团队协作', link: '/guide/team-collaboration' }
          ]
        }
      ],
      '/api/': [
        {
          text: 'API 参考',
          items: [
            { text: '概述', link: '/api/overview' },
            { text: '认证', link: '/api/authentication' },
            { text: '端点', link: '/api/endpoints' }
          ]
        }
      ],
      '/architecture/': [
        {
          text: '架构说明',
          items: [
            { text: '概述', link: '/architecture/overview' },
            { text: '后端架构', link: '/architecture/backend' },
            { text: '前端架构', link: '/architecture/frontend' },
            { text: 'CLI 架构', link: '/architecture/cli' }
          ]
        }
      ],
      '/deployment/': [
        {
          text: '部署指南',
          items: [
            { text: 'Docker 部署', link: '/deployment/docker' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/anthropics/claude-code' }
    ],

    search: {
      provider: 'local'
    },

    outline: 'deep'
  }
})
