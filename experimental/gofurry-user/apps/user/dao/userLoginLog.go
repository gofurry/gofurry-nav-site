package dao

import (
	"github.com/gofurry/gofurry-user/apps/user/models"
	"github.com/gofurry/gofurry-user/common/abstract"
)

var newUserLogDao = new(userLogDao)

func init() {
	newUserLogDao.Init()
	newUserLogDao.Mode = models.GfLoginLog{}
}

type userLogDao struct{ abstract.Dao }

func GetUserLogDao() *userLogDao { return newUserLogDao }
