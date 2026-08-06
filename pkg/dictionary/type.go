package dictionary

import (
	"context"
	"fmt"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"strings"

	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var Dict IDictionary

type dictionary struct {
	printerID *message.Printer
	printerEn *message.Printer
}

type IDictionary interface {
	SetDictionaryMessage(ctx context.Context, dict string, args ...interface{}) string
}

func New(config config.DictionaryConfig) (IDictionary, error) {
	var err error

	files := strings.Split(config.Path, ".")
	if len(files) < 2 {
		err = fmt.Errorf("invalid filename")
		return nil, err
	}

	viper.SetConfigFile(config.Path)
	viper.SetConfigType("json")

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = fmt.Errorf("fatal error config file: config file not found")
			return nil, err
		}

		err = fmt.Errorf("fatal error config file: %w", err)
		return nil, err
	}

	var library Library
	if err := viper.Unmarshal(&library); err != nil {
		panic(err)
	}

	for _, lib := range library.Library {
		_ = message.SetString(language.Indonesian, lib.Key, lib.Translation.ID)
		_ = message.SetString(language.English, lib.Key, lib.Translation.EN)
	}

	return &dictionary{
		printerID: message.NewPrinter(language.Indonesian),
		printerEn: message.NewPrinter(language.English),
	}, nil
}
