package project

import (
	"bytes"
	_ "embed"
	"github.com/FriendsOfShopware/shopware-cli/extension"
	"github.com/FriendsOfShopware/shopware-cli/logging"
	"github.com/FriendsOfShopware/shopware-cli/shop"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

//go:embed templates/Dockerfile.tpl
var dockerFileTemplate string

var dockerBuildCmd = &cobra.Command{
	Use:   "build [name]",
	Short: "Build Docker Image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shopCfg, err := shop.ReadConfig(projectConfigPath, true)
		if err != nil {
			return err
		}

		if shopCfg.Docker.PHP.PhpVersion == "" {
			projectRoot, err := os.Getwd()

			if err != nil {
				return err
			}

			constraint, err := extension.GetShopwareProjectConstraint(projectRoot)
			if err != nil {
				return err
			}

			phpVersion, err := extension.GetPhpVersion(cmd.Context(), constraint)

			if err != nil {
				return err
			}

			shopCfg.Docker.PHP.PhpVersion = phpVersion
			logging.FromContext(cmd.Context()).Infof("No PHP version set, using PHP version %s", phpVersion)
		}

		var buf bytes.Buffer

		if err := template.
			Must(template.New("Dockerfile").
				Parse(dockerFileTemplate)).
			Execute(&buf, *shopCfg.Docker); err != nil {
			return err
		}

		if err := os.WriteFile("Dockerfile", buf.Bytes(), os.ModePerm); err != nil {
			return err
		}

		shopCfg.Docker.ExcludePaths = append(shopCfg.Docker.ExcludePaths, "/var", ".git", "node_modules", ".idea")

		if err := os.WriteFile(".dockerignore", []byte(strings.Join(shopCfg.Docker.ExcludePaths, "\n")), os.ModePerm); err != nil {
			return err
		}

		return runTransparentCommand(exec.CommandContext(cmd.Context(), "docker", "build", "-t", args[0], "."))
	},
}

func init() {
	dockerRootCmd.AddCommand(dockerBuildCmd)
}
