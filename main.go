//go:generate go install -v github.com/kevinburke/go-bindata/go-bindata
//go:generate go-bindata -prefix res/ -pkg assets -o assets/assets.go res/Opera.lnk
//go:generate go install -v github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo -icon=res/papp.ico -manifest=res/papp.manifest
package main

import (
	"os"
	"path"

	"github.com/portapps/opera-portable/assets"
	"github.com/portapps/portapps/v3"
	"github.com/portapps/portapps/v3/pkg/log"
	"github.com/portapps/portapps/v3/pkg/registry"
	"github.com/portapps/portapps/v3/pkg/shortcut"
	"github.com/portapps/portapps/v3/pkg/utl"
)

type config struct {
	Cleanup bool `yaml:"cleanup" mapstructure:"cleanup"`
}

var (
	app *portapps.App
	cfg *config
)

func init() {
	var err error

	// Default config
	cfg = &config{
		Cleanup: false,
	}

	// Init app
	if app, err = portapps.NewWithCfg("opera-portable", "Opera", cfg); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize application. See log file for more info.")
	}
}

func main() {
	utl.CreateFolder(app.DataPath)
	app.Process = utl.PathJoin(app.AppPath, "opera.exe")
	app.Args = []string{
		"--user-data-dir=" + app.DataPath,
		"--no-default-browser-check",
		"--disable-logging",
		"--disable-breakpad",
		"--disable-machine-id",
		"--disable-encryption-win",
		"--no-sandbox",
		"--enable-unsafe-webgpu",
	}

	// Cleanup on exit
	if cfg.Cleanup {
		defer func() {
			utl.Cleanup([]string{
				path.Join(os.Getenv("APPDATA"), "Opera"),
				path.Join(os.Getenv("LOCALAPPDATA"), "Opera"),
			})
		}()
	}

	// Copy default shortcut
	shortcutPath := path.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Opera Portable.lnk")
	defaultShortcut, err := assets.Asset("Opera.lnk")
	if err != nil {
		log.Error().Err(err).Msg("Cannot load asset Opera.lnk")
	}
	err = os.WriteFile(shortcutPath, defaultShortcut, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Cannot write default shortcut")
	}

	// Update default shortcut
	err = shortcut.Create(shortcut.Shortcut{
		ShortcutPath:     shortcutPath,
		TargetPath:       app.Process,
		Arguments:        shortcut.Property{Clear: true},
		Description:      shortcut.Property{Value: "Opera Portable by Portapps and Adeii"},
		IconLocation:     shortcut.Property{Value: app.Process},
		WorkingDirectory: shortcut.Property{Value: app.AppPath},
	})
	if err != nil {
		log.Error().Err(err).Msg("Cannot create shortcut")
	}
	defer func() {
		if err := os.Remove(shortcutPath); err != nil {
			log.Error().Err(err).Msg("Cannot remove shortcut")
		}
	}()

	// Registry keys
	regsPath := utl.CreateFolder(app.RootPath, "reg")
	bsRegKey := registry.Key{
		Key:  `HKCU\SOFTWARE\Opera`,
		Arch: "32",
	}
	bbdRegKey := registry.Key{
		Key:  `HKCU\SOFTWARE\Opera-Browser-Development`,
		Arch: "32",
	}

	if err := bsRegKey.Import(utl.PathJoin(regsPath, "Opera.reg")); err != nil {
		log.Error().Err(err).Msg("Cannot import registry key")
	}
	if err := bbdRegKey.Import(utl.PathJoin(regsPath, "Opera-Browser-Development.reg")); err != nil {
		log.Error().Err(err).Msg("Cannot import registry key")
	}

	defer func() {
		if err := bsRegKey.Export(utl.PathJoin(regsPath, "Opera.reg")); err != nil {
			log.Error().Err(err).Msg("Cannot export registry key")
		}
		if err := bbdRegKey.Export(utl.PathJoin(regsPath, "Opera-Browser-Development.reg")); err != nil {
			log.Error().Err(err).Msg("Cannot export registry key")
		}
		if cfg.Cleanup {
			if err := bsRegKey.Delete(true); err != nil {
				log.Error().Err(err).Msg("Cannot remove registry key")
			}
			if err := bbdRegKey.Delete(true); err != nil {
				log.Error().Err(err).Msg("Cannot remove registry key")
			}
		}
	}()

	defer app.Close()
	app.Launch(os.Args[1:])
}
