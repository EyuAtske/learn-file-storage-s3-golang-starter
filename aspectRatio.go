package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error){
	type aspect struct{
		Streams []struct{
			Width int `json:"width"`
			Height int `json:"height"`
		}`json:"streams"`
	}
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var buf bytes.Buffer
	var asp aspect
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil{
		return "", err
	}
	err := json.Unmarshal(buf.Bytes(), &asp)
	if err != nil{
		return "", err
	}

	if asp.Streams[0].Width == 0 || asp.Streams[0].Height == 0 {
		return "", fmt.Errorf("no valid video dimensions found")
	}

	ratio := GetAspectRatio(asp.Streams[0].Width, asp.Streams[0].Height)
	var aspectRatio string
	if ratio > 175 && ratio < 180{
		aspectRatio = "16:9"
	}else if ratio > 54 && ratio < 58{
		aspectRatio = "9:16"
	}else{
		aspectRatio = "other"
	}
	return aspectRatio, nil
}

func GetAspectRatio(width, height int) (int) {
	ratio := (width*100/height)
	return ratio
}