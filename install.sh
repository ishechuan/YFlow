#!/bin/bash

# =============================================================================
# YFlow 一键安装脚本
# =============================================================================
# 功能：
#   - 自动下载并安装 YFlow 最新版本
#   - 支持保留配置升级
#   - 委托执行 deploy.sh 的所有功能
# =============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# 常量定义
REPO_OWNER="ishechuan"
REPO_NAME="YFlow"
INSTALL_DIR="$HOME/yflow"
SCRIPT_VERSION="1.0.0"
RELEASE_API_URL="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
INSTALL_SCRIPT_URL="https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/master/install.sh"

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

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

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
    echo -e "${BLUE}║${NC}   ${YELLOW}         国际化管理平台 - 一键安装脚本 v${SCRIPT_VERSION}${NC}           ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}                                                                ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# 检查系统依赖
check_dependencies() {
    log_step "检查系统依赖..."

    local missing_deps=()
    local warnings=()

    if ! command -v curl &> /dev/null; then
        log_error "curl 未安装"
        missing_deps+=("curl")
    fi

    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        missing_deps+=("docker")
    else
        if ! docker info &> /dev/null; then
            log_warn "Docker 已安装但服务未运行"
            warnings+=("docker-service")
        fi
    fi

    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose 未安装"
        missing_deps+=("docker-compose")
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "缺少必要依赖: ${missing_deps[*]}"
        echo ""
        echo "请安装以下依赖后重新运行:"
        echo "  - curl: 绝大多数 Linux 系统已预装"
        echo "  - Docker: https://docs.docker.com/get-docker/"
        echo "  - Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi

    if [ ${#warnings[@]} -ne 0 ]; then
        log_warn "部分依赖需要额外配置"
        for warn in "${warnings[@]}"; do
            case "$warn" in
                docker-service)
                    echo "  - 请确保 Docker 服务已启动"
                    ;;
            esac
        done
        echo ""
    fi

    log_success "所有依赖检查通过"
}

# 获取最新版本信息
get_latest_release() {
    log_step "获取 YFlow 最新版本信息..."

    local response
    response=$(curl -fsSL -o /dev/null -w "%{http_code}" "$RELEASE_API_URL" 2>/dev/null) || response="000"

    if [ "$response" != "200" ]; then
        log_error "无法连接 GitHub API: HTTP $response"
        log_info "请检查网络连接或稍后重试"
        exit 1
    fi

    local release_info
    release_info=$(curl -fsSL "$RELEASE_API_URL")

    LATEST_VERSION=$(echo "$release_info" | grep '"tag_name"' | head -1 | sed 's/.*"v\?\([0-9.]*\)".*/\1/')
    RELEASE_URL=$(echo "$release_info" | grep '"browser_download_url"' | grep -E '\.(tar\.gz|zip)$' | head -1 | cut -d '"' -f 4)

    if [ -z "$LATEST_VERSION" ] || [ -z "$RELEASE_URL" ]; then
        log_error "无法解析版本信息"
        exit 1
    fi

    log_info "最新版本: v${LATEST_VERSION}"
}

# 检查已安装版本
check_installed_version() {
    if [ -f "$INSTALL_DIR/VERSION" ]; then
        INSTALLED_VERSION=$(cat "$INSTALL_DIR/VERSION" 2>/dev/null | sed 's/v//')
        log_info "当前安装版本: v${INSTALLED_VERSION:-未知}"
    elif [ -d "$INSTALL_DIR/.git" ]; then
        INSTALLED_VERSION="git"
        log_info "当前通过 Git 安装"
    else
        INSTALLED_VERSION=""
        log_info "未检测到已安装版本"
    fi
}

# 备份配置
backup_config() {
    if [ -f "$INSTALL_DIR/.env" ]; then
        BACKUP_ENV="/tmp/yflow.env.backup.$(date +%Y%m%d%H%M%S)"
        cp "$INSTALL_DIR/.env" "$BACKUP_ENV"
        log_info "已备份配置文件到: $BACKUP_ENV"
    fi
}

# 恢复配置
restore_config() {
    if [ -f "$BACKUP_ENV" ]; then
        cp "$BACKUP_ENV" "$INSTALL_DIR/.env"
        log_success "已恢复配置文件"
        rm -f "$BACKUP_ENV"
    fi
}

# 下载并解压
download_and_extract() {
    log_step "下载 YFlow v${LATEST_VERSION}..."

    local temp_dir
    temp_dir=$(mktemp -d)
    local archive_path="$temp_dir/yflow.tar.gz"

    trap "rm -rf $temp_dir" EXIT

    echo -n "  下载中... "

    local http_code
    http_code=$(curl -fsSL -o "$archive_path" -w "%{http_code}" "$RELEASE_URL" 2>/dev/null)

    if [ "$http_code" != "200" ]; then
        echo "失败"
        log_error "下载失败: HTTP $http_code"
        exit 1
    fi

    echo "完成"

    log_step "解压文件..."

    local extract_dir="$temp_dir/extracted"
    mkdir -p "$extract_dir"

    tar -xzf "$archive_path" -C "$extract_dir" --strip-components=1 2>/dev/null || {
        tar -xzf "$archive_path" -C "$extract_dir" 2>/dev/null || {
            log_error "解压失败"
            exit 1
        }
    }

    log_step "安装到 $INSTALL_DIR..."

    mkdir -p "$INSTALL_DIR"

    cp -r "$extract_dir"/* "$INSTALL_DIR/" 2>/dev/null || {
        log_error "文件复制失败，请检查权限"
        exit 1
    }

    echo "$LATEST_VERSION" > "$INSTALL_DIR/VERSION"

    if [ -d "$INSTALL_DIR/.git" ]; then
        rm -rf "$INSTALL_DIR/.git"
    fi

    log_success "安装完成"
}

# 安装完成提示
print_install_complete() {
    local domain=$(grep "DOMAIN=" "$INSTALL_DIR/.env" 2>/dev/null | cut -d'=' -f2)
    domain=${domain:-localhost}

    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                                ║${NC}"
    echo -e "${GREEN}║${NC}                     ${YELLOW}安装完成!${NC}                           ${GREEN}║${NC}"
    echo -e "${GREEN}║                                                                ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${BLUE}安装目录:${NC} $INSTALL_DIR"
    echo -e "  ${BLUE}版本:${NC} v${LATEST_VERSION}"
    echo ""
    echo -e "  ${CYAN}下一步操作:${NC}"
    echo ""
    echo -e "  ${GREEN}1.${NC} 进入安装目录:"
    echo -e "       ${MAGENTA}cd $INSTALL_DIR${NC}"
    echo ""
    echo -e "  ${GREEN}2.${NC} 运行部署脚本进行配置:"
    echo -e "       ${MAGENTA}./deploy/deploy.sh${NC}"
    echo ""
    echo -e "  ${GREEN}3.${NC} 或者使用 install.sh 直接部署:"
    echo -e "       ${MAGENTA}./install.sh --help${NC}"
    echo ""
    echo -e "  ${CYAN}常用命令:${NC}"
    echo -e "    查看帮助:     ${MAGENTA}./install.sh --help${NC}"
    echo -e "    启动服务:     ${MAGENTA}./install.sh --start${NC}"
    echo -e "    查看状态:     ${MAGENTA}./install.sh --status${NC}"
    echo -e "    查看日志:     ${MAGENTA}./install.sh --logs${NC}"
    echo -e "    停止服务:     ${MAGENTA}./install.sh --stop${NC}"
    echo ""
}

# 检查并启动服务
check_and_start() {
    if [ ! -f "$INSTALL_DIR/.env" ]; then
        log_error "未检测到配置文件，请先运行 ./deploy/deploy.sh 进行配置"
        exit 1
    fi

    if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
        log_error "docker-compose.yml 不存在，可能安装不完整"
        exit 1
    fi
}

# 委托执行 deploy.sh
delegate_to_deploy() {
    cd "$INSTALL_DIR"

    local script_path="$INSTALL_DIR/deploy/deploy.sh"

    if [ ! -f "$script_path" ]; then
        log_error "deploy.sh 不存在: $script_path"
        exit 1
    fi

    chmod +x "$script_path"
    exec "$script_path" "$@"
}

# 显示帮助
show_help() {
    print_banner
    echo -e "${BLUE}YFlow 一键安装脚本${NC}"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  (无)        下载并安装最新版本"
    echo "  --install   下载并安装最新版本 (同无参数)"
    echo "  --update    升级到最新版本 (保留配置)"
    echo "  --start     启动服务"
    echo "  --stop      停止所有服务"
    echo "  --restart   重启所有服务"
    echo "  --status    查看服务状态"
    echo "  --logs      查看实时日志 (Ctrl+C 退出)"
    echo "  --logs <服务>  查看指定服务日志"
    echo "  --reset     重置所有数据 (危险!)"
    echo "  --help      显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                  # 下载并安装"
    echo "  $0 --update         # 升级到最新版本"
    echo "  $0 --start          # 启动服务"
    echo "  $0 --logs backend   # 查看后端日志"
    echo "  $0 --status         # 查看服务状态"
    echo ""
    echo "安装位置: $INSTALL_DIR"
    echo ""
}

# 主函数
main() {
    local command=${1:-}

    case "$command" in
        --help|-h)
            show_help
            exit 0
            ;;
        --install|"")
            print_banner
            check_dependencies
            get_latest_release
            check_installed_version

            if [ -n "$INSTALLED_VERSION" ] && [ "$INSTALLED_VERSION" != "git" ]; then
                if [ "$INSTALLED_VERSION" = "$LATEST_VERSION" ]; then
                    log_info "当前已安装最新版本 v${LATEST_VERSION}"
                    echo ""
                    read -p "是否重新安装? (y/N): " -n 1 -r
                    echo ""
                    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                        log_info "取消安装"
                        exit 0
                    fi
                else
                    log_warn "当前安装版本: v${INSTALLED_VERSION}"
                    log_info "最新版本: v${LATEST_VERSION}"
                    echo ""
                    read -p "是否升级? (Y/n): " -n 1 -r
                    echo ""
                    if [[ ! $REPLY =~ ^[Yy]$ ]] && [[ ! -z $REPLY ]]; then
                        log_info "取消升级"
                        exit 0
                    fi
                    backup_config
                    download_and_extract
                    restore_config
                    print_install_complete
                    exit 0
                fi
            fi

            backup_config
            download_and_extract
            restore_config
            print_install_complete
            ;;
        --update)
            print_banner
            check_dependencies
            get_latest_release
            check_installed_version

            if [ -z "$INSTALLED_VERSION" ]; then
                log_warn "未检测到已安装版本，将执行安装"
                backup_config
                download_and_extract
                restore_config
                print_install_complete
            else
                log_info "升级从 v${INSTALLED_VERSION} 到 v${LATEST_VERSION}..."
                backup_config
                download_and_extract
                restore_config
                print_install_complete
            fi
            ;;
        --start)
            check_dependencies
            check_and_start
            delegate_to_deploy --start
            ;;
        --stop)
            check_dependencies
            delegate_to_deploy --stop
            ;;
        --restart)
            check_dependencies
            delegate_to_deploy --restart
            ;;
        --status)
            delegate_to_deploy --status
            ;;
        --logs)
            check_dependencies
            delegate_to_deploy --logs "$2"
            ;;
        --reset)
            check_dependencies
            delegate_to_deploy --reset
            ;;
        *)
            log_error "未知命令: $command"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
