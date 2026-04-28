package mysql

func InsertUser(){
	//判断用户是否存在

	//执行SQL语句入库
}

func QueryUserById(){

}

//判断用户是否存在
func CheckUserExist(username string) (bool,error){
	//查询数据库，判断用户是否存在
	sqlstr:="select count(user_id) from user where username=?"
	var count int
	err:=db.Get(&count,sqlstr,username)
	if err != nil {
		return false, err
	}
	return count > 0, nil
	
}