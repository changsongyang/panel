package apps

import (
	"github.com/go-chi/chi/v5"

	_ "github.com/TheTNB/panel/internal/apps/benchmark"
	_ "github.com/TheTNB/panel/internal/apps/docker"
	_ "github.com/TheTNB/panel/internal/apps/fail2ban"
	_ "github.com/TheTNB/panel/internal/apps/frp"
	_ "github.com/TheTNB/panel/internal/apps/gitea"
	_ "github.com/TheTNB/panel/internal/apps/memcached"
	_ "github.com/TheTNB/panel/internal/apps/mysql"
	_ "github.com/TheTNB/panel/internal/apps/nginx"
	_ "github.com/TheTNB/panel/internal/apps/php"
	_ "github.com/TheTNB/panel/internal/apps/phpmyadmin"
	_ "github.com/TheTNB/panel/internal/apps/podman"
	_ "github.com/TheTNB/panel/internal/apps/postgresql"
	_ "github.com/TheTNB/panel/internal/apps/pureftpd"
	_ "github.com/TheTNB/panel/internal/apps/redis"
	_ "github.com/TheTNB/panel/internal/apps/rsync"
	_ "github.com/TheTNB/panel/internal/apps/s3fs"
	_ "github.com/TheTNB/panel/internal/apps/supervisor"
	_ "github.com/TheTNB/panel/internal/apps/toolbox"
	"github.com/TheTNB/panel/pkg/apploader"
)

func Boot(r chi.Router) {
	apploader.Boot(r)
}
