#!/bin/bash

# =============================================================================
# YFlow 一键部署脚本
# =============================================================================
# 支持的功能：
#   - 交互式部署
#   - 停止/重启服务
#   - 查看状态和日志
#   - 重置数据
# =============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

# 打印横幅
print_banner() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                                                                ║${NC}"
    echo -e "${BLUE}║${NC}   ${GREEN}  ██████╗ ██████╗ ███████╗███╗   ██╗ █████╗  ██████╗██╗  ██╗${NC}   ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}   ${GREEN} ██╔════╝██╔══██╗██╔════╝████╗  ██║██╔══██╗██╔════╝██║ ██╔╝${NC}   ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}   ${GREEN} ██║     ██████╔╝█████╗  ██╔██╗ ██║███████║██║     █████╔╝ ${NC}   ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}   ${GREEN} ██║     ██╔══██╗██╔══╝  ██║╚██╗██║██╔══██║██║     ██╔═██╗ ${NC}   ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}   ${GREEN} ╚██████╗██║  ██║███████╗██║ ╚████║██║  ██║╚██████╗██║  ██╗${NC}   ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}   ${GREEN}  ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═══╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝${NC}   ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}                                                                ║${NC}"
    echo -e "${BLUE}║${NC}   ${YELLOW}         国际化管理平台 - 一键部署脚本 v1.0${NC}                   ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}                                                                ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# 检查依赖
check_dependencies() {
    log_step "检查系统依赖..."

    local missing_deps=()

    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        missing_deps+=("docker")
    fi

    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        missing_deps+=("docker-compose")
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "缺少必要依赖: ${missing_deps[*]}"
        echo ""
        echo "请安装以下依赖后重新运行:"
        echo "  - Docker: https://docs.docker.com/get-docker/"
        echo "  - Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi

    log_info "✅ Docker 和 Docker Compose 已安装"
}

# 检查 Docker 服务状态
check_docker_service() {
    log_step "检查 Docker 服务状态..."

    if ! docker info &> /dev/null; then
        log_error "Docker 服务未运行，请启动 Docker 后重试"
        exit 1
    fi

    # 检查并修复 Docker 凭据配置（常见于从 Windows 迁移到 Linux）
    local docker_config="$HOME/.docker/config.json"
    if [ -f "$docker_config" ]; then
        if grep -q 'desktop.exe' "$docker_config" 2>/dev/null; then
            log_warn "检测到不兼容的 Docker 凭据配置，正在修复..."
            mkdir -p "$HOME/.docker"
            cat > "$docker_config" << 'DOCKEREOF'
{
	"auths": {},
	"credsStore": ""
}
DOCKEREOF
            log_info "已修复 Docker 凭据配置"
        fi
    fi

    log_info "✅ Docker 服务运行正常"
}

# 生成随机密码
generate_password() {
    local length=${1:-32}
    openssl rand -base64 "$length" | head -c "$length" | tr -dc 'a-zA-Z0-9' | head -c "$length"
}

# 交互式配置
interactive_config() {
    print_banner
    echo -e "${YELLOW}欢迎使用 YFlow 一键部署脚本${NC}"
    echo ""
    echo "此脚本将引导您完成 YFlow 的部署配置。"
    echo ""

    # 检查是否已存在 .env 文件
    if [ -f "$PROJECT_DIR/.env" ]; then
        log_warn "检测到已存在的 .env 配置文件"
        read -p "是否重新配置? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "使用现有配置"
            return 0
        fi
    fi

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}                    配置步骤${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    # 数据库配置
    echo -e "${CYAN}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${CYAN}│  1. 数据库配置                                              │${NC}"
    echo -e "${CYAN}└─────────────────────────────────────────────────────────────┘${NC}"
    echo ""

    local db_password=""
    while true; do
        read -s -p "     MySQL Root 密码 (至少8位): " db_password
        echo ""
        if [ ${#db_password} -ge 8 ]; then
            read -s -p "     确认密码: " db_password_confirm
            echo ""
            if [ "$db_password" = "$db_password_confirm" ]; then
                break
            else
                log_error "两次输入的密码不一致，请重新输入"
            fi
        else
            log_error "密码长度至少8位"
        fi
    done

    # 管理员账户配置
    echo ""
    echo -e "${CYAN}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${CYAN}│  2. 管理员账户配置                                          │${NC}"
    echo -e "${CYAN}└─────────────────────────────────────────────────────────────┘${NC}"
    echo ""

    read -p "     用户名 [admin]: " admin_username
    admin_username=${admin_username:-admin}

    read -s -p "     密码 [自动生成]: " admin_password
    echo ""
    if [ -z "$admin_password" ]; then
        admin_password=$(generate_password 12)
        log_info "已生成管理员密码: $admin_password"
    fi

    # 服务端口配置
    echo ""
    echo -e "${CYAN}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${CYAN}│  3. 服务端口配置                                            │${NC}"
    echo -e "${CYAN}└─────────────────────────────────────────────────────────────┘${NC}"
    echo ""

    read -p "     前端端口 [8081]: " frontend_port
    frontend_port=${frontend_port:-8081}

    # 机器翻译配置
    echo ""
    echo -e "${CYAN}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${CYAN}│  4. 机器翻译配置 (LibreTranslate)                           │${NC}"
    echo -e "${CYAN}└─────────────────────────────────────────────────────────────┘${NC}"
    echo ""

    log_warn "LibreTranslate 机器翻译服务需要约 1GB 内存"
    read -p "     是否启动机器翻译服务? [Y/n]: " -n 1 -r
    echo ""
    enable_mt=${REPLY:-y}
    if [[ ! $enable_mt =~ ^[Yy]$ ]]; then
        enable_mt="no"
    else
        enable_mt="yes"
    fi

    # 域名配置
    echo ""
    echo -e "${CYAN}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${CYAN}│  5. 域名与 HTTPS 配置                                       │${NC}"
    echo -e "${CYAN}└─────────────────────────────────────────────────────────────┘${NC}"
    echo ""

    echo -e "     ${YELLOW}提示:${NC} 使用 Let's Encrypt 需要："
    echo "       - 有效的域名（DNS 已解析到服务器 IP）"
    echo "       - 开放 80 和 443 端口"
    echo ""
    read -p "     域名 (直接回车使用 localhost): " domain
    domain=${domain:-localhost}

    # 生成 .env 文件
    log_step "生成配置文件..."

    local jwt_secret=$(openssl rand -base64 64 | tr -dc 'a-zA-Z0-9' | head -c 64)
    local jwt_refresh_secret=$(openssl rand -base64 64 | tr -dc 'a-zA-Z0-9' | head -c 64)
    local cli_api_key=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | head -c 32)

    cat > "$PROJECT_DIR/.env" << EOF
# =============================================================================
# YFlow 配置文件
# =============================================================================
# 此文件由 deploy.sh 自动生成
# 生成时间: $(date '+%Y-%m-%d %H:%M:%S')
# =============================================================================

# Database Configuration
DB_DRIVER=mysql
DB_ROOT_PASSWORD=$db_password
DB_USERNAME=root
DB_PASSWORD=$db_password
DB_HOST=db
DB_PORT=3306
DB_NAME=yflow

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_PREFIX=yflow:

# JWT Configuration
# IMPORTANT: 这些密钥已自动生成，请妥善保管
JWT_SECRET=$jwt_secret
JWT_EXPIRATION_HOURS=24
JWT_REFRESH_SECRET=$jwt_refresh_secret
JWT_REFRESH_EXPIRATION_HOURS=168

# CLI API Key Configuration
# IMPORTANT: 此密钥已自动生成，请妥善保管
CLI_API_KEY=$cli_api_key

# Admin User Configuration
ADMIN_USERNAME=$admin_username
ADMIN_PASSWORD=$admin_password

# Application Environment
ENV=production
GO_ENV=production

# Frontend Configuration
VITE_API_URL=/api

# LibreTranslate Machine Translation Configuration
ENABLE_LIBRETRANSLATE=$enable_mt
LIBRE_TRANSLATE_URL=http://libretranslate:5000

# Domain Configuration
DOMAIN=$domain
EOF

    log_info "✅ 配置文件 .env 已生成"
}

# 启动服务
start_services() {
    log_step "启动 YFlow 服务..."

    # 检查 .env 是否存在
    if [ ! -f "$PROJECT_DIR/.env" ]; then
        log_error ".env 配置文件不存在，请先运行 ./deploy/deploy.sh 进行配置"
        exit 1
    fi

    # 检查是否需要启动 LibreTranslate
    local enable_mt=$(grep "ENABLE_LIBRETRANSLATE=" "$PROJECT_DIR/.env" | cut -d'=' -f2)

    if [ "$enable_mt" = "no" ]; then
        log_info "跳过 LibreTranslate 机器翻译服务 (已在配置中禁用)"
    fi

    # 拉取最新镜像
    log_step "拉取最新镜像..."
    docker compose pull

    # 启动服务
    log_step "启动容器..."
    docker compose up -d

    # 等待服务启动
    log_step "等待服务启动..."
    sleep 10

    # 检查服务状态
    check_services_status
}

# 检查服务状态
check_services_status() {
    echo ""
    log_step "检查服务状态..."

    local services=("db" "backend" "frontend" "caddy")
    local all_running=true

    for service in "${services[@]}"; do
        local container_name=$(docker ps --format '{{.Names}}' | grep -E "^yflow-${service}(-[0-9]+)?$" | head -1)
        if [ -z "$container_name" ]; then
            container_name="yflow-${service}"
        fi

        local status=$(docker inspect --format='{{.State.Status}}' "${container_name}" 2>/dev/null || echo "unknown")
        local health=$(docker inspect --format='{{.State.Health.Status}}' "${container_name}" 2>/dev/null || echo "")

        if [ "$status" = "running" ]; then
            if [ -n "$health" ] && [ "$health" != "healthy" ]; then
                echo -e "  ${YELLOW}✓${NC} $service: $status ($health)"
            else
                echo -e "  ${GREEN}✓${NC} $service: 运行中"
            fi
        else
            echo -e "  ${RED}✗${NC} $service: $status"
            all_running=false
        fi
    done

    if [ "$all_running" = true ]; then
        log_info "✅ 所有服务启动成功"
    else
        log_warn "部分服务可能未正常启动，请检查日志"
    fi
}

# 停止服务
stop_services() {
    log_step "停止 YFlow 服务..."
    docker compose stop
    log_info "✅ 服务已停止"
}

# 重启服务
restart_services() {
    log_step "重启 YFlow 服务..."
    docker compose restart
    log_info "✅ 服务已重启"
}

# 查看日志
view_logs() {
    local service=${1:-}
    if [ -n "$service" ]; then
        docker compose logs -f "$service"
    else
        docker compose logs -f
    fi
}

# 查看状态
status_services() {
    echo ""
    docker compose ps
    echo ""
}

# 重置数据
reset_data() {
    echo ""
    log_warn "⚠️  警告：此操作将删除所有数据！"
    echo ""
    read -p "确定要继续吗? (输入 DELETE 确认): " -r
    echo ""

    if [ "$REPLY" = "DELETE" ]; then
        log_step "删除所有数据并重置..."
        docker compose down -v
        rm -f "$PROJECT_DIR/.env"
        log_info "✅ 数据已重置，请重新运行 ./deploy/deploy.sh 进行配置"
    else
        log_info "已取消"
    fi
}

# 显示帮助
show_help() {
    echo ""
    echo -e "${BLUE}YFlow 一键部署脚本${NC}"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  (无)        交互式部署向导"
    echo "  --start     启动服务 (使用现有配置)"
    echo "  --stop      停止所有服务"
    echo "  --restart   重启所有服务"
    echo "  --status    查看服务状态"
    echo "  --logs      查看实时日志 (Ctrl+C 退出)"
    echo "  --logs <服务>  查看指定服务日志"
    echo "  --reset     重置所有数据 (危险!)"
    echo "  --help      显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0              # 交互式部署"
    echo "  $0 --start      # 使用现有配置启动"
    echo "  $0 --logs backend  # 查看后端日志"
    echo "  $0 --reset      # 重置所有数据"
    echo ""
}

# 主函数
main() {
    local command=${1:-}

    case "$command" in
        --start)
            check_dependencies
            check_docker_service
            start_services
            ;;
        --stop)
            check_dependencies
            stop_services
            ;;
        --restart)
            check_dependencies
            restart_services
            ;;
        --status)
            status_services
            ;;
        --logs)
            check_dependencies
            view_logs "$2"
            ;;
        --reset)
            check_dependencies
            reset_data
            ;;
        --help|-h)
            show_help
            ;;
        "")
            check_dependencies
            check_docker_service
            interactive_config
            start_services
            print_completion_info
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 打印完成信息
print_completion_info() {
    local domain=$(grep "DOMAIN=" "$PROJECT_DIR/.env" | cut -d'=' -f2)
    local admin_username=$(grep "ADMIN_USERNAME=" "$PROJECT_DIR/.env" | cut -d'=' -f2)
    local admin_password=$(grep "ADMIN_PASSWORD=" "$PROJECT_DIR/.env" | cut -d'=' -f4 | tr -d '[:space:]')

    # 格式化访问地址
    local access_url="https://$domain"
    if [ "$domain" = "localhost" ]; then
        access_url="http://localhost"
    fi

    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                                ║${NC}"
    echo -e "${GREEN}║${NC}                     ${YELLOW}部署完成!${NC}                           ${GREEN}║${NC}"
    echo -e "${GREEN}║                                                                ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${BLUE}访问地址:${NC}   $access_url"
    echo -e "  ${BLUE}管理员账户:${NC} $admin_username"
    echo -e "  ${BLUE}管理员密码:${NC} $admin_password"
    echo ""
    if [ "$domain" != "localhost" ]; then
        echo -e "  ${GREEN}✓${NC} HTTPS 证书由 Let's Encrypt 自动提供"
        echo -e "  ${YELLOW}提示:${NC} 首次访问可能需要几秒钟获取证书"
        echo ""
    fi
    echo -e "  ${YELLOW}提示:${NC} 首次登录后请及时修改管理员密码"
    echo ""
    echo -e "  ${CYAN}常用命令:${NC}"
    echo -e "    查看日志: ${GREEN}./deploy/deploy.sh --logs${NC}"
    echo -e "    重启服务: ${GREEN}./deploy/deploy.sh --restart${NC}"
    echo -e "    停止服务: ${GREEN}./deploy/deploy.sh --stop${NC}"
    echo -e "    查看状态: ${GREEN}./deploy/deploy.sh --status${NC}"
    echo ""
    echo -e "  ${CYAN}配置文件:${NC} $PROJECT_DIR/.env"
    echo ""
}

# 运行主函数
main "$@"
