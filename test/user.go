package test

import . "github.com/doytowin/go-query/core"

type UserEntity struct {
	Id    int     `json:"id"`
	Score *int    `json:"score"`
	Memo  *string `json:"memo"`
}

func (u UserEntity) GetTableName() string {
	return "User"
}

type AccountOr struct {
	Username *string
	Email    *string
	Mobile   *string
}

type UserQuery struct {
	PageQuery
	IdGt      *int
	IdIn      *[]int
	ScoreLt   *int
	MemoNull  bool
	MemoLike  *string
	AccountOr *AccountOr
	Deleted   *bool
}

type TestEntity struct {
	Id int
}

func (e TestEntity) GetTableName() string {
	return "t_user"
}
