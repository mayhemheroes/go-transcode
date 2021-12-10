package hlsproxy

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/m1k1o/go-transcode/internal/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ManagerCtx struct {
	logger  zerolog.Logger
	baseUrl string
	prefix  string
	config  *Config

	cache   map[string]*utils.Cache
	cacheMu sync.RWMutex

	cleanup   bool
	cleanupMu sync.RWMutex
	shutdown  chan struct{}
}

func New(baseUrl string, prefix string, config *Config) *ManagerCtx {
	// use default config values
	if config == nil {
		config = &Config{
			CacheCleanupPeriod: 4 * time.Second,
			SegmentExpiration:  60 * time.Second,
			PlaylistExpiration: 1 * time.Second,
		}
	}

	// ensure it ends with slash
	baseUrl = strings.TrimSuffix(baseUrl, "/")
	baseUrl += "/"

	return &ManagerCtx{
		logger:  log.With().Str("module", "hlsproxy").Str("submodule", "manager").Logger(),
		baseUrl: baseUrl,
		prefix:  prefix,
		config:  config,
		cache:   map[string]*utils.Cache{},
	}
}

func (m *ManagerCtx) Shutdown() {
	m.cleanupStop()
}

func (m *ManagerCtx) ServePlaylist(w http.ResponseWriter, r *http.Request) {
	url := m.baseUrl + strings.TrimPrefix(r.URL.String(), m.prefix)

	cache, ok := m.getFromCache(url)
	if !ok {
		resp, err := http.Get(url)
		if err != nil {
			m.logger.Err(err).Msg("unable to get HTTP")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 && resp.StatusCode >= 300 {
			defer resp.Body.Close()

			m.logger.Err(err).Int("code", resp.StatusCode).Msg("invalid HTTP response")
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}

		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			m.logger.Err(err).Msg("unadle to read response body")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var re = regexp.MustCompile(`(?m:^(https?\:\/\/[^\/]+)?\/)`)
		text := re.ReplaceAllString(string(buf), m.prefix)

		cache = m.saveToCache(url, strings.NewReader(text), m.config.PlaylistExpiration)
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.WriteHeader(200)

	cache.ServeHTTP(w)
}

func (m *ManagerCtx) ServeMedia(w http.ResponseWriter, r *http.Request) {
	url := m.baseUrl + strings.TrimPrefix(r.URL.String(), m.prefix)

	cache, ok := m.getFromCache(url)
	if !ok {
		resp, err := http.Get(url)
		if err != nil {
			m.logger.Err(err).Msg("unable to get HTTP")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if resp.StatusCode < 200 && resp.StatusCode >= 300 {
			defer resp.Body.Close()

			m.logger.Err(err).Int("code", resp.StatusCode).Msg("invalid HTTP response")
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}

		cache = m.saveToCache(url, resp.Body, m.config.SegmentExpiration)
	}

	w.Header().Set("Content-Type", "video/MP2T")
	w.WriteHeader(200)

	cache.ServeHTTP(w)
}
