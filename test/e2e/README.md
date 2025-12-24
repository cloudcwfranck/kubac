# End-to-End Testing

This directory contains e2e test configurations and scripts for kubac.

## Local Testing

### Prerequisites

- Go 1.21+
- kubectl
- kind OR k3d

### Running Tests

#### Kind

```bash
cd test/kind
./test-kind.sh
```

This will:
1. Create a kind cluster
2. Build and install kubac
3. Deploy the demo application
4. Run the verification suite
5. Run load and chaos tests

#### K3d

```bash
cd test/k3d
./test-k3d.sh
```

Same workflow as kind, but using k3d.

### Manual Testing

To manually test kubac after creating a cluster:

```bash
# Build kubac
go build -o kubac ./cmd/kubac

# Run doctor checks
./kubac doctor

# Initialize
./kubac init --cluster-profile=local --mode=direct

# Install platform
./kubac install

# Deploy demo
./kubac demo deploy

# Verify
./kubac verify

# Test scaling
./kubac demo load --duration=60s

# Test self-healing
./kubac demo chaos
```

## CI/CD

E2e tests run automatically in GitHub Actions:
- Nightly: Every day at 2 AM UTC
- On-demand: Via workflow dispatch

See `.github/workflows/nightly-e2e.yaml` for details.
