package pmb

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ini "gopkg.in/ini.v1"
)

type ConfigClient interface {
	GetAll() (map[string]string, error)
	Unset(system bool, key string) error
	Set(system bool, key, value string) error
	Get(key string) (string, error)
}

type ConfigGetter interface {
	Get(key string) (string, error)
}

type RealConfigClient struct {
	systemConfig string
	userConfig   string
}

func (rcc *RealConfigClient) Set(system bool, key, value string) error {
	var configPath string

	if system {
		configPath = rcc.systemConfig
	} else {
		configPath = rcc.userConfig
	}

	cfg, err := ini.LooseLoad(configPath)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	section, k := splitKey(key)

	_, err = cfg.Section(section).NewKey(k, value)
	if err != nil {
		return fmt.Errorf("unable to set new key: %v", err)
	}

	err = cfg.SaveToIndent(configPath, "    ")
	if err != nil {
		return fmt.Errorf("unable to save: %v", err)
	}

	return nil
}

func (rcc *RealConfigClient) Unset(system bool, key string) error {
	var configPath string

	if system {
		configPath = rcc.systemConfig
	} else {
		configPath = rcc.userConfig
	}

	cfg, err := ini.LooseLoad(configPath)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	section, k := splitKey(key)

	cfg.Section(section).DeleteKey(k)
	if len(cfg.Section(section).Keys()) == 0 {
		cfg.DeleteSection(section)
	}

	err = cfg.SaveToIndent(configPath, "    ")
	if err != nil {
		return fmt.Errorf("unable to save: %v", err)
	}

	return nil
}

func (rcc *RealConfigClient) Get(key string) (string, error) {
	cfg, err := ini.LooseLoad(rcc.systemConfig, rcc.userConfig)
	if err != nil {
		return "", fmt.Errorf("failure to load config: %v", err)
	}

	section, k := splitKey(key)
	ck := cfg.Section(section).Key(k)

	return ck.Value(), nil
}

func (rcc *RealConfigClient) GetAll() (map[string]string, error) {
	all := make(map[string]string)

	cfg, err := ini.LooseLoad(rcc.systemConfig, rcc.userConfig)
	if err != nil {
		return all, fmt.Errorf("failure to load config: %v", err)
	}

	for _, section := range cfg.Sections() {
		if len(section.Keys()) > 0 {
			for _, key := range section.Keys() {
				all[fmt.Sprintf("%s.%s", section.Name(), key.Name())] = key.Value()
			}
		}
	}

	return all, nil
}

func splitKey(key string) (string, string) {
	parts := strings.SplitN(key, ".", 2)

	return parts[0], parts[1]
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func NewDefaultConfigClient() (*RealConfigClient, error) {
	var baseDir string
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); len(xdgConfigHome) > 0 {
		baseDir = filepath.Join(xdgConfigHome, "pmb")
	} else {
		var home string
		if home = os.Getenv("HOME"); len(home) == 0 {
			return nil, fmt.Errorf("$HOME environment variable not found")
		}
		baseDir = filepath.Join(home, ".config", "pmb")
		os.MkdirAll(baseDir, 0755)
	}

	systemHome := "/etc"
	if pmbSystemEnv := os.Getenv("PMB_SYSTEM_CONFIG"); len(pmbSystemEnv) > 0 {
		systemHome = pmbSystemEnv
	}

	var err error
	systemConfigPath := filepath.Join(systemHome, "pmbconfig")
	if fileExists(systemConfigPath) {
		systemConfigPath, err = filepath.EvalSymlinks(systemConfigPath)
		if err != nil {
			return nil, err
		}
	}

	userConfigPath := filepath.Join(baseDir, "config")
	if fileExists(userConfigPath) {
		userConfigPath, err = filepath.EvalSymlinks(userConfigPath)
		if err != nil {
			return nil, err
		}
	}

	configClient := RealConfigClient{
		systemConfig: systemConfigPath,
		userConfig:   userConfigPath,
	}

	return &configClient, nil
}
