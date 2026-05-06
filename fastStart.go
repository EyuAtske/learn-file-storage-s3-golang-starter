package main

import "os/exec"

func processVideoForFastStart(filePath string) (string, error){
	outPutPath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outPutPath)
	if err := cmd.Run(); err != nil{
		return "", err
	}
	return outPutPath, nil
}
