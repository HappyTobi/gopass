package root

import (
	"context"
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/backend"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

// Store is the public facing password store
type Store struct {
	cfg     *config.Config
	mounts  map[string]*sub.Store
	path    string // path to the root store
	store   *sub.Store
	version string
}

// New creates a new store
func New(ctx context.Context, cfg *config.Config) (*Store, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}
	if cfg.Root != nil && cfg.Root.Path == "" {
		return nil, errors.Errorf("need path")
	}
	r := &Store{
		cfg:     cfg,
		mounts:  make(map[string]*sub.Store, len(cfg.Mounts)),
		path:    cfg.Root.Path,
		version: cfg.Version,
	}

	// create the base store
	if !backend.HasCryptoBackend(ctx) {
		ctx = backend.WithCryptoBackendString(ctx, cfg.Root.CryptoBackend)
	}
	if !backend.HasSyncBackend(ctx) {
		ctx = backend.WithSyncBackendString(ctx, cfg.Root.SyncBackend)
	}
	s, err := sub.New(ctx, "", r.Path(), config.Directory())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize the root store at '%s': %s", r.Path(), err)
	}
	r.store = s

	// initialize all mounts
	for alias, sc := range cfg.Mounts {
		path := fsutil.CleanPath(sc.Path)
		if err := r.addMount(ctx, alias, path, sc); err != nil {
			out.Red(ctx, "Failed to initialize mount %s (%s). Ignoring: %s", alias, path, err)
			continue
		}
	}

	// check for duplicate mounts
	if err := r.checkMounts(); err != nil {
		return nil, errors.Errorf("checking mounts failed: %s", err)
	}

	return r, nil
}

// Exists checks the existence of a single entry
func (r *Store) Exists(ctx context.Context, name string) bool {
	_, store, name := r.getStore(ctx, name)
	return store.Exists(ctx, name)
}

// IsDir checks if a given key is actually a folder
func (r *Store) IsDir(ctx context.Context, name string) bool {
	_, store, name := r.getStore(ctx, name)
	return store.IsDir(ctx, name)
}

func (r *Store) String() string {
	ms := make([]string, 0, len(r.mounts))
	for alias, sub := range r.mounts {
		ms = append(ms, alias+"="+sub.String())
	}
	return fmt.Sprintf("Store(Path: %s, Mounts: %+v)", r.store.Path(), strings.Join(ms, ","))
}

// Path returns the store path
func (r *Store) Path() string {
	return r.path
}

// Alias always returns an empty string
func (r *Store) Alias() string {
	return ""
}

// Store returns the storage backend for the given mount point
func (r *Store) Store(ctx context.Context, name string) backend.Store {
	_, sub, _ := r.getStore(ctx, name)
	return sub.Store()
}
