package pureftpd

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/go-rat/chix"
	"github.com/spf13/cast"

	"github.com/TheTNB/panel/internal/app"
	"github.com/TheTNB/panel/internal/service"
	"github.com/TheTNB/panel/pkg/firewall"
	"github.com/TheTNB/panel/pkg/io"
	"github.com/TheTNB/panel/pkg/shell"
	"github.com/TheTNB/panel/pkg/systemctl"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// List 获取用户列表
func (s *Service) List(w http.ResponseWriter, r *http.Request) {
	listRaw, err := shell.Execf("pure-pw list")
	if err != nil {
		service.Success(w, chix.M{
			"total": 0,
			"items": []User{},
		})
	}

	listArr := strings.Split(listRaw, "\n")
	var users []User
	for _, v := range listArr {
		if len(v) == 0 {
			continue
		}

		match := regexp.MustCompile(`(\S+)\s+(\S+)`).FindStringSubmatch(v)
		users = append(users, User{
			Username: match[1],
			Path:     strings.Replace(match[2], "/./", "/", 1),
		})
	}

	paged, total := service.Paginate(r, users)

	service.Success(w, chix.M{
		"total": total,
		"items": paged,
	})
}

// Create 创建用户
func (s *Service) Create(w http.ResponseWriter, r *http.Request) {
	req, err := service.Bind[Create](r)
	if err != nil {
		service.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	if !strings.HasPrefix(req.Path, "/") {
		req.Path = "/" + req.Path
	}
	if !io.Exists(req.Path) {
		service.Error(w, http.StatusUnprocessableEntity, "目录不存在")
		return
	}

	if err = io.Chmod(req.Path, 0755); err != nil {
		service.Error(w, http.StatusUnprocessableEntity, "修改目录权限失败")
		return
	}
	if err = io.Chown(req.Path, "www", "www"); err != nil {
		service.Error(w, http.StatusUnprocessableEntity, "修改目录权限失败")
		return
	}
	if _, err = shell.Execf(`yes '%s' | pure-pw useradd '%s' -u www -g www -d '%s'`, req.Password, req.Username, req.Path); err != nil {
		service.Error(w, http.StatusInternalServerError, "%v", err)
		return
	}
	if _, err = shell.Execf("pure-pw mkdb"); err != nil {
		service.Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	service.Success(w, nil)
}

// Delete 删除用户
func (s *Service) Delete(w http.ResponseWriter, r *http.Request) {
	req, err := service.Bind[Delete](r)
	if err != nil {
		service.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	if _, err = shell.Execf("pure-pw userdel '%s' -m", req.Username); err != nil {
		service.Error(w, http.StatusInternalServerError, "%v", err)
		return
	}
	if _, err = shell.Execf("pure-pw mkdb"); err != nil {
		service.Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	service.Success(w, nil)
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(w http.ResponseWriter, r *http.Request) {
	req, err := service.Bind[ChangePassword](r)
	if err != nil {
		service.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	if _, err = shell.Execf(`yes '%s' | pure-pw passwd '%s' -m`, req.Password, req.Username); err != nil {
		service.Error(w, http.StatusInternalServerError, "%v", err)
		return
	}
	if _, err = shell.Execf("pure-pw mkdb"); err != nil {
		service.Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	service.Success(w, nil)
}

// GetPort 获取端口
func (s *Service) GetPort(w http.ResponseWriter, r *http.Request) {
	port, err := shell.Execf(`cat %s/server/pure-ftpd/etc/pure-ftpd.conf | grep "Bind" | awk '{print $2}' | awk -F "," '{print $2}'`, app.Root)
	if err != nil {
		service.Error(w, http.StatusInternalServerError, "获取PureFtpd端口失败")
		return
	}

	service.Success(w, cast.ToInt(port))
}

// UpdatePort 设置端口
func (s *Service) UpdatePort(w http.ResponseWriter, r *http.Request) {
	req, err := service.Bind[UpdatePort](r)
	if err != nil {
		service.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	if _, err = shell.Execf(`sed -i "s/Bind.*/Bind 0.0.0.0,%d/g" %s/server/pure-ftpd/etc/pure-ftpd.conf`, req.Port, app.Root); err != nil {
		service.Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	fw := firewall.NewFirewall()
	err = fw.Port(firewall.FireInfo{
		PortStart: req.Port,
		PortEnd:   req.Port,
		Protocol:  "tcp",
	}, firewall.OperationAdd)
	if err != nil {
		service.Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	if err = systemctl.Restart("pure-ftpd"); err != nil {
		service.Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	service.Success(w, nil)
}
