package config_test

import (
	"os"
	"time"

	"code.cloudfoundry.org/debugserver"
	loggingclient "code.cloudfoundry.org/diego-logging-client"
	"code.cloudfoundry.org/durationjson"
	"code.cloudfoundry.org/lager/v3/lagerflags"
	"code.cloudfoundry.org/locket/cmd/locket/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LocketConfig", func() {
	var configFilePath, configData string

	BeforeEach(func() {
		configData = `{
			"log_level": "debug",
			"listen_address": "1.2.3.4:9090",
			"database_driver": "mysql",
			"max_open_database_connections": 1000,
			"db_connection_timeout": "30s",
			"db_read_timeout": "600s",
			"db_write_timeout": "600s",
			"max_database_connection_lifetime": "1h",
			"database_connection_string": "stuff",
			"debug_address": "some-more-stuff",
			"ca_file": "i am a ca file",
			"cert_file": "i am a cert file",
			"key_file": "i am a key file",
			"sql_ca_cert_file": "/var/vcap/jobs/locket/config/sql.ca",
			"sql_enable_identity_verification": true,
      "report_interval":"1s",
			"loggregator": {
				"loggregator_api_port": 1234,
				"loggregator_ca_path": "/var/ca_cert",
				"loggregator_cert_path": "/var/cert_path",
				"loggregator_key_path": "/var/key_path",
				"loggregator_source_id": "my-source-id",
				"loggregator_instance_id": "1",
				"loggregator_job_origin": "Earth"
			}
		}`
	})

	JustBeforeEach(func() {
		configFile, err := os.CreateTemp("", "config-file")
		Expect(err).NotTo(HaveOccurred())

		n, err := configFile.WriteString(configData)
		Expect(err).NotTo(HaveOccurred())
		Expect(n).To(Equal(len(configData)))

		configFilePath = configFile.Name()
	})

	AfterEach(func() {
		err := os.RemoveAll(configFilePath)
		Expect(err).NotTo(HaveOccurred())
	})

	It("correctly parses the config file", func() {
		locketConfig, err := config.NewLocketConfig(configFilePath)
		Expect(err).NotTo(HaveOccurred())

		config := config.LocketConfig{
			DatabaseDriver:                "mysql",
			ListenAddress:                 "1.2.3.4:9090",
			DatabaseConnectionString:      "stuff",
			DBConnectionTimeout:           durationjson.Duration(30 * time.Second),
			DBReadTimeout:                 durationjson.Duration(600 * time.Second),
			DBWriteTimeout:                durationjson.Duration(600 * time.Second),
			MaxOpenDatabaseConnections:    1000,
			MaxDatabaseConnectionLifetime: durationjson.Duration(time.Hour),
			LagerConfig: lagerflags.LagerConfig{
				LogLevel: "debug",
			},
			DebugServerConfig: debugserver.DebugServerConfig{
				DebugAddress: "some-more-stuff",
			},
			CaFile:                        "i am a ca file",
			CertFile:                      "i am a cert file",
			KeyFile:                       "i am a key file",
			SQLCACertFile:                 "/var/vcap/jobs/locket/config/sql.ca",
			SQLEnableIdentityVerification: true,
			LoggregatorConfig: loggingclient.Config{
				APIPort:    1234,
				CACertPath: "/var/ca_cert",
				CertPath:   "/var/cert_path",
				KeyPath:    "/var/key_path",
				JobOrigin:  "Earth",
				SourceID:   "my-source-id",
				InstanceID: "1",
			},
			ReportInterval: durationjson.Duration(time.Second),
		}

		Expect(locketConfig).To(Equal(config))
	})

	Context("when the file does not exist", func() {
		It("returns an error", func() {
			_, err := config.NewLocketConfig("foobar")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the file does not contain valid json", func() {
		BeforeEach(func() {
			configData = "{{"
		})

		It("returns an error", func() {
			_, err := config.NewLocketConfig(configFilePath)
			Expect(err).To(HaveOccurred())
		})
	})
})
