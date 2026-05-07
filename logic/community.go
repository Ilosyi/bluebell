package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"database/sql"
	"errors"
)

var (
	// 通过变量引用 dao 层函数，方便单元测试时替换。
	getCommunityListFromMySQL       = mysql.GetCommunityList
	getCommunityDetailByIDFromMySQL = mysql.GetCommunityDetailByID
)

// GetCommunityList 获取社区列表。
// 这个函数本身几乎没有额外业务规则，因此直接把请求转发给 dao 层。
func GetCommunityList() ([]*models.Community, error) {
	return getCommunityListFromMySQL()
}

// GetCommunityDetail 获取指定社区详情。
// 这里的主要价值是把 dao 层返回的“数据库层错误”转换成更稳定的领域错误。
func GetCommunityDetail(id int64) (*models.CommunityDetail, error) {
	detail, err := getCommunityDetailByIDFromMySQL(id)
	if err != nil {
		// 对上层来说，我们更关心“社区是否存在”，
		// 而不关心底层到底是 sql.ErrNoRows 还是 dao 手工返回的“无效的ID”。
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "无效的ID" {
			return nil, ErrCommunityNotFound
		}
		return nil, err
	}
	return detail, nil
}
