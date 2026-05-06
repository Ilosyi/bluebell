package mysql

import (
	"bluebell/models"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// CreatePost 将帖子数据插入MySQL数据库
// 参数 p 包含帖子的所有字段（ID由雪花算法生成，已在logic层赋值）
func CreatePost(p *models.Post) (err error) {
	sqlStr := `insert into post(
	post_id, title, content, author_id, community_id)
	values (?, ?, ?, ?, ?)
	`
	_, err = db.Exec(sqlStr, p.ID, p.Title, p.Content, p.AuthorID, p.CommunityID)
	return
}

// GetPostById 根据帖子ID查询单条帖子
// 返回帖子的完整信息，包括标题、内容、作者ID、社区ID、创建时间
func GetPostById(pid int64) (data *models.Post, err error) {
	data = new(models.Post)
	sqlstr := "SELECT " +
		"post_id, title, content, author_id, community_id,create_time FROM post " +
		"WHERE post_id=?"
	err = db.Get(data, sqlstr, pid)
	return
}

// GetPostListByIDs 根据ID列表批量查询帖子（单条SQL + FIND_IN_SET保持排序）
// 使用sqlx.In将IN子句参数化，避免SQL注入
// 使用FIND_IN_SET保持Redis返回的帖子排序（按分数或时间倒序）
func GetPostListByIDs(ids []int64) (posts []*models.Post, err error) {
	if len(ids) == 0 {
		return
	}
	//将ids转为逗号分隔的字符串，用于FIND_IN_SET保持排序
	idsStr := make([]string, 0, len(ids))
	for _, id := range ids {
		idsStr = append(idsStr, strconv.FormatInt(id, 10))
	}
	//SQL模板：IN子句的?会被sqlx.In展开为(?,?,?)，FIND_IN_SET的?保持原样
	sqlStr := `select post_id, title, content, author_id, community_id, create_time
		from post
		where post_id in (?)
		order by FIND_IN_SET(post_id, ?)`
	//sqlx.In自动展开IN子句的占位符，并返回新的SQL和参数列表
	query, args, err := sqlx.In(sqlStr, ids, strings.Join(idsStr, ","))
	if err != nil {
		return nil, err
	}
	//Rebind将通用占位符?转为MySQL的?格式（兼容不同数据库驱动）
	query = db.Rebind(query)
	posts = make([]*models.Post, 0, len(ids))
	err = db.Select(&posts, query, args...)
	return
}

// GetPostList 分页查询帖子列表（旧版，直接从MySQL分页）
// 按创建时间倒序排列，offset由page和size计算得出
func GetPostList(page, size int64) (posts []*models.Post, err error) {
	sqlstr := "SELECT " +
		"post_id, title, content, author_id, community_id,create_time FROM post " +
		"ORDER BY create_time DESC LIMIT ?,?"
	offset := (page - 1) * size
	posts = make([]*models.Post, 0, 10)
	err = db.Select(&posts, sqlstr, offset, size)
	return
}
