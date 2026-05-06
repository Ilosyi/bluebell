package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
)

var (
	getCommunityListFromMySQL       = mysql.GetCommunityList
	getCommunityDetailByIDFromMySQL = mysql.GetCommunityDetailByID
)

func GetCommunityList() ([]*models.Community, error) {

	return getCommunityListFromMySQL()

}

func GetCommunityDetail(id int64) (*models.CommunityDetail, error) {
	return getCommunityDetailByIDFromMySQL(id)
}
