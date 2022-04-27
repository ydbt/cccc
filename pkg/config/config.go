package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"

	toml "github.com/pelletier/go-toml"
	yaml "gopkg.in/yaml.v2"
)

type SystemSignal struct {
	status int8
	exit   chan struct{}
	wg     sync.WaitGroup
	code   chan os.Signal
	msg    string
	ctx    context.Context
	cancel context.CancelFunc
}

func NewSystemSignal() *SystemSignal {
	ss := &SystemSignal{
		status: 0,
		code:   make(chan os.Signal),
		exit:   make(chan struct{}),
	}
	signal.Notify(ss.code, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	ss.ctx, ss.cancel = context.WithCancel(context.Background())
	return ss
}
func (ss *SystemSignal) RegistTask() (wait <-chan struct{}, close func()) {
	ss.wg.Add(1)
	wait = ss.exit
	close = func() {
		ss.wg.Done()
	}
	return
}
func (ss *SystemSignal) ContextExit() context.Context {
	return ss.ctx
}
func (ss *SystemSignal) NotifyExit() {
	ss.code <- syscall.SIGQUIT
}
func (ss *SystemSignal) WaitExit() {
	sg := <-ss.code
	ss.msg = fmt.Sprintf("process kill by signal: %s", sg.String())
	close(ss.exit)
	ss.cancel()
	ss.status = 1
}
func (ss *SystemSignal) Exited() bool {
	return ss.status == 1
}

// 配置初始之后的校验
type ConfigI interface {
	Normalize()
	RedepInit(*SystemSignal)
}

type ConfigServer struct {
	SS     *SystemSignal
	Module map[string]interface{}
}

func NewConfigServer(ctx context.Context) *ConfigServer {
	c := &ConfigServer{
		SS:     NewSystemSignal(),
		Module: make(map[string]interface{}),
	}
	return c
}

func (c *ConfigServer) Regist(name string, cfg interface{}) error {
	if _, ok := c.Module[name]; ok {
		return errors.New(fmt.Sprintf("%s already Regist", name))
	} else {
		c.Module[name] = cfg
	}
	fmt.Println("[Regist] config->", name)
	return nil
}

func (s *ConfigServer) LoadConfigFile(configFile string) error {
	pos := strings.LastIndex(configFile, ".")
	suffix := "yaml"
	if pos != -1 {
		suffix = configFile[pos+1:]
	}
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	s.parseConfig(suffix, data)
	return nil
}
func (s *ConfigServer) parseConfig(suffix string, data []byte) {
	normalize_config(suffix, data, s.Module)
	for _, v := range s.Module {
		if i, ok := v.(ConfigI); ok {
			i.Normalize()
			i.RedepInit(s.SS)
		}
	}
}

type normalizeConfig struct {
	OriSrc         map[interface{}]interface{}     // 原始配置解析的map结构
	IdxSrc         map[string]interface{}          // 根据键前缀快速查找值
	Modules        interface{}                     // 配置对象
	original_value func(keys []string) interface{} // 根据键前缀查找配置值
}

func normalize_config(suffix string, data []byte, cfg map[string]interface{}) {
	nc := &normalizeConfig{
		OriSrc:  make(map[interface{}]interface{}),
		IdxSrc:  make(map[string]interface{}),
		Modules: cfg,
	}
	nc.original_value = nc.original_value_v2
	switch suffix {
	case "json", "js":
		err := json.Unmarshal(data, &nc.OriSrc)
		fmt.Println("[normalize_config]jsonErr:", err)
	case "yaml", "yml":
		err := yaml.Unmarshal(data, &nc.OriSrc)
		fmt.Println("[normalize_config]yamlErr:", err)
	case "toml", "ini":
		err := toml.Unmarshal(data, &nc.OriSrc)
		fmt.Println("[normalize_config]tomlErr:", err)
	}
	nc.normalize_src([]string{}, nc.OriSrc)
	for k, v := range cfg {
		nc.normalize_merge([]string{k}, reflect.ValueOf(v))
	}
}

func (nc *normalizeConfig) new_prefixs(prefixA []string, prefixB ...string) []string {
	size := len(prefixA)
	t := make([]string, size+len(prefixB))
	copy(t, prefixA)
	copy(t[size:], prefixB)
	return t
}

func (nc *normalizeConfig) original_value_v2(keys []string) interface{} {
	key := strings.Join(keys, " ")
	if v, ok := nc.IdxSrc[key]; ok {
		return v
	} else {
		return nil
	}
}

func (nc *normalizeConfig) original_value_v1(keys []string) interface{} {
	m := nc.OriSrc
	var v interface{}
	ok := false
	for i, s_i := 0, len(keys); i < s_i; i++ {
		if v, ok = m[keys[i]]; ok {
			if m, ok = v.(map[interface{}]interface{}); !ok {
				return nil // 类型值不是结构体或者map结构
			}
			k := interface{}(keys[i])
			if v, ok = m[k]; !ok {
				return nil // 键匹配错误
			}
		} else {
			return nil
		}
	}
	return v
}
func (nc *normalizeConfig) normalize_src(prefixs []string, src map[interface{}]interface{}) {
	for k, v := range src {
		nc.IdxSrc[strings.Join(nc.new_prefixs(prefixs, k.(string)), " ")] = v
		if x, ok := v.(map[interface{}]interface{}); ok {
			nc.normalize_src(nc.new_prefixs(prefixs, k.(string)), x)
		} else if y, ok := v.([]interface{}); ok {
			for i, s_i := 0, len(y); i < s_i; i++ {
				if z, ok := y[i].(map[interface{}]interface{}); ok {
					nc.normalize_src(nc.new_prefixs(prefixs, k.(string), fmt.Sprint(i)), z)
				} else {
					nc.IdxSrc[strings.Join(nc.new_prefixs(prefixs, k.(string), fmt.Sprint(i)), " ")] = y[i]
				}
			}
		}
	}
}
func (nc *normalizeConfig) normalize_merge(prefixs []string, v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if !v.CanSet() {
		return
	}
	t := v.Type()
	// fmt.Println("normalize_merge", t, t.Name(), v)
	switch t.Kind() {
	case reflect.Struct:
		for i, s_i := 0, t.NumField(); i < s_i; i++ {
			f := t.Field(i)
			tag := f.Tag.Get("cccc")
			nc.normalize_merge(nc.new_prefixs(prefixs, tag), v.Field(i))
		}
	case reflect.Array:
		srcV := nc.original_value(prefixs)
		if srcV == nil {
			break
		}
		if y, ok := srcV.([]interface{}); !ok {
			break // 原始数据，数组断言失败
		} else {
			s_i := len(y)
			for i := 0; i < s_i; i++ {
				nc.normalize_merge(nc.new_prefixs(prefixs, fmt.Sprint(i)), v.Index(i))
			}
		}
	case reflect.Slice:
		srcV := nc.original_value(prefixs)
		if srcV == nil {
			break
		}
		if y, ok := srcV.([]interface{}); !ok {
			break // 原始数据，数组断言失败
		} else {
			s_i := len(y)
			x := reflect.MakeSlice(t, s_i, s_i)
			v.Set(x)
			for i := 0; i < s_i; i++ {
				nc.normalize_merge(nc.new_prefixs(prefixs, fmt.Sprint(i)), x.Index(i))
			}
		}
	case reflect.Map:
		srcV := nc.original_value(prefixs)
		if srcV == nil {
			break
		}
		if y, ok := srcV.(map[interface{}]interface{}); !ok {
			break // 原始数据，数组断言失败
		} else {
			x := reflect.MakeMap(t)
			for a, b := range y {
				b1 := reflect.ValueOf(b)
				nc.normalize_merge(nc.new_prefixs(prefixs, fmt.Sprint(a)), b1)
				x.SetMapIndex(reflect.ValueOf(a), b1)
			}
		}
	case reflect.String:
		srcV := nc.original_value(prefixs)
		if srcV == nil {
			break
		}
		v.Set(reflect.ValueOf(srcV))
	default:
		srcV := nc.original_value(prefixs)
		if srcV == nil {
			break
		}
		b := reflect.ValueOf(srcV)
		if t == b.Type() {
			v.Set(b)
		} else {
			v.Set(b.Convert(t))
		}
	}
}
