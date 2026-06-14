package di

import (
	"github.com/vukyn/isme/internal/constants"

	"github.com/sarulabs/di/v2"
	"github.com/vukyn/kuery/log"
	"github.com/vukyn/kuery/medioa"
)

// defineMedioaClient wires the kuery/medioa SDK as an App-scoped singleton. The
// API key is read once from config and stays server-side — handlers and the
// media usecase only ever receive the constructed client, never the raw key.
//
// When MEDIOA_API_KEY is blank the SDK's New rejects the config; we hold an
// absent (typed-nil) client so isme still boots without a key configured —
// only avatar uploads fail (the media usecase surfaces a clear 502).
func defineMedioaClient() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_MEDIOA_CLIENT,
		Scope: di.App,
		Build: func(ctn di.Container) (any, error) {
			cfg := GetConfig(ctn)
			client, err := medioa.New(medioa.Config{
				BaseURL: cfg.Medioa.BaseURL,
				APIKey:  cfg.Medioa.APIKey,
			})
			if err != nil {
				// Return a typed-nil client (no error) so the container holds an
				// absent client; the media usecase maps the absence to a 502
				// rather than crashing isme at boot when no key is set.
				log.New().Warnf("Medioa client not initialized: %v", err)
				return (*medioa.Client)(nil), nil
			}
			log.New().Debugf("Medioa client initialized with base URL: %s", cfg.Medioa.BaseURL)
			return client, nil
		},
		Close: func(obj any) error {
			log.New().Debug("Medioa client destroyed")
			return nil
		},
	}
	return def
}

// GetMedioaClient returns the medioa SDK client. The returned client may be nil
// when MEDIOA_API_KEY is unset — callers must treat a nil client as a
// configuration error (the media usecase does).
func GetMedioaClient(ctn di.Container) (*medioa.Client, error) {
	service, err := ctn.SafeGet(constants.CONTAINER_NAME_MEDIOA_CLIENT)
	if err != nil {
		return nil, err
	}
	client, _ := service.(*medioa.Client)
	return client, nil
}
