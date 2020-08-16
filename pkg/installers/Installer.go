package installers

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iwind/TeaGo/files"
	"github.com/iwind/TeaGo/maps"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// 安装工具
type Installer struct {
	Dir    string
	Master string
	Id     string
	Key    string
	IP     string
}

func NewInstaller() *Installer {
	return &Installer{}
}

func (this *Installer) Start() (isInstalled bool, err error) {
	if len(this.Master) == 0 {
		return false, errors.New("'master' should not be empty")
	}

	if !strings.HasPrefix(this.Master, "http://") && !strings.HasPrefix(this.Master, "https://") {
		return false, errors.New("'master' should starts with 'http://' or 'https'")
	}

	if len(this.Dir) == 0 {
		return false, errors.New("'dir' should not be empty")
	}

	// 是否有别的版本正在运行
	{
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:7778/status", nil)
		if err != nil {
			return false, err
		}
		client := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return true, errors.New("agent already run with port '7778'")
			}
		}
	}

	// 检测IP
	{
		urlString := this.Master + "/api/agent/ip"
		req, err := http.NewRequest(http.MethodGet, urlString, nil)
		if err != nil {
			return false, err
		}
		client := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return false, errors.New("'" + urlString + "' respond a invalid status code:" + fmt.Sprintf("%d", resp.StatusCode))
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		m := maps.Map{}
		err = json.Unmarshal(data, &m)
		if err != nil {
			return false, err
		}

		this.IP = m.GetString("ip")
	}

	// 下载
	{
		urlString := this.Master + "/api/agent/upgrade"
		req, err := http.NewRequest(http.MethodGet, urlString, nil)
		if err != nil {
			return false, err
		}
		req.Header.Set("Tea-Agent-Id", this.Id)
		req.Header.Set("Tea-Agent-Key", this.Key)
		req.Header.Set("Tea-Agent-Version", "0.0.0")
		req.Header.Set("Tea-Agent-Os", runtime.GOOS)
		req.Header.Set("Tea-Agent-Arch", runtime.GOARCH)
		client := &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return false, errors.New("'" + urlString + "' respond a invalid status code:" + fmt.Sprintf("%d", resp.StatusCode))
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		if len(data) < 1024 {
			return false, errors.New(string(data))
		}

		// 目录是否存在
		targetDir := this.Dir + "/agent"
		for _, subDir := range []string{"bin", "configs", "configs/agents", "logs", "plugins"} {
			subFile := files.NewFile(targetDir + "/" + subDir)
			if subDir == "bin" && subFile.Exists() {
				err = subFile.DeleteAll() // 清空目录
				if err != nil {
					return false, err
				}
			}
			err = subFile.MkdirAll()
			if err != nil {
				return false, errors.New("mkdir '" + subDir + "' error:" + err.Error())
			}
		}

		// 写入可执行文件
		exeFile := files.NewFile(targetDir + "/bin/teaweb-agent")
		err = exeFile.Write(data)
		if err != nil {
			return false, err
		}
		exeFile.Chmod(0777)

		// 写入配置文件
		confFile := files.NewFile(targetDir + "/configs/agent.conf")
		err = confFile.WriteString(`master: ` + this.Master + `           # TeaWeb access address
id: ` + this.Id + `                    # Agent ID
key: ` + this.Key + `   # Agent Key
`)
		if err != nil {
			return false, err
		}

		// 启动
		cmd := exec.Command("bin/teaweb-agent", "start")
		cmd.Dir = targetDir
		err = cmd.Start()
		if err != nil {
			return false, err
		}

		// 返回成功
	}

	return true, nil
}
