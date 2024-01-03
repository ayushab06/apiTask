package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"test/models"
	"test/utility"

	"github.com/gorilla/mux"
	"github.com/spf13/cast"
)

func GetNote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	note, err := models.GetNoteByID(cast.ToInt(id))
	if err != nil {
		utility.Respond(http.StatusInternalServerError, "some error at our end", &w, false)
	} else {
		data, _ := json.Marshal(note.Content)
		utility.RespondStruct(data, &w, true)
	}
}

func GetNotes(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("UserName")
	if err != nil {
		utility.Respond(http.StatusInternalServerError, "some error at our end", &w, false)
		return
	}
	user, err := models.GetUserByUserName(c.Value)
	if err != nil {
		utility.Respond(http.StatusInternalServerError, "some error at our end", &w, false)
		return
	}
	note, err := models.GetNotesByUserID(user.Id)
	if err != nil {
		utility.Respond(http.StatusInternalServerError, "some error at our end", &w, false)
	} else {
		data, _ := json.Marshal(note)
		utility.RespondStruct(data, &w, true)
	}
}

func CreateNote(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("UserName")
	if err != nil {
		utility.Respond(http.StatusInternalServerError, "some error at our end", &w, false)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utility.Respond(http.StatusBadRequest, "wrong format", &w, false)
		return
	}
	var params struct {
		Content string `json:"content"`
	}

	err = json.Unmarshal(body, &params)
	if err != nil {
		utility.Respond(http.StatusBadRequest, "wrong format", &w, false)
		return
	}

	user, err := models.GetUserByUserName(c.Value)

	note := models.Note{UserID: user.Id, Content: params.Content}
	err = note.InsertToDB()

	w.Write([]byte("Note created successfully"))
}

func UpdateNote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var paramsJ struct {
		Content string `json:"content"`
	}

	fmt.Print(r.Body)

	err := json.NewDecoder(r.Body).Decode(&paramsJ)
	if err != nil {
		http.Error(w, "Unable to parse request body", http.StatusBadRequest)
		return
	}

	note, err := models.GetNoteByID(cast.ToInt(id))

	fmt.Println(id, paramsJ.Content)
	err = note.UpdateNote(paramsJ.Content)
	if err != nil {
		utility.Respond(http.StatusInternalServerError, "some error at our end", &w, false)
		return
	}

	w.Write([]byte("Note updated successfully"))
}

func DeleteNoteByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, "Unable to parse request body", http.StatusBadRequest)
		return
	}

	err = models.DeleteNoteByID(cast.ToInt(id))
	if err != nil {
		utility.Respond(http.StatusInternalServerError, "some error at our end", &w, false)
		return
	}

	w.Write([]byte("Note deleted successfully"))
}

func ShareNoteWithUser(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID       int    `json:"id"`
		UserName string `json:"username"`
	}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, "Unable to parse request body", http.StatusBadRequest)
		return
	}

	note, err := models.GetNoteByID(cast.ToInt(params.ID))
	if err != nil {
		utility.Respond(http.StatusInternalServerError, "some error at our end", &w, false)
		return
	}

	user, err := models.GetUserByUserName(params.UserName)
	if err != nil {
		utility.Respond(http.StatusInternalServerError, "some error at our end", &w, false)
		return
	}

	note.UserID = user.Id

	newNote := models.Note{UserID: user.Id, Content: note.Content}
	err = newNote.InsertToDB()
	if err != nil {
		utility.Respond(http.StatusInternalServerError, "some error at our end", &w, false)
		return
	}

	w.Write([]byte("Note shared successfully"))
}

func SearchNotes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	note, err := models.SearchNotesByQuery(query)
	if err != nil {
		utility.Respond(http.StatusInternalServerError, "some error at our end", &w, false)
	} else {
		data, _ := json.Marshal(note)
		utility.RespondStruct(data, &w, true)
	}
}
