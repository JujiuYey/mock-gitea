package data

import (
	"fmt"
	"mockgitea/internal/models"
)

func BuildUsers() []models.GiteaUser {
	definitions := []struct {
		Login    string
		FullName string
	}{
		{Login: "zhangsan", FullName: "张三"},
		{Login: "lisi", FullName: "李四"},
		{Login: "wangwu", FullName: "王五"},
		{Login: "zhaoliu", FullName: "赵六"},
		{Login: "chenxi", FullName: "陈曦"},
		{Login: "yangfan", FullName: "杨帆"},
		{Login: "zhoumo", FullName: "周墨"},
		{Login: "sunqian", FullName: "孙倩"},
		{Login: "wenyu", FullName: "文宇"},
		{Login: "linran", FullName: "林然"},
	}

	Users := make([]models.GiteaUser, 0, len(definitions))
	for i, definition := range definitions {
		Users = append(Users, models.GiteaUser{
			ID:        int64(i + 1),
			Login:     definition.Login,
			FullName:  definition.FullName,
			Email:     fmt.Sprintf("%s@example.com", definition.Login),
			AvatarURL: fmt.Sprintf("https://api.dicebear.com/7.x/avataaars/svg?seed=%s", definition.Login),
		})
	}

	return Users
}
