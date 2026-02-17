package server

// The serverError helper writes a log entry at Error Level (including the request
// method and URI as attributes), then sends a generic 500 Internal Server Error
// response to the user.

// NOTE : only here for doc

// func (serv *server) serverError(w http.ResponseWriter, r *http.Request, err error) {
// 	var (
// 		method = r.Method
// 		uri    = r.URL.RequestURI()
// 	)

// 	serv.logger.Error(err.Error(), "method", method, "uri", uri)
// 	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
// }

// // func (serv *server) clientError(w http.ResponseWriter, status int) {
// // 	http.Error(w, http.StatusText(status), status)
// // }
