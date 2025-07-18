# SMTP Bridge Service

This service provides a secure SMTP bridge for sending emails through configured SMTP servers with high- and low-priority queues. It serves as an intermediary between other services and external SMTP providers, offering load balancing and failover capabilities.

## Configuration

### Required Environment Variable

- `CONFIG_FILE_PATH` - Path to the YAML configuration file

### Optional Environment Variables (Secret Overrides)

The following environment variables can be used to override secrets from the configuration file:

#### API Keys

- `API_KEYS` - Override API keys for accessing the SMTP bridge (comma-separated list)

#### SMTP Server Credentials

For each SMTP server defined in your configuration, you can override the username and password using environment variables with the following naming pattern:

**Format**:

- Username: `SMTP_SERVER_USERNAME_FOR_{NORMALIZED_HOST}_{NORMALIZED_PORT}`
- Password: `SMTP_SERVER_PASSWORD_FOR_{NORMALIZED_HOST}_{NORMALIZED_PORT}`

**Host/Port Normalization Rules**:

- Convert to uppercase
- Replace any non-alphanumeric characters with underscores
- Remove leading/trailing underscores

**Examples**:

- SMTP server: `smtp.gmail.com:587` → Environment variables:
  - `SMTP_SERVER_USERNAME_FOR_SMTP_GMAIL_COM_587`
  - `SMTP_SERVER_PASSWORD_FOR_SMTP_GMAIL_COM_587`
- SMTP server: `mail-relay.example.org:25` → Environment variables:
  - `SMTP_SERVER_USERNAME_FOR_MAIL_RELAY_EXAMPLE_ORG_25`
  - `SMTP_SERVER_PASSWORD_FOR_MAIL_RELAY_EXAMPLE_ORG_25`

## Configuration File Example

```yaml
# Logging configuration
logging:
  log_level: "info"
  include_src: true
  log_to_file: true
  filename: "smtp-bridge.log"
  max_size: 100
  max_age: 28
  max_backups: 3
  compress_old_logs: true
  include_build_info: "once"

# Gin web server configuration
gin_config:
  debug_mode: false
  allow_origins:
    - "https://api.example.com"
    - "https://management.example.com"
  port: "8080"

# API keys for securing access to the SMTP bridge
api_keys:
  - "secure_api_key_1"
  - "secure_api_key_2"

# SMTP server configuration with priority levels
smtp_server_config:
  # High priority servers (used for critical emails)
  high_prio:
    from: '"System Notifications" <noreply@example.com>'
    sender: "system@example.com"
    replyTo:
      - "support@example.com"
    servers:
      - host: "smtp.gmail.com"
        port: "587"
        connections: 1
        insecureSkipVerify: false
        sendTimeout: 30
        auth:
          user: "smtp_username"
          password: "smtp_password"
      - host: "smtp.sendgrid.net"
        port: "587"
        connections: 4
        insecureSkipVerify: false
        sendTimeout: 30
        auth:
          user: "apikey"
          password: "sendgrid_api_key"

  # Low priority servers (used for bulk/non-critical emails)
  low_prio:
    from: '"Bulk Notifications" <bulk@example.com>'
    sender: "bulk@example.com"
    replyTo:
      - "no-reply@example.com"
    servers:
      - host: "smtp.mailgun.org"
        port: "587"
        connections: 2
        insecureSkipVerify: false
        sendTimeout: 60
        auth:
          user: "postmaster@mg.example.com"
          password: "mailgun_password"
      - host: "smtp-relay.internal.com"
        port: "25"
        connections: 2
        insecureSkipVerify: true
        sendTimeout: 60
        auth:
          user: ""
          password: ""
```

## Configuration Options

### Logging Configuration

- `log_level`: Log level (debug, info, warn, error)
- `include_src`: Include source file information in logs
- `log_to_file`: Whether to write logs to a file
- `filename`: Log file name
- `max_size`: Maximum log file size in MB
- `max_age`: Maximum age of log files in days
- `max_backups`: Maximum number of log file backups
- `compress_old_logs`: Whether to compress old log files
- `include_build_info`: Build info inclusion mode (never, once, always)

### Gin Configuration

- `debug_mode`: Enable debug mode for development
- `allow_origins`: CORS allowed origins
- `port`: HTTP server port

### SMTP Server Configuration

Each SMTP server supports:

- `host`: SMTP server hostname
- `port`: SMTP server port
- `connections`: Maximum concurrent connections to this server
- `insecureSkipVerify`: Skip TLS certificate verification (use with caution)
- `sendTimeout`: Timeout for sending emails in seconds
- `auth.user`: Username for SMTP authentication (can be empty for anonymous)
- `auth.password`: Password for SMTP authentication (can be empty for anonymous)

### Server List Configuration

Each server list (high_prio/low_prio) supports:

- `from`: Default "From" header for emails
- `sender`: Default "Sender" header for emails
- `replyTo`: Default "Reply-To" addresses
- `servers`: Array of SMTP server configurations

## Usage

1. Create a configuration file based on the example above
2. Set the `CONFIG_FILE_PATH` environment variable to point to your config file
3. Optionally set any secret override environment variables
4. Ensure the log file directory exists and is writable
5. Run the SMTP bridge service
