package require

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

// NodeJS module search algorithm described by
// https://nodejs.org/api/modules.html#modules_all_together
func (r *RequireModule) resolve(path string) (resolvedPath string, err error) {
	origPath, path := path, filepathClean(path)
	if path == "" {
		return "", IllegalModuleNameError
	}

	resolvedPath, err = r.loadNative(path)
	if err == nil {
		return resolvedPath, nil
	}

	start := r.resolveStart
	if strings.HasPrefix(origPath, "/") {
		start = "/"
	}

	if strings.HasPrefix(origPath, "./") ||
		strings.HasPrefix(origPath, "/") || strings.HasPrefix(origPath, "../") ||
		origPath == "." || origPath == ".." {
		p := filepath.Join(start, path)
		resolvedPath, err = r.loadAsFileOrDirectory(p)
		if err == nil {
			return resolvedPath, nil
		}
		return "", InvalidModuleError
	}

	p := filepath.Dir(start)
	resolvedPath, err = r.loadNodeModules(path, p)
	if err == nil {
		return resolvedPath, nil
	}

	return "", InvalidModuleError
}

func (r *RequireModule) loadNative(path string) (string, error) {
	path = filepathClean(path)
	if _, exists := r.r.native[path]; exists {
		return path, nil
	}
	if _, exists := native[path]; exists {
		return path, nil
	}
	return "", InvalidModuleError
}

func (r *RequireModule) loadAsFileOrDirectory(path string) (string, error) {
	path = filepathClean(path)
	resolvedPath, err := r.loadAsFile(path)
	if err == nil {
		return resolvedPath, nil
	}
	resolvedPath, err = r.loadAsDirectory(path)
	if err == nil {
		return resolvedPath, nil
	}
	return "", InvalidModuleError
}

func (r *RequireModule) loadAsFile(path string) (string, error) {
	path = filepathClean(path)
	_, err := r.r.getSource(path)
	if err == nil {
		return path, nil
	}
	p := path + ".js"
	_, err = r.r.getSource(p)
	if err == nil {
		return p, nil
	}
	p = path + ".json"
	_, err = r.r.getSource(p)
	if err == nil {
		return p, nil
	}
	return "", InvalidModuleError
}

func (r *RequireModule) loadIndex(path string) (string, error) {
	path = filepathClean(path)
	p := filepath.Join(path, "index.js")
	_, err := r.r.getSource(p)
	if err == nil {
		return p, nil
	}
	p = filepath.Join(path, "index.json")
	_, err = r.r.getSource(p)
	if err == nil {
		return p, nil
	}
	return "", InvalidModuleError
}

func (r *RequireModule) loadAsDirectory(path string) (string, error) {
	path = filepathClean(path)
	p := filepath.Join(path, "package.json")
	buf, err := r.r.getSource(p)
	if err != nil {
		return r.loadIndex(path)
	}
	var pkg struct {
		Main string
	}
	err = json.Unmarshal(buf, &pkg)
	if err != nil || len(pkg.Main) == 0 {
		return r.loadIndex(path)
	}
	m := filepath.Join(path, pkg.Main)
	resolvedPath, err := r.loadAsFile(m)
	if err == nil {
		return resolvedPath, nil
	}
	resolvedPath, err = r.loadIndex(m)
	if err == nil {
		return resolvedPath, nil
	}
	return "", InvalidModuleError
}

func (r *RequireModule) loadNodeModules(path, start string) (string, error) {
	dirs, err := r.nodeModulesPaths(start)
	if err != nil {
		return "", err
	}
	for _, dir := range dirs {
		p := filepath.Join(dir, path)
		resolvedPath, err := r.loadAsFileOrDirectory(p)
		if err == nil {
			return resolvedPath, nil
		}
	}
	return "", InvalidModuleError
}

func (r *RequireModule) nodeModulesPaths(start string) ([]string, error) {
	dirs := r.r.globalFolders
	prev, dir := "", filepath.Dir(start)
	for prev != dir {
		if filepath.Base(dir) != "node_modules" {
			dirs = append(dirs, filepath.Join(dir, "node_modules"))
		}
		prev, dir = dir, filepath.Dir(dir)
	}
	return dirs, nil
}
