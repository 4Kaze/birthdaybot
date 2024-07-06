package adapters

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/4Kaze/birthdaybot/common"
)

const (
	VIDEO_WIDTH       = 872
	VIDEO_HEIGHT      = 480
	BACKGROUND_1      = "background-1.png"
	BACKGROUND_2      = "background-2.png"
	PART_1_VIDEO      = "part-1.mp4"
	PART_3_VIDEO      = "part-3.mp4"
	PART_6_VIDEO      = "part-6.mp4"
	PART_7_VIDEO      = "part-7.mp4"
	TRANSPARENT_IMAGE = "transparent.png"
	CAKE              = "birthday-cake.png"
	AUDIO_TRACK       = "audio.mp3"

	PART_2_FILTER = `[1]scale=400:400, split[avatar1][avatar2]; \
	[avatar2]geq=lum='if(lte(sqrt((X-W/2)*(X-W/2)+(Y-H/2)*(Y-H/2)),min(W,H)/2),255,0)':a=255 [mask]; \
	[avatar1][mask] alphamerge[maskedavatar]; \
	[0][maskedavatar] overlay=(main_w-overlay_w)/2-5:(main_h-overlay_h-60)[background with avatar]; \
	[background with avatar][2]overlay=x=(W-w)/2:y=H-h[avatar with cake]; \
	[avatar with cake]scale='%[1]v*4':'%[2]v*4',zoompan=z='zoom+0.0005':d=700:x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)',scale=%[1]v:%[2]v`
	PART_4_FILTER = `[1]scale=400:400, split[avatar1][avatar2]; \
	[avatar2]geq=lum='if(lte(sqrt((X-W/2)*(X-W/2)+(Y-H/2)*(Y-H/2)),min(W,H)/2),255,0)':a=255 [mask]; \
	[avatar1][mask] alphamerge[masked avatar]; \
	[masked avatar]rotate=-0.05:c=none[rotated avatar]; \
	[0][rotated avatar] overlay=(main_w-overlay_w)/2-t*5:main_h-overlay_h-10-t*10`
	PART_5_FILTER = `[1]scale=400:400, split[avatar1][avatar2]; \
	[avatar2]geq=lum='if(lte(sqrt((X-W/2)*(X-W/2)+(Y-H/2)*(Y-H/2)),min(W,H)/2),255,0)':a=255 [mask]; \
	[avatar1][mask] alphamerge[masked avatar]; \
	[masked avatar]rotate=-0.1:c=none[rotated avatar]; \
	[0][rotated avatar] overlay=(main_w-overlay_w)/2-20+'if(between(t,0,0.6),sin(t*25)*10,0)':main_h-overlay_h-5-'if(between(t,0,0.8),cos(t*10)*5,0)`
	PART_6_FILTER = `[1]scale=24:24, split[avatar1][avatar2]; \
	[avatar2]geq=lum='if(lte(sqrt((X-W/2)*(X-W/2)+(Y-H/2)*(Y-H/2)),min(W,H)/2),255,0)':a=255 [mask]; \
	[avatar1][mask] alphamerge[masked avatar]; \
	[masked avatar] scale=4*iw:4*ih[scaled avatar]; \
	[0] scale='%[1]v*4':'%[2]v*4'[scaled video];\
	[scaled video][scaled avatar] overlay=339*4:310*4+t*48*4,scale=%[1]v:%[2]v`
	PART_7_FILTER = `[1]scale=200:200, split[avatar1][avatar2]; \
	[avatar2]geq=lum='if(lte(sqrt((X-W/2)*(X-W/2)+(Y-H/2)*(Y-H/2)),min(W,H)/2),255,0)':a=255 [mask]; \
	[avatar1][mask] alphamerge[masked avatar]; \
	[2]scale='%[1]v*4':'%[2]v*4'[scaled background]; \
	[scaled background][masked avatar]overlay=x=(W-w)/2-2:y=504[overlayed avatar]; \
	[overlayed avatar]scale='%[1]v*4':'%[2]v*4',zoompan=z='if(lte(zoom,1.0),1.5,max(1.0001,zoom-0.0011))':d=300:x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)+time*0.8',scale=%[1]v:%[2]v[zoomed avatar]; \
	[0][zoomed avatar] overlay=0:0:enable='between(t,0,8.51)'`
)

var creators = []func(VideoGenerator, string, string) error{VideoGenerator.createPartTwo, VideoGenerator.createPartFour, VideoGenerator.createPartFive, VideoGenerator.createPartSix, VideoGenerator.createPartSeven}

type VideoGenerator struct {
	resourceDir string
}

func NewVideoGenerator(resourceDir string) *VideoGenerator {
	return &VideoGenerator{resourceDir: resourceDir}
}

func (generator VideoGenerator) CreateVideo(pathToProfilePicture string) (string, error) {
	log.Printf("Generating a video with profile picture: %v\n", pathToProfilePicture)
	tmpDir, err := os.MkdirTemp("", "*")
	if err != nil {
		common.ErrorLogger.Printf("Failed to create a temp dir: %v\n", err)
		return "", err
	}
	files := make([]string, len(creators))
	for index, creator := range creators {
		log.Printf("Generating video %v...\n", index+1)
		fileName := fmt.Sprintf("%v.mp4", index)
		filePath := filepath.Join(tmpDir, fileName)
		files[index] = filePath
		err := creator(generator, pathToProfilePicture, filePath)
		if err != nil {
			return "", err
		}
	}
	mergedVideoPath := filepath.Join(tmpDir, "merged.mp4")
	log.Println("Merging videos...")
	err = generator.mergeVideos(files, tmpDir, mergedVideoPath)
	if err != nil {
		return "", err
	}
	finalVideoPath := filepath.Join(tmpDir, "final.mp4")
	log.Println("Adding audio...")
	err = generator.addAudio(mergedVideoPath, finalVideoPath)
	if err != nil {
		return "", err
	}
	log.Printf("Generated a video: %v, for profile picture: %v\n", finalVideoPath, pathToProfilePicture)
	return finalVideoPath, nil
}

func (generator VideoGenerator) createPartTwo(profilePicturePath string, outputFilePath string) error {
	return execCommand(
		"ffmpeg",
		"-i",
		filepath.Join(generator.resourceDir, BACKGROUND_1),
		"-i",
		profilePicturePath,
		"-i",
		filepath.Join(generator.resourceDir, CAKE),
		"-filter_complex",
		convertToOneLine(fmt.Sprintf(PART_2_FILTER, VIDEO_WIDTH, VIDEO_HEIGHT)),
		"-t",
		"8",
		outputFilePath,
	)
}

func (generator VideoGenerator) createPartFour(profilePicturePath string, outputFilePath string) error {
	return execCommand(
		"ffmpeg",
		"-loop",
		"1",
		"-i",
		filepath.Join(generator.resourceDir, BACKGROUND_2),
		"-i",
		profilePicturePath,
		"-filter_complex",
		convertToOneLine(PART_4_FILTER),
		"-t",
		"0.9",
		outputFilePath,
	)
}

func (generator VideoGenerator) createPartFive(profilePicturePath string, outputFilePath string) error {
	return execCommand(
		"ffmpeg",
		"-loop",
		"1",
		"-i",
		filepath.Join(generator.resourceDir, BACKGROUND_2),
		"-i",
		profilePicturePath,
		"-filter_complex",
		convertToOneLine(PART_5_FILTER),
		"-t",
		"1.1",
		outputFilePath,
	)
}

func (generator VideoGenerator) createPartSix(profilePicturePath string, outputFilePath string) error {
	return execCommand(
		"ffmpeg",
		"-i",
		filepath.Join(generator.resourceDir, PART_6_VIDEO),
		"-i",
		profilePicturePath,
		"-filter_complex",
		convertToOneLine(fmt.Sprintf(PART_6_FILTER, VIDEO_WIDTH, VIDEO_HEIGHT)),
		outputFilePath,
	)
}

func (generator VideoGenerator) createPartSeven(profilePicturePath string, outputFilePath string) error {
	return execCommand(
		"ffmpeg",
		"-i",
		filepath.Join(generator.resourceDir, PART_7_VIDEO),
		"-i",
		profilePicturePath,
		"-i",
		filepath.Join(generator.resourceDir, TRANSPARENT_IMAGE),
		"-filter_complex",
		convertToOneLine(fmt.Sprintf(PART_7_FILTER, VIDEO_WIDTH, VIDEO_HEIGHT)),
		outputFilePath,
	)
}

func (generator VideoGenerator) mergeVideos(files []string, tempDir string, outputFilePath string) error {
	part1Video := filepath.Join(generator.resourceDir, PART_1_VIDEO)
	part3Video := filepath.Join(generator.resourceDir, PART_3_VIDEO)
	allFiles := make([]string, 3, 7)
	allFiles[0] = part1Video
	allFiles[1] = files[0]
	allFiles[2] = part3Video
	allFiles = append(allFiles, files[1:]...)
	listFilePath := filepath.Join(tempDir, "videos.txt")
	err := createFileList(allFiles, listFilePath)
	if err != nil {
		common.ErrorLogger.Printf("Failed to create a file list: %v\n", err)
		return err
	}
	err = execCommand(
		"ffmpeg",
		"-f",
		"concat",
		"-safe",
		"0",
		"-i",
		listFilePath,
		"-c",
		"copy",
		outputFilePath,
	)
	if err != nil {
		return err
	}
	return nil
}

func createFileList(files []string, listFilePath string) error {
	fileList := make([]byte, 0, 256)
	for _, file := range files {
		fileList = append(fileList, "file "...)
		fileList = append(fileList, file...)
		fileList = append(fileList, '\n')
	}
	return os.WriteFile(listFilePath, fileList, fs.ModePerm)
}

func (generator VideoGenerator) addAudio(videoFilePath string, outputFilePath string) error {
	return execCommand(
		"ffmpeg",
		"-i",
		videoFilePath,
		"-i",
		filepath.Join(generator.resourceDir, AUDIO_TRACK),
		"-c:v",
		"copy",
		"-map",
		"0:v:0",
		"-map",
		"1:a:0",
		outputFilePath,
	)
}

func execCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if output, err := cmd.CombinedOutput(); err != nil {
		common.ErrorLogger.Printf("Failed to execute command: %v due to: %v.\nFull output: %v\n", cmd.String(), err, string(output))
		return fmt.Errorf(string(output))
	}
	return nil
}

func convertToOneLine(str string) string {
	return strings.ReplaceAll(str, "\\\n", "")
}
