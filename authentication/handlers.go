package authentication

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func ServeTestAuth(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("./authentication/test_firebase.html")
	w.Write(b)
}

func ServeTestAuthOnSuccess(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("authorization")

	payload, _ := ioutil.ReadAll(r.Body)

	infos := map[string]interface{}{
		"token":   token,
		"payload": payload,
	}
	writeJSON(w, infos, 200)
}

func writeJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	switch v := data.(type) {
	case []byte:
		w.Write(v)
	case string:
		w.Write([]byte(v))
	default:
		json.NewEncoder(w).Encode(data)
	}
}
