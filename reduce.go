package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/antonmedv/fx/internal/engine"
)

//go:embed npm/index.js
var src []byte

var runtimes = [3]string{"bun", "node", "deno"}

var runconfig = map[string]struct{ args, env []string }{
	"node": {args: []string{}, env: []string{"NODE_OPTIONS=--max-old-space-size=16384"}},
	"deno": {args: []string{"run", "-A"}, env: []string{"V8_FLAGS=--max-old-space-size=16384"}},
	"bun":  {args: []string{"run", "-bi", "--smol"}, env: []string{"BUN_JSC_forceRAMSize=17179869184"}},
}

func reduce(fns []string) {
	script := path.Join(os.TempDir(), fmt.Sprintf("fx-%v.js", version))
	_, err := os.Stat(script)
	if os.IsNotExist(err) {
		err = os.WriteFile(script, src, 0644)
		if err != nil {
			panic(err)
		}
	}

	var bin string
	env := os.Environ()
	args := append([]string{script}, fns...)
	for _, runtime := range runtimes {
		bin, err = exec.LookPath(runtime)
		if err == nil {
			run := runconfig[runtime]
			env = append(run.env, env...)
			args = append(run.args, args...)
			break
		}
	}

	if err != nil {
		engine.Reduce(fns)
		return
	}

	cmd := exec.Command(bin, args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	switch err := cmd.Run().(type) {
	case nil:
		os.Exit(0)
	case *exec.ExitError:
		os.Exit(err.ExitCode())
	default:
		panic(err)
	}
}
