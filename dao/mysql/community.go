package mysql

import (
	"bluebell/models"
	"database/sql"
	"errors"

	"go.uber.org/zap"
)

// GetCommunityList 查询所有社区列表
func GetCommunityList() (communities []*models.Community, err error) {
	sqlstr := "SELECT community_id, community_name FROM community"
	err = db.Select(&communities, sqlstr)
	if errors.Is(err, sql.ErrNoRows) {
		zap.L().Warn("no community in db")
		err = nil
	}
	return
}

// GetCommunityDetailByID 根据id查询社区详情
func GetCommunityDetailByID(id int64) (detail *models.CommunityDetail, err error) {
	detail = new(models.CommunityDetail)
	sqlstr := "SELECT community_id, community_name, introduction, create_time FROM community WHERE community_id=?"
	err = db.Get(detail, sqlstr, id)
	if errors.Is(err, sql.ErrNoRows) {
		zap.L().Warn("no community in db")
		err = errors.New("无效的ID")
	}
	return
}
