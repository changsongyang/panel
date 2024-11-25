package data

import (
	"fmt"
	"slices"

	"github.com/samber/do/v2"

	"github.com/TheTNB/panel/internal/app"
	"github.com/TheTNB/panel/internal/biz"
	"github.com/TheTNB/panel/internal/http/request"
	"github.com/TheTNB/panel/pkg/db"
)

type databaseRepo struct{}

func NewDatabaseRepo() biz.DatabaseRepo {
	return do.MustInvoke[biz.DatabaseRepo](injector)
}

func (r databaseRepo) List(page, limit uint) ([]*biz.Database, int64, error) {
	var databaseServer []*biz.DatabaseServer
	if err := app.Orm.Model(&biz.DatabaseServer{}).Order("id desc").Find(&databaseServer).Error; err != nil {
		return nil, 0, err
	}

	database := make([]*biz.Database, 0)
	for _, server := range databaseServer {
		switch server.Type {
		case biz.DatabaseTypeMysql:
			mysql, err := db.NewMySQL(server.Username, server.Password, fmt.Sprintf("%s:%d", server.Host, server.Port))
			if err == nil {
				if databases, err := mysql.Databases(); err == nil {
					for item := range slices.Values(databases) {
						database = append(database, &biz.Database{
							Name:     item.Name,
							Server:   server.Name,
							ServerID: server.ID,
							Encoding: item.CharSet,
						})
					}
				}
			}
		case biz.DatabaseTypePostgresql:
			postgres, err := db.NewPostgres(server.Username, server.Password, server.Host, server.Port)
			if err == nil {
				if databases, err := postgres.Databases(); err == nil {
					for item := range slices.Values(databases) {
						database = append(database, &biz.Database{
							Name:     item.Name,
							Server:   server.Name,
							ServerID: server.ID,
							Encoding: item.Encoding,
						})
					}
				}
			}
		}
	}

	return database[(page-1)*limit:], int64(len(database)), nil
}

func (r databaseRepo) Create(req *request.DatabaseCreate) error {
	server, err := NewDatabaseServerRepo().Get(req.ServerID)
	if err != nil {
		return err
	}

	switch server.Type {
	case biz.DatabaseTypeMysql:
		mysql, err := db.NewMySQL(server.Username, server.Password, fmt.Sprintf("%s:%d", server.Host, server.Port))
		if err != nil {
			return err
		}
		if err = mysql.UserCreate(req.Username, req.Password); err != nil {
			return err
		}
		if err = mysql.DatabaseCreate(req.Name); err != nil {
			return err
		}
		if err = mysql.PrivilegesGrant(req.Username, req.Name); err != nil {
			return err
		}
	case biz.DatabaseTypePostgresql:
		postgres, err := db.NewPostgres(server.Username, server.Password, server.Host, server.Port)
		if err != nil {
			return err
		}
		if err = postgres.UserCreate(req.Username, req.Password); err != nil {
			return err
		}
		if err = postgres.DatabaseCreate(req.Name); err != nil {
			return err
		}
		if err = postgres.PrivilegesGrant(req.Username, req.Name); err != nil {
			return err
		}
	}

	return nil
}

func (r databaseRepo) Delete(serverID uint, name string) error {
	server, err := NewDatabaseServerRepo().Get(serverID)
	if err != nil {
		return err
	}

	switch server.Type {
	case biz.DatabaseTypeMysql:
		mysql, err := db.NewMySQL(server.Username, server.Password, fmt.Sprintf("%s:%d", server.Host, server.Port))
		if err != nil {
			return err
		}
		return mysql.DatabaseDrop(name)
	case biz.DatabaseTypePostgresql:
		postgres, err := db.NewPostgres(server.Username, server.Password, server.Host, server.Port)
		if err != nil {
			return err
		}
		return postgres.DatabaseDrop(name)
	}

	return nil
}
