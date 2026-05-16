package abstract

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	database "github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/db"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/models"
	"gorm.io/gorm"
)

type Dao[T any] struct {
	db  *gorm.DB
	ctx context.Context
}

func NewDao[T any](ctx context.Context) (*Dao[T], common.Error) {
	engine := database.Databases.DB(database.Admin)
	if engine == nil {
		return nil, common.NewDaoError("database is not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	return &Dao[T]{
		db:  engine.WithContext(ctx),
		ctx: ctx,
	}, nil
}

func NewDaoWithDB[T any](ctx context.Context, tx *gorm.DB) (*Dao[T], common.Error) {
	if tx == nil {
		return nil, common.NewDaoError("transaction db instance is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	return &Dao[T]{
		db:  tx.WithContext(ctx),
		ctx: ctx,
	}, nil
}

func (dao *Dao[T]) DB() *gorm.DB {
	return dao.db
}

func (dao *Dao[T]) Add(record *T) common.Error {
	if record == nil {
		return common.NewDaoError("record is nil")
	}
	if err := dao.db.Create(record).Error; err != nil {
		return wrapDAOError("create record failed", err)
	}
	return nil
}

func (dao *Dao[T]) UpdateById(id int64, record *T, omitFields ...string) (int64, common.Error) {
	if id <= 0 {
		return 0, common.NewValidationError("id must be greater than 0")
	}
	if record == nil {
		return 0, common.NewDaoError("record is nil")
	}

	query := dao.db.Model(new(T)).Where("id = ?", id)
	if len(omitFields) > 0 {
		query = query.Omit(omitFields...)
	}

	result := query.Updates(record)
	if result.Error != nil {
		return 0, wrapDAOError("update record failed", result.Error)
	}
	return result.RowsAffected, nil
}

func (dao *Dao[T]) UpdateByIdSelective(id int64, record *T, omitFields ...string) (int64, common.Error) {
	if id <= 0 {
		return 0, common.NewValidationError("id must be greater than 0")
	}
	if record == nil {
		return 0, common.NewDaoError("record is nil")
	}

	query := dao.db.Model(new(T)).Where("id = ?", id)
	if len(omitFields) > 0 {
		query = query.Omit(omitFields...)
	}

	result := query.Save(record)
	if result.Error != nil {
		return 0, wrapDAOError("save record failed", result.Error)
	}
	return result.RowsAffected, nil
}

func (dao *Dao[T]) DeleteById(id int64) (int64, common.Error) {
	if id <= 0 {
		return 0, common.NewValidationError("id must be greater than 0")
	}

	result := dao.db.Delete(new(T), id)
	if result.Error != nil {
		return 0, wrapDAOError("delete record failed", result.Error)
	}
	return result.RowsAffected, nil
}

func (dao *Dao[T]) DeleteByIds(idList []int64) (int64, common.Error) {
	if len(idList) == 0 {
		return 0, common.NewValidationError("id list is empty")
	}

	result := dao.db.Where("id IN ?", idList).Delete(new(T))
	if result.Error != nil {
		return 0, wrapDAOError("delete records failed", result.Error)
	}
	return result.RowsAffected, nil
}

func (dao *Dao[T]) GetById(id int64) (*T, common.Error) {
	if id <= 0 {
		return nil, common.NewValidationError("id must be greater than 0")
	}

	var record T
	if err := dao.db.First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewDaoError("record not found")
		}
		return nil, wrapDAOError("query record failed", err)
	}
	return &record, nil
}

func (dao *Dao[T]) Count(conditions ...interface{}) (int64, common.Error) {
	var total int64
	query := dao.db.Model(new(T))
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	if err := query.Count(&total).Error; err != nil {
		return 0, wrapDAOError("count records failed", err)
	}
	return total, nil
}

func (dao *Dao[T]) PageQuery(page, pageSize int, conditions ...interface{}) ([]T, int64, common.Error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	total, err := dao.Count(conditions...)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []T{}, 0, nil
	}

	var records []T
	query := dao.db.Model(new(T))
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, wrapDAOError("page query failed", err)
	}

	return records, total, nil
}

func (dao *Dao[T]) PageQueryFormatted(pageReq *models.PageReq, conditions ...interface{}) (models.PageResponse, common.Error) {
	if pageReq == nil {
		pageReq = &models.PageReq{}
	}
	pageReq.InitPageIfAbsent()

	data, total, err := dao.PageQuery(pageReq.PageNum, pageReq.PageSize, conditions...)
	if err != nil {
		return models.PageResponse{}, err
	}

	return models.PageResponse{
		Total: total,
		Data:  data,
	}, nil
}

func (dao *Dao[T]) GetTableName() string {
	stmt := &gorm.Statement{DB: dao.db}
	if err := stmt.Parse(new(T)); err != nil {
		return ""
	}
	return stmt.Schema.Table
}

func (dao *Dao[T]) BeginTx() (*Dao[T], common.Error) {
	tx := dao.db.Begin()
	if tx.Error != nil {
		return nil, wrapDAOError("begin transaction failed", tx.Error)
	}
	txDao, err := NewDaoWithDB[T](dao.ctx, tx)
	if err != nil {
		return nil, err
	}
	return txDao, nil
}

func (dao *Dao[T]) CommitTx() common.Error {
	if err := dao.db.Commit().Error; err != nil {
		return wrapDAOError("commit transaction failed", err)
	}
	return nil
}

func (dao *Dao[T]) RollbackTx() common.Error {
	if err := dao.db.Rollback().Error; err != nil {
		return wrapDAOError("rollback transaction failed", err)
	}
	return nil
}

func (dao *Dao[T]) SoftDeleteById(id int64) (int64, common.Error) {
	if id <= 0 {
		return 0, common.NewValidationError("id must be greater than 0")
	}

	result := dao.db.Model(new(T)).Where("id = ?", id).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP"))
	if result.Error != nil {
		return 0, wrapDAOError("soft delete record failed", result.Error)
	}
	return result.RowsAffected, nil
}

func wrapDAOError(message string, err error) common.Error {
	slog.Error(message, "error", err)
	return common.NewDaoError(fmt.Sprintf("%s: %v", message, err))
}
