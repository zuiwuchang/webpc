package daemon

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"

	"gitlab.com/king011/webpc/configure"
	"gitlab.com/king011/webpc/logger"
	"gitlab.com/king011/webpc/shell"
)

// Run run as deamon
func Run(release bool) {
	if release {
		gin.SetMode(gin.ReleaseMode)
	}
	cnf := configure.Single().HTTP
	l, e := net.Listen(`tcp`, cnf.Addr)
	if e != nil {
		if ce := logger.Logger.Check(zap.FatalLevel, `listen error`); ce != nil {
			ce.Write(
				zap.Error(e),
			)
		}
		os.Exit(1)
	}
	if cnf.TLS() {
		if ce := logger.Logger.Check(zap.InfoLevel, `https work`); ce != nil {
			ce.Write(
				zap.String(`addr`, cnf.Addr),
			)
		}
		if !logger.Logger.OutConsole() {
			log.Println(`https work`, cnf.Addr)
		}
	} else {
		if ce := logger.Logger.Check(zap.InfoLevel, `http work`); ce != nil {
			ce.Write(
				zap.String(`addr`, cnf.Addr),
			)
		}
		if !logger.Logger.OutConsole() {
			log.Println(`http work`, cnf.Addr)
		}
	}
	shell.Single().Restore()

	router := newGIN()

	if cnf.TLS() {
		e = http.ServeTLS(l, router, cnf.CertFile, cnf.KeyFile)
	} else {
		e = http.Serve(l, router)
	}
	if ce := logger.Logger.Check(zap.FatalLevel, `serve error`); ce != nil {
		ce.Write(
			zap.Error(e),
		)
	}
	os.Exit(1)
}
