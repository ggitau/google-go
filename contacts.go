package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
)

type Contact struct {
	Id        bson.ObjectId `bson:"_id"`
	FirstName string        `bson:"firstName"`
	LastName  string        `bson:"lastName"`
	Phone     string        `bson:"phone"`
	Email     string        `bson:"email"`
}
type AddressBook struct {
	Contacts []Contact
}

var router = new(mux.Router)
var templates = template.Must(template.ParseGlob("partials/*"))

//homeHandler handles requests made with the url /
func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("homeHandler called")

	session, err := mgo.Dial("localhost")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer session.Close()

	c := session.DB("contacts_db").C("contacts")
	var contacts []Contact
	c.Find(bson.M{}).Select(bson.M{"firstName": 1, "lastName": 1, "_id": 1}).All(&contacts)
	for _, v := range contacts {
		log.Println(v.FirstName)
	}
	addressBook := AddressBook{Contacts: contacts}

	templates.ExecuteTemplate(w, "index.html", addressBook)
}

//newContactHandler handles requests made using the url /new.
func newContactHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("newContactHandler called")
	templates.ExecuteTemplate(w, "contactform.html", nil)
}

//saveContactHandler handles requests made using the /save url
func saveContactHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	decoder := schema.NewDecoder()
	contact := new(Contact)

	err = decoder.Decode(contact, r.PostForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	contact.Id = bson.NewObjectId()

	session, err := mgo.Dial("localhost")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer session.Close()

	c := session.DB("contacts_db").C("contacts")

	err = c.Insert(contact)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// The showContactHandler function handles /contact/{contactId} requests
func showContactHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contactId := vars["contactId"]

	log.Printf("Got the contactId %s\n", contactId)

	session, err := mgo.Dial("localhost")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer session.Close()

	c := session.DB("contacts_db").C("contacts")
	contact := Contact{}
	c.Find(bson.M{"_id": bson.ObjectIdHex(contactId)}).One(&contact)

	log.Printf("Found contact %s", contact.FirstName)
	templates.ExecuteTemplate(w, "contact.html", contact)
}

func main() {
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/new", newContactHandler)
	router.HandleFunc("/save", saveContactHandler)
	router.HandleFunc("/contacts/{contactId}", showContactHandler)

	http.Handle("/", router)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("img"))))

	http.ListenAndServe(":8080", nil)
}
