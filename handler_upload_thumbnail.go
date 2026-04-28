package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"

	// "encoding/base64"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("thumbnail")
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
	if mediaType != "image/jpeg" || mediaType != "image/png"{
		respondWithError(w, http.StatusBadRequest, "Not of type mime", nil)
		return
	}

	// data, err := io.ReadAll(file)
	// if err != nil{
	// 	respondWithError(w, http.StatusInternalServerError, "Error reading file", err)
	// 	return
	// }

	metadata, err := cfg.db.GetVideo(videoID)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Unable to get metadata", err)
		return
	}
	if userID != metadata.UserID{
		respondWithError(w, http.StatusUnauthorized, "authenticated user is not the video owner", err)
		return
	}
	//store in file system
	fileExe := strings.Split(mediaType, "/")
	fmt.Print(fileExe[1])
	fileName := fmt.Sprintf("%s.%s", videoID, fileExe[1])
	fPath := filepath.Join(cfg.assetsRoot,fileName)
	newFile, err := os.Create(fPath)
	defer newFile.Close()
	if err != nil{
		respondWithError(w, http.StatusUnauthorized, "Failed to create file", err)
		return
	}
	_, err = io.Copy(newFile, file)
	if err != nil{
		respondWithError(w, http.StatusUnauthorized, "Failed to copy to the new file", err)
		return
	}
	url := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, videoID, fileExe[1])

	//store in database
	// url := base64.StdEncoding.EncodeToString(data)

	// dataUrl := fmt.Sprintf("data:<%s>;base64,<%s>", mediaType, url)
	metadata.ThumbnailURL = &url

	err = cfg.db.UpdateVideo(metadata)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Unable to update metadata", err)
		return
	}

	respondWithJSON(w, http.StatusOK, metadata)
}
