package httpservers

import (
	"context"
	"encoding/json"
	"fmt"

	//	cors "github.com/OliveTin/OliveTin/internal/cors"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	config "github.com/OliveTin/OliveTin/internal/config"
	updatecheck "github.com/OliveTin/OliveTin/internal/updatecheck"
	ngrok "github.com/ngrok/libngrok-go"
)

type webUISettings struct {
	Rest             string
	ThemeName        string
	ShowFooter       bool
	ShowNavigation   bool
	ShowNewVersions  bool
	AvailableVersion string
	CurrentVersion   string
}

func findWebuiDir() string {
	directoriesToSearch := []string{
		"./webui",
		"/var/www/olivetin/",
		"/etc/OliveTin/webui/",
	}

	for _, dir := range directoriesToSearch {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			log.WithFields(log.Fields{
				"dir": dir,
			}).Infof("Found the webui directory")

			return dir
		}
	}

	log.Warnf("Did not find the webui directory, you will probably get 404 errors.")

	return "./webui" // Should not exist
}

func generateWebUISettings(w http.ResponseWriter, r *http.Request) {
	jsonRet, _ := json.Marshal(webUISettings{
		Rest:             cfg.ExternalRestAddress + "/api/",
		ThemeName:        cfg.ThemeName,
		ShowFooter:       cfg.ShowFooter,
		ShowNavigation:   cfg.ShowNavigation,
		ShowNewVersions:  cfg.ShowNewVersions,
		AvailableVersion: updatecheck.AvailableVersion,
		CurrentVersion:   updatecheck.CurrentVersion,
	})

	_, err := w.Write([]byte(jsonRet))

	if err != nil {
		log.Warnf("Could not write webui settings: %v", err)
	}
}

func startWebUIServer(cfg *config.Config) {
	log.WithFields(log.Fields{
		"address": cfg.ListenAddressWebUI,
	}).Info("Starting WebUI server")

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(findWebuiDir())))
	mux.HandleFunc("/webUiSettings.json", generateWebUISettings)

	ctx := context.Background()
	opts := ngrok.ConnectOptions().
		WithAuthToken(os.Getenv("NGROK_TOKEN"))
	sess, _ := ngrok.Connect(ctx, opts)

	tun, _ := sess.StartTunnel(ctx, ngrok.HTTPOptions())
	l := tun.AsHTTP()
	fmt.Println("url: ", l.URL())

	log.Fatal(l.Serve(ctx, mux))
}
