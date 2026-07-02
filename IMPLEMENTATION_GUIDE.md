# Terraform Drift Detector - Implementation Guide

## Overview

Driftctl is a cloud-agnostic Terraform drift detection platform that continuously compares Terraform state files against live cloud infrastructure to identify configuration drift without requiring `terraform plan` or `terraform apply` operations.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    User Interfaces                           │
├──────────────────────┬──────────────────┬───────────────────┤
│  CLI (driftctl)      │  REST API        │  Web Dashboard    │
│  - scan              │  /api/v1/*       │  http://localhost │
│  - report            │                  │  :8080            │
│  - workspace         │                  │                   │
│  - schedule          │                  │                   │
└──────┬───────────────┴────────┬─────────┴─────────┬─────────┘
       │                        │                   │
       └────────────┬───────────┴───────────────────┘
                    │
              ┌─────▼──────────────────────┐
              │   Scanner (Orchestrator)   │
              │  - Coordinates scan flow   │
              │  - Manages state/cloud     │
              └─────┬──────────────────────┘
                    │
        ┌───────────┴──────────────┐
        │                          │
   ┌────▼──────────┐      ┌───────▼────────┐
   │ State Reader  │      │ Cloud Provider │
   │ - Local       │      │ - AWS EC2      │
   │ - S3          │      │ - AWS VPC      │
   │ - HTTP        │      │ - AWS SG       │
   │ - Extractor   │      │ - AWS S3       │
   └────┬──────────┘      │ - Extensible   │
        │                 └───────┬────────┘
        │                         │
        └───────────┬─────────────┘
                    │
              ┌─────▼──────────────┐
              │  Drift Engine      │
              │  - Compare logic   │
              │  - Tag filtering   │
              │  - Attribute diff  │
              └─────┬──────────────┘
                    │
              ┌─────▼──────────────┐
              │  Persistence       │
              │  - SQLite DB       │
              │  - Workspaces      │
              │  - Scans           │
              │  - Schedules       │
              └────────────────────┘
```

## Project Structure

```
cmd/
├── driftctl/            # CLI entry point
│   └── main.go         # Commands: scan, report, workspace, schedule
└── drift-server/        # REST API + Dashboard server
    └── main.go         # Server initialization + scheduler

internal/
├── api/                 # REST API handlers
│   ├── server.go       # HTTP server setup
│   └── handlers.go     # Endpoint implementations
├── config/             # Configuration management
│   └── config.go       # YAML parsing
├── drift/              # Drift detection engine
│   ├── engine.go       # Compare logic
│   └── engine_test.go  # Tests
├── model/              # Data models
│   └── resource.go     # Resource, DriftFinding, etc.
├── output/             # Output formatters
│   └── formatter.go    # JSON, table, CLI output
├── providers/          # Cloud provider abstraction
│   ├── provider.go     # Provider interface + registry
│   └── aws/            # AWS implementation
│       ├── aws.go      # Provider setup
│       ├── fetch_ec2.go
│       ├── fetch_s3.go
│       └── fetch_*.go  # Other resource fetchers
├── scan/               # Scan orchestration
│   ├── scanner.go      # Main scan coordinator
│   └── scan_test.go
├── scheduler/          # Cron scheduling
│   └── scheduler.go    # Schedule management
├── state/              # Terraform state handling
│   ├── reader.go       # Read from local/S3/HTTP
│   ├── s3.go          # S3-specific logic
│   ├── extractor.go   # Parse state → Resource
│   └── extractor_test.go
└── store/              # Data persistence
    ├── store.go        # Interface definition
    └── sqlite.go       # SQLite implementation

web/
├── index.html          # Dashboard UI
└── static/
    ├── app.js         # Dashboard JS client
    └── style.css      # Dashboard styles

configs/
└── driftctl.yaml       # Example configuration

testdata/
└── state/              # Sample terraform state files
```

## Key Concepts

### Resource Model
Every resource (whether from Terraform state or cloud API) is normalized into a canonical `Resource` struct:
- **ID**: Unique identifier (terraform resource address or cloud resource ID)
- **Type**: Resource type (e.g., `aws_instance`, `aws_security_group`)
- **Attributes**: Map of key-value pairs (instance type, security group rules, etc.)
- **Tags**: Key-value tags for categorization and filtering
- **Region**: AWS region or equivalent
- **Source**: "state" or "cloud"

### Drift Finding Types
1. **Missing in Cloud** (Critical): Resource exists in state but not in cloud
2. **Extra in Cloud** (Warning): Resource exists in cloud but not in state
3. **Attribute Changed** (Warning): Resource exists in both but an attribute differs
4. **Tags Changed** (Info): Resource exists in both but tags differ

### Scan Flow
1. **Read State**: Load terraform.tfstate from local file, S3, or HTTP
2. **Extract Expected**: Parse state JSON into Resource objects
3. **Fetch Actual**: Query cloud provider APIs for actual resources
4. **Compare**: Run drift engine to identify differences
5. **Report**: Generate DriftReport with findings and summary
6. **Persist**: Save to SQLite for history and dashboard
7. **Output**: Format results as JSON, table, or HTML dashboard

## Implementation Tasks

### Phase 1: Core Infrastructure (In Progress)
- [x] Project structure and module setup
- [x] Model definitions (Resource, DriftFinding, DriftReport)
- [x] Drift comparison engine
- [x] State reader (local, S3, HTTP backends)
- [x] State parser/extractor
- [x] AWS provider skeleton
- [x] SQLite persistence layer
- [ ] **TODO**: Complete AWS EC2 fetcher
- [ ] **TODO**: Complete AWS VPC/Subnet fetcher
- [ ] **TODO**: Complete AWS Security Group fetcher
- [ ] **TODO**: S3 bucket fetcher

### Phase 2: API & CLI (Partial)
- [x] CLI skeleton with Cobra
- [x] Scan command
- [x] Report command
- [x] Workspace command
- [x] Schedule command
- [ ] **TODO**: REST API handlers (workspaces, scans, reports)
- [ ] **TODO**: Graceful error handling and validation
- [ ] **TODO**: API authentication (optional API key)

### Phase 3: Scheduler & Server
- [x] Cron scheduler setup
- [ ] **TODO**: Complete scheduler implementation
- [ ] **TODO**: REST API server with dashboard
- [ ] **TODO**: Scheduled scan execution

### Phase 4: Dashboard
- [ ] **TODO**: React/Vue frontend (or vanilla JS enhancement)
- [ ] **TODO**: Real-time scan status
- [ ] **TODO**: Historical drift reports
- [ ] **TODO**: Drift visualization and filtering

### Phase 5: Extensions
- [ ] **TODO**: Additional cloud providers (Azure, GCP, etc.)
- [ ] **TODO**: Notification integration (Slack, email, webhooks)
- [ ] **TODO**: Terraform plan integration
- [ ] **TODO**: Multi-region aggregation
- [ ] **TODO**: Drift remediation suggestions

## Configuration

### driftctl.yaml Structure
```yaml
database: driftctl.db              # SQLite database path

api:
  addr: ":8080"                    # API listen address
  api_key: "optional-key"          # Optional API authentication

workspaces:
  - name: prod                     # Workspace identifier
    provider: aws                  # Cloud provider
    state:
      backend: s3                  # State backend (local, s3, http)
      bucket: my-tf-state         # S3 bucket (if S3)
      key: prod/terraform.tfstate # S3 key (if S3)
      region: us-east-1           # S3 region (if S3)
    regions:
      - us-east-1                 # Cloud regions to scan
      - us-west-2
    compare:
      ignore_tags:                # Tags to ignore in comparison
        - environment
        - managed-by
      ignore_attributes:          # Attributes to ignore
        - state
        - last_modified
    schedule:
      cron: "0 */6 * * *"        # Cron for periodic scans
```

## API Endpoints

```
GET    /health                              # Health check
GET    /api/v1/workspaces                  # List all workspaces
POST   /api/v1/workspaces                  # Create workspace
GET    /api/v1/workspaces/{id}            # Get workspace
PUT    /api/v1/workspaces/{id}            # Update workspace
DELETE /api/v1/workspaces/{id}            # Delete workspace

POST   /api/v1/workspaces/{id}/scans      # Trigger manual scan
GET    /api/v1/scans                      # List recent scans
GET    /api/v1/scans/{id}/report          # Get scan report
GET    /api/v1/scans/{id}/report?format=json|table

PUT    /api/v1/workspaces/{id}/schedules  # Set cron schedule
GET    /api/v1/workspaces/{id}/schedules  # Get schedule
```

## CLI Usage Examples

```bash
# Build
make build

# Test scan (state-only, no cloud queries)
./bin/driftctl scan --state testdata/state/sample.tfstate --provider aws --skip-cloud

# Live AWS scan
export AWS_REGION=us-east-1
./bin/driftctl scan --state /path/to/terraform.tfstate --provider aws --region us-east-1

# Scan from workspace config
./bin/driftctl scan --config configs/driftctl.yaml --workspace prod

# View saved report
./bin/driftctl report <scan-id> --output json

# List workspaces
./bin/driftctl workspace list

# Set scan schedule
./bin/driftctl schedule create --workspace prod --cron "0 6 * * *"

# Start API + dashboard
make run-server
# Open http://localhost:8080
```

## Testing Strategy

1. **Unit Tests**: Test individual components in isolation
   - Drift engine comparison logic
   - State parser/extractor
   - Provider fetchers (mock cloud responses)
   - Output formatters

2. **Integration Tests**: Test component interactions
   - Full scan flow with mock data
   - State reader with various backends
   - Database persistence

3. **E2E Tests**: Full stack with test fixtures
   - CLI end-to-end scans
   - API endpoint tests
   - Real AWS integration (optional, in CI)

## Performance Considerations

1. **State Parsing**: Stream-parse large state files to avoid full memory load
2. **Cloud API Calls**: Batch API calls where possible, use pagination
3. **Comparison**: Use efficient map-based lookups for resource matching
4. **Database**: Index on scan_id, workspace_id, timestamp for fast queries
5. **Caching**: Cache cloud API responses within a scan (don't re-fetch same resource)

## Security

1. **AWS Credentials**: Use standard AWS SDK credential chain
2. **API Key**: Optional bearer token for API authentication
3. **State File Access**: Respect IAM permissions and S3 access controls
4. **Database**: SQLite file permissions should be restrictive (0600)
5. **Secrets**: Never log API keys, state contents, or resource sensitive data

## Extensibility

### Adding a Cloud Provider

1. Create `internal/providers/azure/azure.go`
2. Implement `CloudProvider` interface:
   ```go
   type CloudProvider interface {
       Name() string
       FetchResources(ctx context.Context, expected []model.Resource, regions []string) ([]model.Resource, error)
       SupportedTypes() []string
   }
   ```
3. Implement fetchers for each resource type
4. Register in `providers.DefaultRegistry()`

### Adding Resource Types

1. Update state extractor to recognize new resource types
2. Add cloud fetcher for the resource type
3. Update comparison logic if needed (custom attribute handling)
4. Add test fixtures and tests

## Next Steps

1. **Complete AWS Fetchers**: Implement EC2, VPC, SG, S3 fetchers
2. **API Layer**: Build REST handlers and validation
3. **Dashboard**: Create responsive web UI for report viewing
4. **Scheduler**: Finalize cron schedule execution
5. **Testing**: Add comprehensive test coverage
6. **Documentation**: Add provider-specific guides and examples
