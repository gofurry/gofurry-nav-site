package abstract

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofurry/gofurry-game-collector/common"
	"github.com/gofurry/gofurry-game-collector/common/models"
	database "github.com/gofurry/gofurry-game-collector/roof/db"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

/*
 * @Desc: 统一增删改查接口
 * @author: 福狼
 * @version: v1.0.1
 */

// Dao 抽象 Dao 层
type Dao[T any] struct {
	db  *gorm.DB        // GORM 实例
	ctx context.Context // 支持超时/链路追踪
}

// NewDao 创建 Dao 实例
func NewDao[T any](ctx context.Context) *Dao[T] {
	db := database.Orm.DB()
	if db == nil {
		panic("数据库连接未初始化")
	}

	// 默认上下文
	if ctx == nil {
		ctx = context.Background()
	}

	return &Dao[T]{
		db:  db.WithContext(ctx), // 绑定上下文
		ctx: ctx,
	}
}

// NewDaoWithDB 自定义 DB 实例
func NewDaoWithDB[T any](ctx context.Context, tx *gorm.DB) *Dao[T] {
	if tx == nil {
		panic("DB 实例不能为空")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return &Dao[T]{
		db:  tx.WithContext(ctx),
		ctx: ctx,
	}
}

// ========== 通用 CRUD 方法 ==========

// DB 暴露gorm.DB
func (dao *Dao[T]) DB() *gorm.DB {
	return dao.db
}

// Add 新增记录
func (dao *Dao[T]) Add(record *T) common.GFError {
	if record == nil {
		err := errors.New("新增记录不能为空")
		slog.Error("[DAO Add] 参数错误", "error", err)
		return common.NewDaoError(err.Error())
	}

	// 获取表名
	tableName := dao.GetTableName

	// 执行新增
	result := dao.db.Create(record)
	if err := result.Error; err != nil {
		// 解析 PostgreSQL 错误
		errMsg, isPgErr := parsePgError(err)
		if !isPgErr {
			errMsg = fmt.Sprintf("新增记录失败 [表:%s]: %v", tableName, err)
		}

		slog.Error("[DAO Add] 新增失败",
			"table", tableName,
			"error", err,
			"pg_code", getPgErrorCode(err),
		)
		return common.NewDaoError(errMsg)
	}

	slog.Info("[DAO Add] 新增成功",
		"table", tableName,
		"rows_affected", result.RowsAffected,
	)
	return nil
}

// UpdateById 根据 ID 更新
func (dao *Dao[T]) UpdateById(id int64, record *T, omitFields ...string) (int64, common.GFError) {
	if id <= 0 {
		err := errors.New("ID 必须大于 0")
		slog.Error("[DAO UpdateById] 参数错误", "error", err, "id", id)
		return 0, common.NewDaoError(err.Error())
	}
	if record == nil {
		err := errors.New("更新记录不能为空")
		slog.Error("[DAO UpdateById] 参数错误", "error", err)
		return 0, common.NewDaoError(err.Error())
	}

	tableName := dao.GetTableName
	db := dao.db.Model(new(T)).Where("id = ?", id)

	// 忽略指定字段
	omitFields = append(omitFields, "create_time")
	db = db.Omit(omitFields...)

	// 执行更新
	result := db.Updates(record)
	if err := result.Error; err != nil {
		errMsg, isPgErr := parsePgError(err)
		if !isPgErr {
			errMsg = fmt.Sprintf("更新记录失败 [表:%s, ID:%d]: %v", tableName, id, err)
		}

		slog.Error("[DAO UpdateById] 更新失败",
			"table", tableName,
			"id", id,
			"error", err,
			"pg_code", getPgErrorCode(err),
		)
		return 0, common.NewDaoError(errMsg)
	}

	slog.Info("[DAO UpdateById] 更新成功",
		"table", tableName,
		"id", id,
		"rows_affected", result.RowsAffected,
	)
	return result.RowsAffected, nil
}

// UpdateByIdSelective 全量更新
func (dao *Dao[T]) UpdateByIdSelective(id int64, record *T, omitFields ...string) (int64, common.GFError) {
	if id <= 0 || record == nil {
		err := errors.New("参数错误：ID 必须大于 0 且记录不能为空")
		slog.Error("[DAO UpdateByIdSelective] 参数错误", "error", err, "id", id)
		return 0, common.NewDaoError(err.Error())
	}

	tableName := dao.GetTableName
	db := dao.db.Model(new(T)).Where("id = ?", id)

	omitFields = append(omitFields, "create_time")
	db = db.Omit(omitFields...)

	// 使用 Save 全量更新
	result := db.Save(record)
	if err := result.Error; err != nil {
		errMsg, isPgErr := parsePgError(err)
		if !isPgErr {
			errMsg = fmt.Sprintf("全量更新失败 [表:%s, ID:%d]: %v", tableName, id, err)
		}

		slog.Error("[DAO UpdateByIdSelective] 更新失败",
			"table", tableName,
			"id", id,
			"error", err,
		)
		return 0, common.NewDaoError(errMsg)
	}

	return result.RowsAffected, nil
}

// DeleteById 物理删除单条记录
func (dao *Dao[T]) DeleteById(id int64) (int64, common.GFError) {
	if id <= 0 {
		err := errors.New("ID 必须大于 0")
		slog.Error("[DAO DeleteById] 参数错误", "error", err, "id", id)
		return 0, common.NewDaoError(err.Error())
	}

	tableName := dao.GetTableName
	result := dao.db.Where("id = ?", id).Delete(new(T))

	if err := result.Error; err != nil {
		errMsg := fmt.Sprintf("删除记录失败 [表:%s, ID:%d]: %v", tableName, id, err)
		slog.Error("[DAO DeleteById] 删除失败",
			"table", tableName,
			"id", id,
			"error", err,
		)
		return 0, common.NewDaoError(errMsg)
	}

	slog.Info("[DAO DeleteById] 删除成功",
		"table", tableName,
		"id", id,
		"rows_affected", result.RowsAffected,
	)
	return result.RowsAffected, nil
}

// DeleteByIds 批量物理删除
func (dao *Dao[T]) DeleteByIds(idList []int64) (int64, common.GFError) {
	if len(idList) == 0 {
		err := errors.New("ID 列表不能为空")
		slog.Error("[DAO DeleteByIds] 参数错误", "error", err)
		return 0, common.NewDaoError(err.Error())
	}

	// 校验 ID 合法性
	for _, id := range idList {
		if id <= 0 {
			err := fmt.Errorf("无效 ID: %d", id)
			slog.Error("[DAO DeleteByIds] 参数错误", "error", err)
			return 0, common.NewDaoError(err.Error())
		}
	}

	tableName := dao.GetTableName
	result := dao.db.Where("id IN (?)", idList).Delete(new(T))

	if err := result.Error; err != nil {
		errMsg := fmt.Sprintf("批量删除失败 [表:%s]: %v", tableName, err)
		slog.Error("[DAO DeleteByIds] 删除失败",
			"table", tableName,
			"id_list", idList,
			"error", err,
		)
		return 0, common.NewDaoError(errMsg)
	}

	slog.Info("[DAO DeleteByIds] 批量删除成功",
		"table", tableName,
		"id_count", len(idList),
		"rows_affected", result.RowsAffected,
	)
	return result.RowsAffected, nil
}

// GetById 根据 ID 查询单条记录
func (dao *Dao[T]) GetById(id int64) (*T, common.GFError) {
	if id <= 0 {
		err := errors.New("ID 必须大于 0")
		slog.Error("[DAO GetById] 参数错误", "error", err, "id", id)
		return nil, common.NewDaoError(err.Error())
	}

	var record T
	tableName := dao.db.Model(&record).Statement.Table

	result := dao.db.Where("id = ?", id).Take(&record)
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errMsg := fmt.Sprintf("记录不存在 [表:%s, ID:%d]", tableName, id)
			slog.Warn("[DAO GetById] 记录不存在",
				"table", tableName,
				"id", id,
			)
			return nil, common.NewDaoError(errMsg)
		}

		errMsg := fmt.Sprintf("查询记录失败 [表:%s, ID:%d]: %v", tableName, id, err)
		slog.Error("[DAO GetById] 查询失败",
			"table", tableName,
			"id", id,
			"error", err,
		)
		return nil, common.NewDaoError(errMsg)
	}

	return &record, nil
}

// Count 统计总数
func (dao *Dao[T]) Count(conditions ...interface{}) (int64, common.GFError) {
	var count int64
	tableName := dao.GetTableName

	db := dao.db.Model(new(T))
	if len(conditions) > 0 {
		db = db.Where(conditions[0], conditions[1:]...)
	}

	result := db.Count(&count)
	if err := result.Error; err != nil {
		errMsg := fmt.Sprintf("统计数量失败 [表:%s]: %v", tableName, err)
		slog.Error("[DAO Count] 统计失败",
			"table", tableName,
			"conditions", conditions,
			"error", err,
		)
		return 0, common.NewDaoError(errMsg)
	}

	slog.Debug("[DAO Count] 统计成功",
		"table", tableName,
		"conditions", conditions,
		"count", count,
	)
	return count, nil
}

// PageQuery 分页查询
func (dao *Dao[T]) PageQuery(page, pageSize int, conditions ...interface{}) ([]T, int64, common.GFError) {
	// 参数校验
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 { // 限制最大页大小
		pageSize = 20
	}

	var list []T
	tableName := dao.GetTableName

	// 统计总数
	total, err := dao.Count(conditions...)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return list, 0, nil
	}

	// 分页查询
	offset := (page - 1) * pageSize
	db := dao.db.Model(new(T))
	if len(conditions) > 0 {
		db = db.Where(conditions[0], conditions[1:]...)
	}

	result := db.Offset(offset).Limit(pageSize).Find(&list)
	if err := result.Error; err != nil {
		errMsg := fmt.Sprintf("分页查询失败 [表:%s]: %v", tableName, err)
		slog.Error("[DAO PageQuery] 查询失败",
			"table", tableName,
			"page", page,
			"page_size", pageSize,
			"conditions", conditions,
			"error", err,
		)
		return nil, 0, common.NewDaoError(errMsg)
	}

	return list, total, nil
}

// PageQueryFormatted 格式化版分页查询
func (dao *Dao[T]) PageQueryFormatted(pageReq *models.PageReq, conditions ...interface{}) (models.PageResponse, common.GFError) {
	// 初始化分页参数
	if pageReq == nil {
		pageReq = &models.PageReq{}
	}
	pageReq.InitPageIfAbsent()

	// 获取表名
	tableName := dao.GetTableName

	// 统计总数
	var total int64
	db := dao.db.Model(new(T))
	if len(conditions) > 0 {
		db = db.Where(conditions[0], conditions[1:]...)
	}
	countResult := db.Count(&total)
	if countResult.Error != nil {
		errMsg := fmt.Sprintf("分页查询-统计总数失败 [表:%s]: %v", tableName, countResult.Error)
		slog.Error("[DAO PageQuery] 统计总数失败",
			"table", tableName,
			"page_num", pageReq.PageNum,
			"page_size", pageReq.PageSize,
			"error", countResult.Error,
		)
		return models.PageResponse{}, common.NewDaoError(errMsg)
	}

	if total == 0 {
		return models.PageResponse{
			Total: 0,
			Data:  []T{},
		}, nil
	}

	// 分页查询
	offset := (pageReq.PageNum - 1) * pageReq.PageSize
	var dataList []T
	queryResult := dao.db.Model(new(T)).
		Where(conditions[0], conditions[1:]...).
		Offset(offset).
		Limit(pageReq.PageSize).
		Find(&dataList)

	if queryResult.Error != nil {
		errMsg := fmt.Sprintf("分页查询-获取数据失败 [表:%s]: %v", tableName, queryResult.Error)
		slog.Error("[DAO PageQuery] 获取数据失败",
			"table", tableName,
			"error", queryResult.Error,
		)
		return models.PageResponse{}, common.NewDaoError(errMsg)
	}

	return models.PageResponse{
		Total: total,
		Data:  dataList,
	}, nil
}

// ========== 工具方法 ==========

// parsePgError 解析 PostgreSQL 错误码
func parsePgError(err error) (string, bool) {
	pe, ok := err.(*pgconn.PgError)
	if !ok {
		return "", false
	}

	switch pe.Code {
	case "23502": // 非空约束
		return fmt.Sprintf("必要字段不能为空 [PG_CODE:%s]", pe.Code), true
	case "23505": // 唯一约束
		return fmt.Sprintf("数据重复 [PG_CODE:%s, 约束:%s]", pe.Code, pe.ConstraintName), true
	case "23503": // 外键约束
		return fmt.Sprintf("外键约束失败 [PG_CODE:%s, 约束:%s]", pe.Code, pe.ConstraintName), true
	case "22001": // 字符串超长
		return fmt.Sprintf("字段长度超过限制 [PG_CODE:%s]", pe.Code), true
	default:
		return fmt.Sprintf("数据库错误 [PG_CODE:%s]: %s", pe.Code, pe.Message), true
	}
}

// getPgErrorCode 获取 PostgreSQL 错误码
func getPgErrorCode(err error) string {
	pe, ok := err.(*pgconn.PgError)
	if !ok {
		return ""
	}
	return pe.Code
}

func (dao *Dao[T]) GetTableName() string {
	return dao.db.Model(new(T)).Statement.Table
}

// ========== 扩展方法 ==========

// BeginTx 开启事务
func (dao *Dao[T]) BeginTx() (*Dao[T], common.GFError) {
	tx := dao.db.Begin()
	if tx.Error != nil {
		slog.Error("[DAO BeginTx] 开启事务失败", "error", tx.Error)
		return nil, common.NewDaoError("开启事务失败: " + tx.Error.Error())
	}
	return NewDaoWithDB[T](dao.ctx, tx), nil
}

// CommitTx 提交事务
func (dao *Dao[T]) CommitTx() common.GFError {
	if err := dao.db.Commit().Error; err != nil {
		slog.Error("[DAO CommitTx] 提交事务失败", "error", err)
		return common.NewDaoError("提交事务失败: " + err.Error())
	}
	return nil
}

// RollbackTx 回滚事务
func (dao *Dao[T]) RollbackTx() common.GFError {
	if err := dao.db.Rollback().Error; err != nil {
		slog.Error("[DAO RollbackTx] 回滚事务失败", "error", err)
		return common.NewDaoError("回滚事务失败: " + err.Error())
	}
	return nil
}

// SoftDeleteById 软删除（需模型包含 deleted_at 字段）
func (dao *Dao[T]) SoftDeleteById(id int64) (int64, common.GFError) {
	if id <= 0 {
		err := errors.New("ID 必须大于 0")
		slog.Error("[DAO SoftDeleteById] 参数错误", "error", err, "id", id)
		return 0, common.NewDaoError(err.Error())
	}

	tableName := dao.GetTableName
	result := dao.db.Model(new(T)).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))

	if err := result.Error; err != nil {
		errMsg := fmt.Sprintf("软删除失败 [表:%s, ID:%d]: %v", tableName, id, err)
		slog.Error("[DAO SoftDeleteById] 软删除失败",
			"table", tableName,
			"id", id,
			"error", err,
		)
		return 0, common.NewDaoError(errMsg)
	}

	return result.RowsAffected, nil
}
