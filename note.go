package main

// func credshandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != "GET" {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	credsData, err := os.ReadFile("users.json")
// 	if err != nil {
// 		http.Error(w, "Could not read credentials", http.StatusInternalServerError)
// 		return
// 	}
// 	var cr Credentials
// 	err = json.Unmarshal(credsData, &cr)
// 	if err != nil {
// 		http.Error(w, "Could not parse credentials", http.StatusInternalServerError)
// 		fmt.Fprintf(w, "%s\n", err)
// 		return
// 	}
// 	fmt.Fprintf(w, "Stored crds\nUsername: %s\nPassword: %s\n", cr.Username, cr.Password)
// }
