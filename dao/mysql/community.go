package mysql

import (
	"bluebell/models"
	"database/sql"
	"errors"

	"go.uber.org/zap"
)

// GetCommunityList 查询所有社区列表。
// 因为首页只需要社区 ID 和名称，所以 SQL 也只查这两列。
func GetCommunityList() (communities []*models.Community, err error) {
	sqlstr := "SELECT community_id, community_name FROM community"
	err = db.Select(&communities, sqlstr)
	if errors.Is(err, sql.ErrNoRows) {
		// Select 查询空结果通常不会返回 ErrNoRows，这里保留防御性处理。
		zap.L().Warn("no community in db")
		err = nil
	}
	return
}

// GetCommunityDetailByID 根据社区 ID 查询社区详情。
// 这里使用 db.Get，表示“预期只返回一行”。
func GetCommunityDetailByID(id int64) (detail *models.CommunityDetail, err error) {
	detail = new(models.CommunityDetail)
	sqlstr := "SELECT community_id, community_name, introduction, create_time FROM community WHERE community_id=?"
	err = db.Get(detail, sqlstr, id)
	if errors.Is(err, sql.ErrNoRows) {
		zap.L().Warn("no community in db")
		// 这里没有直接往外暴露 sql.ErrNoRows，而是转换成更贴近业务语义的错误。
		err = errors.New("无效的ID")
	}
	return
}
