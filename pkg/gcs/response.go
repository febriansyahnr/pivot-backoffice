package gcs

type Response struct {
	PublicUrl string `json:"publicUrl"`
	SignedUrl string `json:"signedUrl"`
}

type UploadMultipart struct {
	PublicURL  string
	Bucket     string
	ObjectName string
}
