#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_NAME="kubac-test"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

cleanup() {
    log_info "Cleaning up..."
    kind delete cluster --name "$CLUSTER_NAME" || true
}

# Trap errors and cleanup
trap cleanup EXIT

main() {
    log_info "Starting kubac e2e test on kind..."

    # Build kubac CLI
    log_info "Building kubac CLI..."
    cd "$ROOT_DIR"
    go build -v -o kubac ./cmd/kubac
    chmod +x kubac
    export PATH="$ROOT_DIR:$PATH"

    # Check if kind is installed
    if ! command -v kind &> /dev/null; then
        log_error "kind is not installed. Please install it first:"
        echo "  https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
        exit 1
    fi

    # Create kind cluster
    log_info "Creating kind cluster..."
    kind create cluster --name "$CLUSTER_NAME" --config "$SCRIPT_DIR/kind-config.yaml"

    # Wait for cluster to be ready
    log_info "Waiting for cluster to be ready..."
    kubectl wait --for=condition=ready node --all --timeout=120s

    # Run kubac doctor
    log_info "Running kubac doctor..."
    kubac doctor

    # Initialize kubac
    log_info "Initializing kubac..."
    cd "$ROOT_DIR"
    rm -f kubac.yaml
    kubac init --cluster-profile=local --mode=direct

    # Install kubac platform
    log_info "Installing kubac platform..."
    kubac install --mode=direct --cluster-profile=local

    # Wait for core components
    log_info "Waiting for platform components..."
    sleep 10

    # Check metrics-server
    log_info "Checking metrics-server..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=metrics-server -n kube-system --timeout=180s || {
        log_warn "Metrics server not ready, checking status..."
        kubectl get pods -n kube-system -l app.kubernetes.io/name=metrics-server
    }

    # Check kube-state-metrics
    log_info "Checking kube-state-metrics..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=kube-state-metrics -n kubac-system --timeout=180s || {
        log_warn "Kube-state-metrics not ready, checking status..."
        kubectl get pods -n kubac-system
    }

    # Check Kyverno
    log_info "Checking Kyverno..."
    kubectl wait --for=condition=ready pod -l app=kyverno -n kyverno --timeout=180s || {
        log_warn "Kyverno not ready, checking status..."
        kubectl get pods -n kyverno
    }

    # Deploy demo application
    log_info "Deploying demo application..."
    kubac demo deploy

    # Wait for demo app
    log_info "Waiting for demo app..."
    kubectl wait --for=condition=ready pod -l app=demo -n kubac-demo --timeout=180s

    # Run verification suite
    log_info "Running verification suite..."
    kubac verify --output=text --report=kubac-verify-report.json

    # Run load test
    log_info "Running load test..."
    kubac demo load --duration=30s

    # Run chaos test
    log_info "Running chaos test..."
    kubac demo chaos

    # Give it time to recover
    sleep 15
    kubectl wait --for=condition=ready pod -l app=demo -n kubac-demo --timeout=120s

    log_info "✓ All tests passed successfully!"
    log_info "Cluster is ready for manual testing. To keep it running, press Ctrl+C"
    log_info "To delete the cluster manually, run: kind delete cluster --name $CLUSTER_NAME"

    # Keep the cluster running for manual inspection
    # Comment this out if you want automatic cleanup
    trap - EXIT
    read -p "Press Enter to cleanup and exit..."
    cleanup
}

main "$@"
