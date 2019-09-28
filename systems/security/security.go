// Package security contains implementation of an API security validators.
package security

import (
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/user"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems"
	"go-home.io/x/server/systems/logger"
)

// Implements security provider.
type provider struct {
	sync.Mutex

	userStorage user.IUserStorage
	logger      common.ILoggerProvider
	secret      common.ISecretProvider
	roles       []*bakedRole
	cache       *cache.Cache
}

// ConstructSecurityProvider has all data required for a new security provider.
type ConstructSecurityProvider struct {
	PluginLogger  common.ILoggerProvider
	Secret        common.ISecretProvider
	Loader        providers.IPluginLoaderProvider
	Roles         []*providers.SecRole
	UserRawConfig []byte
	UserProvider  string
}

// Helper type for pre-baked role.
type bakedRole struct {
	Name  string
	Rules []*providers.BakedRule
	Users []glob.Glob
}

// NewSecurityProvider constructs new security provider.
func NewSecurityProvider(ctor *ConstructSecurityProvider) providers.ISecurityProvider {
	storage, log := loadUserStorage(ctor)
	prov := &provider{
		userStorage: storage,
		logger:      log,
		secret:      ctor.Secret,
		roles:       make([]*bakedRole, 0),
		cache:       cache.New(5*time.Minute, 10*time.Minute),
	}

	prov.processRoles(ctor.Roles)
	return prov
}

// Loads user storage.
// If plugin fails to load, falls back to default file system provider.
func loadUserStorage(ctor *ConstructSecurityProvider) (user.IUserStorage, common.ILoggerProvider) {
	if "" == ctor.UserProvider || "basic" == ctor.UserProvider {
		return loadBasicAuthStorage(ctor)
	}

	loggerCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.PluginLogger,
		Provider:     ctor.UserProvider,
		System:       systems.SysSecurity.String(),
	}

	loggerProvider := logger.NewPluginLogger(loggerCtor)

	initData := &user.InitDataUserStorage{
		Secret:    ctor.Secret,
		Logger:    loggerProvider,
		RawConfig: ctor.UserRawConfig,
	}

	pluginRequest := &providers.PluginLoadRequest{
		ExpectedType:   user.TypeUserStorage,
		SystemType:     systems.SysSecurity,
		InitData:       initData,
		RawConfig:      ctor.UserRawConfig,
		PluginProvider: ctor.UserProvider,
	}

	storage, err := ctor.Loader.LoadPlugin(pluginRequest)
	if err != nil {
		loggerProvider.Error("Failed to load user storage, defaulting to basic", err)
		return loadBasicAuthStorage(ctor)
	}

	return storage.(user.IUserStorage), loggerProvider
}

// Loads default file system provider.
//noinspection GoUnhandledErrorResult
func loadBasicAuthStorage(ctor *ConstructSecurityProvider) (user.IUserStorage, common.ILoggerProvider) {
	loggerCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.PluginLogger,
		Provider:     "basic",
		System:       systems.SysSecurity.String(),
	}

	loggerProvider := logger.NewPluginLogger(loggerCtor)
	loggerProvider.Info("Loading default user storage")

	initData := &user.InitDataUserStorage{
		Secret: ctor.Secret,
		Logger: loggerProvider,
	}

	prov := &basicAuthProvider{}
	prov.Init(initData) // nolint: gosec, errcheck
	return prov, loggerProvider
}

// GetUser returns found user with allowed roles if any.
func (p *provider) GetUser(headers map[string][]string) (providers.IAuthenticatedUser, error) {
	p.Lock()
	defer p.Unlock()

	usr, err := p.userStorage.Authorize(headers)
	if err != nil {
		return nil, errors.Wrap(err, "auth failed")
	}

	authData, ok := p.cache.Get(usr)
	if ok {
		return authData.(*AuthenticatedUser), nil
	}

	authUser := &AuthenticatedUser{
		Username: usr,
		Rules:    make(map[providers.SecSystem][]*providers.BakedRule),
	}

	for _, v := range p.roles {
		found := false
		for _, u := range v.Users {
			if u.Match(usr) {
				found = true
				break
			}
		}

		if found {
			for _, r := range v.Rules {
				_, ok := authUser.Rules[r.System]
				if !ok {
					authUser.Rules[r.System] = make([]*providers.BakedRule, 0)
				}

				authUser.Rules[r.System] = append(authUser.Rules[r.System], r)
			}
		}
	}

	p.cache.Set(usr, authUser, cache.DefaultExpiration)

	return authUser, nil
}

// Processes configured roles and pre-complies regexps.
func (p *provider) processRoles(roles []*providers.SecRole) {
	p.roles = make([]*bakedRole, 0)
	for _, v := range roles {
		role := &bakedRole{
			Name:  v.Name,
			Users: make([]glob.Glob, 0),
			Rules: make([]*providers.BakedRule, 0),
		}

		for _, o := range v.Users {
			o := o
			reg, err := glob.Compile(o)
			if err != nil {
				p.logger.Warn("Failed to compile role's user regexp", "regexp", o,
					common.LogRoleNameToken, v.Name)
				continue
			}

			role.Users = append(role.Users, reg)
		}

		if 0 == len(role.Users) {
			p.logger.Warn("Skipping role since users are empty", common.LogRoleNameToken, v.Name)
			continue
		}

		for _, o := range v.Rules {
			o := o
			rule := p.processRule(&o, v.Name)
			if nil == rule {
				continue
			}

			role.Rules = append(role.Rules, rule)
		}

		if 0 == len(role.Rules) {
			p.logger.Warn("Skipping role since rules are empty", common.LogRoleNameToken, v.Name)
			continue
		}

		p.roles = append(p.roles, role)
	}
}

// Processing config's role rules.
func (p *provider) processRule(rule *providers.SecRoleRule, roleName string) *providers.BakedRule {
	system, err := getSystem(rule.System)
	if err != nil {
		return nil
	}

	baked := &providers.BakedRule{
		Resources: make([]glob.Glob, 0),
		Get:       false,
		Command:   false,
		History:   false,
		System:    system,
	}

	for _, v := range rule.Resources {
		reg, err := glob.Compile(v)
		if err != nil {
			p.logger.Warn("Failed to compile role's resource regexp", "regexp",
				v, common.LogRoleNameToken, roleName)
			continue
		}

		baked.Resources = append(baked.Resources, reg)
	}

	if 0 == len(baked.Resources) {
		p.logger.Warn("Skipping role since resources are empty", common.LogRoleNameToken, roleName)
		return nil
	}

	p.prepareVerbs(rule, baked)
	return baked
}

// Processes verbs.
func (p *provider) prepareVerbs(rule *providers.SecRoleRule, baked *providers.BakedRule) {
	rule.Verbs = make([]providers.SecVerb, 0)

	for _, v := range rule.StrVerb {
		verb, err := getVerb(v)
		if err != nil {
			continue
		}

		rule.Verbs = append(rule.Verbs, verb)
	}

	for _, v := range rule.Verbs {
		switch v {
		case providers.SecVerbAll:
			baked.Get = true
			baked.Command = true
			baked.History = true
			return
		case providers.SecVerbGet:
			baked.Get = true
		case providers.SecVerbCommand:
			baked.Command = true
		case providers.SecVerbHistory:
			baked.History = true
		}
	}
}

// Gets correct system.
func getSystem(in string) (providers.SecSystem, error) {
	system, err := providers.SecSystemString(in)
	if err != nil && isAll(in) {
		err = nil
		system = providers.SecSystemAll
	}

	return system, err
}

// Gets correct verb.
func getVerb(in string) (providers.SecVerb, error) {
	system, err := providers.SecVerbString(in)
	if err != nil && isAll(in) {
		err = nil
		system = providers.SecVerbAll
	}

	return system, err
}

// Checks whether rule is "all".
func isAll(expr string) bool {
	return "*" == expr
}
