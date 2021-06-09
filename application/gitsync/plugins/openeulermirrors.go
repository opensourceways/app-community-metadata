package plugins

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opensourceways/app-community-metadata/application/gitsync"
	"io/ioutil"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

const InfrastructureRepo = "https://gitee.com/openeuler/infrastructure"

type OpenEulerMirrorsPlugin struct {
	repos []string
}

func NewOpenEulerMirrorsPlugin() gitsync.Plugin {
	return &OpenEulerMirrorsPlugin{}
}

func (h *OpenEulerMirrorsPlugin) GetMeta() *gitsync.PluginMeta {
	return &gitsync.PluginMeta{
		Name:        "mirrors",
		Group:       "openeuler",
		Description: "get all openeuler mirror information",
		Repos: []gitsync.GitMeta{
			{
				Repo:       InfrastructureRepo,
				Branch:     "master",
				SubModules: "recursive",
				Schema:     gitsync.Https,
				WatchFiles: []string{
					"mirrors",
				},
			},
		},
	}
}

func (h *OpenEulerMirrorsPlugin) Load(files map[string][]string) error {
	mirrors := []string{}
	if files, ok := files[InfrastructureRepo]; ok {
		if len(files) > 0 {
			//walk the yaml file to collect all mirror sites
			err := filepath.Walk(files[0], func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.Mode().IsRegular() {
					return nil
				}

				if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
					f, err := os.Open(path)
					if err != nil {
						return err
					}
					defer f.Close()
					bytes, err := ioutil.ReadAll(f)
					if err != nil {
						return err
					}
					m, err := yaml.YAMLToJSON(bytes)
					if err != nil {
						return err
					}
					mirrors = append(mirrors, string(m))
				}
				return nil
			})
			if err != nil {
				return err
			}
			h.repos = mirrors
		}
	}
	return nil
}

func (h *OpenEulerMirrorsPlugin) RegisterEndpoints(group *gin.RouterGroup) {
	group.GET("/all", h.ReadMirrorYamls)
}

func (h *OpenEulerMirrorsPlugin) ReadMirrorYamls(c *gin.Context) {
	if len(h.repos) == 0 {
		c.Data(200, "application/json", []byte("[]"))
	} else {
		c.Data(200, "application/json", []byte(fmt.Sprintf("[%s]",
			strings.Join(h.repos, ","))))
	}

}
