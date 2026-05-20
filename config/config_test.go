package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	Cfg *Config
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) SetupSuite() {}

func (suite *ConfigTestSuite) TearDownSuite() {}

func (suite *ConfigTestSuite) BeforeTest(suiteName, testName string) {
	clearBrandEnvVars()
	Set(Empty())
	suite.Cfg = Get()
	suite.Cfg.Env = "production"
}

func (suite *ConfigTestSuite) AfterTest(suiteName, testName string) {
	clearBrandEnvVars()
}

func clearBrandEnvVars() {
	for _, env := range os.Environ() {
		key, _, _ := strings.Cut(env, "=")
		if strings.HasPrefix(key, "WAKAPI_") || strings.HasPrefix(key, "WAKA3X_") {
			os.Unsetenv(key)
		}
	}
}

func (suite *ConfigTestSuite) TestLoadOidcProviders() {
	oidcMock1, _ := mockoidc.Run()
	defer oidcMock1.Shutdown()
	oidcMock2, _ := mockoidc.Run()
	defer oidcMock2.Shutdown()

	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_0_NAME", "testprovider1")
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_0_DISPLAY_NAME", "Test Provider 1")
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_0_CLIENT_ID", oidcMock1.ClientID)
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_0_CLIENT_SECRET", oidcMock1.ClientSecret)
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_0_ENDPOINT", oidcMock1.Addr()+"/oidc")
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_1_NAME", "testprovider2")
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_1_CLIENT_ID", oidcMock2.ClientID)
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_1_CLIENT_SECRET", oidcMock2.ClientSecret)
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_1_ENDPOINT", oidcMock2.Addr()+"/oidc")

	cfg := Load("", "")
	oidcCfg := cfg.Security.OidcProviders

	suite.Len(oidcCfg, 2)
	suite.Equal("testprovider1", oidcCfg[0].Name)
	suite.Equal("Test Provider 1", oidcCfg[0].DisplayName)
	suite.Equal("Test Provider 1", oidcCfg[0].String())
	suite.Equal(oidcMock1.ClientID, oidcCfg[0].ClientID)
	suite.Equal(oidcMock1.ClientSecret, oidcCfg[0].ClientSecret)
	suite.Equal(oidcMock1.Addr()+"/oidc", oidcCfg[0].Endpoint)
	suite.Equal("testprovider2", oidcCfg[1].Name)
	suite.Equal("", oidcCfg[1].DisplayName)
	suite.Equal("Testprovider2", oidcCfg[1].String())
	suite.Equal(oidcMock2.ClientID, oidcCfg[1].ClientID)
	suite.Equal(oidcMock2.ClientSecret, oidcCfg[1].ClientSecret)
	suite.Equal(oidcMock2.Addr()+"/oidc", oidcCfg[1].Endpoint)

	p1, err1 := GetOidcProvider("testprovider1")
	suite.NoError(err1)
	suite.Equal("Test Provider 1", p1.DisplayName)

	p2, err2 := GetOidcProvider("testprovider2")
	suite.NoError(err2)
	suite.Equal("Testprovider2", p2.DisplayName)
}

func (suite *ConfigTestSuite) TestLoadWaka3xEnvAliases() {
	suite.T().Setenv("WAKA3X_PUBLIC_URL", "https://waka3x.example")
	suite.T().Setenv("WAKA3X_DB_NAME", "waka3x_test.db")
	suite.T().Setenv("WAKA3X_IMPORT_HOSTS_WHITELIST", "api.example.com,*.internal.test")

	cfg := Load("", "")

	suite.Equal("https://waka3x.example", cfg.Server.PublicUrl)
	suite.Equal("waka3x_test.db", cfg.Db.Name)
	suite.Equal([]string{"api.example.com", "*.internal.test"}, cfg.App.ImportHostsWhitelist)
}

func (suite *ConfigTestSuite) TestWakapiEnvTakesPrecedenceOverWaka3xAlias() {
	suite.T().Setenv("WAKA3X_PUBLIC_URL", "https://waka3x.example")
	suite.T().Setenv("WAKAPI_PUBLIC_URL", "https://wakapi.example")

	cfg := Load("", "")

	suite.Equal("https://wakapi.example", cfg.Server.PublicUrl)
}

func (suite *ConfigTestSuite) TestLoadWaka3xOidcProviderAliases() {
	oidcMock, _ := mockoidc.Run()
	defer oidcMock.Shutdown()

	suite.T().Setenv("WAKA3X_OIDC_PROVIDERS_0_NAME", "waka3x-provider")
	suite.T().Setenv("WAKA3X_OIDC_PROVIDERS_0_CLIENT_ID", oidcMock.ClientID)
	suite.T().Setenv("WAKA3X_OIDC_PROVIDERS_0_CLIENT_SECRET", oidcMock.ClientSecret)
	suite.T().Setenv("WAKA3X_OIDC_PROVIDERS_0_ENDPOINT", oidcMock.Addr()+"/oidc")

	cfg := Load("", "")

	suite.Len(cfg.Security.OidcProviders, 1)
	suite.Equal("waka3x-provider", cfg.Security.OidcProviders[0].Name)
	suite.Equal(oidcMock.ClientID, cfg.Security.OidcProviders[0].ClientID)
}

func (suite *ConfigTestSuite) TestLoadWaka3xSecretFileAlias() {
	secretFile := filepath.Join(suite.T().TempDir(), "password_salt")
	suite.Require().NoError(os.WriteFile(secretFile, []byte("secret-from-file\n"), 0600))
	suite.T().Setenv("WAKA3X_PASSWORD_SALT_FILE", secretFile)

	cfg := Load("", "")

	suite.Equal("secret-from-file", cfg.Security.PasswordSalt)
}

func (suite *ConfigTestSuite) TestResolveConfigFilesLayersDefaultsConfigAndLocal() {
	dir := suite.T().TempDir()
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, DefaultConfigReferencePath), []byte("env: production\n"), 0600))
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, DefaultConfigPath), []byte("env: staging\n"), 0600))
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, LocalConfigPath), []byte("env: local\n"), 0600))

	cwd, err := os.Getwd()
	suite.Require().NoError(err)
	suite.Require().NoError(os.Chdir(dir))
	defer os.Chdir(cwd)

	suite.Equal([]string{LocalConfigPath, DefaultConfigPath, DefaultConfigReferencePath}, resolveConfigFiles(DefaultConfigPath))
}

func (suite *ConfigTestSuite) TestLoadSparseConfigKeepsDefaultReferenceValues() {
	dir := suite.T().TempDir()
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, DefaultConfigReferencePath), []byte(`
env: production
server:
  listen_ipv4: 127.0.0.1
  listen_ipv6: ::1
  public_url: http://localhost:3000
app:
  heartbeat_max_age: 4320h
  date_format: Mon, 02 Jan 2006
  datetime_format: Mon, 02 Jan 2006 15:04
  leaderboard_scope: 7_days
  aggregation_time: '0 15 2 * * *'
  report_time_weekly: '0 0 18 * * 5'
  leaderboard_generation_time: '0 0 6 * * *,0 0 18 * * *'
db:
  dialect: sqlite3
  name: waka3x.db
  max_conn: 10
mail:
  provider: smtp
security:
  insecure_cookies: true
`), 0600))
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, DefaultConfigPath), []byte(`
server:
  public_url: https://waka3x.example
security:
  password_salt: private-salt
`), 0600))

	cwd, err := os.Getwd()
	suite.Require().NoError(err)
	suite.Require().NoError(os.Chdir(dir))
	defer os.Chdir(cwd)

	cfg := Load(DefaultConfigPath, "")

	suite.Equal("127.0.0.1", cfg.Server.ListenIpV4)
	suite.Equal("::1", cfg.Server.ListenIpV6)
	suite.Equal("https://waka3x.example", cfg.Server.PublicUrl)
	suite.Equal("private-salt", cfg.Security.PasswordSalt)
	suite.Equal("4320h", cfg.App.HeartbeatMaxAge)
}

func (suite *ConfigTestSuite) TestLocalConfigOverridesExplicitConfig() {
	dir := suite.T().TempDir()
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, DefaultConfigReferencePath), []byte(`
env: production
server:
  listen_ipv4: 127.0.0.1
  listen_ipv6: ::1
  public_url: http://localhost:3000
app:
  heartbeat_max_age: 4320h
  date_format: Mon, 02 Jan 2006
  datetime_format: Mon, 02 Jan 2006 15:04
  leaderboard_scope: 7_days
  aggregation_time: '0 15 2 * * *'
  report_time_weekly: '0 0 18 * * 5'
  leaderboard_generation_time: '0 0 6 * * *,0 0 18 * * *'
db:
  dialect: sqlite3
  name: waka3x.db
  max_conn: 10
mail:
  provider: smtp
security:
  insecure_cookies: true
`), 0600))
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, DefaultConfigPath), []byte(`
server:
  public_url: https://config.example
`), 0600))
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, LocalConfigPath), []byte(`
server:
  public_url: https://local.example
security:
  password_salt: local-salt
`), 0600))

	cwd, err := os.Getwd()
	suite.Require().NoError(err)
	suite.Require().NoError(os.Chdir(dir))
	defer os.Chdir(cwd)

	cfg := Load(DefaultConfigPath, "")

	suite.Equal("https://local.example", cfg.Server.PublicUrl)
	suite.Equal("local-salt", cfg.Security.PasswordSalt)
}

func (suite *ConfigTestSuite) TestOidcProviderConfigValidate() {
	testCases := []struct {
		name   string
		config oidcProviderConfig
		err    string
	}{
		{
			name: "valid",
			config: oidcProviderConfig{
				Name:         "test-provider-1",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Endpoint:     "https://provider.com/oidc",
			},
			err: "",
		},
		{
			name: "valid with http",
			config: oidcProviderConfig{
				Name:         "test-provider-1",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Endpoint:     "http://provider.com/oidc",
			},
			err: "",
		},
		{
			name: "invalid name with spaces",
			config: oidcProviderConfig{
				Name: "test provider",
			},
			err: "invalid provider name 'test provider', must only contain alphanumeric characters or '-'",
		},
		{
			name: "invalid name with underscore",
			config: oidcProviderConfig{
				Name: "test_provider",
			},
			err: "invalid provider name 'test_provider', must only contain alphanumeric characters or '-'",
		},
		{
			name: "missing client id",
			config: oidcProviderConfig{
				Name:         "test-provider",
				ClientSecret: "client-secret",
				Endpoint:     "https://provider.com/oidc",
			},
			err: "provider 'test-provider' is missing client id",
		},
		{
			name: "missing client secret",
			config: oidcProviderConfig{
				Name:     "test-provider",
				ClientID: "client-id",
				Endpoint: "https://provider.com/oidc",
			},
			err: "provider 'test-provider' is missing client secret",
		},
		{
			name: "missing endpoint",
			config: oidcProviderConfig{
				Name:         "test-provider",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
			},
			err: "provider 'test-provider' is missing endpoint",
		},
		{
			name: "invalid endpoint scheme",
			config: oidcProviderConfig{
				Name:         "test-provider",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Endpoint:     "ftp://provider.com/oidc",
			},
			err: "provider 'test-provider' is missing endpoint",
		},
		{
			name: "endpoint without scheme",
			config: oidcProviderConfig{
				Name:         "test-provider",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Endpoint:     "provider.com/oidc",
			},
			err: "provider 'test-provider' is missing endpoint",
		},
	}
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.err == "" {
				suite.NoError(err)
			} else {
				suite.EqualError(err, tc.err)
			}
		})
	}
}

func (suite *ConfigTestSuite) TestIsDev() {
	suite.True(IsDev("dev"))
	suite.True(IsDev("development"))
	suite.False(IsDev("prod"))
	suite.False(IsDev("production"))
	suite.False(IsDev("anything else"))
}

func (suite *ConfigTestSuite) TestMysqlConnectionString() {
	c := &dbConfig{
		Host:     "test_host",
		Port:     9999,
		User:     "test_user",
		Password: "test_password",
		Name:     "test_name",
		Dialect:  "mysql",
		Charset:  "utf8mb4",
		MaxConn:  10,
		Compress: true,
	}
	suite.Equal(fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=%s&compress=true&sql_mode=ANSI_QUOTES",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		"Local",
	), mysqlConnectionString(c))
}

func (suite *ConfigTestSuite) TestMysqlConnectionStringSocket() {
	c := &dbConfig{
		Socket:   "/var/run/mysql.sock",
		User:     "test_user",
		Password: "test_password",
		Name:     "test_name",
		Dialect:  "mysql",
		Charset:  "utf8mb4",
		MaxConn:  10,
		Compress: true,
	}
	suite.Equal(fmt.Sprintf(
		"%s:%s@unix(%s)/%s?charset=utf8mb4&parseTime=true&loc=%s&compress=true&sql_mode=ANSI_QUOTES",
		c.User,
		c.Password,
		c.Socket,
		c.Name,
		"Local",
	), mysqlConnectionString(c))
}

func (suite *ConfigTestSuite) TestPostgresConnectionString() {
	c := &dbConfig{
		Host:     "test_host",
		Port:     9999,
		User:     "test_user",
		Password: "test_password",
		Name:     "test_name",
		Dialect:  "postgres",
		MaxConn:  10,
	}
	suite.Equal(fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		c.Host,
		c.Port,
		c.User,
		c.Name,
		c.Password,
	), postgresConnectionString(c))
}

func (suite *ConfigTestSuite) TestSqliteConnectionString() {
	c := &dbConfig{
		Name:    "test_name",
		Dialect: "sqlite3",
	}
	suite.True(strings.HasPrefix(sqliteConnectionString(c), c.Name))
	suite.Contains(strings.ToLower(sqliteConnectionString(c)), "journal_mode=wal")
}

func (suite *ConfigTestSuite) TestPublicNetUrl() {
	suite.T().Setenv("WAKAPI_PUBLIC_URL", "https://wakapi.dev")
	cfg := Load("", "")
	suite.NotNil(cfg.Server.PublicNetUrl)
	suite.Equal("wakapi.dev", cfg.Server.PublicNetUrl.Hostname())
	suite.Equal("https", cfg.Server.PublicNetUrl.Scheme)
}

func (suite *ConfigTestSuite) TestIsImportHostWhitelisted() {
	testCases := []struct {
		name      string
		whitelist []string
		host      string
		expected  bool
	}{
		{
			name:      "empty whitelist",
			whitelist: []string{},
			host:      "google.com",
			expected:  true,
		},
		{
			name:      "exact match",
			whitelist: []string{"google.com"},
			host:      "google.com",
			expected:  true,
		},
		{
			name:      "no match",
			whitelist: []string{"google.com"},
			host:      "bing.com",
			expected:  false,
		},
		{
			name:      "wildcard prefix",
			whitelist: []string{"*.google.com"},
			host:      "api.google.com",
			expected:  true,
		},
		{
			name:      "wildcard suffix",
			whitelist: []string{"google.*"},
			host:      "google.de",
			expected:  true,
		},
		{
			name:      "wildcard both sides",
			whitelist: []string{"*google*"},
			host:      "my-google-app.com",
			expected:  true,
		},
		{
			name:      "multiple entries",
			whitelist: []string{"bing.com", "*.google.com"},
			host:      "api.google.com",
			expected:  true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if len(tc.whitelist) > 0 {
				suite.T().Setenv("WAKAPI_IMPORT_HOSTS_WHITELIST", strings.Join(tc.whitelist, ","))
			} else {
				os.Unsetenv("WAKAPI_IMPORT_HOSTS_WHITELIST")
			}

			cfg := Load("", "")
			suite.Equal(tc.expected, cfg.App.IsImportHostWhitelisted(tc.host))
		})
	}
}
