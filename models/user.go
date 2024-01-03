package models

import (
	"log"

	"github.com/astaxie/beego/orm"
)

type User struct {
	Id       int    `orm:"column(id);auto"`
	Username string `orm:"column(username);null"`
	Password string `orm:"column(password);null"`
}

func init() {
	orm.RegisterModel(new(User))
}

func (u *User) InsertToDB() error {
	myOrm := orm.NewOrm()
	_, err := myOrm.Insert(u)
	if err != nil {
		log.Fatal("Error in Insert: ", err)
	}
	return err
}

func GetUserByUserName(username string) (User, error) {
	o := orm.NewOrm()
	var user User
	err := o.QueryTable(new(User)).Filter("username", username).One(&user)
	return user, err
}
