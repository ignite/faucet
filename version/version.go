package version

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"text/tabwriter"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/v48/github"
	"github.com/ignite/cli/v28/ignite/pkg/cmdrunner/exec"
	"github.com/ignite/cli/v28/ignite/pkg/cmdrunner/step"
	"github.com/ignite/cli/v28/ignite/pkg/xexec"
)

const (
	versionDev     = "development"
	versionNightly = "nightly"
)

// Version is the semantic version of Faucet.
var Version = versionDev

type Info struct {
	Version         string
	GoVersion       string
	BuildDate       string
	SourceHash      string
	OS              string
	Arch            string
	Uname           string
	CWD             string
	BuildFromSource bool
}

// CheckNext checks whether there is a new version of Faucet.
func CheckNext(ctx context.Context) (isAvailable bool, version string, err error) {
	if Version == versionDev || Version == versionNightly {
		return false, "", nil
	}

	tagName, err := getLatestReleaseTag(ctx)
	if err != nil {
		return false, "", err
	}

	currentVersion, err := semver.ParseTolerant(Version)
	if err != nil {
		return false, "", err
	}

	latestVersion, err := semver.ParseTolerant(tagName)
	if err != nil {
		return false, "", err
	}

	isAvailable = latestVersion.GT(currentVersion)

	return isAvailable, tagName, nil
}

func getLatestReleaseTag(ctx context.Context) (string, error) {
	latest, _, err := github.
		NewClient(nil).
		Repositories.
		GetLatestRelease(ctx, "ignite", "faucet")
	if err != nil {
		return "", err
	}

	if latest.TagName == nil {
		return "", nil
	}

	return *latest.TagName, nil
}

// fromSource check if the binary was build from source using the Faucet version.
func fromSource() bool {
	return Version == versionDev
}

// resolveDevVersion creates a string for version printing if the version being used is "development".
// the version will be of the form "LATEST-dev" where LATEST is the latest tagged release.
func resolveDevVersion(ctx context.Context) string {
	// do nothing if built with specific tag
	if Version != versionDev && Version != versionNightly {
		return Version
	}

	tag, err := getLatestReleaseTag(ctx)
	if err != nil {
		return Version
	}

	if Version == versionDev {
		return tag + "-dev"
	}
	if Version == versionNightly {
		return tag + "-nightly"
	}

	return Version
}

// Long generates a detailed version info.
func Long(ctx context.Context) (string, error) {
	var (
		w = &tabwriter.Writer{}
		b = &bytes.Buffer{}
	)

	info, err := GetInfo(ctx)
	if err != nil {
		return "", err
	}

	write := func(k, v string) {
		fmt.Fprintf(w, "%s:\t%s\n", k, v)
	}
	w.Init(b, 0, 8, 0, '\t', 0)

	write("Version", info.Version)
	write("Build date", info.BuildDate)
	write("Source hash", info.SourceHash)
	write("Your OS", info.OS)
	write("Your arch", info.Arch)
	write("Your go version", info.GoVersion)
	write("Your uname -a", info.Uname)

	if info.CWD != "" {
		write("Your cwd", info.CWD)
	}

	if err := w.Flush(); err != nil {
		return "", err
	}

	return b.String(), nil
}

// GetInfo gets the Faucet info.
func GetInfo(ctx context.Context) (Info, error) {
	var (
		info     Info
		modified bool

		date = "undefined"
		head = "undefined"
	)
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, kv := range buildInfo.Settings {
			switch kv.Key {
			case "vcs.revision":
				head = kv.Value
			case "vcs.time":
				date = kv.Value
			case "vcs.modified":
				modified = kv.Value == "true"
			}
		}
		if modified {
			// add * suffix to head to indicate the sources have been modified.
			head += "*"
		}
	}

	goVersionBuf := &bytes.Buffer{}
	if err := exec.Exec(ctx, []string{"go", "version"}, exec.StepOption(step.Stdout(goVersionBuf))); err != nil {
		return info, err
	}

	var (
		unameCmd = "uname"
		uname    = ""
	)
	if xexec.IsCommandAvailable(unameCmd) {
		unameBuf := &bytes.Buffer{}
		unameBuf.Reset()
		if err := exec.Exec(ctx, []string{unameCmd, "-a"}, exec.StepOption(step.Stdout(unameBuf))); err != nil {
			return info, err
		}
		uname = strings.TrimSpace(unameBuf.String())
	}

	info.Uname = uname
	info.Version = resolveDevVersion(ctx)
	info.BuildDate = date
	info.SourceHash = head
	info.OS = runtime.GOOS
	info.Arch = runtime.GOARCH
	info.GoVersion = strings.TrimSpace(goVersionBuf.String())
	info.BuildFromSource = fromSource()

	if cwd, err := os.Getwd(); err == nil {
		info.CWD = cwd
	}

	return info, nil
}
