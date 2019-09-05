package functions

import (
	"database/sql"
	"fmt"
	"github.com/op/go-logging"
	"github.com/vjeantet/jodaTime"
	"net"
	"os"
	"time"
)

func GetIpAdd() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		_, _ = os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}
	var ipadd = ""
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipadd = ipnet.IP.String()
			}
		}
	}
	return ipadd
}

//noinspection SqlDialectInspection,SqlNoDataSourceInspection
func InsertLogCron(scheduler string, message string, conn *sql.DB) {
	stmt, err := conn.Prepare("INSERT INTO logcron(scheduler,ip_address, message,tgl_proses) VALUES(?,?,?,?)")
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
	ipAdd := GetIpAdd()
	tglProses := jodaTime.Format("YYYY-MM-dd HH:mm:ss", time.Now())
	stmt.Exec(scheduler, ipAdd, message, tglProses)
}

func ParseTimeScheduler(Time string) string {
	t, err := time.Parse("15:04", Time)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return t.Format("15:04")
}

func Logger() *logging.Logger {
	var log = logging.MustGetLogger("main")
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ► %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)
	return log
}
