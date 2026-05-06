package mysql

import (
	"bluebell/models"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type postBundleRow struct {
	PostID              int64     `db:"post_id"`
	AuthorID            int64     `db:"author_id"`
	CommunityID         int64     `db:"community_id"`
	Status              int32     `db:"status"`
	Title               string    `db:"title"`
	Content             string    `db:"content"`
	CreateTime          time.Time `db:"create_time"`
	AuthorName          string    `db:"author_name"`
	CommunityName       string    `db:"community_name"`
	CommunityIntro      string    `db:"community_intro"`
	CommunityCreateTime time.Time `db:"community_create_time"`
}

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
	if errors.Is(err, sql.ErrNoRows) {
		err = sql.ErrNoRows
	}
	return
}

func CountPosts() (int64, error) {
	var total int64
	err := db.Get(&total, "SELECT COUNT(post_id) FROM post")
	return total, err
}

// GetPostBundleByID 通过一条 JOIN 查询把帖子、作者、社区一次性查出来。
// 这样 logic 层在组装详情时就不需要再分别查 user/community，避免 N+1 查询。
func GetPostBundleByID(pid int64) (*models.ApiPostDetail, error) {
	sqlStr := `
		SELECT
			p.post_id,
			p.author_id,
			p.community_id,
			p.status,
			p.title,
			p.content,
			p.create_time,
			u.username AS author_name,
			c.community_name,
			c.introduction AS community_intro,
			c.create_time AS community_create_time
		FROM post p
		JOIN user u ON p.author_id = u.user_id
		JOIN community c ON p.community_id = c.community_id
		WHERE p.post_id = ?
	`

	row := new(postBundleRow)
	if err := db.Get(row, sqlStr, pid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return buildPostDetailFromRow(row), nil
}

// GetPostBundlesByIDs 批量查询帖子列表所需的聚合数据，并用 FIND_IN_SET 保持 Redis 返回的排序。
func GetPostBundlesByIDs(ids []int64) ([]*models.ApiPostDetail, error) {
	if len(ids) == 0 {
		return []*models.ApiPostDetail{}, nil
	}

	idsStr := make([]string, 0, len(ids))
	for _, id := range ids {
		idsStr = append(idsStr, strconv.FormatInt(id, 10))
	}

	sqlStr := `
		SELECT
			p.post_id,
			p.author_id,
			p.community_id,
			p.status,
			p.title,
			p.content,
			p.create_time,
			u.username AS author_name,
			c.community_name,
			c.introduction AS community_intro,
			c.create_time AS community_create_time
		FROM post p
		JOIN user u ON p.author_id = u.user_id
		JOIN community c ON p.community_id = c.community_id
		WHERE p.post_id IN (?)
		ORDER BY FIND_IN_SET(p.post_id, ?)
	`

	query, args, err := sqlx.In(sqlStr, ids, strings.Join(idsStr, ","))
	if err != nil {
		return nil, err
	}
	query = db.Rebind(query)

	rows := make([]*postBundleRow, 0, len(ids))
	if err := db.Select(&rows, query, args...); err != nil {
		return nil, err
	}

	result := make([]*models.ApiPostDetail, 0, len(rows))
	for _, row := range rows {
		result = append(result, buildPostDetailFromRow(row))
	}
	return result, nil
}

// buildPostDetailFromRow 把 JOIN 查询结果映射成接口层使用的 ApiPostDetail。
func buildPostDetailFromRow(row *postBundleRow) *models.ApiPostDetail {
	return &models.ApiPostDetail{
		AuthorName: row.AuthorName,
		Post: &models.Post{
			ID:          row.PostID,
			AuthorID:    row.AuthorID,
			CommunityID: row.CommunityID,
			Status:      row.Status,
			Title:       row.Title,
			Content:     row.Content,
			CreateTime:  row.CreateTime,
		},
		CommunityDetail: &models.CommunityDetail{
			ID:           row.CommunityID,
			Name:         row.CommunityName,
			Introduction: row.CommunityIntro,
			CreateTime:   row.CommunityCreateTime,
		},
	}
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
