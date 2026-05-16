package abstract

import "github.com/gofurry/gofurry-game-backend/common/util"

/*
 * @Desc: 公共模型
 * @author: 福狼
 * @version: v1.0.0
 */

type Model interface {
	GetId() int64
	SetId(id int64)
}

type IdModel struct {
	ID int64 `gorm:"primaryKey;column:id" json:"id,string"`
}

type DefaultModel struct {
	IdModel
	Name string `gorm:"column:name;not null" json:"name"`
}

func (dm *DefaultModel) GetId() int64   { return dm.ID }
func (dm *DefaultModel) SetId(id int64) { dm.ID = id }

func (dm *DefaultModel) GetName() string     { return dm.Name }
func (dm *DefaultModel) SetName(name string) { dm.Name = name }

func (im *IdModel) GetId() int64   { return im.ID }
func (im *IdModel) SetId(id int64) { im.ID = id }

func (im *IdModel) IsNull() bool {
	if im.ID == 0 {
		return true
	}
	return false
}

func (im *IdModel) SetNewId() {
	im.ID = util.GenerateId()
}

type oauthModel interface {
	GetId() string
	GetSecret() string
}

type Oauth struct {
	ClientId     string
	ClientSecret string
	RedirectUrl  string
}

func (o *Oauth) GetId() string     { return o.ClientId }
func (o *Oauth) GetSecret() string { return o.ClientSecret }
