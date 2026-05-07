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
	// 这个结构体不是对外响应模型，而是“SQL JOIN 查询结果的中间承载体”。
	// 因为一条 JOIN 会同时查出帖子、作者、社区三类字段，
	// 直接映射到 models.ApiPostDetail 不够方便，所以先用这个 row 结构承接。
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
	post_id, title, content, author_id, community_id, status)
	values (?, ?, ?, ?, ?, ?)
	`
	_, err = db.Exec(sqlStr, p.ID, p.Title, p.Content, p.AuthorID, p.CommunityID, p.Status)
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
	// 只统计已发布帖子，草稿不算公开帖子总数。
	err := db.Get(&total, "SELECT COUNT(post_id) FROM post WHERE status=? ", models.PostStatusPublished)
	return total, err
}

// SearchPostBundles 根据关键词搜索公开帖子，并返回聚合后的详情结构。
// 搜索范围包括：标题、正文、作者用户名、作者昵称、社区名。
func SearchPostBundles(p *models.ParamPostList) ([]*models.ApiPostDetail, error) {
	// buildSearchPostWhere 负责拼装 WHERE 子句和对应参数。
	where, args := buildSearchPostWhere(p)
	orderBy := "p.create_time DESC"
	if p.Order == "score" {
		// 当前搜索结果仍按发布时间排序。
		// 如果未来要支持“搜索结果按热度排”，需要额外设计 Redis/MySQL 混合排序方案。
		orderBy = "p.create_time DESC"
	}
	// LIMIT ?,? 的参数分别是 offset 和 size。
	args = append(args, (p.Page-1)*p.Size, p.Size)

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
	` + where + `
		ORDER BY ` + orderBy + `
		LIMIT ?,?
	`

	rows := make([]*postBundleRow, 0, p.Size)
	if err := db.Select(&rows, sqlStr, args...); err != nil {
		return nil, err
	}
	// 把中间行结构转换成 API 层真正要返回的结构。
	result := make([]*models.ApiPostDetail, 0, len(rows))
	for _, row := range rows {
		result = append(result, buildPostDetailFromRow(row))
	}
	return result, nil
}

// CountSearchPosts 返回某个搜索条件下的总帖子数。
// 这个总数会参与计算 total_pages / has_more。
func CountSearchPosts(p *models.ParamPostList) (int64, error) {
	where, args := buildSearchPostWhere(p)
	sqlStr := `
		SELECT COUNT(p.post_id)
		FROM post p
		JOIN user u ON p.author_id = u.user_id
		JOIN community c ON p.community_id = c.community_id
	` + where
	var total int64
	err := db.Get(&total, sqlStr, args...)
	return total, err
}

// buildSearchPostWhere 动态拼接搜索 SQL 的 WHERE 子句。
// 这样可以根据是否有 community_id / keyword 灵活组合条件。
func buildSearchPostWhere(p *models.ParamPostList) (string, []any) {
	// 初始条件永远是“只搜已发布帖子”。
	args := []any{models.PostStatusPublished}
	clauses := []string{"p.status = ?"}
	if p.CommunityID > 0 {
		clauses = append(clauses, "p.community_id = ?")
		args = append(args, p.CommunityID)
	}
	keyword := strings.TrimSpace(p.Keyword)
	if keyword != "" {
		// 用 LIKE 做模糊搜索。
		// 前后加 % 表示“包含这个关键词”。
		like := "%" + escapeLike(keyword) + "%"
		clauses = append(clauses, `(p.title LIKE ? ESCAPE '\\' OR p.content LIKE ? ESCAPE '\\' OR u.username LIKE ? ESCAPE '\\' OR u.nickname LIKE ? ESCAPE '\\' OR c.community_name LIKE ? ESCAPE '\\')`)
		args = append(args, like, like, like, like, like)
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

// escapeLike 对 LIKE 查询中的特殊字符做转义。
// 否则用户如果输入 % 或 _，会被 MySQL 当成通配符而不是普通字符。
func escapeLike(keyword string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(keyword)
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
		WHERE p.post_id = ? AND p.status = ?
	`

	row := new(postBundleRow)
	if err := db.Get(row, sqlStr, pid, models.PostStatusPublished); err != nil {
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

	// idsStr 用于后面的 FIND_IN_SET 排序。
	// 例如 Redis 给出 [12, 9, 7]，就要让 MySQL 结果按这个顺序返回。
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
		WHERE p.post_id IN (?) AND p.status = ?
		ORDER BY FIND_IN_SET(p.post_id, ?)
	`

	query, args, err := sqlx.In(sqlStr, ids, models.PostStatusPublished, strings.Join(idsStr, ","))
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
	// Rebind 会把通用占位符风格调整成当前驱动支持的风格。
	// 对 MySQL 来说仍然是 ?，但这样写兼容性更好。
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
		"WHERE status=? ORDER BY create_time DESC LIMIT ?,?"
	offset := (page - 1) * size
	posts = make([]*models.Post, 0, 10)
	err = db.Select(&posts, sqlstr, models.PostStatusPublished, offset, size)
	return
}

// GetPostForManageByID 返回当前用户管理帖子所需的详情，草稿和已发布帖子都可查。
func GetPostForManageByID(pid int64) (*models.ApiPostDetail, error) {
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

// GetMyPostBundles 分页查询当前用户指定状态的帖子管理列表。
func GetMyPostBundles(userID int64, status int32, page, size int64) ([]*models.ApiPostDetail, error) {
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
		WHERE p.author_id = ? AND p.status = ?
		ORDER BY p.update_time DESC, p.create_time DESC
		LIMIT ?,?
	`
	// 第 N 页的起始偏移量 = (page - 1) * size。
	offset := (page - 1) * size
	rows := make([]*postBundleRow, 0, size)
	if err := db.Select(&rows, sqlStr, userID, status, offset, size); err != nil {
		return nil, err
	}

	result := make([]*models.ApiPostDetail, 0, len(rows))
	for _, row := range rows {
		result = append(result, buildPostDetailFromRow(row))
	}
	return result, nil
}

// CountMyPosts 返回当前用户某种状态下的帖子总数。
func CountMyPosts(userID int64, status int32) (int64, error) {
	var total int64
	err := db.Get(&total, "SELECT COUNT(post_id) FROM post WHERE author_id=? AND status=?", userID, status)
	return total, err
}

// UpdatePost 更新帖子主体字段。
// 这里只更新标题、正文和社区，不改状态、不改作者。
func UpdatePost(pid int64, p *models.Post) error {
	sqlStr := "update post set title=?, content=?, community_id=? where post_id=?"
	_, err := db.Exec(sqlStr, p.Title, p.Content, p.CommunityID, pid)
	return err
}

// PublishPost 只负责把状态改成“已发布”。
// Redis 索引的补充工作在 logic 层完成。
func PublishPost(pid int64) error {
	_, err := db.Exec("update post set status=? where post_id=?", models.PostStatusPublished, pid)
	return err
}

// DeletePost 从 MySQL 中删除一条帖子记录。
// Redis 里的清理不在这里做，而由 logic 层根据帖子状态决定是否继续清索引。
func DeletePost(pid int64) error {
	_, err := db.Exec("delete from post where post_id=?", pid)
	return err
}
