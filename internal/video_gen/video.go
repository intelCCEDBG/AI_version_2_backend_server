package videogen

import (
	"os"
	"recorder/config"
	"recorder/pkg/fileoperation"
	"recorder/pkg/logger"
	"sort"
	"strconv"
	"strings"
	"time"
)

func GenerateVideo(hour int, minute int, duration int, hostname string, filename string) {
	videoPath := config.Viper.GetString("recording_path")
	files, err := os.ReadDir(videoPath + hostname + "/")
	if err != nil {
		logger.Error("List video dir fail: " + err.Error())
	}
	currentTime := time.Now()
	today_date_string := currentTime.Format("2006-01-02")
	yesterday_date_string := currentTime.AddDate(0, 0, -1).Format("2006-01-02")
	// fmt.Println(today_date_string)
	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}
	sort.Strings(filenames)
	f, err := os.OpenFile(videoPath+hostname+"/"+filename+".m3u8", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 644)
	if err != nil {
		logger.Error("open video file fail: " + err.Error())
		return
	}
	defer f.Close()
	f.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-MEDIA-SEQUENCE:0\n#EXT-X-ALLOW-CACHE:YES\n#EXT-X-TARGETDURATION:10\n")
	var h, m string
	if hour < 10 {
		h = strconv.Itoa(hour)
		h = "0" + h
	} else {
		h = strconv.Itoa(hour)
	}
	if minute < 10 {
		m = strconv.Itoa(minute)
		m = "0" + m
	} else {
		m = strconv.Itoa(minute)
	}
	reqtime := hour*100 + minute
	ctime := time.Now().Hour()*100 + time.Now().Minute()
	var datestring string
	if reqtime > ctime {
		datestring = yesterday_date_string
	} else {
		datestring = today_date_string
	}
	// fmt.Println(datestring)
	for ii, filename := range filenames {
		if strings.Contains(filename, datestring+"_"+h+"-"+m) {
			for i := 0; i < duration/10; i++ {
				if ii+i >= len(filenames) {
					break
				}
				f.WriteString("#EXTINF:10.000000,\n")
				f.WriteString(filenames[ii+i] + "\n")
			}
			break
		}
	}
	f.WriteString("#EXT-X-ENDLIST")
}
func GenerateErrorVideo(hour int, minute int, duration int, hostname string, machine_name string, filename string) {
	logger.Info("Machine " + machine_name + " Fail !" + strconv.Itoa(hour) + " " + strconv.Itoa(minute))
	recording_path := config.Viper.GetString("recording_path")
	errorvideo_path := config.Viper.GetString("error_video_path")
	fileoperation.CreateFolderifNotExist(errorvideo_path + machine_name + "/")
	files, err := os.ReadDir(recording_path + hostname + "/")
	logger.Info(files[0].Name())
	if err != nil {
		logger.Error("List video dir fail: " + err.Error())
	}
	currentTime := time.Now()
	today_date_string := currentTime.Format("2006-01-02")
	yesterday_date_string := currentTime.AddDate(0, 0, -1).Format("2006-01-02")
	// fmt.Println(today_date_string)
	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}
	sort.Strings(filenames)
	f, err := os.OpenFile(errorvideo_path+machine_name+"/"+filename+".m3u8", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Error("open video file fail: " + err.Error())
		return
	}
	f.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-MEDIA-SEQUENCE:0\n#EXT-X-ALLOW-CACHE:YES\n#EXT-X-TARGETDURATION:10\n")
	var h, m string
	if hour < 10 {
		h = strconv.Itoa(hour)
		h = "0" + h
	} else {
		h = strconv.Itoa(hour)
	}
	if minute < 10 {
		m = strconv.Itoa(minute)
		m = "0" + m
	} else {
		m = strconv.Itoa(minute)
	}
	reqtime := hour*100 + minute
	ctime := time.Now().Hour()*100 + time.Now().Minute()
	var datestring string
	if reqtime > ctime {
		datestring = yesterday_date_string
	} else {
		datestring = today_date_string
	}
	logger.Info(datestring)
	// fmt.Println(datestring)
	for ii, filename := range filenames {
		if strings.Contains(filename, datestring+"_"+h+"-"+m) {
			for i := 0; i < duration/10; i++ {
				f.WriteString("#EXTINF:10.000000,\n")
				f.WriteString(filenames[ii+i] + "\n")
				fileoperation.CopyFile(recording_path+hostname+"/"+filenames[ii+i], errorvideo_path+machine_name+"/"+filenames[ii+i])
			}
			break
		}
	}
	f.WriteString("#EXT-X-ENDLIST")
	f.Close()
	fileoperation.CopyFile(errorvideo_path+machine_name+"/"+filename+".m3u8", recording_path+hostname+"/error.m3u8")
}
