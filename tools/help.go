package tools

import (
	"fmt"
	"os"

	"github.com/abdullayev4u/gc2/config"
)

func PrintHelp() {
	path, err := config.Path()
	if err != nil {
		path = "~/" + config.FileName
	}

	fmt.Println(`gc2 — clone a repository into a predictable place, then open it.

Usage:
  gc2 <repo-url> [options]      clone and open
  gc2 where <repo-url>          show where it would land, clone nothing
  gc2 config path               print the config file location
  gc2 config init [--force]     regenerate the config file
  gc2 help | version

Options:
  -d, --depth <n>               clone with --depth n; anything <= 0 clones full history
  -e, --editor <cmd>            editor to open the repo with; "none" skips opening

URLs may be https, ssh or scp-style, with any number of group segments:
  gc2 https://github.com/abdullayev4u/gc2.git
  gc2 git@gitlab.mycompany.com:f1/f2/f3/f4/repo.git

Layout:
  By default a repo lands at <root>/<host>/<group>/.../<repo>, mirroring the
  remote. Edit the config file to change root, drop the host folder, or map long
  group paths onto short local ones:

    gitlab.mycompany.com:
      root: ~/work
      paths:
        "/f1/f2/f3/f4": thatProject     # exact: repos directly in that group
        "/f1/f2/f3/f4/*": thatProject   # trailing *: that group and everything below
        "/company/*/backend": services  # interior *: exactly one segment

  An exact rule beats a "/*" rule, and the most specific rule wins. A target
  starting with ~ or / ignores root and the host folder.

Config file:
  ` + path + `
  (override with GC2_CONFIG)`)

	os.Exit(0)
}
