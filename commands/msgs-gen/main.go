package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

var tplTest = template.Must(template.New("").Parse(
	`// Package {{ .PkgName }} contains message definitions (autogenerated).
package {{ .PkgName }}

import (
	"testing"
)

func TestCompileOk(t *testing.T) {
}
`))

func camelToSnake(in string) string {
	tmp := []rune(in)
	tmp[0] = unicode.ToLower(tmp[0])
	for i := 0; i < len(tmp); i++ {
		if unicode.IsUpper(tmp[i]) {
			tmp[i] = unicode.ToLower(tmp[i])
			tmp = append(tmp[:i], append([]rune{'_'}, tmp[i:]...)...)
		}
	}
	return string(tmp)
}

func shellCommand(cmdstr string) error {
	fmt.Fprintf(os.Stderr, "%s\n", cmdstr)
	cmd := exec.Command("sh", "-c", cmdstr)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func downloadJson(addr string, data interface{}) error {
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(data)
}

type githubFile struct {
	Name        string
	Url         string
	DownloadUrl string `json:"download_url"`
}

func processPackage(name string, addr string) error {
	fmt.Fprintf(os.Stderr, "[%s]\n", name)

	os.Mkdir(filepath.Join("msgs", name), 0755)

	err := func() error {
		f, err := os.Create(filepath.Join("msgs", name, "package_test.go"))
		if err != nil {
			return err
		}
		defer f.Close()

		return tplTest.Execute(f, map[string]interface{}{
			"PkgName": name,
		})
	}()
	if err != nil {
		return err
	}

	var files []githubFile
	err = downloadJson(addr, &files)
	if err != nil {
		return err
	}

	for _, f := range files {
		err := shellCommand(fmt.Sprintf("go run ./commands/msg-import --gopackage=%s --rospackage=%s %s > %s",
			name,
			name,
			f.DownloadUrl,
			filepath.Join("msgs", name, camelToSnake(strings.TrimSuffix(f.Name, ".msg"))+".go")))
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stderr, "\n")
	return nil
}

func processCommonMsgs() error {
	var files []githubFile
	err := downloadJson("https://api.github.com/repos/ros/common_msgs/contents", &files)
	if err != nil {
		return err
	}

	var packages [][2]string

	// get all folders which have the subfolder msg
	for _, f := range files {
		var subfiles []githubFile
		err := downloadJson(f.Url, &subfiles)
		if err != nil {
			return err
		}

		msgDir := func() string {
			for _, f := range subfiles {
				if f.Name == "msg" {
					return f.Url
				}
			}
			return ""
		}()

		if msgDir == "" {
			continue
		}

		packages = append(packages, [2]string{f.Name, msgDir})
	}

	for _, p := range packages {
		err := processPackage(p[0], p[1])
		if err != nil {
			return err
		}
	}

	return nil
}

func run() error {
	err := shellCommand("rm -rf msgs/*/")
	if err != nil {
		return err
	}

	err = processPackage("std_msgs", "https://api.github.com/repos/ros/std_msgs/contents/msg")
	if err != nil {
		return err
	}

	err = processPackage("rosgraph_msgs", "https://api.github.com/repos/ros/ros_comm_msgs/contents/rosgraph_msgs/msg")
	if err != nil {
		return err
	}

	err = processCommonMsgs()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
		os.Exit(1)
	}
}