package cmd

import (
	"fmt"
	"runtime"

	"gitlab.com/king011/webpc/version"

	"github.com/spf13/cobra"
)

const (
	// App .
	App = "webpc"
)

var v bool
var rootCmd = &cobra.Command{
	Use:   App,
	Short: "generate tools create full",
	Run: func(cmd *cobra.Command, args []string) {
		if v {
			fmt.Println(runtime.GOOS, runtime.GOARCH, runtime.Version())
			fmt.Println(version.Tag)
			fmt.Println(version.Commit)
			fmt.Println(version.Date)
		} else {
			fmt.Println(App)
			fmt.Println(runtime.GOOS, runtime.GOARCH, runtime.Version())
			fmt.Println(version.Tag)
			fmt.Println(version.Commit)
			fmt.Println(version.Date)
			fmt.Printf(`Use "%v --help" for more information about this program.
`, App)
		}
	},
}

func init() {
	flags := rootCmd.Flags()
	flags.BoolVarP(&v,
		"version",
		"v",
		false,
		"show version",
	)
}

// Execute run command
func Execute() error {
	return rootCmd.Execute()
}
