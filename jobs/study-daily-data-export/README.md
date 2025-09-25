# Study Daily Data Export Job

This job handles daily data exports for studies, including survey responses and confidential response exports. It also provides cleanup functionality for orphaned task results.

## Configuration

### Required Environment Variable

- `CONFIG_FILE_PATH` - Path to the YAML configuration file

### Optional Environment Variables (Secret Overrides)

The following environment variables can be used to override secrets from the configuration file:

#### Database Credentials

- `STUDY_DB_USERNAME` - Override study database username
- `STUDY_DB_PASSWORD` - Override study database password

#### Confidential Export Task Secrets

For each confidential response export task that has a `name` defined, you can override the `study_global_secret` using environment variables with the following naming pattern:

**Format**: `CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_{NORMALIZED_NAME}`

**Name Normalization Rules**:

- Convert to uppercase
- Replace any non-alphanumeric characters with underscores
- Remove leading/trailing underscores

**Examples**:

- Task name: `"survey-data"` → Environment variable: `CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_SURVEY_DATA`
- Task name: `"weekly report"` → Environment variable: `CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_WEEKLY_REPORT`
- Task name: `"user-feedback_v2"` → Environment variable: `CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_USER_FEEDBACK_V2`
- Task name: `"covid-19 study"` → Environment variable: `CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_COVID_19_STUDY`

**Note**: Only tasks with a defined `name` field will have their secrets overridden. Tasks without names will be skipped. Tasks with the same name will use the same environment variable for override.

## Configuration File Example

```yaml
# Logging configuration
logging:
  log_level: "info"
  include_src: true
  log_to_file: true
  filename: "study-export.log"
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

# Export path for generated files
export_path: "/path/to/export/directory"

# Survey response exports configuration
response_exports:
  retention_days: 30
  override_old: true
  export_tasks:
    - instance_id: "default"
      study_key: "study1"
      survey_keys:
        - "weekly_survey"
        - "monthly_survey"
      extra_context_columns: []
      export_format: "csv"
      separator: "-"
      short_keys: false
      create_empty_file: false
    - instance_id: "research_instance"
      study_key: "covid_study"
      survey_keys:
        - "symptoms_survey"
        - "vaccination_survey"
      extra_context_columns:
        - "location"
      export_format: "json"
      separator: "-"
      short_keys: true
      create_empty_file: true

# Confidential response exports configuration
conf_resp_exports:
  preserve_previous_files: false
  export_tasks:
    - name: "patient-data"
      instance_id: "default"
      study_key: "sensitive_study"
      study_global_secret: "default_secret_here"
      resp_key_filter:
        - "personal_data.Q1"
      export_format: "csv"
    - name: "trial-results"
      instance_id: "research_instance"
      study_key: "clinical_trial"
      study_global_secret: "clinical_secret_here"
      resp_key_filter: []
      export_format: "json"

# Cleanup configuration
clean_up_config:
  clean_orphaned_task_results: true
  filestore_root: "/path/to/filestore"
  instance_ids:
    - "default"
    - "research_instance"
```

## Usage

1. Create a configuration file based on the example above
2. Set the `CONFIG_FILE_PATH` environment variable to point to your config file
3. Optionally set any secret override environment variables
4. Run the study export job

Example:

```bash
export CONFIG_FILE_PATH="/path/to/your/config.yaml"
export STUDY_DB_PASSWORD="secure_database_password"
export CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_PATIENT_DATA="patient_data_secret_key"
export CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_TRIAL_RESULTS="trial_results_secret_key"
./study-daily-data-export
```

## Features

### Survey Response Exports

- **Multiple Export Formats**: Support for CSV and JSON formats
- **Configurable Columns**: Add extra context columns to exports if available
- **Flexible Survey Selection**: Export specific surveys or all surveys from a study
- **Retention Management**: Automatically clean up old export files based on retention days
- **Empty File Creation**: Option to create empty files when no data is available

### Confidential Response Exports

- **Secure Data Handling**: Uses study-specific global secrets for data decryption/processing
- **Filtered Exports**: Export only specific response keys using `resp_key_filter`
- **Multiple Formats**: Support for CSV and JSON export formats
- **Environment-based Secret Override**: Override secrets per export task using environment variables
- **Flexible Naming**: Each export task can have a custom name for identification

### Cleanup Operations

- **Orphaned Task Results**: Clean up orphaned task results from the filestore
- **Multi-Instance Support**: Clean up across multiple study instances
- **Configurable Paths**: Specify custom filestore root paths

## Configuration Details

### Export Formats

- **CSV**: Comma-separated values with configurable separator
- **JSON**: JavaScript Object Notation format

### Export Paths

The job will automatically create the export directory if it doesn't exist. Make sure the process has write permissions to the specified path.
