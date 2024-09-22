//go:generate go install -v github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo -icon=res/papp.ico -manifest=res/papp.manifest
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/Jeffail/gabs"
	"github.com/pkg/errors"
	"github.com/portapps/portapps/v3"
	"github.com/portapps/portapps/v3/pkg/log"
	"github.com/portapps/portapps/v3/pkg/mutex"
	"github.com/portapps/portapps/v3/pkg/utl"
	"github.com/portapps/portapps/v3/pkg/win"
)

type config struct {
	Cleanup           bool   `yaml:"cleanup" mapstructure:"cleanup"`
	MultipleInstances bool   `yaml:"multiple_instances" mapstructure:"multiple_instances"`
	DisableTelemetry  bool   `yaml:"disable_telemetry" mapstructure:"disable_telemetry"`
	GnuPGAgentPath    string `yaml:"gnupg_agent_path" mapstructure:"gnupg_agent_path"`
	Locale            string `yaml:"locale" mapstructure:"locale"`
}

var (
	app *portapps.App
	cfg *config
)

const (
	embeddedGnupgAgentPath = `app\gnupg\bin\gpg.exe`
	defaultLocale          = "en-US"
)

func init() {
	var err error

	// Default config
	cfg = &config{
		Cleanup:           false,
		MultipleInstances: false,
		DisableTelemetry:  false,
		Locale:            defaultLocale,
	}

	// Init app
	if app, err = portapps.NewWithCfg("thunderbird-portable", "Thunderbird", cfg); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize application. See log file for more info.")
	}
}

func main() {
	var err error
	utl.CreateFolder(app.DataPath)
	profileFolder := utl.CreateFolder(app.DataPath, "profile", "default")

	app.Process = utl.PathJoin(app.AppPath, "thunderbird.exe")
	app.Args = []string{
		"--profile",
		profileFolder,
	}

	// Set env vars
	crashreporterFolder := utl.CreateFolder(app.DataPath, "crashreporter")
	pluginsFolder := utl.CreateFolder(app.DataPath, "plugins")
	os.Setenv("MOZ_CRASHREPORTER", "0")
	os.Setenv("MOZ_CRASHREPORTER_DATA_DIRECTORY", crashreporterFolder)
	os.Setenv("MOZ_CRASHREPORTER_DISABLE", "1")
	os.Setenv("MOZ_CRASHREPORTER_NO_REPORT", "1")
	os.Setenv("MOZ_DATA_REPORTING", "0")
	os.Setenv("MOZ_MAINTENANCE_SERVICE", "0")
	os.Setenv("MOZ_PLUGIN_PATH", pluginsFolder)
	os.Setenv("MOZ_UPDATER", "0")

	// Create and check mutex
	mu, err := mutex.Create(app.ID)
	defer mutex.Release(mu)
	if err != nil {
		if !cfg.MultipleInstances {
			log.Error().Msg("You have to enable multiple instances in your configuration if you want to launch another instance")
			if _, err = win.MsgBox(
				fmt.Sprintf("%s portable", app.Name),
				"Other instance detected. You have to enable multiple instances in your configuration if you want to launch another instance.",
				win.MsgBoxBtnOk|win.MsgBoxIconError); err != nil {
				log.Error().Err(err).Msg("Cannot create dialog box")
			}
			return
		} else {
			log.Warn().Msg("Another instance is already running")
		}
	}

	// Cleanup on exit
	if cfg.Cleanup {
		defer func() {
			utl.Cleanup([]string{
				path.Join(os.Getenv("APPDATA"), "Thunderbird"),
			})
		}()
	}

	// Locale
	locale, err := checkLocale()
	if err != nil {
		log.Error().Err(err).Msg("Cannot set locale")
	}

	// GnuPG agent
	var gnupgAgentPath string
	log.Info().Msg("Seeking GnuPG Agent path...")
	if cfg.GnuPGAgentPath != "" {
		gnupgAgentPath = cfg.GnuPGAgentPath
		log.Info().Msgf("Getting GnuPG Agent from YAML cfg: %s", gnupgAgentPath)
	} else if gnupgAgentPath, err = exec.LookPath("gpg.exe"); err == nil {
		log.Info().Msgf("Getting GnuPG Agent from PATH: %s", gnupgAgentPath)
	} else {
		gnupgAgentPath = utl.PathJoin(app.RootPath, embeddedGnupgAgentPath)
		log.Info().Msgf("Getting embedded GnuPG Agent: %s", gnupgAgentPath)
	}
	if gnupgAgentPath != "" {
		gnupgAgentPath = strings.Replace(strings.Replace(gnupgAgentPath, `/`, `\`, -1), `\`, `\\`, -1)
	}

	// Multiple instances
	if cfg.MultipleInstances {
		log.Info().Msg("Multiple instances enabled")
		app.Args = append(app.Args, "--no-remote")
	}

	// Policies
	if err := createPolicies(); err != nil {
		log.Fatal().Err(err).Msg("Cannot create policies")
	}

	// Autoconfig
	prefFolder := utl.CreateFolder(app.AppPath, "defaults/pref")
	autoconfig := utl.PathJoin(prefFolder, "autoconfig.js")
	if err := utl.CreateFile(autoconfig, `//
pref("general.config.filename", "portapps.cfg");
pref("general.config.obscure_value", 0);`); err != nil {
		log.Fatal().Err(err).Msg("Cannot write autoconfig.js")
	}

	// Mozilla cfg
	mozillaCfgPath := utl.PathJoin(app.AppPath, "portapps.cfg")
	mozillaCfgFile, err := os.Create(mozillaCfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create portapps.cfg")
	}
	mozillaCfgData := struct {
		Telemetry      string
		GnuPgAgentPath string
		Locale         string
	}{
		strconv.FormatBool(!cfg.DisableTelemetry),
		gnupgAgentPath,
		locale,
	}
	mozillaCfgTpl := template.Must(template.New("mozillaCfg").Parse(`// Disable updater
lockPref("app.update.enabled", false);
lockPref("app.update.auto", false);
lockPref("app.update.mode", 0);
lockPref("app.update.service.enabled", false);

// Set locale
pref("intl.locale.requested", "{{ .Locale }}");

// Extensions scopes
lockPref("extensions.enabledScopes", 4);
lockPref("extensions.autoDisableScopes", 3);

// Disable check default client
lockPref("mail.shell.checkDefaultClient", false);

// Disable WinSearch integration
lockPref("mail.winsearch.enable", false);
lockPref("mail.winsearch.firstRunDone", true);

// Disable Add-ons compatibility checking
clearPref("extensions.lastAppVersion");

// Don't show 'know your rights' on first run
pref("browser.rights.3.shown", true);
pref("mail.rights.version", 1);

// Disable start page
lockPref("mailnews.start_page.enabled", false);

// Disable calendar notification
lockPref("calendar.integration.notify", false);

// Don't show WhatsNew on first run after every update
pref("mailnews.start_page_override.mstone", "ignore");

// Disable health reporter
lockPref("datareporting.healthreport.service.enabled", {{ .Telemetry }});

// Disable all data upload (Telemetry and FHR)
lockPref("toolkit.telemetry.enabled", {{ .Telemetry }});
lockPref("datareporting.policy.dataSubmissionEnabled", {{ .Telemetry }});

// Disable crash reporter
lockPref("toolkit.crashreporter.enabled", false);

// Set enigmail GnuPG agent path
pref("extensions.enigmail.agentPath", "{{ .GnuPgAgentPath }}");
`))
	if err := mozillaCfgTpl.Execute(mozillaCfgFile, mozillaCfgData); err != nil {
		log.Fatal().Err(err).Msg("Cannot write portapps.cfg")
	}

	// Fix extensions path
	if err := updateAddonStartup(profileFolder); err != nil {
		log.Error().Err(err).Msg("Cannot fix extensions path")
	}

	defer app.Close()
	app.Launch(os.Args[1:])
}

func checkLocale() (string, error) {
	extSourceFile := fmt.Sprintf("%s.xpi", cfg.Locale)
	extDestFile := fmt.Sprintf("langpack-%s@thunderbird.mozilla.org.xpi", cfg.Locale)
	extsFolder := utl.CreateFolder(app.AppPath, "distribution", "extensions")
	localeXpi := utl.PathJoin(app.AppPath, "langs", extSourceFile)

	// If default locale skip (already embedded)
	if cfg.Locale == defaultLocale {
		return cfg.Locale, nil
	}

	// Check .xpi file exists
	if !utl.Exists(localeXpi) {
		return defaultLocale, fmt.Errorf("XPI file does not exist in %s", localeXpi)
	}

	// Copy .xpi
	if err := utl.CopyFile(localeXpi, utl.PathJoin(extsFolder, extDestFile)); err != nil {
		return defaultLocale, err
	}

	return cfg.Locale, nil
}

func createPolicies() error {
	appFile := utl.PathJoin(utl.CreateFolder(app.AppPath, "distribution"), "policies.json")
	dataFile := utl.PathJoin(app.DataPath, "policies.json")
	defaultPolicies := struct {
		Policies map[string]interface{} `json:"policies"`
	}{
		Policies: map[string]interface{}{
			"DisableAppUpdate":        true,
			"DontCheckDefaultBrowser": true,
		},
	}

	jsonPolicies, err := gabs.Consume(defaultPolicies)
	if err != nil {
		return errors.Wrap(err, "Cannot consume default policies")
	}
	log.Debug().Msgf("Default policies: %s", jsonPolicies.String())

	if utl.Exists(dataFile) {
		rawCustomPolicies, err := os.ReadFile(dataFile)
		if err != nil {
			return errors.Wrap(err, "Cannot read custom policies")
		}

		jsonPolicies, err = gabs.ParseJSON(rawCustomPolicies)
		if err != nil {
			return errors.Wrap(err, "Cannot consume custom policies")
		}
		log.Debug().Msgf("Custom policies: %s", jsonPolicies.String())

		jsonPolicies.Set(true, "policies", "DisableAppUpdate")
		jsonPolicies.Set(true, "policies", "DontCheckDefaultBrowser")
	}

	log.Debug().Msgf("Applied policies: %s", jsonPolicies.String())
	err = os.WriteFile(appFile, []byte(jsonPolicies.StringIndent("", "  ")), 0644)
	if err != nil {
		return errors.Wrap(err, "Cannot write policies")
	}

	return nil
}

func updateAddonStartup(profileFolder string) error {
	lz4File := path.Join(profileFolder, "addonStartup.json.lz4")
	if !utl.Exists(lz4File) || app.Prev.RootPath == "" {
		return nil
	}

	lz4Raw, err := mozLz4Decompress(lz4File)
	if err != nil {
		return err
	}

	prevPathLin := strings.Replace(utl.FormatUnixPath(app.Prev.RootPath), ` `, `%20`, -1)
	currPathLin := strings.Replace(utl.FormatUnixPath(app.RootPath), ` `, `%20`, -1)
	lz4Str := strings.Replace(string(lz4Raw), prevPathLin, currPathLin, -1)

	prevPathWin := strings.Replace(strings.Replace(utl.FormatWindowsPath(app.Prev.RootPath), `\`, `\\`, -1), ` `, `%20`, -1)
	currPathWin := strings.Replace(strings.Replace(utl.FormatWindowsPath(app.RootPath), `\`, `\\`, -1), ` `, `%20`, -1)
	lz4Str = strings.Replace(lz4Str, prevPathWin, currPathWin, -1)

	lz4Enc, err := mozLz4Compress([]byte(lz4Str))
	if err != nil {
		return err
	}

	return os.WriteFile(lz4File, lz4Enc, 0644)
}
