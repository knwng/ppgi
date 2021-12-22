package log

import (
	"os"
	"io"
	"fmt"
	"path"
	"runtime"

	log "github.com/sirupsen/logrus"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/zput/zxcTool/ztLog/zt_formatter"

)

func SetLog(logFile string, verbose bool) error {
	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetReportCaller(true)

	writer, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE, 0750)
	if err != nil {
		return err
	}
	log.SetOutput(io.MultiWriter(os.Stdout, writer))

	SetLogFormatter()

	return nil
}

func SetLogFormatter() {
	exampleFormatter := &zt_formatter.ZtFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
		Formatter: nested.Formatter{
			//HideKeys: true,
			FieldsOrder: []string{"component", "category"},
		},
	}

	log.SetFormatter(exampleFormatter)
}
