package _rt_package_name_

// import (
// 	"embed"
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"regexp"
// 	"sort"
// 	"strconv"
// 	"strings"
// 	"sync"
// )

// var gRuntime = newRuntime()

// func OpenRuntime(configPath string, services []Service, dbServiceConfigFS *embed.FS, webFS *embed.FS) error {
// 	for _, service := range services {
// 		gRuntime.AddService(service)
// 	}

// 	return gRuntime.Open(configPath, dbServiceConfigFS, webFS)
// }

// func GetRuntime() *Runtime {
// 	return gRuntime
// }

// func CloseRuntime() error {
// 	return gRuntime.Close()
// }

// func Call(service string, action string, param map[string]any) (ret any, err error) {
// 	return gRuntime.Call(service, action, param)
// }

// func SetUploadChecker(fn UploadCheckFunc) {
// 	gRuntime.SetUploadChecker(fn)
// }

// func SetDownloadChecker(fn DownloadCheckFunc) {
// 	gRuntime.SetDownloadChecker(fn)
// }

// type Runtime struct {
// 	config          *RTConfig
// 	sqlMgr          *SQLManager
// 	cache           IMemCache
// 	serviceMap      map[string]*rtService
// 	actionMap       map[string]*actionConfig
// 	hookMap         map[string][]*hookConfig
// 	initList        []*initConfig
// 	uploadChecker   UploadCheckFunc
// 	downloadChecker DownloadCheckFunc
// }

// func newRuntime() *Runtime {
// 	return &Runtime{
// 		logger: getLoggerByConfig(
// 			RTConfigLog{Kind: "stdout", Level: "info", Config: nil},
// 		),
// 		config:          nil,
// 		sqlMgr:          nil,
// 		cache:           nil,
// 		fsUploadCache:   NewUploadCache(),
// 		serviceMap:      make(map[string]*rtService),
// 		actionMap:       make(map[string]*actionConfig),
// 		hookMap:         make(map[string][]*hookConfig),
// 		initList:        make([]*initConfig, 0),
// 		uploadChecker:   nil,
// 		downloadChecker: nil,
// 	}
// }

// func (p *Runtime) Debugf(format string, a ...any) {
// 	p.logger.Debug(fmt.Sprintf(format, a...))
// }

// func (p *Runtime) Infof(format string, a ...any) {
// 	p.logger.Info(fmt.Sprintf(format, a...))
// }

// func (p *Runtime) Warnf(format string, a ...any) {
// 	p.logger.Warn(fmt.Sprintf(format, a...))
// }

// func (p *Runtime) Errorf(format string, a ...any) {
// 	p.logger.Error(fmt.Sprintf(format, a...))
// }

// func (p *Runtime) Open(configPath string, dbMetaFS *embed.FS, webFS *embed.FS) error {
// 	fnMemoryToBytes := func(memory string) int {
// 		if memory == "" {
// 			return 0
// 		}

// 		// Define a regular expression to match the pattern
// 		re := regexp.MustCompile(`(?i)^(\d+)([gmk])$`)
// 		matches := re.FindStringSubmatch(memory)

// 		// If the format doesn't match, return -1
// 		if len(matches) != 3 {
// 			return -1
// 		}

// 		// Extract the number and the unit
// 		number, err := strconv.Atoi(matches[1])
// 		if err != nil {
// 			return -1
// 		}
// 		unit := strings.ToLower(matches[2])

// 		// Convert the number to seconds based on the unit
// 		switch unit {
// 		case "g":
// 			return number * 1000 * 1000 * 1000
// 		case "m":
// 			return number * 1000 * 1000
// 		case "k":
// 			return number * 1000
// 		default:
// 			return 0
// 		}
// 	}

// 	content := []byte(nil)
// 	e := error(nil)
// 	config := &RTConfig{}

// 	if configPath != "" {
// 		if content, e = os.ReadFile(configPath); e == nil {
// 			p.logger.Info(fmt.Sprintf("found config file %s", configPath))
// 		} else {
// 			return WrapError(e)
// 		}
// 	} else {
// 		if content, e = os.ReadFile("config.json"); e == nil {
// 			p.logger.Info("found config file config.json")
// 		} else {
// 			return WrapError(e)
// 		}
// 	}

// 	if e = json.Unmarshal(content, config); e != nil {
// 		return WrapError(e)
// 	} else {
// 		// Setup Log Config
// 		SetTraceError(config.Log.TraceError)
// 		p.logger = getLoggerByConfig(config.Log)
// 		p.config = config
// 	}

// 	// Open cache
// 	if cache, e := NewLocalMemCache(fnMemoryToBytes(config.MC.Memory)); e != nil {
// 		return WrapError(e)
// 	} else {
// 		p.cache = cache
// 	}

// 	// Open SQL Manager
// 	if manager, e := NewSQLManager(RTConfigDB{
// 		Host:     config.DB.Host,
// 		Port:     config.DB.Port,
// 		User:     config.DB.User,
// 		Password: config.DB.Password,
// 		Name:     config.DB.Name,
// 		Driver:   config.DB.Driver,
// 	}); e != nil {
// 		return WrapError(e)
// 	} else if e := manager.Open(dbMetaFS); e != nil {
// 		return WrapError(e)
// 	} else {
// 		p.sqlMgr = manager
// 	}

// 	// Check and build Service
// 	for _, service := range p.serviceMap {
// 		// Build Service Action Map
// 		if meta := p.sqlMgr.GetService(service.name); meta == nil {
// 			return Errorf("%s \nService %s does not exist ", service.fileLine, service.name)
// 		} else {
// 			for _, action := range service.actionList {
// 				key := fmt.Sprintf("%s.%s", service.name, action.name)

// 				if actionCfg := meta.GetActionConfig(action.name); actionCfg == nil {
// 					return Errorf("%s \nAction %s.%s does not defined in config", action.fileLine, service.name, action.name)
// 				} else if oldAction, ok := p.actionMap[key]; ok {
// 					return Errorf("Duplicated action %s: \n%s\n%s", action.name, action.fileLine, oldAction.fileLine)
// 				} else {
// 					action.tx = actionCfg.Tx
// 					action.editable = actionCfg.Editable
// 					action.useSyncLock = actionCfg.UseSyncLock

// 					p.actionMap[key] = &actionConfig{
// 						name:        action.name,
// 						editable:    action.editable,
// 						useSyncLock: action.useSyncLock,
// 						tx:          action.tx,
// 						fn:          action.fn,
// 						fileLine:    action.fileLine,
// 						authFunc:    action.authFunc,
// 						syncMutex:   &sync.Mutex{},
// 					}
// 				}
// 			}
// 		}

// 		// Build Service Hook Map
// 		for _, hook := range service.hookList {
// 			arr := strings.Split(hook.name, ":")
// 			if len(arr) != 2 {
// 				return Errorf("%s \nInvalid name %s", hook.fileLine, hook.name)
// 			}

// 			hookService := strings.TrimSpace(arr[0])
// 			hookAction := strings.TrimSpace(arr[1])

// 			if p.sqlMgr.GetService(hookService) == nil {
// 				return Errorf("%s \nService %s does not exist", hook.fileLine, hookService)
// 			} else if !ArrayContains([]string{"Add", "Del", "Update"}, hookAction) {
// 				return Errorf("%s \nInvalid name %s", hook.fileLine, hook.name)
// 			} else {
// 				hookList, ok := p.hookMap[hook.name]
// 				if !ok {
// 					hookList = make([]*hookConfig, 0)
// 				}
// 				hookList = append(hookList, &hookConfig{
// 					readOnly: hook.readOnly,
// 					async:    hook.async,
// 					name:     hook.name,
// 					priority: hook.priority,
// 					tx:       hook.tx,
// 					fn:       hook.fn,
// 					fileLine: hook.fileLine,
// 				})
// 				p.hookMap[hook.name] = hookList
// 			}
// 		}
// 	}

// 	// Sort Hook List
// 	for _, list := range p.hookMap {
// 		sort.Slice(list, func(i, j int) bool {
// 			return list[i].priority > list[j].priority
// 		})
// 	}

// 	// Exec Init
// 	p.ExecInit()

// 	// Open upload cache
// 	p.fsUploadCache.Open()

// 	return RunHttpServer(p, webFS)
// }

// func (p *Runtime) Close() error {
// 	if p.sqlMgr != nil {
// 		if e := p.sqlMgr.Close(); e != nil {
// 			return WrapError(e)
// 		} else {
// 			p.sqlMgr = nil
// 		}
// 	}

// 	if p.cache != nil {
// 		if e := p.cache.Close(); e != nil {
// 			return WrapError(e)
// 		} else {
// 			p.cache = nil
// 		}
// 	}

// 	return nil
// }

// func (p *Runtime) AddService(s *rtService) {
// 	if p.config != nil {
// 		panic(Errorf("can not add service after runtime open"))
// 	}

// 	if _, ok := p.serviceMap[s.name]; ok {
// 		panic(Errorf("duplicated service %s", s.name))
// 	}

// 	p.serviceMap[s.name] = s
// 	p.initList = append(p.initList, s.initConfig)
// }

// func (p *Runtime) Call(service string, action string, param map[string]any) (any, error) {
// 	_, ret, err := p.Exec(&httpInput{
// 		Service: service,
// 		Action:  action,
// 		Param:   param,
// 	})
// 	return ret, err
// }

// func (p *Runtime) Exec(input *httpInput) (cookie string, ret any, err error) {
// 	fnExec := func(param *httpInput, meta *actionConfig) (string, any, error) {
// 		ctx := &Context{
// 			ttl:      64,
// 			input:    param,
// 			rt:       p,
// 			cookie:   "",
// 			tx:       p.sqlMgr.NewTransaction(meta.tx, !meta.editable),
// 			dirtyMap: nil,
// 		}
// 		defer func() {
// 			if reason := recover(); reason != nil {
// 				err = WrapError(GetFirstError(err, Errorf("%s\nAction exec error: %s", meta.fileLine, reason)))
// 			}

// 			err = WrapError(GetFirstError(err, ctx.Close(err == nil)))

// 			if err != nil {
// 				ret = nil
// 				cookie = ""
// 			}
// 		}()

// 		if e := meta.authFunc(ctx); e != nil {
// 			cookie = ""
// 			ret = nil
// 			err = e
// 		} else {
// 			ret, err = meta.fn(ctx)
// 			cookie = ctx.cookie
// 		}

// 		return cookie, ret, err
// 	}

// 	key := fmt.Sprintf("%s.%s", input.Service, input.Action)
// 	if meta, ok := p.actionMap[key]; !ok {
// 		return "", nil, Errorf("api: Action %s does not exist", key)
// 	} else if meta.useSyncLock {
// 		meta.syncMutex.Lock()
// 		defer meta.syncMutex.Unlock()
// 		return fnExec(input, meta)
// 	} else {
// 		return fnExec(input, meta)
// 	}
// }

// func (p *Runtime) ExecInit() {
// 	ctx := &Context{
// 		ttl:      64,
// 		input:    &httpInput{},
// 		rt:       p,
// 		tx:       p.sqlMgr.NewTransaction(SqlLevelSerializable, false),
// 		dirtyMap: nil,
// 	}

// 	for _, initConfig := range p.initList {
// 		if initConfig == nil || initConfig.fn == nil {
// 			continue
// 		}

// 		defer func() {
// 			if reason := recover(); reason != nil {
// 				panic(fmt.Sprintf("%s\nInitialize exec error: %s", initConfig.fileLine, reason))
// 			}
// 		}()

// 		if e := initConfig.fn(ctx); e != nil {
// 			panic(ErrorString(e))
// 		}
// 	}

// 	if e := ctx.Close(true); e != nil {
// 		panic(ErrorString(e))
// 	}
// }

// func (p *Runtime) withContext(isolationLevel string, fn func(*Context) error) (ret error) {
// 	ctx := &Context{
// 		ttl:      64,
// 		input:    &httpInput{},
// 		rt:       p,
// 		tx:       p.sqlMgr.NewTransaction(isolationLevel, false),
// 		dirtyMap: nil,
// 	}

// 	defer func() {
// 		if reason := recover(); reason != nil {
// 			ret = WrapError(GetFirstError(ret, Errorf("Context exec error: %s", reason)))
// 		}
// 	}()

// 	ret = fn(ctx)

// 	if e := ctx.Close(ret == nil); e != nil {
// 		ret = GetFirstError(ret, e)
// 	}

// 	return
// }

// func (p *Runtime) WithReadCommittedContext(fn func(*Context) error) error {
// 	return p.withContext(SqlLevelReadCommitted, fn)
// }

// func (p *Runtime) WithRepeatableReadContext(fn func(*Context) error) error {
// 	return p.withContext(SqlLevelRepeatableRead, fn)
// }

// func (p *Runtime) ExecHook(ctx *Context, hookName string, id string) error {
// 	if hookList, ok := p.hookMap[hookName]; ok {
// 		for _, hook := range hookList {
// 			if !hook.async {
// 				ctx.ttl--
// 				defer func() {
// 					ctx.ttl++
// 				}()

// 				var err error

// 				defer func() {
// 					if reason := recover(); reason != nil {
// 						err = WrapError(GetFirstError(err, Errorf("%s\nHook exec error: %s", hook.fileLine, reason)))
// 					}

// 					if err != nil {
// 						p.logger.Error(ErrorString(err))
// 					}
// 				}()

// 				if ctx.ttl < 0 {
// 					err = Errorf("%s: Hook exec %s too deep (maybe looping)", hook.fileLine, hookName)
// 				} else if e := hook.fn(ctx, id); e != nil {
// 					err = e
// 				} else {
// 					err = nil
// 				}

// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// func (p *Runtime) ExecListen(input *httpInput, hookName string, id string, ttl int) {
// 	if hookList, ok := p.hookMap[hookName]; ok {
// 		for _, hook := range hookList {
// 			if hook.async {
// 				if ttl < 1 {
// 					p.logger.Error(ErrorString(Errorf("%s: Listen exec %s too deep (maybe looping)", hook.fileLine, hookName)))
// 					return
// 				}

// 				listenCtx := &Context{
// 					ttl:      ttl - 1,
// 					input:    input,
// 					rt:       p,
// 					tx:       p.sqlMgr.NewTransaction(hook.tx, hook.readOnly),
// 					dirtyMap: nil,
// 				}

// 				var err error

// 				defer func() {
// 					if reason := recover(); reason != nil {
// 						err = WrapError(GetFirstError(err, Errorf("%s\nListen exec error: %s", hook.fileLine, reason)))
// 					}

// 					err = WrapError(GetFirstError(err, listenCtx.Close(err == nil)))

// 					if err != nil {
// 						p.logger.Error(ErrorString(err))
// 					}
// 				}()

// 				err = hook.fn(listenCtx, id)
// 			}
// 		}
// 	}
// }

// func (p *Runtime) SetUploadChecker(fn UploadCheckFunc) {
// 	if p.uploadChecker != nil {
// 		panic(Errorf("duplicated call SetUploadChecker"))
// 	}

// 	p.uploadChecker = fn
// }

// func (p *Runtime) SetDownloadChecker(fn DownloadCheckFunc) {
// 	if p.downloadChecker != nil {
// 		panic(Errorf("duplicated call SetDownloadChecker"))
// 	}

// 	p.downloadChecker = fn
// }

// func (p *Runtime) GetConfig() *RTConfig {
// 	return p.config
// }
