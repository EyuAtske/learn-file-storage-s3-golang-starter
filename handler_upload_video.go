package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1 << 30)
	videoIDString := r.PathValue("videoID")
	videoId, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}
	
	token, err := auth.GetBearerToken(r.Header)
	if err != nil{
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userId, err:= auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil{
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	metadata, err := cfg.db.GetVideo(videoId)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Unable to get metadata", err)
		return
	}
	if userId != metadata.UserID{
		respondWithError(w, http.StatusUnauthorized, "authenticated user is not the video owner", err)
		return
	}
	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()
	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if mediaType == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Content-Type for thumbnail", nil)
		return
	}
	if mediaType != "video/mp4"{
		respondWithError(w, http.StatusBadRequest, "Not of type mime", nil)
		return
	}
	tempFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Unable to create file", err)
		return
	}
	defer os.Remove("tempFile")
	defer tempFile.Close()
	io.Copy(tempFile, file)

	tempFile.Seek(0, io.SeekStart)
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	filename := fmt.Sprintf("%x.mp4", b)
	cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: &cfg.s3Bucket,
		Key: &filename,
		Body: tempFile,
		ContentType: &mediaType,
	})
	videoURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, filename)
	metadata.VideoURL = &videoURL
	err = cfg.db.UpdateVideo(metadata)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Unable to update metadata", err)
		return
	}
}
