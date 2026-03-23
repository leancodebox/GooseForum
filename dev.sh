#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

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
    echo -e "${BLUE}[STEP]${NC} $1"
}

print_banner() {
    echo ""
    echo "============================================"
    echo "       GooseForum Dev Environment         "
    echo "============================================"
    echo ""
}

check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 is not installed. Please install it first."
        exit 1
    fi
}

check_dependencies() {
    log_step "Checking dependencies..."

    check_command go
    check_command node
    check_command pnpm

    log_info "All dependencies are installed."
}

check_node_modules() {
    local dir=$1
    local name=$2

    if [ ! -d "$dir/node_modules" ]; then
        log_warn "$name node_modules not found. Installing..."
        cd "$dir"
        pnpm install
        cd "$SCRIPT_DIR"
    fi
}

start_backend() {
    log_step "Starting Go backend server with air..."

    if [ ! -f "go.mod" ]; then
        log_error "Cannot find Go module. Are you in the right directory?"
        exit 1
    fi

    if ! command -v air &> /dev/null; then
        log_warn "air not found. Installing air..."
        go install github.com/air-verse/air@latest
    fi

    air &
    BACKEND_PID=$!
    log_info "Backend started (PID: $BACKEND_PID) on http://localhost:5234"
}

start_frontend() {
    log_step "Starting Frontend (Vue) dev server..."

    cd resource
    pnpm dev &
    FRONTEND_PID=$!
    log_info "Frontend started (PID: $FRONTEND_PID) on http://localhost:3009"
    cd ..
}

start_admin() {
    log_step "Starting Admin UI (React) dev server..."

    cd admin
    pnpm dev &
    ADMIN_PID=$!
    log_info "Admin UI started (PID: $ADMIN_PID)"
    cd ..
}

cleanup() {
    echo ""
    log_warn "Shutting down services..."

    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
    fi

    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null || true
    fi

    if [ ! -z "$ADMIN_PID" ]; then
        kill $ADMIN_PID 2>/dev/null || true
    fi

    log_info "All services stopped."
    exit 0
}

print_urls() {
    echo ""
    echo "============================================"
    echo "         Services Started Successfully!    "
    echo "============================================"
    echo ""
    echo -e "${GREEN}Frontend:${NC}   http://localhost:3009"
    echo -e "${GREEN}Admin UI:${NC}  http://localhost:5173 (or next available port)"
    echo -e "${GREEN}Backend:${NC}   http://localhost:5234"
    echo ""
    echo "Press Ctrl+C to stop all services"
    echo ""
}

main() {
    print_banner
    check_dependencies

    check_node_modules "resource" "Frontend"
    check_node_modules "admin" "Admin UI"

    trap cleanup SIGINT SIGTERM

    start_backend
    sleep 2

    start_frontend
    start_admin

    sleep 3

    print_urls

    while true; do
        sleep 1
    done
}

main "$@"
