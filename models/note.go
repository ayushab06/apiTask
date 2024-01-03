package models

import (
	"fmt"
	"log"

	"github.com/astaxie/beego/orm"
)

type Note struct {
	Id      int    `orm:"column(id);auto"`
	UserID  int    `orm:"column(userid);null"`
	Content string `orm:"column(content);null"`
}

func init() {

	orm.RegisterModel(new(Note))
}

func (b *Note) InsertToDB() error {
	myOrm := orm.NewOrm()
	_, err := (myOrm).Insert(b)
	if err != nil {
		log.Fatal("Error in Insert: ", err)
	}
	return err
}

func GetNoteByID(id int) (Note, error) {
	o := orm.NewOrm()
	var note Note
	err := o.QueryTable(new(Note)).Filter("id", id).One(&note)
	return note, err
}

func GetNotesByUserID(userid int) ([]Note, error) {
	o := orm.NewOrm()
	var notes []Note
	_, err := o.QueryTable(new(Note)).Filter("userid", userid).All(&notes)
	return notes, err
}

func (b *Note) UpdateNote(content string) error {
	myOrm := orm.NewOrm()
	b.Content = content
	_, err := myOrm.Update(b, "Content")
	if err != nil {
		log.Fatal("Error in Update: ", err)
	}
	return err
}

func DeleteNoteByID(id int) error {
	o := orm.NewOrm()
	_, err := o.QueryTable(new(Note)).Filter("id", id).Delete()
	return err
}

func SearchNotesByQuery(query string) ([]Note, error) {
	o := orm.NewOrm()
	var notes []Note
	fmt.Println(query)

	_, err := o.QueryTable("note").Filter("content__icontains", query).All(&notes)
	if err != nil {
		return nil, err
	}

	return notes, nil
}
