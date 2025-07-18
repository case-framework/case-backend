# Participant API Service

This service provides the participant-facing API for the CASE platform, handling participant authentication, user management, study participation, and survey responses. It serves as the main interface between study participants and the research platform.

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

#### Authentication & Security

- `PARTICIPANT_USER_JWT_SIGN_KEY` - Override JWT signing key for participant user tokens
- `STUDY_GLOBAL_SECRET` - Override the global secret used for study operations

#### Messaging Configuration

- `SMTP_BRIDGE_API_KEY` - Override SMTP bridge API key for email sending
- `SMS_GATEWAY_API_KEY` - Override SMS gateway API key for SMS notifications

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
  filename: "participant-api.log"
  max_size: 100
  max_age: 28
  max_backups: 3
  compress_old_logs: true
  include_build_info: true

# Gin web server configuration
gin_config:
  debug_mode: false
  allow_origins:
    - "https://app.example.com"
    - "https://participant.example.com"
  port: 8070

  # Mutual TLS configuration (optional)
  mtls:
    use: false
    certificate_paths:
      root_ca_cert: "/path/to/root-ca.crt"
      server_cert: "/path/to/server.crt"
      server_key: "/path/to/server.key"

  # OTP configuration for secure endpoints
  otp_configs:
    - route: "/v1/user/password"
      exact: true
      method: "POST"
      max_age: "24h"
      types: ["email"]
    - route: "/v1/user/change-account-email"
      exact: true
      method: "POST"
      max_age: "24h"
      types: ["email"]

# User management configuration
user_management_config:
  # Password hashing parameters (Argon2)
  pw_hashing:
    argon2_memory: 65536      # Memory usage in KB
    argon2_iterations: 4      # Number of iterations
    argon2_parallelism: 2     # Number of parallel threads

  # JWT configuration for participant users
  participant_user_jwt_config:
    sign_key: "<env var PARTICIPANT_USER_JWT_SIGN_KEY>"
    expires_in: "1h"

  # Rate limiting
  max_new_users_per_5_minutes: 10

  # Email verification settings
  email_contact_verification_token_ttl: "48h"

  # Weekday assignment weights for study scheduling
  weekday_assignation_weights:
    "monday": 1
    "tuesday": 1
    "wednesday": 1
    "thursday": 1
    "friday": 1
    "saturday": 1
    "sunday": 1

  # Path to blocked passwords file (optional)
  blocked_passwords_file_path: "/path/to/blocked-passwords.txt"

# List of allowed instance IDs
allowed_instance_ids:
  - "default"
  - "research_instance"
  - "pilot_study"

# Database configurations
db_configs:
  study_db:
    connection_str: "<connection_str>"
    username: "<env var STUDY_DB_USERNAME>"
    password: "<env var STUDY_DB_PASSWORD>"
    connection_prefix: ""
    timeout: 30
    idle_conn_timeout: 45
    max_pool_size: 8
    use_no_cursor_timeout: false
    db_name_prefix: ""
    run_index_creation: false

  participant_user_db:
    connection_str: "<connection_str>"
    username: "<env var PARTICIPANT_USER_DB_USERNAME>"
    password: "<env var PARTICIPANT_USER_DB_PASSWORD>"
    connection_prefix: ""
    timeout: 30
    idle_conn_timeout: 45
    max_pool_size: 8
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

# Study module configuration
study_configs:
  global_secret: "<env var STUDY_GLOBAL_SECRET>"

  # External services for study actions and integrations
  external_services:
    - name: "notification-service"
      url: "https://api.notifications.example.com"
      api_key: "default_notification_key"
      timeout: 30
    - name: "data-export-service"
      url: "https://api.exports.example.com"
      api_key: "default_export_key"
      timeout: 60

# File storage path for participant files
filestore_path: "/var/lib/case/participant-files"

# Messaging configuration
messaging_configs:
  # SMTP bridge configuration for email sending
  smtp_bridge_config:
    url: "https://smtp-bridge.example.com"
    api_key: "<fallback_api_key>"
    request_timeout: 30

  # SMS gateway configuration (optional)
  sms_config:
    api_key: "<fallback_api_key>"
    url: "https://gw.messaging.cm.com/v1.0/message"

  # Global email template constants
  global_email_template_constants:
    "app_name": "CASE Research Platform"
    "support_email": "support@example.com"
    "website_url": "https://app.example.com"
    "privacy_policy_url": "https://app.example.com/privacy"
    "terms_of_service_url": "https://app.example.com/terms"
```

## Usage

1. Create a configuration file based on the example above
2. Set the `CONFIG_FILE_PATH` environment variable to point to your config file
3. Ensure the filestore path exists and is writable
4. Optionally set any secret override environment variables
5. Run the participant API service

Example:

```bash
export CONFIG_FILE_PATH="/path/to/your/config.yaml"
export PARTICIPANT_USER_DB_PASSWORD="secure_user_db_password"
export STUDY_DB_PASSWORD="secure_study_db_password"
export MESSAGING_DB_PASSWORD="secure_messaging_db_password"
export GLOBAL_INFOS_DB_PASSWORD="secure_global_infos_db_password"
export PARTICIPANT_USER_JWT_SIGN_KEY="secure_jwt_signing_key"
export SMTP_BRIDGE_API_KEY="secure_smtp_bridge_key"
export SMS_GATEWAY_API_KEY="secure_sms_gateway_key"
export STUDY_GLOBAL_SECRET="secure_global_secret"
export EXTERNAL_SERVICE_API_KEY_FOR_NOTIFICATION_SERVICE="secure_notification_key"
./participant-api
```
