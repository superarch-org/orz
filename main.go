package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

func getLibc() string {
	switch runtime.GOOS {
	case "linux":
		return "glibc"
	case "netbsd":
		return "libc-12"
	case "freebsd":
		return "libc-7"
	default:
		return "unknown"
	}
}

func getLatest(pkg string) (string, error) {
	first := string(pkg[0])
	url := fmt.Sprintf("https://superarch.org/packages/%s/%s/latest", first, pkg)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("package '%s' not found", pkg)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("server returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

func getDeps(pkg string, version string) (map[string]string, error) {
	first := string(pkg[0])
	url := fmt.Sprintf("https://superarch.org/packages/%s/%s/%s.json", first, pkg, version)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// no deps
		return make(map[string]string), nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var data struct {
		Deps map[string]string `json:"deps"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Deps == nil {
		return make(map[string]string), nil
	}
	return data.Deps, nil
}

func installPackage(pkg string, version string) (string, error) {
	goos := runtime.GOOS
	arch := runtime.GOARCH
	libc := getLibc()
	firstLetter := string(pkg[0])

	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache", "orz", "packages", pkg, version)
	os.MkdirAll(cacheDir, 0755)

	pkgURL := fmt.Sprintf("https://superarch.org/packages/%s/%s/%s/%s/%s/%s.txz",
		firstLetter, pkg, goos, arch, libc, version)
	pkgPath := filepath.Join(cacheDir, version+".txz")

	resp, err := http.Get(pkgURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download returned %d", resp.StatusCode)
	}

	out, err := os.Create(pkgPath)
	if err != nil {
		return "", err
	}

	io.Copy(out, resp.Body)
	out.Close()

	cmd := exec.Command("tar", "-xJf", pkgPath, "-C", cacheDir)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error extracting: %v", err)
	}

	return cacheDir, nil
}

func main() {
	if len(os.Args) < 3 || os.Args[1] != "run" {
		fmt.Println("usage: orz run <package>")
		os.Exit(1)
	}

	pkg := os.Args[2]

	version, err := getLatest(pkg)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	deps, err := getDeps(pkg, version)
	if err != nil {
		fmt.Printf("error getting deps: %v\n", err)
		os.Exit(1)
	}

	var libPaths []string

	for depPkg, depVersion := range deps {
		depDir, err := installPackage(depPkg, depVersion)
		if err != nil {
			fmt.Printf("error installing dep %s: %v\n", depPkg, err)
			os.Exit(1)
		}
		libPaths = append(libPaths, filepath.Join(depDir, "lib"))
	}

	pkgDir, err := installPackage(pkg, version)
	if err != nil {
		fmt.Printf("error installing %s: %v\n", pkg, err)
		os.Exit(1)
	}

	binPath := filepath.Join(pkgDir, pkg)
	env := os.Environ()
	if len(libPaths) > 0 {
		ldPath := strings.Join(libPaths, ":")
		env = append(env, "LD_LIBRARY_PATH="+ldPath)
	}

	syscall.Exec(binPath, []string{pkg}, env)
}
