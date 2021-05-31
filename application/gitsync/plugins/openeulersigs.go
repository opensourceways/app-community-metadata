package plugins

import (
	"github.com/gin-gonic/gin"
	"github.com/opensourceways/app-community-metadata/application/gitsync"
	"io/ioutil"
	"os"
	"sigs.k8s.io/yaml"
)

const CommunityRepo = "https://gitee.com/openeuler/community"

type OpenEulerSigsPlugin struct {
	sigs []byte
}

func NewOpenEulerSigsPlugin() gitsync.Plugin {
	return &OpenEulerSigsPlugin{}
}

func (h *OpenEulerSigsPlugin) GetMeta() *gitsync.PluginMeta {
	return &gitsync.PluginMeta{
		Name:        "sigs",
		Group:       "openeuler",
		Description: "get all sigs information in openEuler community",
		Repos: []gitsync.GitMeta{
			{
				Repo:       CommunityRepo,
				Branch:     "master",
				SubModules: "recursive",
				Schema:     gitsync.Https,
				WatchFiles: []string{
					"sig/sigs.yaml",
				},
			},
		},
	}
}

func (h *OpenEulerSigsPlugin) Load(files map[string][]string) error {
	if files, ok := files[CommunityRepo]; ok {
		if len(files) > 0 {
			f, err := os.Open(files[0])
			if err != nil {
				return err
			}
			defer f.Close()
			bytes, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			h.sigs, err = yaml.YAMLToJSON(bytes)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *OpenEulerSigsPlugin) RegisterEndpoints(group *gin.RouterGroup) {
	group.GET("/overview", h.ReadSigsYaml)
}

func (h *OpenEulerSigsPlugin) ReadSigsYaml(c *gin.Context) {
	c.Data(200, "application/json", h.sigs)
}
