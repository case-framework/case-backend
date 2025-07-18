# Messaging Job

This job handles various messaging-related tasks including processing outgoing emails, scheduled messages, study messages, and researcher notifications.

## Configuration

### Required Environment Variable

- `CONFIG_FILE_PATH` - Path to the YAML configuration file

### Optional Environment Variables (Secret Overrides)

The following environment variables can be used to override secrets from the configuration file:

#### Database Credentials

- `STUDY_DB_USERNAME` - Override study database username
- `STUDY_DB_PASSWORD` - Override study database password
- `PARTICIPANT_USER_DB_USERNAME` - Override participant user database username
- `PARTICIPANT_USER_DB_PASSWORD` - Override participant user database password
- `GLOBAL_INFOS_DB_USERNAME` - Override global infos database username
- `GLOBAL_INFOS_DB_PASSWORD` - Override global infos database password
- `MESSAGING_DB_USERNAME` - Override messaging database username
- `MESSAGING_DB_PASSWORD` - Override messaging database password

#### Other Secrets

- `SMTP_BRIDGE_API_KEY` - Override SMTP bridge API key
- `STUDY_GLOBAL_SECRET` - Override study global secret

## Configuration File Example

```yaml
# Logging configuration
logging:
  log_level: "info"
  include_src: true
  log_to_file: true
  filename: "messaging.log"
  max_size: 100
  max_age: 28
  max_backups: 3
  compress_old_logs: true
  include_build_info: true

# Database configurations
db_configs:
  participant_user_db:
    connection_str: "<connection_str>"
    username: "<env var PARTICIPANT_USER_DB_USERNAME>"
    password: "<env var PARTICIPANT_USER_DB_PASSWORD>"
    connection_prefix: ""
    timeout: 30
    idle_conn_timeout: 45
    max_pool_size: 4
    use_no_cursor_timeout: false
    db_name_prefix: ""
    run_index_creation: false

  global_infos_db:
    connection_str: "<connection_str>"
    username: "<env var GLOBAL_INFOS_DB_USERNAME>"
    password: "<env var GLOBAL_INFOS_DB_PASSWORD>"
    connection_prefix: ""
    timeout: 30
    idle_conn_timeout: 45
    max_pool_size: 4
    use_no_cursor_timeout: false
    db_name_prefix: ""
    run_index_creation: false

  messaging_db:
    connection_str: "<connection_str>"
    username: "<env var MESSAGING_DB_USERNAME>"
    password: "<env var MESSAGING_DB_PASSWORD>"
    connection_prefix: ""
    timeout: 30
    idle_conn_timeout: 45
    max_pool_size: 4
    use_no_cursor_timeout: false
    db_name_prefix: ""
    run_index_creation: false

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
    run_index_creation: false

# Instance IDs for multi-tenant setup
instance_ids:
  - "default"
  - "instance1"
  - "instance2"

# Messaging configuration
messaging_configs:
  smtp_bridge_config:
    url: "http://localhost:8080"
    api_key: "your_smtp_bridge_api_key"
    request_timeout: "90s"

  global_email_template_constants:
    app_name: "Your App Name"
    support_email: "support@example.com"
    base_url: "https://your-app.com"

# Task execution flags
run_tasks:
  process_outgoing_emails: true
  schedule_handler: true
  study_messages_handler: true
  researcher_messages_handler: true

# Timing intervals
intervals:
  last_send_attempt_lock_duration: "20m"
  login_token_ttl: "168h"
  unsubscribe_token_ttl: "8760h"

# Study configuration
study_configs:
  global_secret: "your_global_secret_here"
```

## Usage

1. Create a configuration file based on the example above
2. Set the `CONFIG_FILE_PATH` environment variable to point to your config file
3. Optionally set any secret override environment variables
4. Run the messaging job

Example:

```bash
export CONFIG_FILE_PATH="/path/to/your/config.yaml"
export STUDY_DB_PASSWORD="secure_password"
export SMTP_BRIDGE_API_KEY="your_api_key"
./messaging
```

## Tasks

The messaging job can run the following tasks (controlled by the `run_tasks` configuration):

- **Process Outgoing Emails**: Handles the queue of outgoing emails
- **Schedule Handler**: Processes scheduled messages
- **Study Messages Handler**: Handles automated study-related messages
- **Researcher Messages Handler**: Processes researcher notification messages
