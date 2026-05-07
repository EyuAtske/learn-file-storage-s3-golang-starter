package main

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error){
	presignedClient := s3.NewPresignClient(s3Client)
	presignedRequest, err := presignedClient.PresignGetObject(context.Background(), &s3.GetObjectInput{Bucket: &bucket, Key: &key}, s3.WithPresignExpires(expireTime))
	if err != nil{
		return "", nil
	}
	return presignedRequest.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error){
	if video.VideoURL == nil {
		return video, nil
	}
	splitUrl := strings.Split(*video.VideoURL, ",")
	if len(splitUrl) < 2{
		return video, nil
	}
	presignedUrl, err := generatePresignedURL(cfg.s3Client, splitUrl[0], splitUrl[1], time.Hour)
	if err != nil{
		return video, err
	}
	video.VideoURL = &presignedUrl
	return video, nil
}