package dto

type Income struct {
	MD5 string `json:"md5"`
}

type ErrorResponse struct {
	Error string `json:error`
}