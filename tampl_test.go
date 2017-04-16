package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const testVars = `# tampl test variables
---
ssh:
  hosts:
    some.host:
      Port: 1234
      Compression: 'no'
      IdentityFile: ~/.ssh/id_ed25519
    another.host:
      User: andreas
`

const testTemplateSSH = `{{range $host, $values := .ssh.hosts -}}
Host {{$host}}
{{- range $k, $v := $values}}
	{{$k}} {{$v}}
{{- end}}

{{end}}`

const testTemplateSSHExpected = `Host some.host
	Port 1234
	Compression no
	IdentityFile ~/.ssh/id_ed25519

Host another.host
	User andreas

`

type testResources struct {
	source string
	target string
}

func newTestResources(t *testing.T) testResources {
	source, err := ioutil.TempDir("", "tampl-test-source")
	if err != nil {
		t.Fatal("failed to create test directory: %s", err)
	}
	target, err := ioutil.TempDir("", "tampl-test-target")
	if err != nil {
		t.Fatal("failed to create test directory: %s", err)
	}
	return testResources{source, target}
}

func (res testResources) writeFile(t *testing.T, name, data string) {
	fp := path.Join(res.source, name)
	if err := ioutil.WriteFile(fp, []byte(data), 0644); err != nil {
		t.Fatal("failed to create %q: %s", fp, err)
	}
}

func (res testResources) cleanup(t *testing.T) {
	err := os.RemoveAll(res.source)
	if err != nil {
		t.Logf("failed to remove %q: %s", res.source, err)
	}
	err = os.RemoveAll(res.target)
	if err != nil {
		t.Logf("failed to remove %q: %s", res.target, err)
	}
}

func TestIntegration(t *testing.T) {
	res := newTestResources(t)
	defer res.cleanup(t)

	var err error
	if err = run(res.source, res.target); err == nil {
		t.Fatal("expected to fail on empty source directory")
	}

	res.writeFile(t, VarsFile, testVars)
	if err = run(res.source, res.target); err == nil {
		t.Fatal("expected to fail on missing templates")
	}

	res.writeFile(t, "ssh-config."+TmplExt, testTemplateSSH)
	if err = run(res.source, res.target); err != nil {
		t.Fatal("failed with error: %s", err)
	}

	data, err := ioutil.ReadFile(path.Join(res.target, "ssh-config"))
	if err != nil {
		t.Fatalf("failed to read %q: %s", path.Join(res.target, "ssh-config"), err)
	}
	if string(data) != testTemplateSSHExpected {
		t.Fatalf(`expected:
%s
but was:
%s`, testTemplateSSHExpected, string(data))
	}
}
