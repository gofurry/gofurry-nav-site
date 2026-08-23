package abstract

import "github.com/gofurry/gofurry-admin/pkg/util"

type Model interface {
	GetId() int64
	SetId(id int64)
}

type IdModel struct {
	ID int64 `json:"id,string"`
}

type DefaultModel struct {
	IdModel
	Name string `json:"name"`
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
