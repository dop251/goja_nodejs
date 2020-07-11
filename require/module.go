package require

import (
	"text/template"

	js "github.com/dop251/goja"

	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

type ModuleLoader func(*js.Runtime, *js.Object)
type SourceLoader func(path string) ([]byte, error)

var (
	InvalidModuleError     = errors.New("Invalid module")
	IllegalModuleNameError = errors.New("Illegal module name")
)

var native map[string]ModuleLoader

// Registry contains a cache of compiled modules which can be used by multiple Runtimes
type Registry struct {
	sync.Mutex
	native   map[string]ModuleLoader
	compiled map[string]*js.Program

	srcLoader     SourceLoader
	globalFolders []string
}

type RequireModule struct {
	r            *Registry
	runtime      *js.Runtime
	modules      map[string]*js.Object
	resolveStart string
}

func NewRegistry(opts ...Option) *Registry {
	r := &Registry{}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func NewRegistryWithLoader(srcLoader SourceLoader) *Registry {
	return NewRegistry(WithLoader(srcLoader))
}

type Option func(*Registry)

func WithLoader(srcLoader SourceLoader) Option {
	return func(r *Registry) {
		r.srcLoader = srcLoader
	}
}

// WithGlobalFolders appends the given paths to the registry's list of
// global folders to search if the requested module is not found
// elsewhere.  By default, a registry's global folders list is empty.
// In the reference Node.js implementation, the default global folders
// list is $NODE_PATH, $HOME/.node_modules, $HOME/.node_libraries and
// $PREFIX/lib/node, see
// https://nodejs.org/api/modules.html#modules_loading_from_the_global_folders.
func WithGlobalFolders(globalFolders ...string) Option {
	return func(r *Registry) {
		r.globalFolders = globalFolders
	}
}

// Enable adds the require() function to the specified runtime.
func (r *Registry) Enable(runtime *js.Runtime) *RequireModule {
	rrt := &RequireModule{
		r:       r,
		runtime: runtime,
		modules: make(map[string]*js.Object),
	}

	runtime.Set("require", rrt.require)
	return rrt
}

func (r *Registry) RegisterNativeModule(name string, loader ModuleLoader) {
	r.Lock()
	defer r.Unlock()

	if r.native == nil {
		r.native = make(map[string]ModuleLoader)
	}
	name = filepathClean(name)
	r.native[name] = loader
}

func (r *Registry) getSource(p string) ([]byte, error) {
	srcLoader := r.srcLoader
	if srcLoader == nil {
		srcLoader = ioutil.ReadFile
	}
	return srcLoader(p)
}

func (r *Registry) getCompiledSource(p string) (prg *js.Program, err error) {
	r.Lock()
	defer r.Unlock()

	prg = r.compiled[p]
	if prg == nil {
		if buf, err1 := r.getSource(p); err1 == nil {
			s := string(buf)

			if filepath.Ext(p) == ".json" {
				s = "module.exports = JSON.parse('" + template.JSEscapeString(s) + "')"
			}

			source := "(function(exports, require, module) {" + string(s) + "\n})"
			prg, err = js.Compile(p, source, false)
			if err == nil {
				if r.compiled == nil {
					r.compiled = make(map[string]*js.Program)
				}
				r.compiled[p] = prg
			}
		} else {
			err = err1
		}
	}
	return
}

func (r *RequireModule) loadModule(path string, jsModule *js.Object) error {
	if ldr, exists := r.r.native[path]; exists {
		ldr(r.runtime, jsModule)
		return nil
	}

	if ldr, exists := native[path]; exists {
		ldr(r.runtime, jsModule)
		return nil
	}

	prg, err := r.r.getCompiledSource(path)

	if err != nil {
		return err
	}

	f, err := r.runtime.RunProgram(prg)
	if err != nil {
		return err
	}

	if call, ok := js.AssertFunction(f); ok {
		jsExports := jsModule.Get("exports")
		jsRequire := r.runtime.Get("require")

		origResolveStart := r.resolveStart
		r.resolveStart = filepath.Dir(path)
		defer func() { r.resolveStart = origResolveStart }()

		// Run the module source, with "jsExports" as "this",
		// "jsExports" as the "exports" variable, "jsRequire"
		// as the "require" variable and "jsModule" as the
		// "module" variable (Nodejs capable).
		_, err = call(jsExports, jsExports, jsRequire, jsModule)
		if err != nil {
			return err
		}
	} else {
		return InvalidModuleError
	}

	return nil
}

func (r *RequireModule) require(call js.FunctionCall) js.Value {
	ret, err := r.Require(call.Argument(0).String())
	if err != nil {
		panic(r.runtime.NewGoError(err))
	}
	return ret
}

func filepathClean(p string) string {
	return filepath.Clean(p)
}

// Require can be used to import modules from Go source (similar to JS require() function).
func (r *RequireModule) Require(p string) (ret js.Value, err error) {
	// TODO: if require() called outside of any other require()
	// calls, set resolve start path to
	// filepath.Dir(r.runtime.Program.src.name) (not currently
	// exposed).
	if r.resolveStart == "" {
		r.resolveStart = "."
		defer func() { r.resolveStart = "" }()
	}

	path, err := r.resolve(p)
	if err != nil {
		err = fmt.Errorf("Could not find module '%s': %v", p, err)
		return
	}
	module := r.modules[path]
	if module == nil {
		module = r.runtime.NewObject()
		module.Set("exports", r.runtime.NewObject())
		r.modules[path] = module
		err = r.loadModule(path, module)
		if err != nil {
			delete(r.modules, path)
			err = fmt.Errorf("Could not load module '%s': %v", p, err)
			return
		}
	}
	ret = module.Get("exports")
	return
}

func Require(runtime *js.Runtime, name string) js.Value {
	if r, ok := js.AssertFunction(runtime.Get("require")); ok {
		mod, err := r(js.Undefined(), runtime.ToValue(name))
		if err != nil {
			panic(err)
		}
		return mod
	}
	panic(runtime.NewTypeError("Please enable require for this runtime using new(require.Require).Enable(runtime)"))
}

func RegisterNativeModule(name string, loader ModuleLoader) {
	if native == nil {
		native = make(map[string]ModuleLoader)
	}
	name = filepathClean(name)
	native[name] = loader
}
