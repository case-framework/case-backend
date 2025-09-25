# Study Timer Job

This job handles scheduled timer operations for active studies, including updating study statistics, executing timer-based actions, and optionally cleaning up orphaned task results.

## Configuration

### Required Environment Variable

- `CONFIG_FILE_PATH` - Path to the YAML configuration file

### Optional Environment Variables (Secret Overrides)

The following environment variables can be used to override secrets from the configuration file:

#### Database Credentials

- `STUDY_DB_USERNAME` - Override study database username
- `STUDY_DB_PASSWORD` - Override study database password

#### Study Configuration

- `STUDY_GLOBAL_SECRET` - Override the global secret used for study operations

#### External Service API Keys

For each external service that has a `name` defined, you can override the `api_key` using environment variables with the following naming pattern:

**Format**: `EXTERNAL_SERVICE_API_KEY_FOR_{NORMALIZED_NAME}`

**Name Normalization Rules**:

- Convert to uppercase
- Replace any non-alphanumeric characters with underscores
- Remove leading/trailing underscores

**Examples**:

- Service name: `"notification-api_v2"` → Environment variable: `EXTERNAL_SERVICE_API_KEY_FOR_NOTIFICATION_API_V2`
- Service name: `"webhook-handler"` → Environment variable: `EXTERNAL_SERVICE_API_KEY_FOR_WEBHOOK_HANDLER`

**Note**: Only services with a defined `name` field will have their API keys overridden. Services without names will be skipped.

## Configuration File Example

```yaml
# Logging configuration
logging:
  log_level: "info"
  include_src: true
  log_to_file: true
  filename: "study-timer.log"
  max_size: 100
  max_age: 28
  max_backups: 3
  compress_old_logs: true
  include_build_info: "once" # one of: never, always, once

# Database configurations
db_configs:
  study_db:
    connection_str: "<connection_str>"
    username: "<env var STUDY_DB_USERNAME>"
    password: "<env var STUDY_DB_PASSWORD>"
    connection_prefix: ""
    timeout: 30
    idle_conn_timeout: 45
    max_pool_size: 4
    use_no_cursor_timeout: false
    db_name_prefix: ""

# List of instance IDs to process
instance_ids:
  - "default"
  - "research_instance"
  - "pilot_study"

# Study module configuration
study_configs:
  global_secret: "your_global_secret_here"

  # External services for timer actions
  external_services:
    - name: "email-service"
      url: "https://api.email-provider.com"
      api_key: "default_email_api_key"
      timeout: 30
    - name: "notification-webhook"
      url: "https://webhook.notification-service.com"
      api_key: "default_webhook_key"
      timeout: 15
    - name: "sms-gateway"
      url: "https://api.sms-gateway.com"
      api_key: "default_sms_key"
      timeout: 20

# Cleanup configuration
clean_up_config:
  filestore_path: "/path/to/filestore"
  clean_orphaned_task_results: true
```

## Usage

1. Create a configuration file based on the example above
2. Set the `CONFIG_FILE_PATH` environment variable to point to your config file
3. Optionally set any secret override environment variables
4. Run the study timer job

Example:

```bash
export CONFIG_FILE_PATH="/path/to/your/config.yaml"
export STUDY_DB_PASSWORD="secure_database_password"
export STUDY_GLOBAL_SECRET="secure_global_secret"
export EXTERNAL_SERVICE_API_KEY_FOR_EMAIL_SERVICE="secure_email_api_key"
export EXTERNAL_SERVICE_API_KEY_FOR_SMS_GATEWAY="secure_sms_api_key"
./study-timer
```

## Features

### Study Timer Operations

- **Active Study Processing**: Automatically processes all active studies across configured instances
- **Study Statistics Updates**: Updates participant counts (active, temporary) and response counts for each study
- **Timer-based Actions**: Executes scheduled actions and rules defined in study configurations
- **Multi-instance Support**: Handles multiple study instances in a single execution

### Study Statistics Tracking

- **Active Participant Count**: Tracks the number of currently active participants per study
- **Temporary Participant Count**: Tracks participants with temporary status
- **Response Count**: Monitors total survey responses per study
- **Automatic Updates**: Statistics are updated during each timer execution

### Cleanup Operations

- **Orphaned Task Results**: Optionally clean up orphaned task results from the filestore
- **Multi-Instance Cleanup**: Clean up across all configured study instances
- **Configurable Paths**: Specify custom filestore paths for cleanup operations

### External Service Integration

- **API Key Management**: Secure management of external service API keys
- **Environment-based Overrides**: Override API keys using environment variables
- **Timeout Configuration**: Configurable timeouts for external service calls
- **Multiple Services**: Support for multiple external services per configuration

## Scheduling

This job is typically run on a scheduled basis (e.g., via cron) to ensure regular study timer operations. The frequency depends on your study requirements:

- **Hourly**: For studies requiring frequent timer checks
- **Daily**: For standard study operations and statistics updates
- **Custom**: Based on specific study timer requirements

Example cron job for daily execution at 2 AM:

```bash
0 2 * * * /path/to/study-timer
```

## Configuration Details

### Instance IDs

The job processes all instance IDs listed in the configuration. Each instance represents a separate study environment or tenant.

### External Services

External services are called during timer operations for actions like:

- Sending notifications
- Triggering webhooks
- Integrating with third-party systems
- Custom timer-based actions

### Filestore Cleanup

When enabled, the cleanup operation:

- Identifies orphaned task results in the filestore
- Removes files that are no longer referenced in the database
- Helps maintain optimal storage usage
- Processes all configured instances

Make sure the process has read/write permissions to the filestore path for cleanup operations to work properly.
