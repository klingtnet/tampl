package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"gopkg.in/yaml.v2"
)

// VarsFile is the file name of the YAML defining all template variables
const VarsFile = "_vars.yml"

// Usage message
const Usage = `%s <source> <target>
	source: a directory containing a number of '.tmpl' files and '%s' which stores the variables and values.
	target: the directory where the rendered templates should be placed into.
`

// TmplExt contains the file extension used for template files
const TmplExt = "tmpl"

// Vars is an type alias for the YAML unmarshalling output
type Vars map[string]interface{}

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}

	sourceDir := os.Args[1]
	targetDir := os.Args[2]

	if err := run(sourceDir, targetDir); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func usage() {
	fmt.Printf(Usage, os.Args[0], VarsFile)
}

func run(sourceDir, targetDir string) error {
	varsPath := path.Join(sourceDir, VarsFile)
	vars, err := varsFromFile(varsPath)
	if err != nil {
		return err
	}

	tmpls, err := templates(sourceDir)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(tmpls.Templates()))
	var mtx sync.Mutex
	fails := make([]string, 0)

	for _, tmpl := range tmpls.Templates() {
		outFilepath := path.Join(targetDir, strings.TrimSuffix(tmpl.Name(), "."+TmplExt))
		go func(fpath string, tmpl *template.Template, vars Vars) {
			if err = renderToFile(fpath, tmpl, vars); err != nil {
				fmt.Fprintf(os.Stderr, "failed to render: %q", fpath)
				mtx.Lock()
				fails = append(fails, fpath)
				mtx.Unlock()
			}
			wg.Done()
		}(outFilepath, tmpl, vars)
	}
	wg.Wait()

	if len(fails) != 0 {
		return fmt.Errorf("failed to render templates: %v", fails)
	}
	return nil
}

func templates(dir string) (*template.Template, error) {
	tmplPath := filepath.Join(dir, "*."+TmplExt)
	matches, err := filepath.Glob(tmplPath)
	if err != nil {
		return &template.Template{}, fmt.Errorf("failed to list templates %q: %s", tmplPath, err)
	}
	if len(matches) == 0 {
		return &template.Template{}, fmt.Errorf("no template file found in %q", dir)
	}

	tmpls, err := template.ParseFiles(matches...)
	if err != nil {
		return &template.Template{}, err
	}
	return tmpls, nil
}

func varsFromFile(filepath string) (Vars, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read variables file %q: %s\n", filepath, err)
	}
	return varsFromBytes(data)
}

func varsFromBytes(data []byte) (Vars, error) {
	var vars Vars
	err := yaml.Unmarshal(data, &vars)
	if err != nil {
		return nil, err
	}
	return vars, nil
}

func render(w io.Writer, tmpl *template.Template, vars Vars) error {
	return tmpl.Execute(w, vars)
}

func renderToFile(filepath string, tmpl *template.Template, vars Vars) error {
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create template file %q: %s\n", filepath, err)
	}
	return render(f, tmpl, vars)
}
