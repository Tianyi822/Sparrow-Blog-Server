# 麻雀博客后端系统 (Sparrow-Server)

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-GPL--3.0-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](#)

## 🔥 最新更新

### 🎆 新增功能
- **智能启动体验**: 首次运行自动检测和配置生成
- **环境变量支持**: 支持通过 `SPARROW_BLOG_HOME` 自定义数据目录
- **友好用户界面**: 首次运行显示清晰的配置指导
- **默认配置优化**: 自动生成包含合理默认值的配置文件

## 🛠 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.23+ | 后端开发语言 |
| Gin | 1.10+ | Web 框架 |
| GORM | 1.26+ | ORM 框架 |
| MySQL | 8.0+ | 主数据库 |
| Bleve | 2.5+ | 全文搜索引擎 |
| JWT | 5.2+ | 身份认证 |
| Zap | 1.27+ | 日志框架 |
| Lumberjack | 2.2+ | 日志轮转 |
| Gomail | 2.0+ | 邮件发送 |
