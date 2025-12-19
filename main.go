package main

import (
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

func main() {
	if len(os.Args) < 3 || os.Args[1] != "run" {
		fmt.Println("usage: orz run <package>")
		os.Exit(1)
	}

	goos := runtime.GOOS
	arch := runtime.GOARCH
	libc := getLibc()
	pkg := os.Args[2]

	firstLetter := string(pkg[0])
	baseURL := fmt.Sprintf("https://superarch.org/packages/%s/%s/%s/%s/%s", firstLetter, pkg, goos, arch, libc)

	resp, err := http.Get(baseURL + "/latest")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		fmt.Printf("package '%s' not found for %s/%s/%s\n", pkg, goos, arch, libc)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		fmt.Printf("error: server returned %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading response: %v\n", err)
		os.Exit(1)
	}

	version := strings.TrimSpace(string(body))

	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache", "orz", "packages", pkg, version)
	os.MkdirAll(cacheDir, 0755)

	pkgURL := fmt.Sprintf("%s/%s.txz", baseURL, version)
	pkgPath := filepath.Join(cacheDir, version+".txz")

	resp, err = http.Get(pkgURL)
	if err != nil {
		fmt.Printf("error downloading: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("error: download returned %d\n", resp.StatusCode)
		os.Exit(1)
	}

	out, err := os.Create(pkgPath)
	if err != nil {
		fmt.Printf("error creating file: %v\n", err)
		os.Exit(1)
	}

	io.Copy(out, resp.Body)
	out.Close()

	cmd := exec.Command("tar", "-xJf", pkgPath, "-C", cacheDir)
	if err := cmd.Run(); err != nil {
		fmt.Printf("error extracting: %v\n", err)
		os.Exit(1)
	}

	binPath := filepath.Join(cacheDir, pkg)
	syscall.Exec(binPath, []string{pkg}, os.Environ())
}
