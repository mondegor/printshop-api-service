package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	authcfg "github.com/mondegor/go-components/wire/mrauth/config"
	accesscfg "github.com/mondegor/go-core/mraccess/config"
	"github.com/mondegor/go-core/mrapp"
	extfilecfg "github.com/mondegor/go-core/util/mime/config"
)

const (
	environmentName    = "APPX_ENV"
	defaultEnvironment = "prod"
	detectVersion      = "v0.0.0"
	relativeConfigDir  = "./config/"
)

var regexpEnvironment = regexp.MustCompile(`^[a-z][a-z0-9]+$`)

// Create - создаёт, инициализирует и возвращает конфигурацию приложения.
// Сначала собирает сырую конфигурацию из всех слоёв (createConfig), затем
// корректирует значения (CorrectValuesRealm) и проверяет их валидаторами
// (ValidateRealms, ValidateStorableLanguages, ValidateStorableTimeZones и т.д.).
func Create(args CmdArgs, stdout io.Writer) (cfg Config, err error) {
	if stdout == nil {
		return Config{}, errors.New("stdout is required")
	}

	if cfg, err = createConfig(args); err != nil {
		return Config{}, fmt.Errorf("error readConfig: %w", err)
	}

	cfg.Stdout = stdout

	cfg.AccessControl.Realms = authcfg.CorrectValuesRealm(
		cfg.AccessControl.Realms,
		cfg.AccessControl.DefaultOperationConfirm,
		cfg.AccessControl.OverrideAuthToken,
	)

	// дополнительно проверяется загруженная конфигурация
	if err = authcfg.ValidateRealms(cfg.AccessControl.Realms, cfg.AccessControl.Roles); err != nil {
		return Config{}, err
	}

	if err = authcfg.ValidateSessionThresholds(cfg.AccessControl.SessionSoftThreshold, cfg.AccessControl.SessionHardThreshold); err != nil {
		return Config{}, err
	}

	if err = accesscfg.ValidateActionGroups(cfg.AccessControl.ActionGroups); err != nil {
		return Config{}, err
	}

	if err = extfilecfg.ValidateMimeTypes(cfg.AllowedMimeTypes); err != nil {
		return Config{}, err
	}

	if authcfg.IsJWTUsed(cfg.AccessControl.Realms) {
		if cfg.JWT, err = authcfg.InitJWT(cfg.JWT); err != nil {
			return Config{}, err
		}
	}

	if len(cfg.AppLanguages) == 0 {
		cfg.AppLanguages = append(cfg.AppLanguages, "ru-RU")
	}

	if err = authcfg.ValidateStorableLanguages(cfg.AppLanguages); err != nil {
		return Config{}, err
	}

	if err = authcfg.ValidateStorableTimeZones(cfg.AppTimeZones); err != nil {
		return Config{}, err
	}

	if cfg.UnexpectedErrorHttpStatus < 400 || cfg.UnexpectedErrorHttpStatus > 599 {
		return Config{}, fmt.Errorf("unexpected_error_http_status: min=400, max=599, got=%d", cfg.UnexpectedErrorHttpStatus)
	}

	if cfg.AppVersion == detectVersion {
		if ver := mrapp.Version(); ver != "" {
			cfg.AppVersion = ver
		}
	}

	if cfg.WorkDir != "" {
		if cfg.DBMigrationsDir != "" && !path.IsAbs(cfg.DBMigrationsDir) {
			cfg.DBMigrationsDir = path.Join(cfg.WorkDir, cfg.DBMigrationsDir)
		}

		if !path.IsAbs(cfg.FileProviders.ImageStorage2RootDir) {
			cfg.FileProviders.ImageStorage2RootDir = path.Join(cfg.WorkDir, cfg.FileProviders.ImageStorage2RootDir)
		}

		if !path.IsAbs(cfg.AccessControl.RolesDirPath) {
			cfg.AccessControl.RolesDirPath = path.Join(args.WorkDir, cfg.AccessControl.RolesDirPath)
		}
	}

	return cfg, nil
}

// createConfig - собирает сырую конфигурацию из слоёв. Слои применяются по порядку,
// каждый следующий перекрывает значения предыдущего:
//  1. .env-файл (godotenv.Overload) - если путь задан, значения кладутся в ENV процесса;
//  2. unquoteEnvs() - снимает кавычки с ENV (docker может отдавать значения в кавычках);
//  3. определение окружения: args.Environment -> APPX_ENV -> "prod";
//  4. config.yaml - базовый слой (обязателен);
//  5. config_<env>.yaml - накладывается поверх базового (необязателен);
//     отсутствующие в нём ключи сохраняют значения из config.yaml;
//  6. cleanenv.ReadEnv - ENV-переменные перекрывают yaml по тегам env:"...";
//  7. args.LogLevel - точечный override уровня логирования из аргументов CLI.
func createConfig(args CmdArgs) (cfg Config, err error) {
	cfg.StartedAt = time.Now().UTC()

	// загружаются ENV переменные из .env файла, если он был явно указан
	if args.DotEnvPath != "" {
		if args.WorkDir != "" && !path.IsAbs(args.DotEnvPath) {
			args.DotEnvPath = path.Join(args.WorkDir, args.DotEnvPath)
		}

		if err := godotenv.Overload(args.DotEnvPath); err != nil {
			return Config{}, fmt.Errorf("error reading ENV file '%s': %w", args.DotEnvPath, err)
		}

		cfg.DotEnvPath = args.DotEnvPath
	}

	if err := unquoteEnvs(); err != nil {
		return Config{}, fmt.Errorf("error unquote Envs: %w", err)
	}

	if args.Environment == "" {
		args.Environment = os.Getenv(environmentName)

		if args.Environment == "" {
			args.Environment = defaultEnvironment
		}
	}

	if !regexpEnvironment.MatchString(args.Environment) {
		return Config{}, errors.New("args.Environment must match " + regexpEnvironment.String())
	}

	cfg.Environment = args.Environment
	configDir := relativeConfigDir

	if args.WorkDir != "" {
		configDir = path.Join(args.WorkDir, configDir) + "/"
		cfg.WorkDir = args.WorkDir
	}

	// загружается базовая конфигурация (обязательный слой, задаёт значения по умолчанию)
	if err = parseYAML(configDir+"config.yaml", &cfg, true); err != nil {
		return Config{}, err
	}

	// базовая конфигурация уточняется слоем для текущего окружения (необязательный):
	// заданные в нём ключи перекрывают базовые, а отсутствующие - наследуются из config.yaml.
	// Поэтому, например, config_tests.yaml задаёт только специфичное для тестов, а
	// app_time_zones/totp_issuer/app_languages берутся из config.yaml.
	if err = parseYAML(configDir+"config_"+cfg.Environment+".yaml", &cfg, false); err != nil {
		return Config{}, err
	}

	// основная конфигурация уточняется переменными из ENV окружения
	if err = cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("error reading ENV: %w", err)
	}

	if args.LogLevel != "" {
		cfg.LogLevel = args.LogLevel
	}

	return cfg, nil
}

// parseYAML - декодирует один yaml-файл в уже заполненный cfg (слияние, а не замена):
// присутствующие в файле ключи перекрывают текущие значения (слайсы заменяются целиком),
// отсутствующие - остаются нетронутыми. Если required=false и файла нет, он молча
// пропускается (путь помечается как [SKIPPED]); при required=true отсутствие - ошибка.
func parseYAML(configPath string, cfg *Config, required bool) error {
	fp, err := os.Open(configPath) //nolint:gosec
	if err != nil {
		if !required && errors.Is(err, os.ErrNotExist) {
			cfg.ConfigPaths = append(cfg.ConfigPaths, configPath+" [SKIPPED]")

			return nil
		}

		return fmt.Errorf("open config file '%s': %w", configPath, err)
	}

	defer func() {
		_ = fp.Close()
	}()

	if err = cleanenv.ParseYAML(fp, cfg); err != nil {
		return fmt.Errorf("error parsing config file '%s': %w", configPath, err)
	}

	cfg.ConfigPaths = append(cfg.ConfigPaths, configPath)

	return nil
}

// unquoteEnvs - удаляет двойные кавычки у ENV значений, если они встречаются
// (ENV приходящие через docker, могут оставаться в кавычках).
func unquoteEnvs() error {
	for _, item := range os.Environ() {
		keyValue := strings.SplitN(item, "=", 2)

		// если не указаны ключ/значение, или значение пустое
		if len(keyValue) != 2 || len(keyValue[1]) < 2 {
			continue
		}

		// если значение не заключено в двойные кавычки
		if keyValue[1][0] != '"' || keyValue[1][len(keyValue[1])-1] != '"' {
			continue
		}

		value, err := strconv.Unquote(keyValue[1])
		if err != nil {
			return fmt.Errorf("strconv.Unquote: %w", err)
		}

		_ = os.Setenv(keyValue[0], value)
	}

	return nil
}
