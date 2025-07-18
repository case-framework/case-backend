package smtp_client

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v2"
)

type SmtpServerList struct {
	Servers []SmtpServer `yaml:"servers"`
	From    string       `yaml:"from"`
	Sender  string       `yaml:"sender"`
	ReplyTo []string     `yaml:"replyTo"`
}

type SmtpServer struct {
	Host               string `yaml:"host"`
	Port               string `yaml:"port"`
	Connections        int    `yaml:"connections"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	AuthData           struct {
		Username string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
	SendTimeout int `yaml:"sendTimeout"`
}

// Address URI to smtp server
func (s *SmtpServer) Address() string {
	return s.Host + ":" + s.Port
}

// GetHost returns the hostname of the SMTP server
func (s *SmtpServer) GetHost() string {
	return s.Host
}

// GetPort returns the port of the SMTP server
func (s *SmtpServer) GetPort() string {
	return s.Port
}

// SetUsername sets the username for SMTP authentication
func (s *SmtpServer) SetUsername(username string) {
	s.AuthData.Username = username
}

// SetPassword sets the password for SMTP authentication
func (s *SmtpServer) SetPassword(password string) {
	s.AuthData.Password = password
}

func (sl *SmtpServerList) ReadFromFile(fname string) (err error) {
	yamlFile, err := os.ReadFile(fname)
	if err != nil {
		slog.Error("could not read server config file", slog.String("file", fname), slog.String("error", err.Error()))
		return err
	}
	err = yaml.UnmarshalStrict(yamlFile, &sl)
	return
}
