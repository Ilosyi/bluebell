package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"database/sql"
	"errors"
)

var (
	getCommunityListFromMySQL       = mysql.GetCommunityList
	getCommunityDetailByIDFromMySQL = mysql.GetCommunityDetailByID
)

func GetCommunityList() ([]*models.Community, error) {

	return getCommunityListFromMySQL()

}

func GetCommunityDetail(id int64) (*models.CommunityDetail, error) {
	detail, err := getCommunityDetailByIDFromMySQL(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "无效的ID" {
			return nil, ErrCommunityNotFound
		}
		return nil, err
	}
	return detail, nil
}
