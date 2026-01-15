package config

import (
	"log"
	"runtime"
)

type Config struct {
	TgCfg  TgBotCfg
	AppCfg AppConfig
}

type TgBotCfg struct {
	Token string
}

type AppConfig struct {
	Root           string
	Sep            string
	PandocPath     string
	WkhtmltopdfPdf string
}

func LoadConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatal(err, "cant't load config file")
		return nil
	}

	return cfg
}

func NewConfig() (*Config, error) {
	var (
		p           string
		s           string
		pandoc      string
		wkhtmltopdf string
	)
	System := runtime.GOOS

	switch System {
	case "linux":
		p = `/home/user/programmin/obsidianProject/data/obsidianProject/`
		s = `/`
		pandoc = "pandoc"
		wkhtmltopdf = "wkhtmltopdf"

	case "windows":
		p = `B:\programmin-20260114T065921Z-1-001\programmin\obsidianProject\data\obsidianProject`
		s = `\`
		pandoc = `C:\Program Files\Pandoc\pandoc.exe`
		wkhtmltopdf = `C:\Program Files\wkhtmltopdf\bin\wkhtmltopdf.exe`
	}
	// убрать токен
	return &Config{
		TgCfg: TgBotCfg{Token: "8401341890:AAFGkFp684unx8941oPvHgB_F1j0knDkNAQ"},
		AppCfg: AppConfig{Root: p,
			Sep:            s,
			PandocPath:     pandoc,
			WkhtmltopdfPdf: wkhtmltopdf,
		},
	}, nil
}
