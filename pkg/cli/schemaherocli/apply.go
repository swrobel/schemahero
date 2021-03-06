package schemaherocli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/schemahero/schemahero/pkg/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Apply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply a spec to a database",
		Long:  `...`,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			// to support automaticenv, we can't use cobra required flags
			driver := v.GetString("driver")
			uri := v.GetString("uri")
			ddl := v.GetString("ddl")

			if driver == "" || uri == "" || ddl == "" {
				missing := []string{}
				if driver == "" {
					missing = append(missing, "driver")
				}
				if uri == "" {
					missing = append(missing, "uri")
				}
				if ddl == "" {
					missing = append(missing, "ddl")
				}

				return fmt.Errorf("missing required params: %v", missing)
			}

			fi, err := os.Stat(v.GetString("ddl"))
			if err != nil {
				return err
			}

			db := database.NewDatabase()
			if fi.Mode().IsDir() {
				commands := []string{}
				err := filepath.Walk(v.GetString("ddl"), func(path string, info os.FileInfo, err error) error {
					if info.IsDir() {
						return nil
					}

					f, err := os.Open(path)
					if err != nil {
						return err
					}
					defer f.Close()

					commands := []string{}
					scanner := bufio.NewScanner(f)
					for scanner.Scan() {
						commands = append(commands, scanner.Text())
					}

					if err := scanner.Err(); err != nil {
						return err
					}

					return nil
				})

				if err != nil {
					return err
				}

				if err := db.ApplySync(commands); err != nil {
					return err
				}

				return nil
			} else {
				f, err := os.Open(v.GetString("ddl"))
				if err != nil {
					return err
				}
				defer f.Close()

				commands := []string{}
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					commands = append(commands, scanner.Text())
				}

				if err := scanner.Err(); err != nil {
					return err
				}

				return db.ApplySync(commands)
			}
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")
	cmd.Flags().String("uri", "", "connection string uri to use")
	cmd.Flags().String("ddl", "", "filename or directory name containing the rendered DDL commands to execute")

	return cmd
}
