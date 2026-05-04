package mysql

import "bluebell/models"

func CreatePost(p *models.Post) (err error) {
	sqlstr := "INSERT INTO post(" +
		"post_id, title, content, author_id, community_id) " +
		"VALUES(?,?,?,?,?)"
	_, err = db.Exec(sqlstr, p.ID, p.Title, p.Content, p.AuthorID, p.CommunityID)
	return

}

func GetPostById(pid int64) (data *models.Post, err error) {
	data = new(models.Post)
	sqlstr := "SELECT " +
		"post_id, title, content, author_id, community_id,create_time FROM post " +
		"WHERE post_id=?"
	err = db.Get(data, sqlstr, pid)
	return
}
