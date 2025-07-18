# Management API Configuration

The Management API service provides administrative functionality for the CASE framework. This document describes the configuration options and environment variables.

## Configuration File

The service reads configuration from a YAML file specified by the `CONFIG_FILE_PATH` environment variable.

### Configuration Structure

```yaml
logging:
  log_level: "info"
  include_src: true
  log_to_file: true
  filename: "management-api.log"
  max_size: 1
  max_age: 28
  max_backups: 100
  compress_old_logs: true
  include_build_info: "once"

gin_config:
  debug_mode: false
  allow_origins: []
  port: 8080
  mtls:
    use: false
    certificate_paths:
      ca_cert: ""
      server_cert: ""
      server_key: ""

management_user_jwt_sign_key: "your-jwt-signing-key"
management_user_jwt_expires_in: "1h"

allowed_instance_ids:
  - "default"
  - "research_instance"

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
  management_user_db:
    connection_str: "<connection_str>"
    username: "<env var MANAGEMENT_USER_DB_USERNAME>"
    password: "<env var MANAGEMENT_USER_DB_PASSWORD>"
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

study_configs:
  global_secret: "your-study-global-secret"
  external_services: []

filestore_path: "/path/to/filestore"
daily_file_export_path: "/path/to/exports"
```

## Environment Variables

The following environment variables can override configuration file settings:

### Core Configuration

- `CONFIG_FILE_PATH`: Path to the YAML configuration file
- `GIN_DEBUG_MODE`: Set to "true" to enable debug mode
- `MANAGEMENT_API_LISTEN_PORT`: Port for the API server (default: from config)
- `CORS_ALLOW_ORIGINS`: Comma-separated list of allowed CORS origins
- `FILESTORE_PATH`: Path to the file storage directory

### JWT Configuration

- `MANAGEMENT_USER_JWT_SIGN_KEY`: JWT signing key for management users
- `MANAGEMENT_USER_JWT_EXPIRES_IN`: JWT expiration duration (e.g., "24h", "30m")

### Instance Configuration

- `INSTANCE_IDS`: Comma-separated list of allowed instance IDs

### Database Credentials

Database usernames and passwords can be overridden via environment variables:

**Study Database:**

- `STUDY_DB_USERNAME`: Username for study database
- `STUDY_DB_PASSWORD`: Password for study database

**Participant User Database:**

- `PARTICIPANT_USER_DB_USERNAME`: Username for participant user database
- `PARTICIPANT_USER_DB_PASSWORD`: Password for participant user database

**Global Infos Database:**

- `GLOBAL_INFOS_DB_USERNAME`: Username for global infos database
- `GLOBAL_INFOS_DB_PASSWORD`: Password for global infos database

**Messaging Database:**

- `MESSAGING_DB_USERNAME`: Username for messaging database
- `MESSAGING_DB_PASSWORD`: Password for messaging database

**Management User Database:**

- `MANAGEMENT_USER_DB_USERNAME`: Username for management user database
- `MANAGEMENT_USER_DB_PASSWORD`: Password for management user database

### Study Configuration

- `STUDY_GLOBAL_SECRET`: Global secret for study operations (required)

### External Service API Keys

For external services defined in the configuration, API keys can be overridden using environment variables with the pattern:

```sh
EXTERNAL_SERVICE_API_KEY_<SERVICE_NAME>
```

where `<SERVICE_NAME>` is the uppercased service name with special characters replaced by underscores.

## Required Configuration

The following configuration items are required for the service to start:

1. **CONFIG_FILE_PATH**: Must point to a valid configuration file
2. **STUDY_GLOBAL_SECRET**: Must be set either in config file or environment variable
3. **FILESTORE_PATH**: Must point to an existing directory
4. **Database configurations**: At least one database must be properly configured

## Mutual TLS (Optional)

The service supports mutual TLS authentication. To enable:

```yaml
gin_config:
  mtls:
    use: true
    certificate_paths:
      ca_cert: "/path/to/ca.crt"
      server_cert: "/path/to/server.crt"
      server_key: "/path/to/server.key"
```
