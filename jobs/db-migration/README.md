# DB Migration Job

This job handles various db migration tasks including dropping and creating indexes.

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
- `MANAGEMENT_USER_DB_USERNAME` - Override management user database username
- `MANAGEMENT_USER_DB_PASSWORD` - Override management user database password
- `GLOBAL_INFOS_DB_USERNAME` - Override global infos database username
- `GLOBAL_INFOS_DB_PASSWORD` - Override global infos database password
- `MESSAGING_DB_USERNAME` - Override messaging database username
- `MESSAGING_DB_PASSWORD` - Override messaging database password

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

# Instance IDs for multi-tenant setup
instance_ids:
  - "default"
  - "instance1"
  - "instance2"


# Task execution flags
task_configs:
  drop_indexes:
    study_db: "<all|defaults|none>"
    participant_user_db: "<all|defaults|none>"
    management_user_db: "<all|defaults|none>"
    global_infos_db: "<all|defaults|none>"
    messaging_db: "<all|defaults|none>"

  create_indexes:
    study_db: true
    participant_user_db: true
    management_user_db: true
    global_infos_db: true
    messaging_db: true

  migration_tasks:
    participant_user_contact_infos_fix: false

```

## Usage

1. Create a configuration file based on the example above
2. Set the `CONFIG_FILE_PATH` environment variable to point to your config file
3. Optionally set any secret override environment variables
4. Run the db migration job

Example:

```bash
export CONFIG_FILE_PATH="/path/to/your/config.yaml"

./db-migration
```

## Tasks

The db migration job can run the following tasks (controlled by the `task_configs` configuration):

- **Drop Indexes**: Drops indexes from the specified databases. For each database, the following options are available:
  - `all`: Drops all indexes
  - `defaults`: Drops indexes with the specified name
  - `none`: Does not drop any indexes (default)

- **Create Indexes**: Creates indexes in the specified databases.
