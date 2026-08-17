package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/abdullayev4u/gc2/tools/ostools"
)

func LoadIcons(c *Gc2Cmd, wg *sync.WaitGroup) {
	defer wg.Done()

	if !c.Cfg.SyncDomainIcon {
		return
	}

	// With domainFolder off, or an absolute mapping target, there is no
	// per-host folder to decorate.
	if c.DomainFolderPath == "" {
		return
	}

	// make sure ~/.gc2 exists
	{
		path := filepath.Join(mustHomeDir(), ".gc2")
		err := os.MkdirAll(path, 0o755)
		if err != nil {
			fmt.Println("err init ~/.gc2 folder:", err.Error())
			return
		}

	}

	iconPath := loadDomainIcon(c)
	if iconPath == "" {
		// Every candidate URL failed; setting an empty icon would only
		// produce a confusing warning.
		return
	}

	err := ostools.SetCustomIcon(c.DomainFolderPath, iconPath)

	if err != nil {
		fmt.Printf("Warning: failed to set custom icon: %s\n", err.Error())
	}

}

func loadDomainIcon(c *Gc2Cmd) string {
	urls := []string{
		fmt.Sprintf("https://%s/assets/img/favicon.png", c.Repo_domain),
		fmt.Sprintf("https://%s/fluidicon.png", c.Repo_domain),
		fmt.Sprintf("https://%s/favicon.ico", c.Repo_domain),
	}

	for _, url := range urls {

		// 3. Download the image
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Failed to download icon for %s: %v\n", c.Repo_domain, err)
			continue
		}
		defer resp.Body.Close()

		iconPath := ""
		{
			ext := ".png"
			switch resp.Header.Get("Content-Type") {
			case "image/x-icon", "image/vnd.microsoft.icon":
				ext = ".ico"
			case "image/jpeg":
				ext = ".jpg"
			case "image/gif":
				ext = ".gif"
			case "image/png":
				ext = ".png"
			}

			fileName := fmt.Sprintf("%s_favicon%s", c.Repo_domain, ext)
			iconPath = filepath.Join(mustHomeDir(), ".gc2", fileName)

		}

		// 4. Create the file locally
		out, err := os.Create(iconPath)
		if err != nil {
			fmt.Printf("Failed to create file %s: %v\n", iconPath, err)
			continue
		}
		defer out.Close()

		// 5. Write the body to the file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			fmt.Printf("Failed to save icon data: %v\n", err)
			continue
		}

		return iconPath

	}

	return ""
}
