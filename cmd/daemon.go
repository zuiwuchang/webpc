package cmd

import (
	"log"

	"gitlab.com/king011/webpc/cmd/daemon"
	"gitlab.com/king011/webpc/configure"
	"gitlab.com/king011/webpc/cookie"
	"gitlab.com/king011/webpc/db/manipulator"
	"gitlab.com/king011/webpc/logger"
	"gitlab.com/king011/webpc/mount"
	"gitlab.com/king011/webpc/utils"

	"github.com/spf13/cobra"
)

func init() {
	var filename string
	var release bool
	basePaht := utils.BasePath()
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "run as daemon",
		Run: func(cmd *cobra.Command, args []string) {
			// load configure
			cnf := configure.Single()
			e := cnf.Load(basePaht, filename)
			if e != nil {
				log.Fatalln(e)
			}
			e = cnf.Format()
			if e != nil {
				log.Fatalln(e)
			}

			// init logger
			e = logger.Init(basePaht, &cnf.Logger)
			if e != nil {
				log.Fatalln(e)
			}

			// init cookie
			e = cookie.Init(cnf.Cookie.Filename, cnf.Cookie.MaxAge)
			if e != nil {
				log.Fatalln(e)
			}
			// init db
			e = manipulator.Init(cnf.System.DB)
			if e != nil {
				log.Fatalln(e)
			}
			// init mount
			e = mount.Init(cnf.System.Mount)
			if e != nil {
				log.Fatalln(e)
			}
			// run
			daemon.Run(release)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&filename, "config",
		"c",
		utils.Abs(basePaht, "webpc.jsonnet"),
		"configure file",
	)
	flags.BoolVarP(&release, "release",
		"r",
		false,
		"run as release",
	)

	rootCmd.AddCommand(cmd)
}
