package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/runZeroInc/mustache/v2"
)

var rootCmd = &cobra.Command{
	Use: "mustache [--layout template] [data] template",
	Example: `  $ mustache data.yml template.mustache
  $ cat data.yml | mustache template.mustache
  $ mustache --layout wrapper.mustache data template.mustache
  $ mustache --override over.yml data.yml template.mustache`,
	Args: cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		err := run(cmd, args)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}
	},
}

var (
	layoutFile   string
	overrideFile string
)

func main() {
	rootCmd.Flags().StringVar(&layoutFile, "layout", "", "location of layout file")
	rootCmd.Flags().StringVar(&overrideFile, "override", "", "location of data.yml override yml")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Usage()
	}

	var data any
	var templatePath string
	if len(args) == 1 {
		var err error
		data, err = parseDataFromStdIn()
		if err != nil {
			return err
		}
		templatePath = args[0]
	} else {
		var err error
		data, err = parseDataFromFile(args[0])
		if err != nil {
			return err
		}
		templatePath = args[1]
	}

	if overrideFile != "" {
		override, err := parseDataFromFile(overrideFile)
		if err != nil {
			return err
		}
		for k, v := range override.(map[any]any) {
			data.(map[any]any)[k] = v
		}
	}
	var output string
	var oerr error
	if layoutFile != "" {
		tmpl, err := mustache.New().CompileFile(templatePath)
		if err != nil {
			return err
		}
		layl, err := mustache.New().CompileFile(layoutFile)
		if err != nil {
			return err
		}
		output, oerr = tmpl.RenderInLayout(layl, data)
	} else {
		tmpl, err := mustache.New().CompileFile(templatePath)
		if err != nil {
			return err
		}
		output, oerr = tmpl.Render(data)
	}
	if oerr != nil {
		return oerr
	}
	fmt.Print(output)
	return nil
}

func parseDataFromStdIn() (any, error) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	var data any
	if err := yaml.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func parseDataFromFile(filePath string) (any, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var data any
	if err := yaml.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return data, nil
}
