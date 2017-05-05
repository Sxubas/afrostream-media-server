package ts

import (
	"net/http"
	"strings"
	"mp4"
	"strconv"
	"path"
	"errors"
	"fmt"
)



func TreatM3U8Request(splitDirs []string, jConfig mp4.JsonConfig, VideoIdPath string, videoId string, w http.ResponseWriter) (error) {
	mediaType := splitDirs[0]

	if len(splitDirs) == 1 {
		mainDescriptor := CreateMainDescriptor(jConfig, VideoIdPath)
		//w.Write([]byte(mainDescriptor))
		fmt.Println(mainDescriptor)
	} else if len(splitDirs) == 3 {

		// Request sub descriptor
		switch mediaType {
		case "audio":
			// Find the corresponding language tracks
			lang := splitDirs[1]
			track, err := findLanguageTrack(lang, jConfig.Tracks["audio"])
			if err != nil {
				return err
			}

			// Get the number of segment in this track
			numberOfSegments := getNumberOfSegments(track, jConfig)

			// Create the descriptor
			param := "audio/"  + lang
			audioDescriptor := CreateMediaDescriptor(VideoIdPath, param, "ts", jConfig.SegmentDuration, numberOfSegments)
			w.Write([]byte(audioDescriptor))
			break
		case "video":
			// Get the id stream
			idStr := splitDirs[1]
			track, err := getIdTrack(idStr, jConfig.Tracks["video"])
			if err != nil {
				return err
			}

			// Get the number of segment in this track
			numberOfSegments := getNumberOfSegments(track, jConfig)

			// Create the descriptor
			param := "video/" + idStr
			videoDescriptor := CreateMediaDescriptor(VideoIdPath, param, "ts", jConfig.SegmentDuration, numberOfSegments)
			fmt.Println(videoDescriptor)
			w.Write([]byte(videoDescriptor))
			break
		default:
			return errors.New("No media corresponding")
		}
	} else {
		return errors.New("Incorrect url access")
	}

	return nil
}

func findLanguageTrack(lang string, tracks []mp4.TrackEntry) (mp4.TrackEntry, error){

	// Search for the right audio lang
	for _,track := range tracks{
		if track.Lang == lang {
			return track, nil
		}
	}
	return mp4.TrackEntry{}, errors.New("No language: " + lang)
}

func getNumberOfSegments(track mp4.TrackEntry, jConfig mp4.JsonConfig) (int) {
	numberOfSegments := int(track.Config.Duration) / int(jConfig.SegmentDuration * track.Config.Timescale)
	if int(track.Config.Duration) % int(jConfig.SegmentDuration * track.Config.Timescale) != 0 {
		numberOfSegments++
	}
	return numberOfSegments
}

func getIdTrack(idStr string, tracks []mp4.TrackEntry) (mp4.TrackEntry, error) {

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return mp4.TrackEntry{}, errors.New("Not a valid quality stream: " + idStr)
	}

	if id < 0 || id >= len(tracks) {
		return mp4.TrackEntry{}, errors.New("No stream corresponding to this quality: " + idStr)
	}

	return tracks[id], nil
}

func TreatTSRequest(splitDirs []string, jConfig mp4.JsonConfig, videoIdPath string, videoId string, w http.ResponseWriter) (error) {
	mediaType := splitDirs[0]

	if len(splitDirs) != 3 {
		 return errors.New("Incorrect url access")
	}
	splitFragmentNumber := strings.Split(splitDirs[2], ".")

	if len(splitFragmentNumber) < 2 {
		return errors.New("Incorrect fragment number")
	}

	fragmentNumber, err := strconv.Atoi(splitFragmentNumber[0])

	if err != nil {
		return errors.New("Incorrect fragment number: " + splitFragmentNumber[0])
	}

	var track mp4.TrackEntry

	// Request sub descriptor
	switch mediaType {
	case "audio":
		// Find the corresponding language tracks
		lang := splitDirs[1]
		track, err = findLanguageTrack(lang, jConfig.Tracks["audio"])
		if err != nil {
			return err
		}
		break
	case "video":
		// Get the id stream
		idStr := splitDirs[1]
		track, err = getIdTrack(idStr, jConfig.Tracks["video"])
		if err != nil {
			return err
		}
		break
	default:
		return errors.New("No media corresponding")
	}

	// Get the number of segment in this track
	numberOfSegments := getNumberOfSegments(track, jConfig)

	if fragmentNumber < 0 || fragmentNumber >= numberOfSegments {
		return errors.New("Incorrect fragment number :" + splitFragmentNumber[0])
	}

	filePath := path.Dir(videoIdPath) + "/" + track.File
	fragment := CreateHLSFragmentWithConf(*track.Config, filePath, uint32(fragmentNumber), jConfig.SegmentDuration)
	sizeToWrite := len(fragment)
	w.Header().Set("Content-Length", strconv.Itoa(sizeToWrite))
	for sizeToWrite > 0 {
		num, err := w.Write(fragment)
		if err != nil {
			return err
		}
		sizeToWrite -= num
	}

	return nil
}
