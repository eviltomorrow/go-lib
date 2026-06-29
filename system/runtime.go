package system

import (
	"time"

	"github.com/eviltomorrow/go-lib/timeutil"
	jsoniter "github.com/json-iterator/go"
)

var startup = time.Now()

var (
	LaunchTime = func() string {
		return timeutil.FormatDuration(time.Since(startup))
	}
	Machine   machine
	Network   network
	Process   process
	Directory directory
)

type machine struct {
	hostname string
}

func (m machine) Hostname() string {
	return m.hostname
}

type network struct {
	accessIP string
	bindIP   string
}

func (n network) AccessIP() string {
	return n.accessIP
}

func (n *network) SetAccessIP(value string) {
	n.accessIP = value
}

func (n network) BindIP() string {
	return n.bindIP
}

func (n *network) SetBindIP(value string) {
	n.bindIP = value
}

type process struct {
	name string
	args []string

	pid  int
	ppid int
}

func (p process) Name() string {
	return p.name
}

func (p process) Args() []string {
	return p.args
}

func (p process) Pid() int {
	return p.pid
}

func (p process) PPid() int {
	return p.ppid
}

type directory struct {
	rootDir string

	etcDir string
	appDir string
	usrDir string
	varDir string
	logDir string
	boxDir string

	execDir string
}

func (d directory) EtcDir() string {
	return d.etcDir
}

func (d *directory) SetEtcDir(dir string) {
	d.etcDir = dir
}

func (d directory) BoxDir() string {
	return d.boxDir
}

func (d directory) RootDir() string {
	return d.rootDir
}

func (d directory) AppDir() string {
	return d.appDir
}

func (d directory) UsrDir() string {
	return d.usrDir
}

func (d directory) VarDir() string {
	return d.varDir
}

func (d directory) LogDir() string {
	return d.logDir
}

func (d directory) ExecDir() string {
	return d.execDir
}

func (d *directory) SetAppDir(dir string) {
	d.appDir = dir
}

func (d *directory) SetUsrDir(dir string) {
	d.usrDir = dir
}

func (d *directory) SetVarDir(dir string) {
	d.varDir = dir
}

func (d *directory) SetLogDir(dir string) {
	d.logDir = dir
}

func (d *directory) SetBoxDir(dir string) {
	d.boxDir = dir
}

func String() string {
	data := map[string]interface{}{
		"machine": struct {
			Hostname string `json:"hostname"`
		}{
			Hostname: Machine.hostname,
		},

		"network": struct {
			AccessIP string `json:"access_ip"`
			BindIP   string `json:"bind_ip"`
		}{
			AccessIP: Network.accessIP,
			BindIP:   Network.bindIP,
		},

		"process": struct {
			Name string   `json:"name"`
			Args []string `json:"args"`
			Pid  int      `json:"pid"`
			Ppid int      `json:"ppid"`
		}{
			Name: Process.name,
			Args: Process.args,
			Pid:  Process.pid,
			Ppid: Process.ppid,
		},

		"directory": struct {
			RootDir string `json:"root_dir"`
			AppDir  string `json:"app_dir"`
			UsrDir  string `json:"usr_dir"`
			VarDir  string `json:"var_dir"`
			LogDir  string `json:"log_dir"`
			BoxDir  string `json:"box_dir"`
			ExecDir string `json:"exec_dir"`
		}{
			RootDir: Directory.rootDir,
			AppDir:  Directory.appDir,
			UsrDir:  Directory.usrDir,
			VarDir:  Directory.varDir,
			LogDir:  Directory.logDir,
			BoxDir:  Directory.boxDir,
			ExecDir: Directory.execDir,
		},
	}

	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(data)
	return string(buf)
}
